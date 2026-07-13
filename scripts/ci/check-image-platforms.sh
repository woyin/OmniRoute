#!/bin/sh
set -eu

image=${1:?usage: check-image-platforms.sh IMAGE}
raw=$(docker buildx imagetools inspect --raw "$image")
platforms=$(printf '%s\n' "$raw" | python3 -c '
import json, sys
manifest = json.load(sys.stdin)
for item in manifest.get("manifests", []):
    platform = item.get("platform", {})
    os = platform.get("os")
    architecture = platform.get("architecture")
    if os and architecture and (os, architecture) != ("unknown", "unknown"):
        print(f"{os}/{architecture}")
' | LC_ALL=C sort)
expected=$(printf '%s\n' linux/amd64 linux/arm64)

if [ "$platforms" != "$expected" ]; then
  printf 'expected exactly linux/amd64 and linux/arm64 for %s; found:\n%s\n' "$image" "$platforms" >&2
  exit 1
fi

printf 'verified image platforms for %s:\n%s\n' "$image" "$platforms"
