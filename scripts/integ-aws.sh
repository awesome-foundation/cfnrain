#!/bin/bash
# Integration checks against a real account. Only success* stacks and the
# configured test bucket are created; AWS CLI is the assertion and cleanup oracle.
set -eEuo pipefail

: "${INTEG_BUCKET_NAME:?set INTEG_BUCKET_NAME to the test bucket name}"
INTEG_OUT="${INTEG_OUT:-integ-out}"
RUN_TOKEN="${INTEG_RUN_TOKEN:-${GITHUB_RUN_ID:-local-$(date +%s)-$$}-${GITHUB_RUN_ATTEMPT:-1}}"
STACK_PREFIX="success-${RUN_TOKEN}-"
mkdir -p "$INTEG_OUT"
WORK_DIR=$(mktemp -d)

print_recent_log() {
  local log
  log=$(ls -t "$WORK_DIR"/*.log 2>/dev/null | head -n 1) || return 0
  if [[ -s "$log" ]]; then
    echo "Recent cfnrain output ($log):" >&2
    tail -n 80 "$log" >&2
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
  aws cloudformation list-stacks --query 'StackSummaries[?starts_with(StackName, `success-`) && StackStatus!=`DELETE_COMPLETE` && ParentId==`null`].StackName' --output text
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
      bucket|failure|keep|changeset|unnamed|protected|large|nested) ;;
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
    [[ "$name" =~ ^success-[0-9]+-[0-9]+-(bucket|failure|keep|changeset|unnamed|protected|large|nested)$ ]] || continue
    [[ "$name" == "$STACK_PREFIX"* ]] && continue
    echo "Removing stale integration stack $name"
    delete_stack "$name" || result=1
  done <<< "$stacks"
  return "$result"
}
if [[ "${1:-}" == --cleanup ]]; then
  result=0
  cleanup_prefix "$STACK_PREFIX" || result=1
  rm -rf "$WORK_DIR"
  exit "$result"
fi
on_exit() {
  local result=$?
  trap - EXIT
  if [[ -n "${WATCH_PID:-}" ]]; then
    kill "$WATCH_PID" 2>/dev/null || true
  fi
  cleanup_prefix "$STACK_PREFIX" || result=1
  rm -rf "$WORK_DIR"
  exit "$result"
}
trap on_exit EXIT
trap 'on_error "$LINENO" "$BASH_COMMAND"' ERR
trap 'exit 130' INT
trap 'exit 143' TERM
if [[ "${GITHUB_ACTIONS:-}" == true ]]; then
  sweep_stale
fi

go build ./cmd/cfnrain
./cfnrain ls >/dev/null

# Issue #67: an unexecuted create change set leaves a stack in REVIEW_IN_PROGRESS.
# Preserve a failed deploy, rm, or deletion check while cleaning up this stack.
cleanup_repro_67_on_failure() {
  local status=$?
  trap - EXIT
  if ! aws cloudformation delete-stack --stack-name success-repro-67; then
    echo "Could not request deletion of success-repro-67; checking if it is already absent" >&2
  fi
  if ! aws cloudformation wait stack-delete-complete --stack-name success-repro-67; then
    echo "Could not confirm deletion of success-repro-67" >&2
  fi
  exit "$status"
}
trap cleanup_repro_67_on_failure EXIT
./cfnrain deploy --no-exec test/repro/67/template.yaml success-repro-67 -y
./cfnrain rm -y success-repro-67
aws cloudformation wait stack-delete-complete --stack-name success-repro-67
trap on_exit EXIT

./cfnrain ls

# 1. Create, no-op, previous values, tags, config, and environment parameters.
bucket_stack="${STACK_PREFIX}bucket"
RAIN_DEFAULT_TAG_Team=integ ./cfnrain deploy test/templates/integ-bucket.yaml "$bucket_stack" -y \
  --params "BucketName=$INTEG_BUCKET_NAME,Env=Yes,Secret=hidden" --tags Stage=initial > "$WORK_DIR/bucket-create.log"
assert_status "$bucket_stack" CREATE_COMPLETE
assert_eq "$(stack_param "$bucket_stack" Env)" Yes 'create Env'
assert_eq "$(stack_tag "$bucket_stack" Team)" integ 'default tag'
assert_eq "$(stack_tag "$bucket_stack" Stage)" initial 'initial tag'
RAIN_DEFAULT_TAG_Team=integ ./cfnrain deploy test/templates/integ-bucket.yaml "$bucket_stack" -y \
  --params "BucketName=$INTEG_BUCKET_NAME,Env=Yes,Secret=hidden" --tags Stage=initial > "$WORK_DIR/noop.log"
grep -q 'Deploy was skipped' "$WORK_DIR/noop.log" || fail 'no-op deploy message'
RAIN_DEFAULT_TAG_Team=integ ./cfnrain deploy test/templates/integ-bucket.yaml "$bucket_stack" -y \
  --tags Stage=second > "$WORK_DIR/bucket-update.log"
grep -q 'Successfully updated' "$WORK_DIR/bucket-update.log" || fail 'update message'
assert_eq "$(stack_param "$bucket_stack" Env)" Yes 'previous Env'
assert_eq "$(stack_param "$bucket_stack" BucketName)" "$INTEG_BUCKET_NAME" 'previous bucket name'
cat > "$WORK_DIR/config.yaml" <<'YAML'
Parameters:
  Env: "No"
Tags:
  Stage: config
  Team: config-team
YAML
RAIN_DEFAULT_TAG_Team=integ RAIN_VAR_Env=Yes RAIN_VAR_Note=from-env ./cfnrain deploy test/templates/integ-bucket.yaml "$bucket_stack" -y \
  -c "$WORK_DIR/config.yaml" > "$WORK_DIR/config-update.log"
assert_status "$bucket_stack" UPDATE_COMPLETE
assert_eq "$(stack_param "$bucket_stack" Env)" No 'config parameter precedence'
assert_eq "$(stack_param "$bucket_stack" Note)" from-env 'environment parameter'
assert_eq "$(stack_tag "$bucket_stack" Stage)" config 'config tag'
assert_eq "$(stack_tag "$bucket_stack" Team)" config-team 'config tag precedence'
assert_eq "$(aws s3api get-bucket-tagging --bucket "$INTEG_BUCKET_NAME" --query "TagSet[?Key=='Team'].Value | [0]" --output text)" config-team 'bucket tag propagation'
./cfnrain cat -c "$bucket_stack" > "$WORK_DIR/deployed-config.yaml"
assert_eq "$(yq -r '.Parameters.Note' "$WORK_DIR/deployed-config.yaml")" from-env 'cat config parameter'
assert_eq "$(yq -r '.Tags.Stage' "$WORK_DIR/deployed-config.yaml")" config 'cat config tag'
./cfnrain logs "$bucket_stack" --length 5 >/dev/null
./cfnrain logs "$bucket_stack" --since-user-initiated >/dev/null
./cfnrain logs "$bucket_stack" --chart >/dev/null
./cfnrain cat "$bucket_stack" > "$WORK_DIR/bucket-cat.yaml"
assert_eq "$(yq -r '.Resources.Bucket.Type' "$WORK_DIR/bucket-cat.yaml")" AWS::S3::Bucket 'cat template'
./cfnrain ls "$bucket_stack" > "$WORK_DIR/bucket-ls.log"
grep -q "$bucket_stack" "$WORK_DIR/bucket-ls.log" || fail 'ls summary'
grep -q 'UPDATE_COMPLETE' "$WORK_DIR/bucket-ls.log" || fail 'ls updated status'
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

# 3. Change-set preview and execute on an existing stack. New-stack preview
# cleanup remains excluded until #67 and #95 are fixed.
change_stack="${STACK_PREFIX}changeset"
./cfnrain deploy test/templates/integ-handle.yaml "$change_stack" -y >/dev/null
./cfnrain deploy test/templates/integ-handle-two.yaml "$change_stack" add-handle --no-exec -y > "$WORK_DIR/preview.log"
assert_eq "$(aws cloudformation describe-change-set --stack-name "$change_stack" --change-set-name add-handle --query Status --output text)" CREATE_COMPLETE 'change set status'
./cfnrain ls -c "$change_stack" add-handle > "$WORK_DIR/changeset-ls.log"
grep -q 'SecondHandle' "$WORK_DIR/changeset-ls.log" || fail 'change set listing'
./cfnrain deploy --changeset "$change_stack" add-handle > "$WORK_DIR/changeset-exec.log"
assert_status "$change_stack" UPDATE_COMPLETE
assert_eq "$(aws cloudformation describe-stack-resource --stack-name "$change_stack" --logical-resource-id SecondHandle --query 'StackResourceDetail.ResourceStatus' --output text)" CREATE_COMPLETE 'change set resource'
./cfnrain deploy test/templates/integ-handle-two.yaml "$change_stack" empty-preview --no-exec -y > "$WORK_DIR/empty-preview.log"
grep -q 'Deploy was skipped' "$WORK_DIR/empty-preview.log" || fail 'empty change set message'
./cfnrain rm -c "$change_stack" empty-preview -y >/dev/null
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
module_key="integ-$RUN_TOKEN/module.yaml"
aws s3api put-object --bucket "$artifact_bucket" --key "$module_key" --body test/templates/integ-handle.yaml >/dev/null
cat > "$WORK_DIR/remote-module.yaml" <<YAML
Modules:
  Remote:
    Source: s3://$artifact_bucket/$module_key
YAML
./cfnrain pkg -x "$WORK_DIR/remote-module.yaml" > "$WORK_DIR/remote-module-packaged.yaml"
assert_eq "$(yq -r '.Resources.RemoteHandle.Type' "$WORK_DIR/remote-module-packaged.yaml")" AWS::CloudFormation::WaitConditionHandle 'S3 module source'

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

# 5 and 7 take longer and run on the weekly schedule and release PRs.
if [[ "${INTEG_SLOW:-false}" == true ]]; then
  large_stack="${STACK_PREFIX}large"
  {
    printf 'Description: Large integration template\nMetadata:\n  Padding: "'
    head -c 52000 /dev/zero | tr '\0' x
    printf '"\nResources:\n  Handle:\n    Type: AWS::CloudFormation::WaitConditionHandle\n'
  } > "$WORK_DIR/large.yaml"
  ./cfnrain deploy "$WORK_DIR/large.yaml" "$large_stack" -y >/dev/null
  assert_status "$large_stack" CREATE_COMPLETE
  ./cfnrain rm "$large_stack" -y >/dev/null
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
  ./cfnrain deploy "$WORK_DIR/parent.yaml" "$nested_stack" -y -d > "$WORK_DIR/detach.log"
  grep -q 'Detaching' "$WORK_DIR/detach.log" || fail 'detach message'
  timeout 240 ./cfnrain watch "$nested_stack" > "$WORK_DIR/watch-create.log"
  grep -q 'CREATE_COMPLETE' "$WORK_DIR/watch-create.log" || fail 'watch create status'
  assert_status "$nested_stack" CREATE_COMPLETE
  cp test/templates/integ-handle-two.yaml "$WORK_DIR/child.yaml"
  ./cfnrain deploy "$WORK_DIR/parent.yaml" "$nested_stack" nested-update --no-exec -y > "$WORK_DIR/nested-preview.log"
  grep -q 'SecondHandle' "$WORK_DIR/nested-preview.log" || fail 'nested change set details'
  timeout 240 ./cfnrain watch --wait "$nested_stack" > "$WORK_DIR/watch-update.log" &
  WATCH_PID=$!
  sleep 2
  ./cfnrain deploy --changeset "$nested_stack" nested-update -d > "$WORK_DIR/nested-detach.log"
  wait "$WATCH_PID"
  WATCH_PID=
  grep -q 'UPDATE_COMPLETE' "$WORK_DIR/watch-update.log" || fail 'watch update status'
  assert_status "$nested_stack" UPDATE_COMPLETE
  ./cfnrain logs "$nested_stack" --length 5 >/dev/null
  ./cfnrain rm "$nested_stack" -y >/dev/null
  assert_absent "$nested_stack"
fi

# The packaged modules and Pkl templates are linted in a credential-free job.
./cfnrain pkg -x test/webapp/webapp.yaml > "$INTEG_OUT/webapp.yaml"
./cfnrain pkg test/webapp/webapp.pkl > "$INTEG_OUT/webapp-pkl.yaml"
