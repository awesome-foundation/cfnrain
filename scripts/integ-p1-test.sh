#!/bin/bash
# Offline regression checks. Every AWS call uses a fake; no credentials are read.
set -euo pipefail
# shellcheck source=scripts/integ-cases.sh
source scripts/integ-cases.sh
# shellcheck source=scripts/integ-p1.sh
source scripts/integ-p1.sh
WORK_DIR=$(mktemp -d)
trap 'p1_stop_watch; rm -rf "$WORK_DIR"' EXIT
INTEG_OUT="$WORK_DIR/out"
INTEG_BUCKET_NAME=fixture-bucket
STACK_PREFIX=success-123-1-
MOCK_OWNER=success-123-1-p1-retained
MOCK_VERSION=None
MOCK_ABSENT=false
aws() {
  printf '%s\n' "$*" >> "$WORK_DIR/calls"
  case "$1 $2" in
    's3api get-bucket-tagging')
      if [[ "$MOCK_ABSENT" == true ]]; then echo NoSuchBucket >&2; return 1; fi
      echo "$MOCK_OWNER" ;;
    's3api get-bucket-versioning') echo "$MOCK_VERSION" ;;
    's3api delete-object'|'s3api delete-bucket') return 0 ;;
    *) echo "Unexpected mock call: $*" >&2; return 1 ;;
  esac
}
fail() { echo "$*" >&2; exit 1; }
assert_eq() { [[ "$1" == "$2" ]] || fail "$3"; }
: > "$WORK_DIR/calls"
p1_empty_lab_bucket "$MOCK_OWNER"
grep -q 'delete-object --bucket fixture-bucket --key fixture.txt' "$WORK_DIR/calls"

# A bucket owned by another case is not a cleanup target.
: > "$WORK_DIR/calls"
if p1_empty_lab_bucket success-999-1-p1-retained; then exit 1; fi
if grep -q 'delete-' "$WORK_DIR/calls"; then exit 1; fi

: > "$WORK_DIR/calls"
MOCK_OWNER=unrelated-stack
if p1_empty_lab_bucket "$MOCK_OWNER"; then exit 1; fi
if grep -q 'delete-' "$WORK_DIR/calls"; then exit 1; fi

: > "$WORK_DIR/calls"
MOCK_OWNER=success-123-1-p1-retained
MOCK_VERSION=Enabled
if p1_empty_lab_bucket "$MOCK_OWNER"; then exit 1; fi
if grep -q 'delete-' "$WORK_DIR/calls"; then exit 1; fi

# Missing buckets and unset bootstrap region do not trigger discovery or writes.
: > "$WORK_DIR/calls"
MOCK_ABSENT=true
unset INTEG_BOOTSTRAP_REGION
p1_empty_lab_bucket "$MOCK_OWNER"
p1_bootstrap_cleanup
if grep -Eq 'list-buckets|delete-|ssm ' "$WORK_DIR/calls"; then exit 1; fi

# A listing outage preserves an owned bucket but still attempts bootstrap cleanup.
(
  MOCK_ABSENT=false
  MOCK_VERSION=None
  : > "$WORK_DIR/calls"
  list_test_stacks() { return 1; }
  delete_stack() { exit 1; }
  p1_bootstrap_cleanup() { touch "$WORK_DIR/bootstrap-cleanup-attempted"; }
  if p1_cleanup_lab; then exit 1; fi
  [[ -f "$WORK_DIR/bootstrap-cleanup-attempted" ]]
  if grep -q 'delete-' "$WORK_DIR/calls"; then exit 1; fi
)

# A failed stack deletion preserves the bucket and still cleans bootstrap resources.
(
  MOCK_ABSENT=false
  MOCK_VERSION=None
  : > "$WORK_DIR/calls"
  list_test_stacks() { echo "$MOCK_OWNER"; }
  delete_stack() { touch "$WORK_DIR/stack-delete-attempted"; return 1; }
  p1_bootstrap_cleanup() { touch "$WORK_DIR/bootstrap-after-delete-failure"; }
  if p1_cleanup_lab; then exit 1; fi
  [[ -f "$WORK_DIR/stack-delete-attempted" && -f "$WORK_DIR/bootstrap-after-delete-failure" ]]
  if grep -q 'delete-' "$WORK_DIR/calls"; then exit 1; fi
)

# Each missing lab input must fail before credential setup can start.
python3 - "$WORK_DIR" <<'PYTHON'
import os
from pathlib import Path
import subprocess
import sys
import textwrap

workflow = Path(".github/workflows/integ.yml").read_text()
step = workflow.split("        id: prerequisites\n        run: |\n", 1)[1].split("      - name:", 1)[0]
script = textwrap.dedent(step)
required = ("AWS_INTEG_ROLE_ARN", "AWS_REGION", "INTEG_BUCKET_NAME",
            "INTEG_SERVICE_ROLE_ARN", "INTEG_BOOTSTRAP_REGION")
output = Path(sys.argv[1]) / "prerequisites-output"
for missing in (None, *required):
    output.write_text("")
    env = dict(os.environ, GITHUB_OUTPUT=str(output), **dict.fromkeys(required, "fixture"))
    if missing:
        env[missing] = ""
    result = subprocess.run(["bash", "-eu", "-c", script], env=env,
                            capture_output=True, text=True, timeout=10)
    assert result.returncode == (1 if missing else 0), (missing, result.stderr)
    assert output.read_text() == ("ready=false\n" if missing else "ready=true\n"), missing
    if missing:
        assert "SKIPPED: isolated cases need" in result.stdout, missing
