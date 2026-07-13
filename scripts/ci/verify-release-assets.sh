#!/usr/bin/env bash
set -euo pipefail

release_dir=${1:?usage: verify-release-assets.sh RELEASE_DIR VERSION}
version=${2:?usage: verify-release-assets.sh RELEASE_DIR VERSION}

expected=(
  "checksums.txt"
  "omniroute_${version}_darwin_aarch64.tar.gz"
  "omniroute_${version}_darwin_x86_64.tar.gz"
  "omniroute_${version}_linux_aarch64.tar.gz"
  "omniroute_${version}_linux_x86_64.tar.gz"
  "omniroute_${version}_windows_x86_64.zip"
)

archives=("${expected[@]:1}")
(
  cd "$release_dir"
  sha256sum "${archives[@]}" > checksums.txt
  sha256sum -c checksums.txt
)

actual=()
while IFS= read -r file; do
  actual+=("${file#./}")
done < <(cd "$release_dir" && find . -maxdepth 1 -type f -print | LC_ALL=C sort)

if [[ ${#actual[@]} -ne ${#expected[@]} ]]; then
  printf 'unexpected release assets:\nexpected: %s\nactual:   %s\n' "${expected[*]}" "${actual[*]}" >&2
  exit 1
fi

for i in "${!expected[@]}"; do
  if [[ ${actual[$i]} != "${expected[$i]}" ]]; then
    printf 'unexpected release assets:\nexpected: %s\nactual:   %s\n' "${expected[*]}" "${actual[*]}" >&2
    exit 1
  fi
done
