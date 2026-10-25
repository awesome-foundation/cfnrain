#!/bin/bash

set -eou pipefail

echo "Building rain..."
./scripts/build.sh

SOURCE="internal/cmd/build/tmpl"
OUT="test/templates/build"
FILES="${OUT}/**/*.yaml"
RULES="${SOURCE}/scripts/rules.guard"

echo "Building templates..."
echo "Building bucket..."
./cfnrain build --recommend bucket bucket -o ${OUT}/bucket/bucket.yaml
echo "Building website..."
./cfnrain build --recommend bucket website -o ${OUT}/bucket/website.yaml
echo "Building codecommit pipeline..."
./cfnrain build --recommend pipeline codecommit -o ${OUT}/pipeline/codecommit.yaml
echo "Building s3 pipeline..."
./cfnrain build --recommend pipeline s3 -o ${OUT}/pipeline/s3.yaml
echo "Building ecs cluster..."
./cfnrain build --recommend ecs cluster -o ${OUT}/ecs/cluster.yaml
echo "Building VPC..."
./cfnrain build --recommend vpc vpc -o ${OUT}/vpc/vpc.yaml
echo "Building Event Bridge..."
./cfnrain build --recommend eventbridge central-logs -o ${OUT}/eventbridge/central-logs.yaml
echo "Building Web App..."
./cfnrain build --recommend webapp webapp -o ${OUT}/webapp/webapp.yaml

echo "Linting..."
cfn-lint ${FILES}

echo "Checkov..."
checkov --framework cloudformation --quiet -f ${FILES}

echo "Guard..."
cfn-guard validate --data ${OUT}/bucket --rules ${RULES} --show-summary fail
cfn-guard validate --data ${OUT}/pipeline --rules ${RULES} --show-summary fail
cfn-guard validate --data ${OUT}/ecs --rules ${RULES} --show-summary fail
cfn-guard validate --data ${OUT}/vpc --rules ${RULES} --show-summary fail
cfn-guard validate --data ${OUT}/eventbridge --rules ${RULES} --show-summary fail
cfn-guard validate --data ${OUT}/webapp --rules ${RULES} --show-summary fail

echo "Success"
