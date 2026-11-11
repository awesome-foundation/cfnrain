#!/bin/bash
# P1 helpers are sourced by integ-aws.sh. All bucket cleanup requires ownership tags.

p1_cleanup_buckets() {
  local prefix="$1" bucket owner objects key result=0
  local buckets
  buckets=$(aws s3api list-buckets --query 'Buckets[].Name' --output text) || return 1
  for bucket in $buckets; do
    [[ "$bucket" == "$prefix"* && "$bucket" =~ -p1-(assets|managed|retained)$ ]] || continue
    if [[ "$prefix" == success- ]]; then
      [[ "$bucket" =~ ^success-[0-9]+-[0-9]+-p1-(assets|managed|retained)$ ]] || continue
      [[ "$bucket" == "$STACK_PREFIX"* ]] && continue
    fi
    owner=$(aws s3api get-bucket-tagging --bucket "$bucket" --query "TagSet[?Key=='cfnrain-integ-owner'].Value | [0]" --output text) || { result=1; continue; }
    [[ "$owner" == "$bucket" ]] || { result=1; continue; }
    [[ "$(aws s3api get-bucket-versioning --bucket "$bucket" --query Status --output text)" == None ]] || { result=1; continue; }
    if [[ "$bucket" == *-p1-managed ]]; then
      aws s3api delete-bucket-policy --bucket "$bucket" || { result=1; continue; }
    fi
    # These unversioned buckets belong exclusively to this case. Record the exact
    # keys before deletion. Never empty a shared artifact bucket or another prefix.
    objects=$(aws s3api list-objects-v2 --bucket "$bucket" --query 'Contents[].Key' --output text) || { result=1; continue; }
    for key in $objects; do
      [[ "$key" == None ]] && continue
      if [[ "$bucket" == *-p1-assets ]]; then
        [[ "$key" =~ ^[0-9a-f]{64}(\.[a-z]+)?$ ]] || { result=1; continue; }
      else
        [[ "$key" == fixture.txt ]] || { result=1; continue; }
      fi
      aws s3api delete-object --bucket "$bucket" --key "$key" >/dev/null || result=1
    done
    aws s3api delete-bucket --bucket "$bucket" || result=1
  done
  return "$result"
}

p1_expect_error() {
  local log="$1" code=0
  shift
  "$@" > "$log" 2>&1 || code=$?
  assert_eq "$code" 1 'ordinary command failure exit code'
}

p1_resource() {
  aws cloudformation describe-stack-resource --stack-name "$1" --logical-resource-id "$2" --query 'StackResourceDetail.PhysicalResourceId' --output text
}

