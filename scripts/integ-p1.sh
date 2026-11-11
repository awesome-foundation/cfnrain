#!/bin/bash
# P1 helpers are sourced by integ-aws.sh. All bucket cleanup requires ownership tags.

# The lab has exactly one configured test bucket. Never enumerate account buckets.
# Its stack-name tag is the durable inventory after a Retain deletion or runner loss.
p1_lab_bucket_owner() {
  local owner error="$WORK_DIR/p1-bucket.err"
  if owner=$(aws s3api get-bucket-tagging --bucket "$INTEG_BUCKET_NAME" --query "TagSet[?Key=='cfnrain-integ-owner'].Value | [0]" --output text 2> "$error"); then
    [[ "$owner" =~ ^success-[a-zA-Z0-9-]+-p1-(managed|retained|service)$ ]] || return 2
    printf '%s\n' "$owner"
  elif grep -Eq 'NoSuchBucket|\(404\)' "$error"; then
    return 1
  else
    return 2
  fi
}

p1_empty_lab_bucket() {
  local expected="$1" owner code=0
  owner=$(p1_lab_bucket_owner) || code=$?
  [[ "$code" == 1 ]] && return 0
  [[ "$code" == 0 && "$owner" == "$expected" ]] || return 1
  [[ "$(aws s3api get-bucket-versioning --bucket "$INTEG_BUCKET_NAME" --query Status --output text)" == None ]] || return 1
  # The fixture records this one key before upload. Never delete unknown keys.
  aws s3api delete-object --bucket "$INTEG_BUCKET_NAME" --key fixture.txt >/dev/null
}

p1_cleanup_lab() {
  local stacks name owner code=0 result=0
  # No main-account manifest, rain-bucket lookup, or bucket listing on this path.
  stacks=$(list_test_stacks) || return 1
  while read -r name; do
    [[ "$name" =~ ^success-[a-zA-Z0-9-]+-p1-(managed|retained|service)$ ]] || continue
    delete_stack "$name" || result=1
  done <<< "$stacks"
  owner=$(p1_lab_bucket_owner) || code=$?
  if [[ "$code" == 0 ]]; then
    p1_empty_lab_bucket "$owner" && aws s3api delete-bucket --bucket "$INTEG_BUCKET_NAME" || result=1
  elif [[ "$code" != 1 ]]; then
    result=1
  fi
  p1_bootstrap_cleanup || result=1
  return "$result"
}

p1_expect_error() {
  local log="$1" expected="$2" code=0
  shift 2
  "$@" > "$log" 2>&1 || code=$?
  assert_eq "$code" 1 'ordinary command failure exit code'
  grep -Fq "$expected" "$log" || fail 'command failed for an unexpected reason'
}

p1_resource() {
  aws cloudformation describe-stack-resource --stack-name "$1" --logical-resource-id "$2" --query 'StackResourceDetail.PhysicalResourceId' --output text
}

p1_changesets() {
  local stack original reason mode
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
    p1_expect_error "$WORK_DIR/p1-update-$mode.log" 'failed deploying stack' p1_run ./cfnrain deploy test/templates/integ-failure.yaml "$stack" -y "${flags[@]}"
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

}

p1_removal() {
  local stack mode
  P1_CASE=B4
  # B4: both fixtures reuse the one permitted lab bucket, one stack at a time.
  for mode in managed retained; do
    stack="${STACK_PREFIX}p1-$mode"
    p1_run ./cfnrain deploy "test/templates/integ-p1-$mode.yaml" "$stack" -y --params "BucketName=$INTEG_BUCKET_NAME" > "$WORK_DIR/p1-$mode-create.log" 2>&1
    printf 'P1 removal fixture\n' > "$WORK_DIR/p1-object.txt"
    aws s3api put-object --bucket "$INTEG_BUCKET_NAME" --key fixture.txt --body "$WORK_DIR/p1-object.txt" >/dev/null
    p1_run ./cfnrain rm "$stack" -y --experimental > "$WORK_DIR/p1-$mode-rm.log" 2>&1
    assert_absent "$stack"
    if [[ "$mode" == retained ]]; then
      aws s3api get-object --bucket "$INTEG_BUCKET_NAME" --key fixture.txt "$WORK_DIR/p1-retained.txt" >/dev/null
      cmp "$WORK_DIR/p1-object.txt" "$WORK_DIR/p1-retained.txt"
      p1_empty_lab_bucket "$stack"
      aws s3api delete-bucket --bucket "$INTEG_BUCKET_NAME"
    fi
    # A name-specific check must distinguish absence from access denial.
    local code=0
    p1_lab_bucket_owner > /dev/null || code=$?
    assert_eq "$code" 1 'B4 test bucket absence'
  done
  echo 'SKIPPED: B4 DeleteObject denial deferred, needs bucket-policy permission'
}

