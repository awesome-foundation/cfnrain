#!/bin/bash
#
# The cfnrain integration tests that need no AWS account.

set -eoux pipefail

./scripts/test.sh

# A7: preserve success and error exit codes, with warnings only on stderr.
source scripts/integ-warnings.sh
warning_dir=$(mktemp -d)
trap 'rm -rf "$warning_dir"' EXIT
./cfnrain build AWS::S3::Bucket > "$warning_dir/build.yaml" 2> "$warning_dir/build.err"
assert_warning "$warning_dir/build.yaml" "$warning_dir/build.err"
./cfnrain fmt "$warning_dir/build.yaml" >/dev/null
for command in forecast cc; do
  args=(forecast --experimental "$warning_dir/missing.yaml")
  [[ "$command" != cc ]] || args=(cc deploy --experimental "$warning_dir/missing.yaml" missing -y)
  result=0
  ./cfnrain "${args[@]}" > "$warning_dir/$command.out" 2> "$warning_dir/$command.err" || result=$?
  [[ "$result" == 1 ]]
  assert_warning "$warning_dir/$command.out" "$warning_dir/$command.err"
done
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
./cfnrain pkg -x test/templates/integ-p0/modules.yaml > "$warning_dir/modules.yaml"
./cfnrain fmt test/templates/integ-p0/original.yaml --node-style original > "$warning_dir/original-style.yaml"
for command in 'fmt --write -' 'merge - -' 'diff - -'; do
  result=0
  # Intentional splitting exercises these fixed CLI argument lists.
  # shellcheck disable=SC2086
  ./cfnrain $command < test/templates/integ-handle.yaml >/dev/null 2>&1 || result=$?
  [[ "$result" == 1 ]]
done

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