PYTHON

# A usage error must never satisfy an expected deployment failure.
p1_expect_error "$WORK_DIR/expected.log" 'expected failure' bash -c 'echo expected failure >&2; exit 1'
if (p1_expect_error "$WORK_DIR/usage.log" 'expected failure' bash -c 'echo unknown shorthand flag >&2; exit 1') > /dev/null 2>&1; then exit 1; fi

if run_case wrapper-failure capture bash -c 'echo stdout; echo stderr >&2; exit 1' > /dev/null 2>&1; then exit 1; fi
grep -q 'exit=1 duration=' "$WORK_DIR/wrapper-failure-result.log"
run_case expected-error 1 bash -c 'echo expected-stderr >&2; exit 1'
grep -q expected-stderr "$WORK_DIR/expected-error-stderr.log"
if (run_case unexpected-error 0 bash -c 'exit 1') >/dev/null 2>&1; then exit 1; fi
printf 'https://example.invalid/?signature=secret arn:aws:sts::123456789012:assumed-role/private-role/session\n' > "$WORK_DIR/p1-redaction.log"
preserve_logs
grep -q '<url> arn:aws:sts::000000000000:role/redacted-role' "$INTEG_OUT/logs/p1-redaction.log"

# TERM must kill the whole watcher group, not just its shell wrapper.
mkdir "$WORK_DIR/watch-test"
cat > "$WORK_DIR/watch-test/cfnrain" <<'WATCH'
#!/bin/bash
printf '%s\n' "$BASHPID" > "$WATCH_CHILD_FILE"
echo 'Waiting for stack to begin changing'
exec sleep 300
WATCH
chmod +x "$WORK_DIR/watch-test/cfnrain"
export WATCH_CHILD_FILE="$WORK_DIR/watch-child"
(
  cd "$WORK_DIR/watch-test"
  p1_start_watch offline fake-stack 10
  child=$(cat "$WATCH_CHILD_FILE")
  p1_stop_watch
  deadline=$((SECONDS + 5))
  while kill -0 "$child" 2>/dev/null; do
    # A reaped-or-zombie process is no longer executing watch.
    [[ "$(ps -o stat= -p "$child")" == Z* ]] && break
    (( SECONDS < deadline )) || exit 1
    sleep 0.1
  done
)
# Exercise the real --cleanup entrypoint with no bucket variable or credentials.
# The allowlist must require only ListStacks, never bucket/SSM discovery.
mkdir "$WORK_DIR/bin"
cat > "$WORK_DIR/bin/aws" <<'AWS'
#!/bin/bash
printf '%s\n' "$*" >> "$MOCK_CALL_FILE"
case "$1 $2" in
  'cloudformation list-stacks') exit 0 ;;
  's3api get-bucket-tagging') echo NoSuchBucket >&2; exit 1 ;;
  *) echo 'AccessDenied arn:aws:sts::123456789012:assumed-role/fixture/session' >&2; exit 1 ;;
esac
AWS
chmod +x "$WORK_DIR/bin/aws"
export MOCK_CALL_FILE="$WORK_DIR/entry-calls"
: > "$MOCK_CALL_FILE"
env -u INTEG_BUCKET_NAME -u INTEG_BOOTSTRAP_REGION PATH="$WORK_DIR/bin:$PATH" INTEG_OUT="$WORK_DIR/entry-out" bash scripts/integ-aws.sh --cleanup
[[ "$(wc -l < "$MOCK_CALL_FILE")" == 1 ]]
grep -q '^cloudformation list-stacks ' "$MOCK_CALL_FILE"

# Lab cleanup must not depend on rain-bucket or the main manifest sweep.
: > "$MOCK_CALL_FILE"
env -u INTEG_BOOTSTRAP_REGION PATH="$WORK_DIR/bin:$PATH" INTEG_BUCKET_NAME=fixture-bucket INTEG_OUT="$WORK_DIR/lab-out" bash scripts/integ-aws.sh --cleanup-lab
if grep -Eq 'ssm|list-buckets' "$MOCK_CALL_FILE"; then exit 1; fi

# The outer Python runner redacts AWS errors and waits for diagnostic output.
cat > "$WORK_DIR/bin/aws" <<'AWS'
#!/bin/bash
echo 'AccessDenied arn:aws:sts::123456789012:assumed-role/private-role/session' >&2
exit 1
AWS
if env -u INTEG_BUCKET_NAME -u INTEG_BOOTSTRAP_REGION PATH="$WORK_DIR/bin:$PATH" INTEG_OUT="$WORK_DIR/error-out" bash scripts/integ-aws.sh --cleanup > "$WORK_DIR/public-error" 2>&1; then exit 1; fi
grep -q '000000000000:role/redacted-role' "$WORK_DIR/public-error"
if grep -Eq '123456789012|private-role' "$WORK_DIR/public-error"; then exit 1; fi
python3 scripts/integ-p1-cli-test.py
echo 'P1 offline regression checks passed'
