#!/bin/bash
# Integration checks against a real account. Only success* stacks and the
# configured test bucket are created; AWS CLI is the assertion and cleanup oracle.
set -eEuo pipefail
# Redact every console path, including AWS CLI errors and cleanup failures.
if [[ "${INTEG_REDACTING:-}" != 1 ]]; then
  exec python3 scripts/integ-run-redacted.py "$0" "$@"
fi
source scripts/integ-warnings.sh
source scripts/integ-upload-fixture.sh

MODE="${1:-fast}"
if [[ "$MODE" == fast || "$MODE" == --p1-isolated || "$MODE" == --cleanup-lab ]]; then
  : "${INTEG_BUCKET_NAME:?set INTEG_BUCKET_NAME to the test bucket name}"
fi
INTEG_OUT="${INTEG_OUT:-integ-out}"
RUN_TOKEN="${INTEG_RUN_TOKEN:-${GITHUB_RUN_ID:-local-$(date +%s)-$$}-${GITHUB_RUN_ATTEMPT:-1}}"
[[ "$RUN_TOKEN" =~ ^[a-zA-Z0-9-]+$ ]] || { echo 'invalid run token' >&2; exit 1; }
STACK_PREFIX="success-${RUN_TOKEN}-"
mkdir -p "$INTEG_OUT"
WORK_DIR=$(mktemp -d)
OBJECT_MANIFEST="$INTEG_OUT/objects-$RUN_TOKEN.txt"

# shellcheck source=scripts/integ-cases.sh
source scripts/integ-cases.sh
# shellcheck source=scripts/integ-p1.sh
source scripts/integ-p1.sh

snapshot() {
  local name names
  case "$MODE" in
    --p1-isolated|--cleanup-lab) names='p1-managed p1-retained p1-service' ;;
    --slow) names='p1-preview-execute p1-preview-abandon p1-update-rollback p1-update-keep large nested' ;;
    *) names='pipeline modules original express normal conflict' ;;
  esac
  for name in $names; do
    aws cloudformation describe-stacks --stack-name "$STACK_PREFIX$name" \
      --query 'Stacks[].{Name:StackName,Status:StackStatus,Outputs:Outputs}' > "$WORK_DIR/$name-state.log" 2>/dev/null || true
    aws cloudformation describe-stack-events --stack-name "$STACK_PREFIX$name" \
      --query 'StackEvents[].{LogicalId:LogicalResourceId,Type:ResourceType,Status:ResourceStatus,Reason:ResourceStatusReason}' \
      > "$WORK_DIR/$name-events.log" 2>/dev/null || true
  done
}
# Express completion can precede bucket and tag readiness.
wait_bucket_tag() {
  local bucket="$1" expected="$2" value deadline=$((SECONDS + 120))
  while (( SECONDS < deadline )); do
    if aws s3api head-bucket --bucket "$bucket" 2>> "$WORK_DIR/readiness.log" &&
       value=$(aws s3api get-bucket-tagging --bucket "$bucket" --query "TagSet[?Key=='Revision'].Value | [0]" --output text 2>> "$WORK_DIR/readiness.log") &&
       [[ "$value" == "$expected" ]]; then
      return 0
    fi
    sleep 2 # Bounded state polling, not a fixed readiness assumption.
  done
  fail "bucket tag did not become $expected within 120 seconds"
}
stack_output() { aws cloudformation describe-stacks --stack-name "$1" --query "Stacks[0].Outputs[?OutputKey=='$2'].OutputValue | [0]" --output text; }


print_recent_log() {
  local log
  # Fixed internal filenames contain no whitespace.
  # shellcheck disable=SC2012
  log=$(ls -t "$WORK_DIR"/*.log 2>/dev/null | head -n 1) || return 0
  if [[ -s "$log" ]]; then
    echo "Recent cfnrain output ($log):" >&2
    preserve_logs
    tail -n 80 "$INTEG_OUT/logs/${log##*/}" >&2
  fi
}
fail() { echo "integration assertion failed: $*" >&2; print_recent_log; exit 1; }
on_error() {
  local result=$?
  echo "integration command failed at line $1: $2" >&2
  print_recent_log
  return "$result"
}
assert_eq() { [[ "$1" == "$2" ]] || fail "expected '$2', got '$1': $3"; }
stack_status() { aws cloudformation describe-stacks --stack-name "$1" --query 'Stacks[0].StackStatus' --output text; }
stack_exists() {
  local output
  if output=$(aws cloudformation describe-stacks --stack-name "$1" --query 'Stacks[0].StackStatus' --output text 2>&1); then
    return 0
  fi
  if [[ "$output" == *'does not exist'* ]]; then
    return 1
  fi
  echo "$output" >&2
  return 2
}
assert_status() { assert_eq "$(stack_status "$1")" "$2" "$1 status"; }
assert_absent() {
  local result=0
  stack_exists "$1" || result=$?
  case "$result" in
    1) return 0 ;;
    0) echo "integration assertion failed: $1 still exists" >&2 ;;
    *) echo "unable to confirm deletion of $1" >&2 ;;
  esac
  return 1
}
stack_param() { aws cloudformation describe-stacks --stack-name "$1" --query "Stacks[0].Parameters[?ParameterKey=='$2'].ParameterValue | [0]" --output text; }
stack_tag() { aws cloudformation describe-stacks --stack-name "$1" --query "Stacks[0].Tags[?Key=='$2'].Value | [0]" --output text; }

