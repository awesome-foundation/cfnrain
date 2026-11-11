#!/bin/bash
# Sourced after integ-aws.sh defines assertions and cleanup.
p1_slow() {
  p1_changesets
  p1_bootstrap_existing
  local artifact_bucket upload_prefix large_stack nested_stack child_id reason event_count
  artifact_bucket=$(aws ssm get-parameter --name rain-bucket --query Parameter.Value --output text)
  upload_prefix="integ-p1-$RUN_TOKEN"
  export P1_ARTIFACT_BUCKET="$artifact_bucket"
  export P1_UPLOAD_PREFIX="$upload_prefix"
  printf 'Prefix: %s/\n' "$upload_prefix" > "$WORK_DIR/p1-upload-inventory.log"
  large_stack="${STACK_PREFIX}large"
  {
    printf 'Description: Large integration template\nMetadata:\n  Padding: "'
    head -c 52000 /dev/zero | tr '\0' x
    printf '"\nResources:\n  Handle:\n    Type: AWS::CloudFormation::WaitConditionHandle\n'
  } > "$WORK_DIR/large.yaml"
  run_case "p1-large-command" capture ./cfnrain deploy "$WORK_DIR/large.yaml" "$large_stack" -y --s3-prefix "$upload_prefix" > "$WORK_DIR/p1-large.log" 2>&1
  assert_status "$large_stack" CREATE_COMPLETE
  run_case "p1-large-cat-command" capture ./cfnrain cat "$large_stack" > "$WORK_DIR/p1-large-cat.yaml"
  run_case "large-command" capture ./cfnrain fmt "$WORK_DIR/large.yaml" > "$WORK_DIR/p1-large-expected.yaml"
  diff -u "$WORK_DIR/p1-large-expected.yaml" "$WORK_DIR/p1-large-cat.yaml"
  p1_check_uploads "$artifact_bucket" "$upload_prefix"
  run_case "p1-large-rm-command" capture ./cfnrain rm "$large_stack" -y > "$WORK_DIR/p1-large-rm.log" 2>&1
  assert_absent "$large_stack"

  nested_stack="${STACK_PREFIX}nested"
  cp test/templates/integ-handle.yaml "$WORK_DIR/child.yaml"
  cat > "$WORK_DIR/parent.yaml" <<'YAML'
Description: Integration nested stack
Resources:
  Child:
    Type: AWS::CloudFormation::Stack
    Properties:
      TemplateURL: child.yaml
YAML
  run_case "detach-command" capture ./cfnrain deploy "$WORK_DIR/parent.yaml" "$nested_stack" -y -d --s3-prefix "$upload_prefix" > "$WORK_DIR/detach.log"
  grep -q 'Detaching' "$WORK_DIR/detach.log" || fail 'detach message'
  run_case "watch-create-command" capture timeout 240 ./cfnrain watch --wait "$nested_stack" > "$WORK_DIR/watch-create.log"
  grep -q 'CREATE_COMPLETE' "$WORK_DIR/watch-create.log" || fail 'watch create status'
  assert_status "$nested_stack" CREATE_COMPLETE
  cp test/templates/integ-handle-two.yaml "$WORK_DIR/child.yaml"
  run_case "nested-preview-command" capture ./cfnrain deploy "$WORK_DIR/parent.yaml" "$nested_stack" success-nested-update --no-exec -y --s3-prefix "$upload_prefix" > "$WORK_DIR/nested-preview.log"
  grep -q 'SecondHandle' "$WORK_DIR/nested-preview.log" || fail 'nested change set details'
  p1_start_watch update "$nested_stack" 240
  run_case "nested-detach-command" capture ./cfnrain deploy --changeset "$nested_stack" success-nested-update -d > "$WORK_DIR/nested-detach.log"
  p1_finish_watch
  grep -q 'UPDATE_COMPLETE' "$WORK_DIR/p1-watch-update.log" || fail 'watch update status'
  assert_status "$nested_stack" UPDATE_COMPLETE
  child_id=$(p1_resource "$nested_stack" Child)
  run_case "p1-child-cat-command" capture ./cfnrain cat "$child_id" > "$WORK_DIR/p1-child-cat.yaml"
  run_case "p1-child-expected-command" capture ./cfnrain fmt test/templates/integ-handle-two.yaml > "$WORK_DIR/p1-child-expected.yaml"
  diff -u "$WORK_DIR/p1-child-expected.yaml" "$WORK_DIR/p1-child-cat.yaml"
  run_case "p1-child-events-command" capture ./cfnrain logs "$nested_stack" Child --all --length 5 > "$WORK_DIR/p1-child-events.log"
  grep -q Child "$WORK_DIR/p1-child-events.log" || fail 'B3 resource-filtered events'
  event_count=$(grep -c '/Child (' "$WORK_DIR/p1-child-events.log")
  (( event_count > 0 && event_count <= 5 )) || fail 'B3 event length limit'
  if grep -Eq '/(Handle|SecondHandle) ' "$WORK_DIR/p1-child-events.log"; then
    fail 'B3 resource filter included child resources'
  fi
  run_case "p1-since-command" capture ./cfnrain logs "$nested_stack" --all --since-user-initiated > "$WORK_DIR/p1-since.log"
  grep -q UPDATE_COMPLETE "$WORK_DIR/p1-since.log" || fail 'B3 recent event content'
  if grep -Eq '/Child .*CREATE_(IN_PROGRESS|COMPLETE)' "$WORK_DIR/p1-since.log"; then
    fail 'B3 since-user-initiated included the previous create operation'
  fi
  run_case "p1-chart-command" capture ./cfnrain logs "$nested_stack" --chart > "$WORK_DIR/p1-chart.html"
  grep -q "'Id': 'Child'" "$WORK_DIR/p1-chart.html" || fail 'B3 chart resource'
  grep -q "'Status': 'UPDATE_COMPLETE'" "$WORK_DIR/p1-chart.html" || fail 'B3 chart status'
  p1_expect_error "$WORK_DIR/p1-watch-settled.log" 'not watching unchanging stack' run_case "p1-watch-settled-command" capture ./cfnrain watch "$nested_stack"
  grep -q 'not watching unchanging stack' "$WORK_DIR/p1-watch-settled.log" || fail 'B3 settled watch diagnostic'

  # watch reports terminal failure (exit 0 is the existing watch contract).
  cp test/templates/integ-failure.yaml "$WORK_DIR/child.yaml"
  # Hold the failing operation longer than watch's two-second polling interval.
  sed -i 's/Timeout: "1"/Timeout: "30"/' "$WORK_DIR/child.yaml"
  p1_start_watch failure "$nested_stack" 600
  run_case "p1-nested-failure-command" capture ./cfnrain deploy "$WORK_DIR/parent.yaml" "$nested_stack" -y -d --s3-prefix "$upload_prefix" > "$WORK_DIR/p1-nested-failure.log" 2>&1
  p1_finish_watch
  assert_status "$nested_stack" UPDATE_ROLLBACK_COMPLETE
  grep -q UPDATE_ROLLBACK_COMPLETE "$WORK_DIR/p1-watch-failure.log" || fail 'B3 watch failure state'
  reason=$(aws cloudformation describe-stack-events --stack-name "$child_id" --query "StackEvents[?LogicalResourceId=='NeverSignalled' && ResourceStatus=='CREATE_FAILED'].ResourceStatusReason | [0]" --output text)
  [[ -n "$reason" && "$reason" != None ]] || fail 'B3 missing child failure reason'
  run_case "p1-nested-reason-command" capture ./cfnrain logs "$nested_stack" > "$WORK_DIR/p1-nested-reason.log"
  grep -Fq "$reason" "$WORK_DIR/p1-nested-reason.log" || fail 'B3 nested failure reason'
  grep -q NeverSignalled "$WORK_DIR/p1-nested-reason.log" || fail 'B3 nested failing resource'
  p1_check_uploads "$artifact_bucket" "$upload_prefix"
  run_case "p1-nested-remove-command" capture ./cfnrain rm "$nested_stack" -y >/dev/null
  assert_absent "$nested_stack"
  assert_status "$child_id" DELETE_COMPLETE
  echo 'B3 upload inventory uses the existing artifact lifecycle; this role cannot delete objects'
}
