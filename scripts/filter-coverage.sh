#!/usr/bin/env bash

set -euo pipefail

if [[ $# -eq 0 ]]; then
  echo "usage: $0 PROFILE [PROFILE...]" >&2
  exit 1
fi

temp_profile=

cleanup() {
  if [[ -n "$temp_profile" ]]; then
    rm -f "$temp_profile"
  fi
}
trap cleanup EXIT

for profile in "$@"; do
  if [[ ! -f "$profile" ]]; then
    echo "coverage profile not found: $profile" >&2
    exit 1
  fi

  temp_profile=$(mktemp "${profile}.XXXXXX")
  awk '
    NR == 1 ||
    ($0 !~ /\/examples\// &&
     $0 !~ /\/internal\/adaptertest\// &&
     $0 !~ /\/scripts\/[^:]+\.go:/ &&
     $0 !~ /\/services\/proto\/[^:]+\.pb\.go:/)
  ' "$profile" >"$temp_profile"

  if [[ $(wc -l <"$temp_profile") -le 1 ]]; then
    rm -f "$profile"
    temp_profile=
    continue
  fi

  mv "$temp_profile" "$profile"
  temp_profile=
done