# The list includes deleted stacks for 90 days. Only act on live root stacks.
list_test_stacks() {
  # JMESPath literals must reach the AWS CLI unchanged.
  # shellcheck disable=SC2016
  aws cloudformation list-stacks --query 'StackSummaries[?starts_with(StackName, `success-`) && StackStatus!=`DELETE_COMPLETE` && ParentId==`null`].[StackName]' --output text
}
delete_stack() {
  local name="$1" status protected deadline=$((SECONDS + 900))
  while :; do
    status=$(stack_status "$name") || { assert_absent "$name"; return $?; }
    case "$status" in
      DELETE_IN_PROGRESS)
        timeout 900 aws cloudformation wait stack-delete-complete --stack-name "$name" || return 1
        assert_absent "$name"
        return $?
        ;;
      REVIEW_IN_PROGRESS) break ;; # Empty preview stacks can be deleted by the AWS CLI.
      *_IN_PROGRESS)
        if (( SECONDS >= deadline )); then
          echo "timed out waiting for $name to finish $status" >&2
          return 1
        fi
        sleep 5
        ;;
      *) break ;;
    esac
  done
  if [[ "$MODE" == --p1-isolated || "$MODE" == --cleanup-lab ]]; then
    p1_empty_lab_bucket "$name" || return 1
  fi
  protected=$(aws cloudformation describe-stacks --stack-name "$name" --query 'Stacks[0].EnableTerminationProtection' --output text) || return 1
  if [[ "$protected" == True ]]; then
    aws cloudformation update-termination-protection --stack-name "$name" --no-enable-termination-protection >/dev/null || return 1
  fi
  aws cloudformation delete-stack --stack-name "$name" || return 1
  timeout 900 aws cloudformation wait stack-delete-complete --stack-name "$name" || return 1
  assert_absent "$name"
}
cleanup_prefix() {
  local prefix="$1" name stacks result=0
  stacks=$(list_test_stacks) || { echo 'unable to list stacks for cleanup' >&2; return 1; }
  while read -r name; do
    [[ "$name" == "$prefix"* ]] || continue
    case "${name#"$prefix"}" in
      bucket|failure|keep|changeset|unnamed|protected|large|nested|pipeline|modules|original|express|normal|conflict|repro67|p1-preview-execute|p1-preview-abandon|p1-update-rollback|p1-update-keep) ;;
      *) continue ;;
    esac
    delete_stack "$name" || result=1
  done <<< "$stacks"
  return "$result"
}
sweep_stale() {
  local name stacks result=0
  stacks=$(list_test_stacks) || { echo 'unable to list stacks for stale cleanup' >&2; return 1; }
  while read -r name; do
    if [[ "$name" == success-repro-67 ]]; then
      delete_stack "$name" || result=1
      continue
    fi
    [[ "$name" =~ ^success-[0-9]+-[0-9]+-(bucket|failure|keep|changeset|unnamed|protected|large|nested|pipeline|modules|original|express|normal|conflict|repro67|p1-preview-execute|p1-preview-abandon|p1-update-rollback|p1-update-keep)$ ]] || continue
    [[ "$name" == "$STACK_PREFIX"* ]] && continue
    echo "Removing stale integration stack $name"
    delete_stack "$name" || result=1
  done <<< "$stacks"
  return "$result"
}
cleanup_current() {
  if [[ "$MODE" == --p1-isolated || "$MODE" == --cleanup-lab ]]; then
    p1_cleanup_lab
  else
    cleanup_prefix "$STACK_PREFIX"
  fi
}
if [[ "$MODE" == --cleanup || "$MODE" == --cleanup-lab ]]; then
  result=0
  if [[ "$MODE" == --cleanup && -n "${INTEG_BUCKET_NAME:-}" && ! -d "$INTEG_OUT/logs" ]]; then snapshot; fi
  cleanup_current || result=1
  preserve_logs
  rm -rf "$WORK_DIR"
  exit "$result"
fi
on_exit() {
  local result=$?
  trap - EXIT
  p1_stop_watch
  if (( result != 0 && result != 130 && result != 143 )); then snapshot; fi
  p1_record_uploads || result=1
  cleanup_current || result=1
  preserve_logs
  rm -rf "$WORK_DIR"
  exit "$result"
}
trap on_exit EXIT
trap 'on_error "$LINENO" "$BASH_COMMAND"' ERR
trap 'exit 130' INT
trap 'exit 143' TERM
if [[ "$MODE" == --p1-isolated ]]; then
  go build ./cmd/cfnrain
  # Lab setup provisions the primary artifacts. Never create them implicitly.
  aws ssm get-parameter --name rain-bucket --query Parameter.Value --output text > "$WORK_DIR/p1-primary-bucket"
  aws s3api head-bucket --bucket "$(cat "$WORK_DIR/p1-primary-bucket")" > /dev/null
  p1_cleanup_lab
  p1_bootstrap
  p1_service_role
  p1_removal
  exit 0
