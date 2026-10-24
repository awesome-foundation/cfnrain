#!/bin/bash
#
# cfnrain integration tests. They deploy real stacks.
#
# Set INTEG_BUCKET_NAME to a globally unique S3 bucket name that the
# test stack may create, for example: INTEG_BUCKET_NAME=myname-cfnrain-integ

set -eoux pipefail

: "${INTEG_BUCKET_NAME:?set INTEG_BUCKET_NAME to a unique bucket name for the test stack}"

./scripts/test.sh

./cfnrain ls

./cfnrain deploy test/templates/success.template success-test7 -y --params BucketName="$INTEG_BUCKET_NAME"

./cfnrain cat success-test7

./cfnrain logs success-test7

./cfnrain logs --chart success-test7

./cfnrain rm success-test7 -y

# Unnamed stack
./cfnrain deploy test/templates/success.template -y --params BucketName="$INTEG_BUCKET_NAME"
./cfnrain cat success
./cfnrain ls success
./cfnrain rm -y success

# Change sets
# Create a named changeset
#./cfnrain deploy --no-exec test/templates/success.template success-test7 success-changeset-name -y --params BucketName="$INTEG_BUCKET_NAME"
#./cfnrain ls -c success-test7 success-changeset-name

#./cfnrain rm -c -y success-test7 success-changeset-name
# This leaves the stack as review in progress, and then you can't delete it
# Seems like a bug!
# It also fails if we try to delete the stack, not just the stackset
#./cfnrain rm -y success-test7 

./cfnrain build AWS::S3::Bucket
./cfnrain build -l

./cfnrain fmt test/templates/fmtfindinmap.yaml
./cfnrain fmt test/templates/fmtmultiwithgt.yaml
./cfnrain fmt test/templates/fmtziplinesok.yaml

./cfnrain pkg cft/pkg/tmpl/s3-props-template.yaml
./cfnrain pkg cft/pkg/tmpl/embed-template.yaml
./cfnrain pkg cft/pkg/tmpl/include-template.yaml
./cfnrain pkg cft/pkg/tmpl/s3-template.yaml
./cfnrain pkg cft/pkg/tmpl/s3http-template.yaml
# Given a Template with an Extension value of txt, When Packaged, Then the S3 URI ends '.txt'
./cfnrain pkg cft/pkg/tmpl/s3-extension-template.yaml | yq --exit-status '.Resources.Test.Properties.TheS3URI | test("\.txt$")'

# Make sure merge works
./cfnrain merge test/templates/merge-out-1.yaml test/templates/merge-out-2.yaml

# Make sure build recommendations work
./internal/cmd/build/tmpl/scripts/validate.sh

# Make sure modules package and lint
./cfnrain pkg -x  test/webapp/webapp.yaml | cfn-lint

# Make sure pkl generation works
./cfnrain fmt test/templates/success.template --pkl
./cfnrain fmt test/templates/success.template --pkl --pkl-basic
./cfnrain fmt test/templates/condition-stringlike.yaml --pkl > test/pkl/condition-stringlike.pkl
pkl eval --project-dir test/pkl test/pkl/condition-stringlike.pkl -f yaml
./cfnrain pkg test/webapp/webapp.pkl | cfn-lint

