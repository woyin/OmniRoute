# Release Pipeline Repair Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Publish five usable cross-platform binaries with working SQLite and checksums over downloadable archives.

**Architecture:** Native runners build and smoke-test each supported platform. Windows ARM64 is removed. A final least-privilege job archives verified binaries, hashes final archives, validates checksums, and publishes the release.

**Tech Stack:** Go 1.24, mattn/go-sqlite3, GitHub Actions, MinGW-w64, PowerShell, SHA-256.

---

### Task 1: Add a deterministic SQLite smoke command

**Files:**
- Create: `cmd/omniroute/smoke.go`
- Modify: `cmd/omniroute/main.go`
- Test: `cmd/omniroute/smoke_test.go`

- [ ] Write a failing test invoking a helper that creates a temp SQLite DB, runs migrations through normal startup, writes `key_value(namespace='smoke', key='release')`, reads it back, and closes the DB.
- [ ] Run `go test ./cmd/omniroute -run TestSQLiteSmoke -count=1`; expect failure because helper is absent.
- [ ] Implement `--smoke-sqlite` flag and helper using existing config/DB APIs; print `sqlite smoke: ok` only after write/read succeeds.
- [ ] Run target test, then `go test ./...`, `go vet ./...`, `go build ./...`; expect exit 0.
- [ ] Commit: `feat(cli): add SQLite release smoke check`.

### Task 2: Repair native platform builds

**Files:**
- Modify: `.github/workflows/release.yml`
- Delete: `.goreleaser.yml` if unused after workflow conversion

- [ ] Restrict trigger to `v*-go` and set top-level `permissions: contents: read`.
- [ ] Linux jobs: retain amd64/arm64 CGO builds, execute native amd64 `--smoke-sqlite`; execute arm64 via native ARM runner or QEMU before upload.
- [ ] macOS jobs: build amd64/arm64 with CGO; execute each binary on a matching runner architecture before upload.
- [ ] Windows amd64: install MinGW-w64, set `CGO_ENABLED=1`, `CC=gcc`, build only `omniroute_windows_amd64.exe`, execute `--smoke-sqlite` in PowerShell.
- [ ] Remove Windows ARM64 build and asset references.
- [ ] Give only release job `contents: write`.
- [ ] Validate workflow syntax with Ruby/Python YAML parser and actionlint if available.
- [ ] Commit: `fix(ci): publish SQLite-capable release binaries`.

### Task 3: Hash final archives and validate release assets

**Files:**
- Modify: `.github/workflows/release.yml`
- Create: `scripts/ci/verify-release-assets.sh`

- [ ] Add a shell self-test fixture proving checksum generation uses `release/*.{tar.gz,zip}`, not raw artifacts.
- [ ] Create five archives first, then run `(cd release && sha256sum omniroute_* > checksums.txt)`.
- [ ] Run `(cd release && sha256sum -c checksums.txt)` before upload.
- [ ] Verify exact asset set: four Unix archives, one Windows archive, `checksums.txt`; reject extras/missing files.
- [ ] Commit: `fix(ci): checksum downloadable release archives`.

### Task 4: Publish and verify a new release

- [ ] Push branch after fresh `go test ./...`, `go vet ./...`, `go build ./...`.
- [ ] Create a new immutable tag rather than moving an existing published tag.
- [ ] Watch Release Action to completion.
- [ ] Download all assets, run `sha256sum -c checksums.txt`, unpack and smoke-test locally compatible binary.
- [ ] Confirm Windows ARM64 asset is absent and Windows amd64 smoke passed in CI.