fi
# shellcheck source=scripts/integ-p1-slow.sh
source scripts/integ-p1-slow.sh
if [[ "$MODE" == --slow ]]; then
  go build ./cmd/cfnrain
  sweep_stale
  p1_slow
  exit 0
fi
[[ "$MODE" == fast ]] || fail 'unknown integration mode'
if [[ "${GITHUB_ACTIONS:-}" == true ]]; then
  sweep_stale || echo 'stale stack sweep incomplete; continuing' >&2
fi

go build ./cmd/cfnrain
./cfnrain ls >/dev/null

# Issue #67: use the normal cleanup owner for an unexecuted preview.
./cfnrain deploy --no-exec test/repro/67/template.yaml "${STACK_PREFIX}repro67" -y
./cfnrain rm -y "${STACK_PREFIX}repro67"
assert_absent "${STACK_PREFIX}repro67"

./cfnrain ls

# Create, no-op, previous values, tags, config, and environment parameters.
bucket_stack="${STACK_PREFIX}bucket"
RAIN_DEFAULT_TAG_Team=integ ./cfnrain deploy test/templates/integ-bucket.yaml "$bucket_stack" -y \
  --params "BucketName=$INTEG_BUCKET_NAME,Env=prod,Secret=hidden" --tags Stage=initial > "$WORK_DIR/bucket-create.log" 2>&1
assert_status "$bucket_stack" CREATE_COMPLETE
assert_eq "$(stack_param "$bucket_stack" Env)" prod 'create Env'
assert_eq "$(stack_tag "$bucket_stack" Team)" integ 'default tag'
assert_eq "$(stack_tag "$bucket_stack" Stage)" initial 'initial tag'
RAIN_DEFAULT_TAG_Team=integ ./cfnrain deploy test/templates/integ-bucket.yaml "$bucket_stack" -y \
  --params "BucketName=$INTEG_BUCKET_NAME,Env=prod,Secret=hidden" --tags Stage=initial > "$WORK_DIR/noop.log" 2>&1
grep -q 'Deploy was skipped' "$WORK_DIR/noop.log" || fail 'no-op deploy message'
RAIN_DEFAULT_TAG_Team=integ ./cfnrain deploy test/templates/integ-bucket.yaml "$bucket_stack" -y \
  --tags Stage=second > "$WORK_DIR/bucket-update.log" 2>&1
grep -q 'Successfully updated' "$WORK_DIR/bucket-update.log" || fail 'update message'
assert_eq "$(stack_param "$bucket_stack" Env)" prod 'previous Env'
assert_eq "$(stack_param "$bucket_stack" BucketName)" "$INTEG_BUCKET_NAME" 'previous bucket name'
cat > "$WORK_DIR/config.yaml" <<'YAML'
Parameters:
  Env: dev
Tags:
  Stage: config
  Team: config-team
YAML
RAIN_DEFAULT_TAG_Team=integ RAIN_VAR_Env=prod RAIN_VAR_Note=from-env ./cfnrain deploy test/templates/integ-bucket.yaml "$bucket_stack" -y \
  -c "$WORK_DIR/config.yaml" > "$WORK_DIR/config-update.log" 2>&1
assert_status "$bucket_stack" UPDATE_COMPLETE
assert_eq "$(stack_param "$bucket_stack" Env)" dev 'config parameter precedence'
assert_eq "$(stack_param "$bucket_stack" Note)" from-env 'environment parameter'
assert_eq "$(stack_tag "$bucket_stack" Stage)" config 'config tag'
assert_eq "$(stack_tag "$bucket_stack" Team)" config-team 'config tag precedence'
assert_eq "$(aws s3api get-bucket-tagging --bucket "$INTEG_BUCKET_NAME" --query "TagSet[?Key=='Team'].Value | [0]" --output text)" config-team 'bucket tag propagation'
./cfnrain cat -c "$bucket_stack" > "$WORK_DIR/deployed-config.yaml" 2>&1
assert_eq "$(yq -r '.Parameters.Note' "$WORK_DIR/deployed-config.yaml")" from-env 'cat config parameter'
assert_eq "$(yq -r '.Tags.Stage' "$WORK_DIR/deployed-config.yaml")" config 'cat config tag'
./cfnrain logs "$bucket_stack" --length 5 >/dev/null
./cfnrain logs "$bucket_stack" --since-user-initiated >/dev/null
./cfnrain logs "$bucket_stack" --chart >/dev/null
./cfnrain cat "$bucket_stack" > "$WORK_DIR/bucket-cat.yaml"
assert_eq "$(yq -r '.Resources.Bucket.Type' "$WORK_DIR/bucket-cat.yaml")" AWS::S3::Bucket 'cat template'
./cfnrain ls "$bucket_stack" > "$WORK_DIR/bucket-ls.log" 2>&1
grep -q "$bucket_stack" "$WORK_DIR/bucket-ls.log" || fail 'ls summary'
grep -q 'UPDATE_COMPLETE' "$WORK_DIR/bucket-ls.log" || fail 'ls updated status'
# Explicit CLI values override config and environment values.
RAIN_VAR_Env=dev ./cfnrain deploy test/templates/integ-bucket.yaml "$bucket_stack" -y \
  -c "$WORK_DIR/config.yaml" --params Env=prod --tags Stage=cli > "$WORK_DIR/p1-precedence.log" 2>&1
