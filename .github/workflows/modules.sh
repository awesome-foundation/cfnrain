#!/bin/bash

set -eoux pipefail

# Zip up the modules directory and create a sha256 hash.
# Output goes to release-assets/, not dist/: goreleaser runs this hook
# after cleaning dist/ and then refuses to build into a non-empty dist/.
rm -rf release-assets
mkdir -p release-assets
cd modules
zip -r ../release-assets/modules.zip *
sha256sum -b ../release-assets/modules.zip | cut -d " " -f 1  > ../release-assets/modules.sha256
