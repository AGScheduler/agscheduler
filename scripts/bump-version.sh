#!/usr/bin/env bash

set -euo pipefail

root_dir=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)

usage() {
  echo "usage: $0 <version>" >&2
  exit 1
}

[[ $# -eq 1 ]] || usage

new_version=${1#v}
if [[ ! "$new_version" =~ ^[0-9]+\.[0-9]+\.[0-9]+([.-][0-9A-Za-z.-]+)?$ ]]; then
  echo "invalid version '$1'; expected a semantic version or v-prefixed semantic version" >&2
  exit 1
fi

old_version=$(sed -n 's/^const Version = "\(.*\)"/\1/p' "$root_dir/version.go")
root_module=$(awk '$1 == "module" { print $2; exit }' "$root_dir/go.mod")
[[ -n "$old_version" && -n "$root_module" ]] || {
  echo "failed to read the current version or root module path" >&2
  exit 1
}

if [[ "$old_version" == "$new_version" ]]; then
  echo "version is already $new_version"
  exit 0
fi

mapfile -t modules < <(awk '!/^[[:space:]]*(#|$)/' "$root_dir/scripts/modules.txt")

for module_dir in "${modules[@]}"; do
  go_mod="$root_dir/$module_dir/go.mod"
  [[ -f "$go_mod" ]] || {
    echo "missing go.mod for registered module $module_dir" >&2
    exit 1
  }

  if [[ "$module_dir" == "." ]]; then
    current_module=$root_module
  else
    current_module="$root_module/$module_dir"
  fi

  for dependency_module_dir in "${modules[@]}"; do
    if [[ "$dependency_module_dir" == "." ]]; then
      dependency_module=$root_module
    else
      dependency_module="$root_module/$dependency_module_dir"
    fi
    [[ "$dependency_module" != "$current_module" ]] || continue

    if awk -v dependency="$dependency_module" \
      '{ for (field = 1; field <= NF; field++) if ($field == dependency) found = 1 } END { exit found ? 0 : 1 }' \
      "$go_mod"; then
      (cd "$root_dir/$module_dir" && go mod edit -require="$dependency_module@v$new_version")
    fi
  done
done

perl -pi -e "s/const Version = \"\Q$old_version\E\"/const Version = \"$new_version\"/" "$root_dir/version.go"
perl -pi -e "s/ v\Q$old_version\E =>/ v$new_version =>/g" "$root_dir/go.work"
perl -pi -e "s/v\Q$old_version\E/v$new_version/g" "$root_dir/README.md" "$root_dir/README.zh-CN.md"

echo "updated module version from $old_version to $new_version"
echo "next: run 'make tidy && make check-deps && make test'"
