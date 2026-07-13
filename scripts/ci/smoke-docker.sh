#!/bin/sh
set -eu

valid_health() {
  python3 -c 'import json,sys; data=json.load(sys.stdin); sys.exit(0 if isinstance(data, dict) and data.get("status") == "ok" and data.get("db") == "ok" else 1)'
}

if [ "${SMOKE_DOCKER_SELF_TEST:-}" = 1 ]; then
  printf '%s\n' '{"status":"ok","db":"ok"}' | valid_health
  ! printf '%s\n' '{"status":"ok","db":"error"}' | valid_health
  ! printf '%s\n' '"status=ok db=ok"' | valid_health
  ! printf '%s\n' 'not-json' | valid_health 2>/dev/null
  exit 0
fi

image=${1:?usage: smoke-docker.sh IMAGE}
suffix="$$-$(date +%s)"
container="omniroute-smoke-$suffix"
negative_container="omniroute-smoke-negative-$suffix"
volume="omniroute-smoke-$suffix"
readonly_dir=$(mktemp -d "${TMPDIR:-/tmp}/omniroute-smoke.XXXXXX")
port=
sentinel="$(date +%s)$$"

cleanup() {
  docker rm -f "$container" "$negative_container" >/dev/null 2>&1 || true
  docker volume rm -f "$volume" >/dev/null 2>&1 || true
  chmod 700 "$readonly_dir" 2>/dev/null || true
  rm -rf "$readonly_dir" "$readonly_dir.out"
}
trap cleanup EXIT HUP INT TERM

health() {
  wget -qO- -T 1 --tries=1 "http://127.0.0.1:$port/health"
}

wait_healthy() {
  attempts=0
  while [ "$attempts" -lt 60 ]; do
    if [ "$(docker inspect -f '{{.State.Running}}' "$container" 2>/dev/null || printf false)" != true ]; then
      printf 'container exited before health check succeeded\n' >&2
      docker logs "$container" >&2 || true
      return 1
    fi
    body=$(health 2>/dev/null || true)
    if printf '%s' "$body" | valid_health 2>/dev/null; then
      return 0
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
  port=$(docker inspect -f '{{if eq (len (index .NetworkSettings.Ports "3456/tcp")) 1}}{{(index (index .NetworkSettings.Ports "3456/tcp") 0).HostPort}}{{end}}' "$container")
  case "$port" in
    ''|*[!0-9]*) printf 'failed to resolve single container port: %s\n' "$port" >&2; return 1 ;;
  esac
}

docker volume create "$volume" >/dev/null
start_container
wait_healthy
docker run --rm -e SENTINEL="$sentinel" -v "$volume:/data" --entrypoint /bin/sh "$image" -c \
  'sqlite3 -cmd ".timeout 5000" /data/storage.sqlite "CREATE TABLE IF NOT EXISTS _omniroute_persistence_smoke (value TEXT NOT NULL); DELETE FROM _omniroute_persistence_smoke; INSERT INTO _omniroute_persistence_smoke(value) VALUES ($SENTINEL);"'
docker rm -f "$container" >/dev/null
start_container
wait_healthy
persisted=$(docker run --rm -v "$volume:/data" --entrypoint sqlite3 "$image" /data/storage.sqlite \
  'SELECT value FROM _omniroute_persistence_smoke LIMIT 1;')
if [ "$persisted" != "$sentinel" ]; then
  printf 'SQLite persistence mismatch: expected %s, got %s\n' "$sentinel" "$persisted" >&2
  exit 1
fi

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
