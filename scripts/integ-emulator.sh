#!/usr/bin/env bash
# Local client/server checks. Requires Linux amd64, Go, AWS CLI v2, curl,
# GnuPG, and dpkg-deb. Run from any directory.
# Excluded: module outputs, nested preview child changes, Cloud Control,
# and the type catalog. Track fidelity reports at https://github.com/aws/aws-emulator/issues.
set -euo pipefail
cd "$(dirname "$0")/.."
logs=$(realpath -m "${EMULATOR_LOG_DIR:-emulator-logs}")
mkdir -p "$logs"
exec > >(tee "$logs/commands.log") 2>&1
fail() {
    echo "FAIL: $*" >&2
    exit 1
}
scratch=$(mktemp -d)
pid=
cleanup() {
    if [[ -n "$pid" ]]; then
        kill "$pid" 2>/dev/null || true
        wait "$pid" 2>/dev/null || true
    fi
    rm -rf "$scratch"
}
trap cleanup EXIT
trap 'exit 130' INT
trap 'exit 143' TERM
# Remove inherited AWS settings, proxy routes, and Rain defaults.
for variable in ${!AWS_@} ${!RAIN_@}; do unset "$variable"; done
unset HTTP_PROXY HTTPS_PROXY ALL_PROXY http_proxy https_proxy all_proxy
export NO_PROXY=127.0.0.1,localhost no_proxy=127.0.0.1,localhost

version=1.0.322
sha=ef3c75e5d4196bf2c0170b55b75b01e340e0ef497e636b700a1efe68d68a60d7
fingerprint=98BCF336A3BE53B734E967C52E5BF49BD940F624
package=aws-emulator_${version}-1_amd64.deb
base=https://downloads.aws.com/aws-emulator/packages/aws-emulator/LTS
# Deliberately do not cache executable packages across CI jobs.
for suffix in '' .sig; do
    curl --fail --silent --show-error --location --connect-timeout 15 \
        --retry 3 --retry-all-errors --retry-delay 5 --retry-max-time 240 --max-time 180 "$base/$package$suffix" -o "$scratch/$package$suffix"
done
printf '%s  %s\n' "$sha" "$scratch/$package" | sha256sum --check || fail 'Package SHA-256 mismatch.'
# Source: https://downloads.aws.com/keys/gpg/current.asc (pinned for 1.0.322).
mkdir -m 700 "$scratch/gnupg"
gpg --homedir "$scratch/gnupg" --batch --import scripts/aws-emulator-key.asc
actual=$(gpg --homedir "$scratch/gnupg" --with-colons --fingerprint | awk -F: '$1 == "fpr" {print $10; exit}')
[[ "$actual" == "$fingerprint" ]] || fail "Signing key fingerprint is $actual, expected $fingerprint."
gpg --homedir "$scratch/gnupg" --batch --status-fd 1 --verify "$scratch/$package.sig" "$scratch/$package" > "$scratch/signature" || fail 'Package signature verification failed.'
# Require the pinned primary key (or its signing subkey).
grep -E "^\[GNUPG:\] VALIDSIG $fingerprint |^\[GNUPG:\] VALIDSIG .* $fingerprint$" "$scratch/signature" || fail 'Signature is not from the pinned signing key.'
[[ $(dpkg-deb -f "$scratch/$package" Version) == "$version-1" ]] || fail "Expected package version $version-1."
dpkg-deb -x "$scratch/$package" "$scratch/root"
emulator="$scratch/root/opt/aws-emulator/bin/aws-emulator"
"$emulator" --version
git rev-parse HEAD
# Build with the local Go cache before starting the emulator.
timeout 180 go build -o "$scratch/cfnrain" ./cmd/cfnrain
rain="$scratch/cfnrain"
emulator_home="$scratch/home"
mkdir -p "$emulator_home"
export RAIN_CACHE_DIR="$scratch/rain-cache"
env HOME="$emulator_home" "$emulator" --bind 127.0.0.1 --port 9297 --auth on --telemetry off > "$logs/server.log" 2>&1 &
pid=$!
export AWS_ENDPOINT_URL=https://127.0.0.1:9297
export AWS_CONFIG_FILE="$emulator_home/.aws/emulator/9297/config"
export AWS_SHARED_CREDENTIALS_FILE="$emulator_home/.aws/emulator/9297/emulator_credentials"
export AWS_CA_BUNDLE="$emulator_home/.aws/emulator/9297/ca.pem"
export AWS_REGION=us-east-1 AWS_DEFAULT_REGION=us-east-1
export AWS_EC2_METADATA_DISABLED=true AWS_MAX_ATTEMPTS=1 AWS_PAGER=
ready=false
for ((i=0; i<100; i++)); do
    if ! kill -0 "$pid" 2>/dev/null; then
        cat "$logs/server.log"
        fail 'Emulator exited before readiness. Check server.log; port 9297 may already be in use.'
    fi
    if [[ -f "$AWS_CA_BUNDLE" ]] && curl -fsS --max-time 1 --cacert "$AWS_CA_BUNDLE" "$AWS_ENDPOINT_URL/__admin/health" > /dev/null 2>&1; then
        ready=true
        break
    fi
    sleep 0.1
done
[[ "$ready" == true ]] || fail 'Emulator did not become ready on port 9297. See server.log.'

