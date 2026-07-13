#!/bin/sh
set -eu

data_dir=${DATA_DIR:-/app/data}
probe="$data_dir/.omniroute-write-probe-$$"

if ! mkdir -p "$data_dir" 2>/dev/null \
  || ! test -w "$data_dir" \
  || ! : >"$probe" 2>/dev/null \
  || ! rm -f "$probe" 2>/dev/null; then
  rm -f "$probe" 2>/dev/null || true
  printf 'error: data directory is not writable: %s\n' "$data_dir" >&2
  exit 1
fi

exec "$@"