assert_eq "$(stack_param "$bucket_stack" Env)" prod 'CLI parameter precedence'
assert_eq "$(stack_tag "$bucket_stack" Stage)" cli 'CLI tag precedence'
# The sentinel is test data, never a credential. Check normal output and readback.
if grep -Fq hidden "$WORK_DIR/bucket-create.log" "$WORK_DIR/noop.log" "$WORK_DIR/deployed-config.yaml" "$WORK_DIR/bucket-ls.log" "$WORK_DIR/config-update.log" "$WORK_DIR/bucket-update.log" "$WORK_DIR/p1-precedence.log"; then
  fail 'B1 NoEcho value appeared in normal output'
fi
echo 'SKIPPED: B1 service role runs in the optional isolated P1 job'
./cfnrain rm "$bucket_stack" -y >/dev/null
assert_absent "$bucket_stack"

# 2. Failed create must exit nonzero, then a fixed template must recover.
failed_stack="${STACK_PREFIX}failure"
if ./cfnrain deploy test/templates/integ-failure.yaml "$failed_stack" -y > "$WORK_DIR/failure.log" 2>&1; then
  fail 'failed create returned zero'
fi
grep -q 'failed deploying stack' "$WORK_DIR/failure.log" || fail 'failure message'
assert_status "$failed_stack" ROLLBACK_COMPLETE
./cfnrain deploy test/templates/integ-handle.yaml "$failed_stack" -y >/dev/null
assert_status "$failed_stack" CREATE_COMPLETE
./cfnrain rm "$failed_stack" -y >/dev/null
assert_absent "$failed_stack"
keep_stack="${STACK_PREFIX}keep"
if ./cfnrain deploy test/templates/integ-failure.yaml "$keep_stack" -y --keep > "$WORK_DIR/keep.log" 2>&1; then
  fail '--keep failure returned zero'
fi
assert_status "$keep_stack" CREATE_FAILED
assert_eq "$(aws cloudformation describe-stacks --stack-name "$keep_stack" --query 'Stacks[0].DisableRollback' --output text)" True 'rollback disabled'
./cfnrain rm "$keep_stack" -y >/dev/null
assert_absent "$keep_stack"

# 3. Change-set preview and execute on an existing stack.
change_stack="${STACK_PREFIX}changeset"
./cfnrain deploy test/templates/integ-handle.yaml "$change_stack" -y >/dev/null
./cfnrain deploy test/templates/integ-handle-two.yaml "$change_stack" success-add-handle --no-exec -y > "$WORK_DIR/preview.log"
assert_eq "$(aws cloudformation describe-change-set --stack-name "$change_stack" --change-set-name success-add-handle --query Status --output text)" CREATE_COMPLETE 'change set status'
./cfnrain ls -c "$change_stack" success-add-handle > "$WORK_DIR/changeset-ls.log"
grep -q 'SecondHandle' "$WORK_DIR/changeset-ls.log" || fail 'change set listing'
assert_eq "$(aws cloudformation list-stack-resources --stack-name "$change_stack" --query 'length(StackResourceSummaries)' --output text)" 1 'preview changed resources'
./cfnrain deploy --changeset "$change_stack" success-add-handle > "$WORK_DIR/changeset-exec.log"
assert_status "$change_stack" UPDATE_COMPLETE
assert_eq "$(aws cloudformation describe-stack-resource --stack-name "$change_stack" --logical-resource-id SecondHandle --query 'StackResourceDetail.ResourceStatus' --output text)" CREATE_COMPLETE 'change set resource'
./cfnrain deploy test/templates/integ-handle-two.yaml "$change_stack" success-empty-preview --no-exec -y > "$WORK_DIR/empty-preview.log"
grep -q 'Deploy was skipped' "$WORK_DIR/empty-preview.log" || fail 'empty change set message'
./cfnrain rm -c "$change_stack" success-empty-preview -y >/dev/null
./cfnrain rm "$change_stack" -y >/dev/null
assert_absent "$change_stack"