# Each command records its exit code and output, including expected failures.
check() {
    local name=$1 expected=$2 pattern=$3 code=0
    shift 3
    printf '\n>>> %s: ' "$name"
    printf '%q ' "$@"
    printf '\n'
    timeout --kill-after=5 45 "$@" > "$logs/$name.log" 2>&1 || code=$?
    cat "$logs/$name.log"
    printf 'exit=%s\n' "$code"
    [[ "$code" == "$expected" ]] || fail "$name: exit $code, expected $expected."
    grep -F -- "$pattern" "$logs/$name.log" || fail "$name: expected output not found: $pattern"
}
# Read back deletion through the service, independently of cfnrain's message.
absent() {
    check "absent-$1" 254 "Stack with id $1 does not exist" aws cloudformation describe-stacks --stack-name "$1"
}
# The first AWS API call must identify the emulator account exactly.
check identity 0 000000000000 aws sts get-caller-identity --query Account --output text
[[ $(cat "$logs/identity.log") == 000000000000 ]] || fail 'STS did not return the emulator account.'
check version 0 cfnrain "$rain" --version
check create 0 CREATE_COMPLETE "$rain" deploy test/templates/integ-handle.yaml emulator-handle -y
check update 0 UPDATE_COMPLETE "$rain" deploy test/templates/integ-handle-two.yaml emulator-handle -y
check noop 0 'Deploy was skipped' "$rain" deploy test/templates/integ-handle-two.yaml emulator-handle -y
check preview 0 'changeset created but not executed: success-preview' "$rain" deploy test/templates/integ-handle.yaml emulator-handle success-preview --no-exec -y
check execute 0 UPDATE_COMPLETE "$rain" deploy --changeset emulator-handle success-preview -y
check failure 1 UPDATE_ROLLBACK_COMPLETE "$rain" deploy test/templates/integ-failure.yaml emulator-handle -y
check rollback 0 UPDATE_ROLLBACK_COMPLETE aws cloudformation describe-stacks --stack-name emulator-handle --query 'Stacks[0].StackStatus' --output text
check cat 0 'Integration test stack with no service resources' "$rain" cat emulator-handle
if grep -E 'SecondHandle|NeverSignalled' "$logs/cat.log"; then
    fail 'cat returned resources from the wrong template version.'
fi
check logs 0 UPDATE_ROLLBACK_COMPLETE "$rain" logs emulator-handle
check protected 0 CREATE_COMPLETE "$rain" deploy test/templates/integ-handle.yaml emulator-protected -y -t
check protection 0 True aws cloudformation describe-stacks --stack-name emulator-protected --query 'Stacks[0].EnableTerminationProtection' --output text
check protected-rm 0 'Successfully deleted' "$rain" rm emulator-protected -y
absent emulator-protected
check rm 0 'Successfully deleted' "$rain" rm emulator-handle -y
absent emulator-handle
check bucket 0 CREATE_COMPLETE "$rain" deploy test/templates/integ-p0/bucket.yaml emulator-bucket -y --params BucketName=emulator-test-bucket,Revision=first
check bucket-read 0 first aws s3api get-bucket-tagging --bucket emulator-test-bucket --query "TagSet[?Key=='Revision'].Value" --output text
[[ $(cat "$logs/bucket-read.log") == first ]] || fail 'Bucket Revision tag is not exactly first.'
cat > "$scratch/ssm.yaml" <<'YAML'
Resources:
  Parameter:
    Type: AWS::SSM::Parameter
    Properties:
      Name: emulator-test-parameter
      Type: String
      Value: emulator-test-value
YAML
check ssm 0 CREATE_COMPLETE "$rain" deploy "$scratch/ssm.yaml" emulator-ssm -y
check ssm-read 0 emulator-test-value aws ssm get-parameter --name emulator-test-parameter --query Parameter.Value --output text
printf 'emulator upload bytes\n' > "$scratch/payload.txt"
cat > "$scratch/upload.yaml" <<'YAML'
Resources:
  Handle:
    Type: AWS::CloudFormation::WaitConditionHandle
Outputs:
  Artifact:
    Value: !Rain::S3 payload.txt
YAML
key=$(sha256sum "$scratch/payload.txt" | cut -d ' ' -f 1)
check pkg 0 "s3://rain-artifacts-000000000000-us-east-1/$key" "$rain" pkg "$scratch/upload.yaml"
check download 0 ContentLength aws s3api get-object --bucket rain-artifacts-000000000000-us-east-1 --key "$key" "$scratch/download.txt"
cmp "$scratch/payload.txt" "$scratch/download.txt" || fail 'Downloaded artifact bytes differ.'
check express-create 0 CREATE_COMPLETE "$rain" deploy test/templates/integ-handle.yaml emulator-express -y --express
check express-preview 0 'changeset created but not executed' "$rain" deploy test/templates/integ-handle-two.yaml emulator-express success-express-preview -y --express --no-exec
check express-mode 0 EXPRESS aws cloudformation describe-change-set --stack-name emulator-express --change-set-name success-express-preview --query DeploymentConfig.Mode --output text
check express-update 0 UPDATE_COMPLETE "$rain" deploy --changeset emulator-express success-express-preview -y
check conflict 1 NAME_CONFLICT_VALIDATION "$rain" deploy test/templates/integ-p0/bucket.yaml emulator-conflict success-conflict -y --params BucketName=emulator-test-bucket,Revision=first
for detail in Bucket 'AWS::S3::Bucket' 'already exists'; do
    grep -F "$detail" "$logs/conflict.log" || fail "Conflict detail missing: $detail"
done
for stack in emulator-express emulator-bucket emulator-ssm emulator-conflict; do
    check "rm-$stack" 0 'Successfully deleted' "$rain" rm "$stack" -y
    absent "$stack"
done
echo 'All emulator checks passed.'
