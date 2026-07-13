# Docker Publication Repair Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Reliably publish and verify linux/amd64 and linux/arm64 images with writable, persistent SQLite storage.

**Architecture:** One tag trigger builds both platform images once, publishes GHCR manifests, and verifies runtime behavior. Registry existence checks are per registry; a missing GHCR manifest can be repaired independently. Container startup fails early on unwritable data storage.

**Tech Stack:** Docker Buildx, GitHub Actions, GHCR, Alpine, SQLite, shell smoke tests.

---

### Task 1: Add container runtime smoke script

**Files:**
- Create: `scripts/ci/smoke-docker.sh`
- Modify: `Dockerfile`

- [ ] Write script test that starts image with named volume, waits for `/health`, confirms JSON `status=ok` and `db=ok`, stops/restarts with same volume, confirms `storage.sqlite` persists.
- [ ] Add negative test mounting a non-writable directory; expect container non-zero exit with explicit data-dir error.
- [ ] Implement Docker entrypoint preflight using `test -w "$DATA_DIR"` and a create/remove probe; fail before server start.
- [ ] Set healthcheck timeout consistent with one bounded health attempt (or increase beyond full retry budget).
- [ ] Run local amd64 image smoke; expect all assertions pass.
- [ ] Commit: `fix(docker): verify writable persistent SQLite storage`.

### Task 2: Make registry publication independently repairable

**Files:**
- Modify: `.github/workflows/docker.yml`
- Create: `scripts/ci/check-image-platforms.sh`

- [ ] Add shell tests for platform checker with expected `linux/amd64` and `linux/arm64` manifests.
- [ ] Remove global skip based on one registry.
- [ ] Resolve `publish_ghcr` independently from any other registry state; current Go workflow publishes GHCR only, so GHCR is authoritative.
- [ ] Ensure workflow_dispatch can rebuild/replace a missing GHCR tag.
- [ ] Add `concurrency` keyed by ref to prevent duplicate publication.
- [ ] After push, inspect manifest and fail unless both platforms exist.
- [ ] Commit: `fix(ci): make Docker publication repairable`.

### Task 3: Add post-build runtime gates

**Files:**
- Modify: `.github/workflows/docker.yml`
- Modify: `scripts/ci/smoke-docker.sh`

- [ ] Load or pull built amd64 image and run runtime smoke before marking job successful.
- [ ] Validate arm64 image on native/QEMU runner with the same smoke.
- [ ] Verify `/health`, SQLite initialization, writable volume, and restart persistence.
- [ ] Keep `packages: write` only on push job; test jobs use read-only permissions.
- [ ] Commit: `ci: gate Docker publication on runtime smoke`.

### Task 4: Publish and verify image

- [ ] Push branch and watch Docker Action to completion.
- [ ] Run `docker buildx imagetools inspect ghcr.io/woyin/omniroute:<tag>`; confirm amd64 and arm64.
- [ ] Pull and run tagged image locally with named volume; confirm `/health` and restart persistence.
- [ ] Confirm latest promotion only occurs for stable release policy.