# 4. Upload keys and URLs are checked against the object store, not pkg output alone.
artifact_bucket=$(aws ssm get-parameter --name rain-bucket --query 'Parameter.Value' --output text)
account=$(aws sts get-caller-identity --query Account --output text)
./cfnrain pkg cft/pkg/tmpl/s3-props-template.yaml > "$WORK_DIR/s3-props.yaml"
./cfnrain pkg cft/pkg/tmpl/s3-template.yaml > "$WORK_DIR/s3.yaml"
hash=$(sha256sum cft/pkg/tmpl/s3.txt | cut -d' ' -f1)
assert_eq "$(yq -r '.Resources.Test.Properties.TheS3URI' "$WORK_DIR/s3.yaml")" "s3://$artifact_bucket/$hash" 'S3 URI and SHA-256 key'
aws s3api head-object --bucket "$artifact_bucket" --key "$hash" >/dev/null
record_object "integ-$RUN_TOKEN/$hash"
./cfnrain pkg cft/pkg/tmpl/s3-template.yaml --s3-prefix "integ-$RUN_TOKEN" > "$WORK_DIR/s3-prefix.yaml"
assert_eq "$(yq -r '.Resources.Test.Properties.TheS3URI' "$WORK_DIR/s3-prefix.yaml")" "s3://$artifact_bucket/integ-$RUN_TOKEN/$hash" 'S3 prefix'
aws s3api head-object --bucket "$artifact_bucket" --key "integ-$RUN_TOKEN/$hash" >/dev/null
aws s3api get-object --bucket "$artifact_bucket" --key "integ-$RUN_TOKEN/$hash" "$WORK_DIR/prefixed-upload.txt" >/dev/null
cmp cft/pkg/tmpl/s3.txt "$WORK_DIR/prefixed-upload.txt"
./cfnrain pkg cft/pkg/tmpl/s3-template.yaml --s3-owner "$account" > "$WORK_DIR/s3-owner.yaml"
assert_eq "$(yq -r '.Resources.Test.Properties.TheS3URI' "$WORK_DIR/s3-owner.yaml")" "s3://$artifact_bucket/$hash" 'correct S3 owner'
if ./cfnrain pkg cft/pkg/tmpl/s3-template.yaml --s3-owner 000000000000 > "$WORK_DIR/wrong-owner.log" 2>&1; then
  fail 'wrong S3 owner returned zero'
fi
grep -q 'HeadBucket' "$WORK_DIR/wrong-owner.log" || fail 'wrong S3 owner failure path'
grep -Eiq '403|Forbidden|AccessDenied' "$WORK_DIR/wrong-owner.log" || fail 'wrong S3 owner error'
./cfnrain pkg cft/pkg/tmpl/s3http-template.yaml > "$WORK_DIR/s3http.yaml"
http_url=$(yq -r '.Resources.Test.Properties.TheS3URL' "$WORK_DIR/s3http.yaml")
region="${AWS_REGION:-${AWS_DEFAULT_REGION:-$(aws configure get region)}}"
http_hash=$(sha256sum cft/pkg/tmpl/s3http.txt | cut -d' ' -f1)
assert_eq "$http_url" "https://${artifact_bucket}.s3.${region}.amazonaws.com/${http_hash}" 'regional HTTPS upload URL'
aws s3api head-object --bucket "$artifact_bucket" --key "${http_url##*/}" >/dev/null
./cfnrain pkg cft/pkg/tmpl/s3-extension-template.yaml > "$WORK_DIR/s3-extension.yaml"
[[ "$(yq -r '.Resources.Test.Properties.TheS3URI' "$WORK_DIR/s3-extension.yaml")" == *.txt ]] || fail 'S3 extension'
# A4 (#145): each invocation shares one upload cache across all four references.
printf 'P0 upload cache %s\n' "$RUN_TOKEN" > "$WORK_DIR/cache.txt"
cache_hash=$(sha256sum "$WORK_DIR/cache.txt" | cut -d' ' -f1)
for extension in '' .json .txt; do
  record_object "integ-$RUN_TOKEN/cache/$cache_hash$extension"
done
for directive in S3 S3Http; do
  for order in forward reverse; do
    extensions=('' json txt json)
    [[ "$order" != reverse ]] || extensions=(json txt json '')
    format=URI
    [[ "$directive" != S3Http ]] || format=Http
    # S3Http accepts a scalar. Extension options require S3 with Format: Http.
    write_upload_fixture "$format" "${extensions[@]}" > "$WORK_DIR/cache.yaml"
    run_case "a4-$directive-$order" 0 ./cfnrain pkg "$WORK_DIR/cache.yaml" --s3-prefix "integ-$RUN_TOKEN/cache" --s3-bucket "$artifact_bucket" --s3-owner "$account"
    for i in "${!extensions[@]}"; do
      suffix="${extensions[$i]:+.}${extensions[$i]}"
      key="integ-$RUN_TOKEN/cache/$cache_hash$suffix"
      url=$(yq -r ".Resources.Test.Properties.Value$i" "$WORK_DIR/a4-$directive-$order-stdout.log")
      if [[ "$directive" == S3 ]]; then
        assert_eq "$url" "s3://$artifact_bucket/$key" 'A4 S3 key'
      else
        assert_eq "$url" "https://$artifact_bucket.s3.$region.amazonaws.com/$key" 'A4 HTTPS key'
      fi
      aws s3api head-object --bucket "$artifact_bucket" --key "$key" >/dev/null
      aws s3api get-object --bucket "$artifact_bucket" --key "$key" "$WORK_DIR/cache-result.txt" >/dev/null
      cmp "$WORK_DIR/cache.txt" "$WORK_DIR/cache-result.txt"
    done
  done
done

module_key="integ-$RUN_TOKEN/module.yaml"
record_object "$module_key"
aws s3api put-object --bucket "$artifact_bucket" --key "$module_key" --body test/templates/integ-handle.yaml >/dev/null
cat > "$WORK_DIR/remote-module.yaml" <<YAML
Modules:
  Remote:
    Source: s3://$artifact_bucket/$module_key
YAML
./cfnrain pkg --experimental "$WORK_DIR/remote-module.yaml" > "$WORK_DIR/remote-module-packaged.yaml"
assert_eq "$(yq -r '.Resources.RemoteHandle.Type' "$WORK_DIR/remote-module-packaged.yaml")" AWS::CloudFormation::WaitConditionHandle 'S3 module source'

