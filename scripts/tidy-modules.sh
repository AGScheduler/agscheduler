#!/usr/bin/env bash

set -euo pipefail

root_dir=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
mapfile -t modules < <(awk '!/^[[:space:]]*(#|$)/' "$root_dir/scripts/modules.txt")
root_module=$(awk '$1 == "module" { print $2; exit }' "$root_dir/go.mod")

module_path() {
  local dir=$1
  local rel=${dir#./}
  if [[ "$rel" == "." ]]; then
    printf '%s\n' "$root_module"
  else
    printf '%s/%s\n' "$root_module" "$rel"
  fi
}

for dir in "${modules[@]}"; do
  current_path=$(module_path "$dir")
  echo "tidy $dir"

  for other_dir in "${modules[@]}"; do
    other_path=$(module_path "$other_dir")
    if [[ "$other_path" == "$current_path" ]]; then
      continue
    fi

    other_rel=${other_dir#./}
    target="$root_dir/${other_rel}"
    (cd "$root_dir/$dir" && go mod edit -replace="$other_path=$target")
  done

  tidy_status=0
  (cd "$root_dir/$dir" && GOWORK=off go mod tidy) || tidy_status=$?

  for other_dir in "${modules[@]}"; do
    other_path=$(module_path "$other_dir")
    if [[ "$other_path" == "$current_path" ]]; then
      continue
    fi
    (cd "$root_dir/$dir" && go mod edit -dropreplace="$other_path") 2>/dev/null || true
  done

  if [[ $tidy_status -ne 0 ]]; then
    exit "$tidy_status"
  fi
done
