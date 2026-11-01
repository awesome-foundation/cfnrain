#!/bin/bash
#
# The cfnrain integration tests that need no AWS account.

set -eoux pipefail

./scripts/test.sh

./cfnrain build AWS::S3::Bucket
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
