#!/usr/bin/env bash
set -euo pipefail

verify_assets() {
  local release_dir=$1 version=$2
  local expected=(
    "checksums.txt"
    "omniroute_${version}_darwin_aarch64.tar.gz"
    "omniroute_${version}_darwin_x86_64.tar.gz"
    "omniroute_${version}_linux_aarch64.tar.gz"
    "omniroute_${version}_linux_x86_64.tar.gz"
    "omniroute_${version}_windows_x86_64.zip"
  )
  local archives=("${expected[@]:1}") actual=() file i

  (
    cd "$release_dir"
    sha256sum "${archives[@]}" > checksums.txt
    sha256sum -c checksums.txt
  )

  while IFS= read -r file; do
    actual+=("${file#./}")
  done < <(cd "$release_dir" && find . -maxdepth 1 -type f -print | LC_ALL=C sort)

  if [[ ${#actual[@]} -ne ${#expected[@]} ]]; then
    printf 'unexpected release assets:\nexpected: %s\nactual:   %s\n' "${expected[*]}" "${actual[*]}" >&2
    return 1
  fi

  for i in "${!expected[@]}"; do
    if [[ ${actual[$i]} != "${expected[$i]}" ]]; then
      printf 'unexpected release assets:\nexpected: %s\nactual:   %s\n' "${expected[*]}" "${actual[*]}" >&2
      return 1
    fi
  done
}

self_test() {
  local root version=v0.0.0-go
  root=$(mktemp -d)
  trap 'rm -rf "$root"' RETURN
  mkdir "$root/release"
  touch "$root/release/omniroute_${version}_darwin_aarch64.tar.gz" \
    "$root/release/omniroute_${version}_darwin_x86_64.tar.gz" \
    "$root/release/omniroute_${version}_linux_aarch64.tar.gz" \
    "$root/release/omniroute_${version}_linux_x86_64.tar.gz" \
    "$root/release/omniroute_${version}_windows_x86_64.zip"

  verify_assets "$root/release" "$version" >/dev/null
  [[ $(wc -l < "$root/release/checksums.txt") -eq 5 ]]
  grep -Eq '  omniroute_.*\.(tar\.gz|zip)$' "$root/release/checksums.txt"
  ! grep -Eq '(^|/)(omniroute_(linux|darwin|windows)_[^.]*)$' "$root/release/checksums.txt"

  rm "$root/release/omniroute_${version}_linux_aarch64.tar.gz"
  ! verify_assets "$root/release" "$version" >/dev/null 2>&1
  touch "$root/release/omniroute_${version}_linux_aarch64.tar.gz" "$root/release/extra"
  ! verify_assets "$root/release" "$version" >/dev/null 2>&1
  printf 'verify-release-assets self-test: ok\n'
}

if [[ ${1:-} == --self-test ]]; then
  self_test
  exit
fi

verify_assets "${1:?usage: verify-release-assets.sh RELEASE_DIR VERSION | --self-test}" \
  "${2:?usage: verify-release-assets.sh RELEASE_DIR VERSION | --self-test}"
