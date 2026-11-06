#!/bin/bash
#
# cfnrain integration tests. They deploy real stacks.
#
# Set INTEG_BUCKET_NAME to a globally unique S3 bucket name that the
# test stack may create, for example: INTEG_BUCKET_NAME=myname-cfnrain-integ
#
# CI runs the two halves in separate jobs, so that only the AWS half gets
# credentials. This script runs both, for a local run.

set -eoux pipefail

: "${INTEG_BUCKET_NAME:?set INTEG_BUCKET_NAME to a unique bucket name for the test stack}"
export INTEG_OUT="${INTEG_OUT:-integ-out}"

./scripts/integ-offline.sh
./scripts/integ-aws.sh

# Make sure the packaged modules and Pkl templates lint
cfn-lint "$INTEG_OUT/webapp.yaml" "$INTEG_OUT/webapp-pkl.yaml"