# A3: explicit stdin output must reach CloudFormation.
run_case a3-fmt 0 ./cfnrain fmt - < test/templates/integ-handle.yaml
cp "$WORK_DIR/a3-fmt-stdout.log" "$WORK_DIR/formatted.yaml"
./cfnrain fmt test/templates/integ-handle.yaml > "$WORK_DIR/file-formatted.yaml"
cmp "$WORK_DIR/formatted.yaml" "$WORK_DIR/file-formatted.yaml"
run_case a3-merge 0 ./cfnrain merge "$WORK_DIR/formatted.yaml" - < test/templates/integ-p0/second-handle.yaml
cp "$WORK_DIR/a3-merge-stdout.log" "$WORK_DIR/merged.yaml"
./cfnrain merge "$WORK_DIR/formatted.yaml" test/templates/integ-p0/second-handle.yaml > "$WORK_DIR/file-merged.yaml"
cmp "$WORK_DIR/merged.yaml" "$WORK_DIR/file-merged.yaml"
run_case a3-diff 0 ./cfnrain diff "$WORK_DIR/formatted.yaml" - < "$WORK_DIR/merged.yaml"
grep -q SecondHandle "$WORK_DIR/a3-diff-stdout.log" || fail 'A3 diff missing resource'
run_case a3-deploy 0 ./cfnrain deploy "$WORK_DIR/merged.yaml" "${STACK_PREFIX}pipeline" -y
for logical in Handle SecondHandle; do
  assert_eq "$(aws cloudformation describe-stack-resource --stack-name "${STACK_PREFIX}pipeline" --logical-resource-id "$logical" --query StackResourceDetail.ResourceStatus --output text)" CREATE_COMPLETE 'A3 resource'
done

./cfnrain cat "${STACK_PREFIX}pipeline" > "$WORK_DIR/pipeline-cat.yaml"
yq -o=json '.Resources' "$WORK_DIR/pipeline-cat.yaml" > "$WORK_DIR/actual-resources.json"
yq -o=json '.Resources' "$WORK_DIR/merged.yaml" > "$WORK_DIR/expected-resources.json"
cmp "$WORK_DIR/actual-resources.json" "$WORK_DIR/expected-resources.json"

# A5: assert AWS-evaluated outputs, including parent binding after an update.
module_handle=
for parent in parent-value updated-value; do
  run_case "a5-$parent" 0 ./cfnrain deploy test/templates/integ-p0/modules.yaml "${STACK_PREFIX}modules" --experimental -y --params "Name=$parent"
  status=CREATE_COMPLETE
  [[ "$parent" != updated-value ]] || status=UPDATE_COMPLETE
  assert_status "${STACK_PREFIX}modules" "$status"
  assert_eq "$(stack_output "${STACK_PREFIX}modules" Constant)" constant-local '#141'
  # shellcheck disable=SC2016
  assert_eq "$(stack_output "${STACK_PREFIX}modules" Escaped)" '${Literal}' '#142'
  handle=$(aws cloudformation describe-stack-resource --stack-name "${STACK_PREFIX}modules" --logical-resource-id ChildHandle --query StackResourceDetail.PhysicalResourceId --output text)
  if [[ -n "$module_handle" ]]; then
    assert_eq "$handle" "$module_handle" 'A5 unrelated handle was not replaced'
  fi
  module_handle="$handle"
  assert_eq "$(stack_output "${STACK_PREFIX}modules" Sequence)" "$handle-suffix" '#151'
  assert_eq "$(stack_output "${STACK_PREFIX}modules" Empty)" prefix '#153'
  assert_eq "$(stack_output "${STACK_PREFIX}modules" ParentValue)" "$parent-child" '#176 supplied'
  assert_eq "$(stack_output "${STACK_PREFIX}modules" ParentSub)" "$parent-sub" '#176 Sub'
  assert_eq "$(stack_output "${STACK_PREFIX}modules" DefaultValue)" fallback-default '#176 default'
  assert_eq "$(stack_output "${STACK_PREFIX}modules" Anchor)" xx '#181'
done

# A6: formatting preserves values, but packaging still rejects this input.
run_case a6-fmt 0 ./cfnrain fmt test/templates/integ-p0/original.yaml --node-style original
cp "$WORK_DIR/a6-fmt-stdout.log" "$WORK_DIR/original.yaml"
grep -q 'explicit key comment' "$WORK_DIR/original.yaml" || fail 'A6 comment missing'
./cfnrain fmt "$WORK_DIR/original.yaml" --json > "$WORK_DIR/original.json"
python3 scripts/integ-original-check.py "$WORK_DIR/original.json"
# https://github.com/awesome-foundation/cfnrain/issues/315
# When packaging supports this input, check a real deployment and its values.
assert_absent "${STACK_PREFIX}original"
run_case a6-create 1 ./cfnrain deploy test/templates/integ-p0/original.yaml "${STACK_PREFIX}original" -y --node-style original
grep -Fq 'error converting template to map: yaml: invalid map key:' "$WORK_DIR/a6-create-stderr.log" || fail 'A6 packaging error missing'
assert_absent "${STACK_PREFIX}original"

