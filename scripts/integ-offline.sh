#!/bin/bash
#
# The cfnrain integration tests that need no AWS account.

set -eoux pipefail

./scripts/test.sh
python3 scripts/integ-redact-test.py

# A7: preserve success and error exit codes, with warnings only on stderr.
source scripts/integ-warnings.sh
warning_dir=$(mktemp -d)
trap 'rm -rf "$warning_dir"' EXIT
./cfnrain build AWS::S3::Bucket > "$warning_dir/build.yaml" 2> "$warning_dir/build.err"
assert_warning "$warning_dir/build.yaml" "$warning_dir/build.err"
./cfnrain fmt "$warning_dir/build.yaml" >/dev/null
result=0
./cfnrain forecast --experimental "$warning_dir/missing.yaml" > "$warning_dir/forecast.out" 2> "$warning_dir/forecast.err" || result=$?
[[ "$result" == 1 ]]
assert_warning "$warning_dir/forecast.out" "$warning_dir/forecast.err"
# A7: acknowledged cc hooks use the existing credential-free Go test.
# cc deploy can bootstrap AWS before it reads its input, so never invoke it here.
go test ./internal/cmd/rain -run '^TestDeprecationWarnings$' -count=1
# A7: without acknowledgment the existing failure has no deprecation warning.
for command in forecast cc; do
  args=(forecast "$warning_dir/missing.yaml")
  [[ "$command" != cc ]] || args=(cc deploy "$warning_dir/missing.yaml" missing -y)
  result=0
  ./cfnrain "${args[@]}" > "$warning_dir/noack.out" 2> "$warning_dir/noack.err" || result=$?
  [[ "$result" == 1 ]]
  if grep -q 'is deprecated' "$warning_dir/noack.out" "$warning_dir/noack.err"; then
    exit 1
  fi
done
# A3/A5/A6: fixtures also pass the local input and packaging paths.
./cfnrain fmt - < test/templates/integ-p0/original.yaml > "$warning_dir/original.yaml"
./cfnrain pkg --experimental test/templates/integ-p0/modules.yaml > "$warning_dir/modules.yaml"
./cfnrain fmt test/templates/integ-p0/original.yaml --node-style original > "$warning_dir/original-style.yaml"
for command in 'fmt --write -' 'merge - -' 'diff - -' 'deploy --express --changeset unused unused'; do
  result=0
  # Intentional splitting exercises these fixed CLI argument lists.
  # shellcheck disable=SC2086
  ./cfnrain $command < test/templates/integ-handle.yaml >/dev/null 2>&1 || result=$?
  [[ "$result" == 1 ]]
done

# A3: exercise the complete merge pipeline before any AWS invocation.
./cfnrain fmt - < test/templates/integ-handle.yaml > "$warning_dir/first.yaml"
./cfnrain merge "$warning_dir/first.yaml" - < test/templates/integ-p0/second-handle.yaml > "$warning_dir/merged.yaml"
./cfnrain diff "$warning_dir/first.yaml" - < "$warning_dir/merged.yaml" > "$warning_dir/diff.txt"
grep -q SecondHandle "$warning_dir/diff.txt"
printf 'Resources: [invalid' > "$warning_dir/malformed.yaml"
result=0
./cfnrain fmt - < "$warning_dir/malformed.yaml" >/dev/null 2>&1 || result=$?
[[ "$result" == 1 ]]
# A5 (#161): resource ordering must remain deterministic.
./cfnrain pkg --experimental test/templates/integ-p0/modules.yaml > "$warning_dir/modules-again.yaml"
cmp "$warning_dir/modules.yaml" "$warning_dir/modules-again.yaml"
./cfnrain build -l

./cfnrain fmt test/templates/fmtfindinmap.yaml
./cfnrain fmt test/templates/fmtmultiwithgt.yaml
./cfnrain fmt test/templates/fmtziplinesok.yaml

./cfnrain pkg cft/pkg/tmpl/embed-template.yaml
./cfnrain pkg cft/pkg/tmpl/include-template.yaml

# Make sure merge works
./cfnrain merge test/templates/merge-out-1.yaml test/templates/merge-out-2.yaml

# Make sure build recommendations work
./internal/cmd/build/tmpl/scripts/validate.sh

# Make sure pkl generation works
./cfnrain fmt test/templates/success.template --pkl
./cfnrain fmt test/templates/success.template --pkl --pkl-basic
./cfnrain fmt test/templates/condition-stringlike.yaml --pkl > test/pkl/condition-stringlike.pkl
pkl eval --project-dir test/pkl test/pkl/condition-stringlike.pkl -f yaml
