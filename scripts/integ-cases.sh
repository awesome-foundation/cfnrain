#!/bin/bash
# Shared capture and evidence helpers; assertions are supplied by the caller.
# Artifact objects expire through the existing seven-day bucket lifecycle.
# Keep an inventory for diagnosis. The role deliberately cannot delete objects.
record_object() { printf '%s\n' "$1" >> "$OBJECT_MANIFEST"; }
preserve_logs() {
  mkdir -p "$INTEG_OUT/logs"
  git rev-parse HEAD > "$INTEG_OUT/logs/revision.txt"
  # Exclude templates and presigned handle URLs. Redact identifiers in diagnostics.
  local log
  for log in "$WORK_DIR"/*.log; do
    [[ -f "$log" ]] || continue
    python3 scripts/integ-redact.py < "$log" > "$INTEG_OUT/logs/${log##*/}"
  done
}
# Preserve streams, status, and duration for each named invocation.
run_case() {
  local id="$1" expected="$2" start=$SECONDS result=0 stream remaining=$((2100 - SECONDS))
  shift 2
  (( remaining > 0 )) || fail "case time budget exhausted before $id; remaining cases did not run"
  (( remaining <= 600 )) || remaining=600
  timeout "$remaining" "$@" > "$WORK_DIR/$id-stdout.log" 2> "$WORK_DIR/$id-stderr.log" || result=$?
  printf 'exit=%s duration=%s\n' "$result" "$((SECONDS-start))" > "$WORK_DIR/$id-result.log"
  if [[ "$expected" == capture ]]; then
    cat "$WORK_DIR/$id-stdout.log"
    cat "$WORK_DIR/$id-stderr.log" >&2
    return "$result"
  fi
  if [[ "$result" != "$expected" ]]; then
    preserve_logs
    for stream in stdout stderr; do
      echo "$id $stream:" >&2
      tail -n 80 "$INTEG_OUT/logs/$id-$stream.log" >&2
    done
  fi
  assert_eq "$result" "$expected" "$id exit code"
}