# A1: named previews expose the deployment mode to an independent AWS assertion.
for mode in express normal; do
  flags=()
  [[ "$mode" != express ]] || flags+=(--express)
  stack="${STACK_PREFIX}$mode"
  name="$INTEG_BUCKET_NAME"
  # S3 can briefly reject reuse even after CloudFormation confirms deletion.
  # Retry only that specific failure, never permission or validation failures.
  reuse_deadline=$((SECONDS + 300))
  for attempt in 1 2 3; do
    (( SECONDS < reuse_deadline )) || fail 'A1 bucket reuse deadline exceeded'
    run_case "a1-$mode-preview" 0 ./cfnrain deploy test/templates/integ-p0/bucket.yaml "$stack" success-create-preview -y --no-exec "${flags[@]}" --params "BucketName=$name"
    actual_mode=$(aws cloudformation describe-change-set --stack-name "$stack" --change-set-name success-create-preview --query DeploymentConfig.Mode --output text)
    printf '%s\n' "$actual_mode" > "$WORK_DIR/a1-$mode-mode.log"
    if [[ "$mode" == express ]]; then
      assert_eq "$actual_mode" EXPRESS 'A1 change set mode'
    else
      [[ "$actual_mode" != EXPRESS ]] || fail 'A1 control unexpectedly Express'
    fi
    create_limit=$((reuse_deadline - SECONDS))
    (( create_limit <= 2100 - SECONDS )) || create_limit=$((2100 - SECONDS))
    (( create_limit > 0 )) || fail 'A1 bucket reuse deadline exceeded'
    (( create_limit <= 180 )) || create_limit=180
    create_start=$SECONDS
    create_result=0
    create_log="$WORK_DIR/a1-$mode-create-$attempt"
    timeout "$create_limit" ./cfnrain deploy --changeset "$stack" success-create-preview -y > "$create_log-stdout.log" 2> "$create_log-stderr.log" || create_result=$?
    printf 'exit=%s duration=%s\n' "$create_result" "$((SECONDS-create_start))" > "$create_log-result.log"
    (( create_result != 0 )) || break
    # shellcheck disable=SC2016
    aws cloudformation describe-stack-events --stack-name "$stack" \
      --query 'StackEvents[?ResourceType==`AWS::S3::Bucket` && ResourceStatus==`CREATE_FAILED`].ResourceStatusReason' \
      --output text > "$create_log-events.log"
    if [[ "$create_result" != 1 ]] || ! grep -Eiq 'OperationAborted|conflicting conditional operation' "$create_log-events.log"; then
      fail "A1 $mode create failed with exit $create_result"
    fi
    (( attempt < 3 && SECONDS < reuse_deadline )) || fail 'A1 bucket name remained unavailable'
    delete_stack "$stack"
    sleep 2 # Backoff after a confirmed transient S3 refusal.
  done
  assert_status "$stack" CREATE_COMPLETE
  [[ "$mode" != express ]] || wait_bucket_tag "$name" first
  aws s3api head-bucket --bucket "$name"
  assert_eq "$(aws s3api get-bucket-tagging --bucket "$name" --query "TagSet[?Key=='Revision'].Value | [0]" --output text)" first 'A1 create tag'
  run_case "a1-$mode-update" 0 ./cfnrain deploy test/templates/integ-p0/bucket.yaml "$stack" -y "${flags[@]}" --params "BucketName=$name,Revision=second"
  assert_status "$stack" UPDATE_COMPLETE
  [[ "$mode" != express ]] || wait_bucket_tag "$name" second
  assert_eq "$(aws s3api get-bucket-tagging --bucket "$name" --query "TagSet[?Key=='Revision'].Value | [0]" --output text)" second 'A1 update tag'
  if [[ "$mode" == normal ]]; then
    # A2: the normal control owns the bucket. Early Validation must reject a conflict.
    for validation_mode in normal express; do
      validation_flags=()
      [[ "$validation_mode" != express ]] || validation_flags+=(--express)
      run_case "a2-$validation_mode-conflict" 1 ./cfnrain deploy test/templates/integ-p0/bucket.yaml "${STACK_PREFIX}conflict" success-conflict -y "${validation_flags[@]}" --params "BucketName=$INTEG_BUCKET_NAME"
      assert_eq "$(aws cloudformation describe-change-set --stack-name "${STACK_PREFIX}conflict" --change-set-name success-conflict --query Status --output text)" FAILED 'A2 failed change set'
      aws cloudformation describe-events --stack-name "${STACK_PREFIX}conflict" --change-set-name success-conflict > "$WORK_DIR/a2-$validation_mode-events.log"
      python3 - "$WORK_DIR" "$validation_mode" <<'PYTHON'
import json, pathlib, sys
p = pathlib.Path(sys.argv[1])
events = json.loads((p/f'a2-{sys.argv[2]}-events.log').read_text())['OperationEvents']
text = (p/f'a2-{sys.argv[2]}-conflict-stdout.log').read_text() + (p/f'a2-{sys.argv[2]}-conflict-stderr.log').read_text()
failures = [e for e in events if e.get('ValidationStatus') == 'FAILED' and e.get('ValidationFailureMode') != 'WARN']
assert failures, 'A2 blocked: region returned no Early Validation failures'
for event in failures:
    for field in ('LogicalResourceId', 'ResourceType', 'ValidationName', 'ValidationStatusReason'):
        assert event.get(field) and event[field] in text, (field, 'missing from CLI details')
