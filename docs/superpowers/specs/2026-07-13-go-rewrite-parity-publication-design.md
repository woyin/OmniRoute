# Go Rewrite Parity and Publication Design

## Objective

Ensure `go-rewrite` provides verified behavioral parity with `main` for supported functionality and publishes usable GitHub Release binaries and multi-architecture Docker images.

## Scope

### Supported binary platforms

- Linux x86_64
- Linux aarch64
- macOS x86_64
- macOS aarch64
- Windows x86_64

Windows ARM64 is excluded until a maintained CGO-compatible toolchain is available and verified with SQLite.

### Docker platforms

- linux/amd64
- linux/arm64

### Parity definition

Parity requires matching HTTP method, path, status behavior, response contract, persistence side effects, authentication boundary, and streaming semantics. A route that only returns an empty success object is not equivalent when `main` performs real work.

## Release Architecture

Each platform builds on a compatible native runner where practical:

- Linux builds on Linux runners with CGO enabled.
- macOS builds on macOS runners with CGO enabled.
- Windows x86_64 builds on a Windows runner with MinGW and CGO enabled.

Every build job must run a SQLite smoke test that creates a temporary database, runs migrations, writes a record, reads it back, and exits successfully. Native-platform binaries should also execute `--help` or a bounded server-start smoke test.

Build jobs upload raw binaries with read-only repository permissions. A final release job downloads all binaries, creates the five final archives, calculates SHA-256 hashes over those archives, and uploads the archives plus `checksums.txt` to GitHub Release. `sha256sum -c checksums.txt` must work directly against downloaded Release assets.

Only tags matching `v*-go` trigger this release line. Tags containing `-alpha`, `-beta`, or `-rc` are prereleases; `-go` alone is a stable Go-rewrite suffix.

## Docker Publication Architecture

Docker publication must treat Docker Hub and GHCR independently. Existing state in one registry must not suppress publication or repair in the other. A retry must be able to recreate a missing or incomplete manifest in either registry.

The workflow publishes linux/amd64 and linux/arm64 manifests and verifies the resulting manifest contains both architectures. Before publication completion, each platform image must pass:

1. Container startup.
2. `/health` returns HTTP 200.
3. SQLite initializes in the configured data directory.
4. Data directory is writable.
5. A restart with the same volume preserves the database.

For bind mounts, startup must fail clearly when the data directory is not writable; warning-and-continue behavior is not accepted. Healthcheck timeout must exceed the script's maximum retry budget or the script must use a single bounded address strategy.

## Parity Verification

The parity audit compares `main` and `go-rewrite` across:

- Provider IDs and aliases.
- Route method/path pairs.
- Authentication requirements.
- Request validation.
- Status codes and response shapes.
- SQLite side effects.
- SSE and WebSocket framing.
- Management operations currently represented by placeholders.

Confirmed mismatches become implementation tasks. Placeholder and fabricated-success handlers cannot be counted as parity. Verification must exercise representative endpoint families end to end against temporary databases.

## Failure Handling

- Any platform build or SQLite smoke failure blocks Release creation.
- Missing Docker architecture or failed container smoke blocks manifest completion.
- Release publication is idempotent and replaces assets for the same tag only when explicitly rerun.
- Registry publication retries operate per registry.
- CI logs identify platform, architecture, failed command, and artifact name.

## Security and Permissions

- Build jobs: `contents: read` only.
- Release job: `contents: write` only.
- Docker publish jobs: `packages: write` only where GHCR upload occurs.
- No Redis, registry, or signing credentials are embedded in artifacts.
- Checksums cover downloadable archives.

## Acceptance Criteria

1. `go test ./...`, `go vet ./...`, and `go build ./...` pass.
2. Parity audit has no unresolved high-confidence contract mismatch in the declared supported scope.
3. GitHub Release contains exactly five supported platform archives plus `checksums.txt`.
4. Every binary passes platform-appropriate SQLite smoke validation.
5. Release checksums validate downloaded archives directly.
6. GHCR and Docker Hub manifests contain linux/amd64 and linux/arm64.
7. Published Docker image passes health, writable SQLite, and persistence smoke tests.
8. Release and Docker workflows can repair one missing registry or asset without rebuilding unrelated targets unnecessarily.
