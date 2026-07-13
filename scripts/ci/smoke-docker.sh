#!/bin/sh
set -eu

image=${1:?usage: smoke-docker.sh IMAGE}
suffix="$$-$(date +%s)"
container="omniroute-smoke-$suffix"
negative_container="omniroute-smoke-negative-$suffix"
volume="omniroute-smoke-$suffix"
readonly_dir=$(mktemp -d "${TMPDIR:-/tmp}/omniroute-smoke.XXXXXX")
port=

cleanup() {
  docker rm -f "$container" "$negative_container" >/dev/null 2>&1 || true
  docker volume rm -f "$volume" >/dev/null 2>&1 || true
  chmod 700 "$readonly_dir" 2>/dev/null || true
  rm -rf "$readonly_dir" "$readonly_dir.out"
}
trap cleanup EXIT HUP INT TERM

health() {
  wget -qO- "http://127.0.0.1:$port/health"
}

wait_healthy() {
  attempts=0
  while [ "$attempts" -lt 60 ]; do
    body=$(health 2>/dev/null || true)
    if printf '%s' "$body" | grep -Eq '"status"[[:space:]]*:[[:space:]]*"ok"' \
      && printf '%s' "$body" | grep -Eq '"db"[[:space:]]*:[[:space:]]*"ok"'; then
      return 0
    fi
    if ! docker inspect "$container" >/dev/null 2>&1; then
      printf 'container exited before health check succeeded\n' >&2
      return 1
    fi
    attempts=$((attempts + 1))
    sleep 1
  done
  printf 'health check timed out\n' >&2
  docker logs "$container" >&2 || true
  return 1
}

start_container() {
  docker run -d --name "$container" -p 127.0.0.1::3456 -v "$volume:/app/data" "$image" >/dev/null
  port=$(docker port "$container" 3456/tcp | sed 's/.*://')
}

docker volume create "$volume" >/dev/null
start_container
wait_healthy
docker run --rm -v "$volume:/data" --entrypoint /bin/sh "$image" -c 'test -f /data/storage.sqlite'
docker rm -f "$container" >/dev/null
start_container
wait_healthy

chmod 555 "$readonly_dir"
if docker run --name "$negative_container" -v "$readonly_dir:/app/data" "$image" >"$readonly_dir.out" 2>&1; then
  printf 'container unexpectedly started with non-writable data directory\n' >&2
  exit 1
fi
if ! grep -q 'data directory is not writable' "$readonly_dir.out"; then
  printf 'missing explicit non-writable data directory error\n' >&2
  docker logs "$negative_container" >&2 || true
  exit 1
fi
rm -f "$readonly_dir.out"

printf 'Docker persistence smoke passed: %s\n' "$image"
