#!/bin/sh
set -eu

data_dir=${DATA_DIR:-/app/data}
probe=

cleanup_probe() {
  [ -z "$probe" ] || rm -f -- "$probe" 2>/dev/null || true
}
trap cleanup_probe EXIT HUP INT TERM

if ! mkdir -p "$data_dir" 2>/dev/null \
  || ! test -w "$data_dir" \
  || ! probe=$(mktemp "$data_dir/.omniroute-write-probe.XXXXXX" 2>/dev/null) \
  || ! rm -f -- "$probe" 2>/dev/null; then
  printf 'error: data directory is not writable: %s\n' "$data_dir" >&2
  exit 1
fi
probe=
trap - EXIT HUP INT TERM

exec "$@"