p1_cases() {
  local stack original reason mode bucket
  # B2: execute a new-stack preview and abandon a separate preview.
  for mode in execute abandon; do
    stack="${STACK_PREFIX}p1-preview-$mode"
    p1_run ./cfnrain deploy test/templates/integ-handle.yaml "$stack" initial --no-exec -y > "$WORK_DIR/p1-preview-$mode.log" 2>&1
    assert_status "$stack" REVIEW_IN_PROGRESS
    assert_eq "$(aws cloudformation list-stack-resources --stack-name "$stack" --query 'length(StackResourceSummaries)' --output text)" 0 'preview created no resources'
    if [[ "$mode" == execute ]]; then
      p1_run ./cfnrain deploy --changeset "$stack" initial > "$WORK_DIR/p1-preview-execute.log" 2>&1
      assert_status "$stack" CREATE_COMPLETE
    fi
    p1_run ./cfnrain rm "$stack" -y > "$WORK_DIR/p1-preview-rm.log" 2>&1
    assert_absent "$stack"
  done

  # B2: failed updates preserve the original handle. Recovery must update it.
  for mode in rollback keep; do
    local flags=()
    stack="${STACK_PREFIX}p1-update-$mode"
    [[ "$mode" == keep ]] && flags=(--keep)
    p1_run ./cfnrain deploy test/templates/integ-handle.yaml "$stack" -y > "$WORK_DIR/p1-create-$mode.log" 2>&1
    original=$(p1_resource "$stack" Handle)
    p1_expect_error "$WORK_DIR/p1-update-$mode.log" p1_run ./cfnrain deploy test/templates/integ-failure.yaml "$stack" -y "${flags[@]}"
    if [[ "$mode" == keep ]]; then
      assert_status "$stack" UPDATE_FAILED
    else
      assert_status "$stack" UPDATE_ROLLBACK_COMPLETE
    fi
    [[ "$(p1_resource "$stack" Handle)" == "$original" ]] || fail 'B2 original handle changed'
    reason=$(aws cloudformation describe-stack-events --stack-name "$stack" --query "StackEvents[?LogicalResourceId=='NeverSignalled' && ResourceStatus=='CREATE_FAILED'].ResourceStatusReason | [0]" --output text)
    [[ -n "$reason" && "$reason" != None ]] || fail 'B2 missing failure reason'
    p1_run ./cfnrain logs "$stack" NeverSignalled > "$WORK_DIR/p1-failure-events.log" 2>&1
    grep -Fq "$reason" "$WORK_DIR/p1-failure-events.log" || fail 'B2 logs omit AWS failure reason'
    if [[ "$mode" == rollback ]]; then
      p1_run ./cfnrain deploy test/templates/integ-handle-two.yaml "$stack" -y > "$WORK_DIR/p1-recovery.log" 2>&1
      assert_status "$stack" UPDATE_COMPLETE
      [[ "$(p1_resource "$stack" Handle)" == "$original" ]] || fail 'B2 recovery replaced handle'
    fi
    p1_run ./cfnrain rm "$stack" -y > "$WORK_DIR/p1-update-rm.log" 2>&1
    assert_absent "$stack"
  done

  # B4: metadata empties a nonempty bucket. Retain leaves its object intact.
  for mode in managed retained; do
    stack="${STACK_PREFIX}p1-$mode"
    bucket="$stack"
    p1_run ./cfnrain deploy "test/templates/integ-p1-$mode.yaml" "$stack" -y --params "BucketName=$bucket" > "$WORK_DIR/p1-$mode-create.log" 2>&1
    printf 'P1 removal fixture\n' > "$WORK_DIR/p1-object.txt"
    aws s3api put-object --bucket "$bucket" --key fixture.txt --body "$WORK_DIR/p1-object.txt" >/dev/null
    if [[ "$mode" == managed ]]; then
      # B4: a bucket policy denies only deletion of this case's recorded object.
      cat > "$WORK_DIR/p1-deny.json" <<POLICY
{"Version":"2012-10-17","Statement":[{"Effect":"Deny","Principal":"*","Action":"s3:DeleteObject","Resource":"arn:aws:s3:::$bucket/fixture.txt"}]}
POLICY
      aws s3api put-bucket-policy --bucket "$bucket" --policy "file://$WORK_DIR/p1-deny.json"
      p1_expect_error "$WORK_DIR/p1-rm-denied.log" p1_run ./cfnrain rm "$stack" -y -x
      grep -Eq 'AccessDenied|Access Denied|Forbidden' "$WORK_DIR/p1-rm-denied.log" || fail 'B4 denied deletion diagnostic'
      assert_status "$stack" CREATE_COMPLETE
      aws s3api head-object --bucket "$bucket" --key fixture.txt >/dev/null
      aws s3api delete-bucket-policy --bucket "$bucket"
    fi
    p1_run ./cfnrain rm "$stack" -y -x > "$WORK_DIR/p1-$mode-rm.log" 2>&1
    assert_absent "$stack"
    if [[ "$mode" == retained ]]; then
      aws s3api get-object --bucket "$bucket" --key fixture.txt "$WORK_DIR/p1-retained.txt" >/dev/null
      cmp "$WORK_DIR/p1-object.txt" "$WORK_DIR/p1-retained.txt"
      p1_cleanup_buckets "$stack"
    else
      local names
      names=$(aws s3api list-buckets --query 'Buckets[].Name' --output text)
      [[ " $names " != *"$bucket"* ]] || fail 'B4 managed bucket remains'
    fi
  done

  # B4: repeat bootstrap without changing the shared artifact infrastructure.
  local before after
  before=$(aws ssm get-parameter --name rain-bucket --query 'Parameter.[Value,Version]' --output text)
  p1_run ./cfnrain bootstrap -y > "$WORK_DIR/p1-bootstrap.log" 2>&1
  p1_run ./cfnrain bootstrap -y >> "$WORK_DIR/p1-bootstrap.log" 2>&1
  after=$(aws ssm get-parameter --name rain-bucket --query 'Parameter.[Value,Version]' --output text)
  [[ "$before" == "$after" ]] || fail 'B4 bootstrap changed existing parameter'
}

