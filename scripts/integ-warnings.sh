#!/bin/bash
# Shared A7 stream assertions. This file does not invoke AWS.
assert_warning() {
  local stdout="$1" stderr="$2"
  ! grep -q 'is deprecated' "$stdout" || return 1
  [[ $(grep -c 'is deprecated and will be removed in cfnrain 2.0' "$stderr") == 1 ]] || return 1
  grep -q 'https://github.com/awesome-foundation/cfnrain/discussions/20' "$stderr"
}
