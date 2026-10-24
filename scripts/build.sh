#!/bin/bash

set -eoux pipefail

go build ./cmd/cfnrain
staticcheck ./...
go vet ./...