# B4: a persistent reservation records ownership before bootstrap can create
# anything. It also lets --cleanup and the next stale sweep recover an interruption.
p1_bootstrap_cleanup() {
  [[ -n "${INTEG_BOOTSTRAP_REGION:-}" ]] || return 0
  local bucket value error="$WORK_DIR/p1-bootstrap-cleanup.err"
  if ! bucket=$(aws ssm get-parameter --region "$INTEG_BOOTSTRAP_REGION" --name /cfnrain-integ/p1-bootstrap --query Parameter.Value --output text 2> "$error"); then
    grep -q ParameterNotFound "$error" && return 0
    return 1
  fi
  [[ "$bucket" == "rain-artifacts-$(aws sts get-caller-identity --query Account --output text)-$INTEG_BOOTSTRAP_REGION" ]] || return 1
  if value=$(aws ssm get-parameter --region "$INTEG_BOOTSTRAP_REGION" --name rain-bucket --query Parameter.Value --output text 2> "$error"); then
    [[ "$value" == "$bucket" ]] || return 1
    aws ssm delete-parameter --region "$INTEG_BOOTSTRAP_REGION" --name rain-bucket || return 1
  else
    grep -q ParameterNotFound "$error" || return 1
  fi
  # No objects are uploaded by this case. A nonempty bucket must fail cleanup.
  if aws s3api head-bucket --region "$INTEG_BOOTSTRAP_REGION" --bucket "$bucket" 2> "$error"; then
    aws s3api delete-bucket --region "$INTEG_BOOTSTRAP_REGION" --bucket "$bucket" || return 1
  else
    grep -q '(404)' "$error" || return 1
  fi
  aws ssm delete-parameter --region "$INTEG_BOOTSTRAP_REGION" --name /cfnrain-integ/p1-bootstrap
}

p1_bootstrap() {
  if [[ -z "${INTEG_BOOTSTRAP_REGION:-}" ]]; then
    echo 'SKIPPED: B4 first-use bootstrap needs INTEG_BOOTSTRAP_REGION'
    return
  fi
  local bucket first second error="$WORK_DIR/p1-bootstrap.err"
  bucket="rain-artifacts-$(aws sts get-caller-identity --query Account --output text)-$INTEG_BOOTSTRAP_REGION"
  if aws ssm get-parameter --region "$INTEG_BOOTSTRAP_REGION" --name rain-bucket > /dev/null 2> "$error"; then
    fail 'B4 isolated region already has rain-bucket'
  fi
  grep -q ParameterNotFound "$error" || fail 'B4 cannot establish parameter absence'
  if aws s3api head-bucket --region "$INTEG_BOOTSTRAP_REGION" --bucket "$bucket" 2> "$error"; then
    fail 'B4 isolated artifact bucket already exists'
  fi
  grep -q '(404)' "$error" || fail 'B4 cannot establish bucket absence'
  aws ssm put-parameter --region "$INTEG_BOOTSTRAP_REGION" --name /cfnrain-integ/p1-bootstrap --type String --value "$bucket" > /dev/null
  p1_run ./cfnrain bootstrap -y --region "$INTEG_BOOTSTRAP_REGION" > "$WORK_DIR/p1-first-bootstrap.log" 2>&1
  first=$(aws ssm get-parameter --region "$INTEG_BOOTSTRAP_REGION" --name rain-bucket --query 'Parameter.[Value,Version]' --output text)
  [[ "$first" == "$bucket"$'\t1' ]] || fail 'B4 first-use compatibility name or version mismatch'
  p1_run ./cfnrain bootstrap -y --region "$INTEG_BOOTSTRAP_REGION" >> "$WORK_DIR/p1-first-bootstrap.log" 2>&1
  second=$(aws ssm get-parameter --region "$INTEG_BOOTSTRAP_REGION" --name rain-bucket --query 'Parameter.[Value,Version]' --output text)
  [[ "$first" == "$second" ]] || fail 'B4 second bootstrap changed parameter'
  aws s3api head-bucket --region "$INTEG_BOOTSTRAP_REGION" --bucket "$bucket" >/dev/null
  p1_bootstrap_cleanup
}

