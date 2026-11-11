#!/bin/bash
# Offline ownership checks. The aws function below never contacts AWS.
set -euo pipefail
# shellcheck source=scripts/integ-p1.sh
source scripts/integ-p1.sh
WORK_DIR=$(mktemp -d)
trap 'rm -rf "$WORK_DIR"' EXIT
STACK_PREFIX=success-123-1-
MOCK_BUCKET=success-123-1-p1-retained
MOCK_OWNER="$MOCK_BUCKET"
MOCK_VERSION=None
MOCK_KEYS=fixture.txt
aws() {
  printf '%s\n' "$*" >> "$WORK_DIR/calls"
  case "$1 $2" in
    's3api list-buckets') printf '%s\n' "$MOCK_BUCKET" 'unrelated-bucket' ;;
    's3api get-bucket-tagging') echo "$MOCK_OWNER" ;;
    's3api get-bucket-versioning') echo "$MOCK_VERSION" ;;
    's3api list-objects-v2') echo "$MOCK_KEYS" ;;
    's3api delete-object'|'s3api delete-bucket') return 0 ;;
    *) echo "Unexpected mock call: $*" >&2; return 1 ;;
  esac
}
: > "$WORK_DIR/calls"
p1_cleanup_buckets "$STACK_PREFIX"
grep -q "delete-object --bucket $MOCK_BUCKET --key fixture.txt" "$WORK_DIR/calls"
if grep -q 'delete-.*unrelated-bucket' "$WORK_DIR/calls"; then exit 1; fi

: > "$WORK_DIR/calls"
MOCK_OWNER=someone-else
if p1_cleanup_buckets "$STACK_PREFIX"; then exit 1; fi
if grep -q 'delete-' "$WORK_DIR/calls"; then exit 1; fi

: > "$WORK_DIR/calls"
MOCK_OWNER="$MOCK_BUCKET"
MOCK_VERSION=Enabled
if p1_cleanup_buckets "$STACK_PREFIX"; then exit 1; fi
if grep -q 'delete-' "$WORK_DIR/calls"; then exit 1; fi

: > "$WORK_DIR/calls"
MOCK_VERSION=None
MOCK_KEYS=unexpected.txt
if p1_cleanup_buckets "$STACK_PREFIX"; then exit 1; fi
if grep -q 'delete-object' "$WORK_DIR/calls"; then exit 1; fi

: > "$WORK_DIR/calls"
MOCK_KEYS=fixture.txt
p1_cleanup_buckets success-
if grep -q 'delete-' "$WORK_DIR/calls"; then exit 1; fi

# Captured streams, exit code and redaction must survive an ordinary failure.
INTEG_OUT="$WORK_DIR/out"
if p1_run bash -c 'echo stdout; echo stderr >&2; exit 1' > /dev/null 2>&1; then exit 1; fi
grep -q $'\t1\t' "$WORK_DIR/p1-results.tsv"
printf 'https://example.invalid/?signature=secret 123456789012 hidden\n' > "$WORK_DIR/p1-redaction.log"
p1_save_evidence
grep -q '\[URL\] \[ACCOUNT\] \[NOECHO\]' "$INTEG_OUT/p1/p1-redaction.log"
echo 'P1 offline ownership checks passed'
