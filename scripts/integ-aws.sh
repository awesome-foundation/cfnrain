#!/bin/bash
#
# The cfnrain integration tests that need an AWS account. They deploy real stacks.
#
# Set INTEG_BUCKET_NAME to a globally unique S3 bucket name that the
# test stack may create, for example: INTEG_BUCKET_NAME=myname-cfnrain-integ
# Packaged templates go to INTEG_OUT (default: integ-out), so that a later
# step can lint them without AWS credentials.

set -eoux pipefail

: "${INTEG_BUCKET_NAME:?set INTEG_BUCKET_NAME to a unique bucket name for the test stack}"
INTEG_OUT="${INTEG_OUT:-integ-out}"
mkdir -p "$INTEG_OUT"

go build ./cmd/cfnrain

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

# These upload to the artifact bucket
./cfnrain pkg cft/pkg/tmpl/s3-props-template.yaml
./cfnrain pkg cft/pkg/tmpl/s3-template.yaml
./cfnrain pkg cft/pkg/tmpl/s3http-template.yaml
# Given a Template with an Extension value of txt, When Packaged, Then the S3 URI ends '.txt'
./cfnrain pkg cft/pkg/tmpl/s3-extension-template.yaml | yq --exit-status '.Resources.Test.Properties.TheS3URI | test("\.txt$")'

# Modules and Pkl package and upload. integ.sh lints the output.
./cfnrain pkg -x test/webapp/webapp.yaml > "$INTEG_OUT/webapp.yaml"
./cfnrain pkg test/webapp/webapp.pkl > "$INTEG_OUT/webapp-pkl.yaml"
