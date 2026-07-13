#!/bin/sh
set -eu

repo_root=$(CDPATH= cd -- "$(dirname "$0")/../.." && pwd)
checker="$repo_root/scripts/ci/check-image-platforms.sh"
tmp=$(mktemp -d "${TMPDIR:-/tmp}/check-image-platforms-test.XXXXXX")
trap 'rm -rf "$tmp"' EXIT HUP INT TERM

mkdir -p "$tmp/bin"
cat >"$tmp/bin/docker" <<'EOF'
#!/bin/sh
cat "$IMAGETOOLS_RAW"
EOF
chmod +x "$tmp/bin/docker"

cat >"$tmp/pass.json" <<'EOF'
{"manifests":[
  {"platform":{"architecture":"amd64","os":"linux"}},
  {"platform":{"os":"linux","architecture":"arm64"}}
]}
EOF
IMAGETOOLS_RAW="$tmp/pass.json" PATH="$tmp/bin:$PATH" "$checker" ghcr.io/example/image:test

cat >"$tmp/missing.json" <<'EOF'
{"manifests":[{"platform":{"os":"linux","architecture":"amd64"}}]}
EOF
if IMAGETOOLS_RAW="$tmp/missing.json" PATH="$tmp/bin:$PATH" "$checker" ghcr.io/example/image:test >"$tmp/missing.out" 2>&1; then
  printf 'checker accepted manifest missing linux/arm64\n' >&2
  exit 1
fi
grep -q 'expected exactly linux/amd64 and linux/arm64' "$tmp/missing.out"

cat >"$tmp/extra.json" <<'EOF'
{"manifests":[
  {"platform":{"os":"linux","architecture":"amd64"}},
  {"platform":{"os":"linux","architecture":"arm64"}},
  {"platform":{"os":"linux","architecture":"s390x"}}
]}
EOF
if IMAGETOOLS_RAW="$tmp/extra.json" PATH="$tmp/bin:$PATH" "$checker" ghcr.io/example/image:test >"$tmp/extra.out" 2>&1; then
  printf 'checker accepted unexpected platform\n' >&2
  exit 1
fi
grep -q 'expected exactly linux/amd64 and linux/arm64' "$tmp/extra.out"

printf 'check-image-platforms tests passed\n'
