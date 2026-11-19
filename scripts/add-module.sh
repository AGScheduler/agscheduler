#!/usr/bin/env bash

set -euo pipefail

root_dir=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)

usage() {
  echo "usage: $0 <module-directory> [dependency@version ...]" >&2
  exit 1
}

[[ $# -ge 1 ]] || usage

module_dir=$1
shift
module_dir=${module_dir#./}
module_dir=${module_dir%/}

[[ -n "$module_dir" && "$module_dir" != "." ]] || usage
case "/$module_dir/" in
  */../* | */./*)
    echo "module directory must not contain '.' or '..' segments" >&2
    exit 1
    ;;
esac

root_module=$(awk '$1 == "module" { print $2; exit }' "$root_dir/go.mod")
go_version=$(awk '$1 == "go" { print $2; exit }' "$root_dir/go.mod")
current_version=$(sed -n 's/^const Version = "\(.*\)"/\1/p' "$root_dir/version.go")
module_path="$root_module/$module_dir"
target_dir="$root_dir/$module_dir"

[[ -n "$root_module" && -n "$go_version" && -n "$current_version" ]] || {
  echo "failed to read module metadata from the repository root" >&2
  exit 1
}

if [[ -e "$target_dir/go.mod" ]]; then
  echo "$module_dir already contains go.mod" >&2
  exit 1
fi

if grep -Fxq "$module_dir" "$root_dir/scripts/modules.txt"; then
  echo "$module_dir is already registered in scripts/modules.txt" >&2
  exit 1
fi

dependencies=()
allowed_dependencies=()
for dependency in "$@"; do
  if [[ "$dependency" != *@* ]]; then
    echo "dependency '$dependency' must include a version, for example vendor/pkg@v1.2.3" >&2
    exit 1
  fi
  dependencies+=("$dependency")
  dependency_path=${dependency%@*}
  if [[ "$dependency_path" != "$root_module" && "$dependency_path" != "$root_module/"* ]]; then
    allowed_dependencies+=("$dependency_path")
  fi
done

mkdir -p "$target_dir"
(
  cd "$target_dir"
  go mod init "$module_path" >/dev/null
  go mod edit -go="$go_version" -require="$root_module@v$current_version"
  for dependency in "${dependencies[@]}"; do
    go mod edit -require="$dependency"
  done
)

printf '%s\n' "$module_dir" >>"$root_dir/scripts/modules.txt"
printf '%s' "$module_dir" >>"$root_dir/scripts/module-deps.txt"
for dependency in "${allowed_dependencies[@]}"; do
  printf ' %s' "$dependency" >>"$root_dir/scripts/module-deps.txt"
done
printf '\n' >>"$root_dir/scripts/module-deps.txt"

(
  cd "$root_dir"
  go work edit \
    -use="./$module_dir" \
    -replace="$module_path@v$current_version=./$module_dir"
)

echo "created module $module_path"
echo "registered $module_dir in go.work, scripts/modules.txt, and scripts/module-deps.txt"
echo "next: add source files, then run 'make tidy && make check-deps'"