# B3: each uploaded object must have a SHA-256 key and matching bytes.
p1_check_uploads() {
  local bucket="$1" key hash count=0
  local keys
  keys=$(aws s3api list-objects-v2 --bucket "$bucket" --query 'Contents[].Key' --output text)
  for key in $keys; do
    [[ "$key" != None ]] || continue
    aws s3api get-object --bucket "$bucket" --key "$key" "$WORK_DIR/p1-download.yaml" >/dev/null
    hash=$(sha256sum "$WORK_DIR/p1-download.yaml" | cut -d' ' -f1)
    [[ "${key%%.*}" == "$hash" ]] || fail 'B3 uploaded bytes differ from SHA-256 key'
    count=$((count + 1))
  done
  (( count > 0 )) || fail 'B3 no uploaded templates'
}

# B1: only the isolated job can attach a service role to this stack.
p1_service_role() {
  if [[ -z "${INTEG_SERVICE_ROLE_ARN:-}" ]]; then
    echo 'SKIPPED: B1 service role needs INTEG_SERVICE_ROLE_ARN'
    return
  fi
  local stack="${STACK_PREFIX}p1-service" arn
  p1_run ./cfnrain deploy test/templates/integ-handle.yaml "$stack" -y --role-arn "$INTEG_SERVICE_ROLE_ARN" > "$WORK_DIR/p1-role-create.log" 2>&1
  assert_status "$stack" CREATE_COMPLETE
  p1_run ./cfnrain deploy test/templates/integ-handle-two.yaml "$stack" -y --role-arn "$INTEG_SERVICE_ROLE_ARN" > "$WORK_DIR/p1-role-update.log" 2>&1
  assert_status "$stack" UPDATE_COMPLETE
  [[ "$(aws cloudformation describe-stacks --stack-name "$stack" --query 'Stacks[0].RoleARN' --output text)" == "$INTEG_SERVICE_ROLE_ARN" ]] || fail 'B1 service role mismatch'
  arn=$(aws cloudformation describe-stacks --stack-name "$stack" --query 'Stacks[0].StackId' --output text)
  [[ "$(cut -d: -f4 <<< "$arn")" == "$AWS_REGION" ]] || fail 'B1 deployment region mismatch'
  p1_run ./cfnrain rm "$stack" -y > "$WORK_DIR/p1-role-rm.log" 2>&1
  assert_absent "$stack"
}

# Keep command output private until redaction. Arguments can contain a role ARN.
p1_run() {
  local stamp start code=0
  stamp=$(mktemp "$WORK_DIR/p1-command-XXXXXX")
  start=$SECONDS
  "$@" > "$stamp.stdout" 2> "$stamp.stderr" || code=$?
  printf '%s\t%s\t%s\t%s\n' "${stamp##*/}" "$code" "$((SECONDS - start))" "${TEST_SHA:-$(git rev-parse HEAD)}" >> "$WORK_DIR/p1-results.tsv"
  cat "$stamp.stdout"
  cat "$stamp.stderr" >&2
  return "$code"
}

p1_redact() {
  # WaitCondition handles contain signed URLs. Drop them before archiving logs.
  sed -E 's#https?://[^[:space:]"<>]+#[URL]#g; s/[0-9]{12}/[ACCOUNT]/g; s/hidden/[NOECHO]/g'
}

p1_save_evidence() {
  local file
  mkdir -p "$INTEG_OUT/p1"
  for file in "$WORK_DIR"/p1-*.log "$WORK_DIR"/p1-*.stdout "$WORK_DIR"/p1-*.stderr "$WORK_DIR"/p1-results.tsv; do
    [[ -f "$file" ]] || continue
    p1_redact < "$file" > "$INTEG_OUT/p1/${file##*/}"
  done
}