p1_bootstrap_existing() {
  local before after
  P1_CASE=B4-existing
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
  [[ "$INTEG_BOOTSTRAP_REGION" != "${AWS_REGION:-}" ]] || return 1
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
  P1_CASE=B4-first-use
  [[ "$INTEG_BOOTSTRAP_REGION" != "$AWS_REGION" ]] || fail 'B4 bootstrap region must differ from the preprovisioned primary region'
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

# B3 uses only Get/Put/List in the existing artifact bucket. The main account
# forbids DeleteObject. Record run-owned keys for the configured lifecycle audit.
p1_check_uploads() {
  local bucket="$1" prefix="$2" key hash count=0 keys
  keys=$(aws s3api list-objects-v2 --bucket "$bucket" --prefix "$prefix/" --query 'Contents[].Key' --output text)
  for key in $keys; do
    [[ "$key" != None ]] || continue
    [[ "$key" == "$prefix/"* ]] || fail 'B3 object escaped its prefix'
    aws s3api get-object --bucket "$bucket" --key "$key" "$WORK_DIR/p1-download.yaml" >/dev/null
    hash=$(sha256sum "$WORK_DIR/p1-download.yaml" | cut -d' ' -f1)
    [[ "${key##*/}" == "$hash" || "${key##*/}" == "$hash.template" ]] || fail 'B3 uploaded bytes differ from SHA-256 key'
    printf '%s\n' "$key" >> "$WORK_DIR/p1-upload-inventory.log"
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
  P1_CASE=B1-role
  p1_run ./cfnrain deploy test/templates/integ-p1-managed.yaml "$stack" -y --params "BucketName=$INTEG_BUCKET_NAME" --role-arn "$INTEG_SERVICE_ROLE_ARN" > "$WORK_DIR/p1-role-create.log" 2>&1
  assert_status "$stack" CREATE_COMPLETE
  p1_run ./cfnrain deploy test/templates/integ-p1-managed.yaml "$stack" -y --tags Stage=role-update --role-arn "$INTEG_SERVICE_ROLE_ARN" > "$WORK_DIR/p1-role-update.log" 2>&1
  assert_status "$stack" UPDATE_COMPLETE
  [[ "$(aws cloudformation describe-stacks --stack-name "$stack" --query 'Stacks[0].RoleARN' --output text)" == "$INTEG_SERVICE_ROLE_ARN" ]] || fail 'B1 service role mismatch'
  arn=$(aws cloudformation describe-stacks --stack-name "$stack" --query 'Stacks[0].StackId' --output text)
  [[ "$(cut -d: -f4 <<< "$arn")" == "$AWS_REGION" ]] || fail 'B1 deployment region mismatch'
  assert_eq "$(aws s3api get-bucket-tagging --bucket "$INTEG_BUCKET_NAME" --query "TagSet[?Key=='Stage'].Value | [0]" --output text)" role-update 'B1 service role resource update'
  p1_run ./cfnrain rm "$stack" -y --experimental > "$WORK_DIR/p1-role-rm.log" 2>&1
  assert_absent "$stack"
}

# Keep command output private until redaction. Arguments can contain a role ARN.
p1_run() {
  local stamp start code=0
  stamp=$(mktemp "$WORK_DIR/p1-command-XXXXXX")
  start=$SECONDS
  "$@" > "$stamp.stdout" 2> "$stamp.stderr" || code=$?
  printf '%s\t%s\t%s\t%s\t%s\n' "${P1_CASE:-P1}" "${stamp##*/}" "$code" "$((SECONDS - start))" "${TEST_SHA:-$(git rev-parse HEAD)}" >> "$WORK_DIR/p1-results.tsv"
  cat "$stamp.stdout"
  p1_redact < "$stamp.stderr" >&2
  return "$code"
}

p1_redact() {
  # WaitCondition handles contain signed URLs. Drop them before archiving logs.
  local bucket_pattern="${INTEG_BUCKET_NAME:-}"
  # Bucket names contain only DNS characters. Escape dots for the regexp.
  bucket_pattern=${bucket_pattern//./\\.}
  local rules='s#https?://[^[:space:]"<>]+#[URL]#g; s/\b[0-9]{12}\b/[ACCOUNT]/g; s/hidden/[NOECHO]/g'
  if [[ -n "$bucket_pattern" ]]; then
    rules="s/$bucket_pattern/[TEST_BUCKET]/g; $rules"
  fi
  sed -E "$rules"
}

p1_save_evidence() {
  local file
  if [[ -n "${P1_ARTIFACT_BUCKET:-}" && -n "${P1_UPLOAD_PREFIX:-}" ]]; then
    aws s3api list-objects-v2 --bucket "$P1_ARTIFACT_BUCKET" --prefix "$P1_UPLOAD_PREFIX/" --query 'Contents[].Key' --output text >> "$WORK_DIR/p1-upload-inventory.log" || return 1
  fi
  mkdir -p "$INTEG_OUT/p1"
  for file in "$WORK_DIR"/p1-*.log "$WORK_DIR"/p1-*.stdout "$WORK_DIR"/p1-*.stderr "$WORK_DIR"/p1-results.tsv; do
    [[ -f "$file" ]] || continue
    p1_redact < "$file" > "$INTEG_OUT/p1/${file##*/}"
  done
}

# Capture AWS stderr before forwarding it. Raw diagnostics remain private in WORK_DIR.
aws() {
  local error code=0
  error=$(mktemp "$WORK_DIR/p1-aws-XXXXXX.err")
  command aws "$@" 2> "$error" || code=$?
  p1_redact < "$error" >&2
  return "$code"
}

# Start timeout itself in a new process group so TERM stops watch and its children.
p1_start_watch() {
  local label="$1" stack="$2" limit="$3"
  WATCH_LOG="$WORK_DIR/p1-watch-$label.log"
  WATCH_STARTED=$SECONDS
  setsid timeout "$limit" ./cfnrain watch --wait --debug "$stack" > "$WATCH_LOG" 2>&1 &
  WATCH_PID=$!
  local deadline=$((SECONDS + 30))
  until grep -q 'Waiting for stack to begin changing' "$WATCH_LOG"; do
    kill -0 "$WATCH_PID" 2>/dev/null || fail 'B3 watcher exited before deployment'
    (( SECONDS < deadline )) || fail 'B3 watcher did not become ready'
    sleep 0.2 # Poll the readiness message, not a fixed startup delay.
  done
}

p1_stop_watch() {
  [[ -n "${WATCH_PID:-}" ]] || return 0
  kill -TERM -- "-$WATCH_PID" 2>/dev/null || true
  wait "$WATCH_PID" 2>/dev/null || true
  WATCH_PID=
}

p1_finish_watch() {
  local code=0
  wait "$WATCH_PID" || code=$?
  WATCH_PID=
  printf 'B3-watch\t%s\t%s\t%s\t%s\n' "${WATCH_LOG##*/}" "$code" "$((SECONDS - WATCH_STARTED))" "${TEST_SHA:-$(git rev-parse HEAD)}" >> "$WORK_DIR/p1-results.tsv"
  [[ "$code" == 0 ]] || fail 'B3 watcher did not finish successfully'
}

p1_aws_wait() {
  local error code=0
  error=$(mktemp "$WORK_DIR/p1-wait-XXXXXX.err")
  timeout 900 aws "$@" 2> "$error" || code=$?
  p1_redact < "$error" >&2
  return "$code"
}