PYTHON
      assert_eq "$(aws s3api get-bucket-tagging --bucket "$INTEG_BUCKET_NAME" --query "TagSet[?Key=='Revision'].Value | [0]" --output text)" second 'A2 owner unchanged'

      delete_stack "${STACK_PREFIX}conflict"
    done

    # A7: forecast reads a cheap bucket fixture and keeps its warning on stderr.
    run_case a7-forecast 0 ./cfnrain forecast test/templates/integ-p0/bucket.yaml "${STACK_PREFIX}normal" --experimental --params "BucketName=$INTEG_BUCKET_NAME,Revision=second"
    assert_warning "$WORK_DIR/a7-forecast-stdout.log" "$WORK_DIR/a7-forecast-stderr.log"
  fi
  delete_stack "$stack"
done

# A7: the artifact bucket already exists. Missing input must fail before Cloud Control.
run_case a7-cc-error 1 ./cfnrain cc deploy "$WORK_DIR/missing.yaml" "${STACK_PREFIX}cc" --experimental -y
assert_warning "$WORK_DIR/a7-cc-error-stdout.log" "$WORK_DIR/a7-cc-error-stderr.log"
grep -q 'error packaging template' "$WORK_DIR/a7-cc-error-stderr.log" || fail 'A7 cc did not reach input validation'

# A7: cc state reads only this run's recorded S3 state object. No cc resource is created.
cc_key="integ-$RUN_TOKEN/deployments/${STACK_PREFIX}cc.yaml"
record_object "$cc_key"
aws s3api put-object --bucket "$artifact_bucket" --key "$cc_key" --body test/templates/integ-handle.yaml >/dev/null
run_case a7-cc-state 0 ./cfnrain cc state "${STACK_PREFIX}cc" --experimental --s3-prefix "integ-$RUN_TOKEN"
assert_warning "$WORK_DIR/a7-cc-state-stdout.log" "$WORK_DIR/a7-cc-state-stderr.log"
./cfnrain fmt "$WORK_DIR/a7-cc-state-stdout.log" > "$WORK_DIR/cc-state.yaml"
./cfnrain fmt test/templates/integ-handle.yaml > "$WORK_DIR/cc-expected.yaml"
cmp "$WORK_DIR/cc-state.yaml" "$WORK_DIR/cc-expected.yaml"

# 8. The unnamed stack path uses the filename and returns the deployed template.
unnamed_stack="${STACK_PREFIX}unnamed"
cp test/templates/integ-handle.yaml "$WORK_DIR/$unnamed_stack.yaml"
./cfnrain deploy "$WORK_DIR/$unnamed_stack.yaml" -y >/dev/null
assert_status "$unnamed_stack" CREATE_COMPLETE
./cfnrain fmt "$WORK_DIR/$unnamed_stack.yaml" > "$WORK_DIR/unnamed-expected.yaml"
./cfnrain cat "$unnamed_stack" > "$WORK_DIR/unnamed-cat.yaml"
diff -u "$WORK_DIR/unnamed-expected.yaml" "$WORK_DIR/unnamed-cat.yaml"
./cfnrain ls "$unnamed_stack" > "$WORK_DIR/unnamed-ls.log"
grep -q "$unnamed_stack" "$WORK_DIR/unnamed-ls.log" || fail 'unnamed stack summary'
grep -q 'CREATE_COMPLETE' "$WORK_DIR/unnamed-ls.log" || fail 'unnamed stack status'
./cfnrain rm "$unnamed_stack" -y >/dev/null
assert_absent "$unnamed_stack"

# 6. rm must disable termination protection before deleting the stack.
protected_stack="${STACK_PREFIX}protected"
./cfnrain deploy test/templates/integ-handle.yaml "$protected_stack" -y -t >/dev/null
assert_eq "$(aws cloudformation describe-stacks --stack-name "$protected_stack" --query 'Stacks[0].EnableTerminationProtection' --output text)" True 'termination protection enabled'
./cfnrain rm "$protected_stack" -y >/dev/null
assert_absent "$protected_stack"

# Local callers can request both suites. CI uses a separate --slow job.
if [[ "${INTEG_SLOW:-false}" == true ]]; then
  p1_slow
fi

# The packaged modules and Pkl templates are linted in a credential-free job.
# Never place raw templates in the artifact directory, even on partial failure.
./cfnrain pkg --experimental test/webapp/webapp.yaml > "$WORK_DIR/webapp.yaml"
python3 scripts/integ-redact.py --template < "$WORK_DIR/webapp.yaml" > "$INTEG_OUT/webapp.yaml"
./cfnrain pkg test/webapp/webapp.pkl > "$WORK_DIR/webapp-pkl.yaml"
python3 scripts/integ-redact.py --template < "$WORK_DIR/webapp-pkl.yaml" > "$INTEG_OUT/webapp-pkl.yaml"
