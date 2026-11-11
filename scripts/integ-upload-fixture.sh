#!/bin/bash
# Write repeated upload references with different extensions in one template.
write_upload_fixture() {
  local format="$1"
  shift
  local extension i=0
  printf 'Resources:\n  Test:\n    Type: A::B::C\n    Properties:\n'
  for extension in "$@"; do
    printf '      Value%s: !Rain::S3\n        Path: cache.txt\n        Format: %s\n' "$i" "$format"
    if [[ -n "$extension" ]]; then
      printf '        Extension: %s\n' "$extension"
    fi
    i=$((i + 1))
  done
}
