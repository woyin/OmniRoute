# Go Rewrite Behavioral Parity Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Replace unsupported 1:1 claims with an automated behavioral parity gate, then close every supported route/provider/stream mismatch.

**Architecture:** Generate inventories from source/runtime, detect placeholders and fabricated success, compare provider contracts, and run differential E2E fixtures against isolated main and Go servers. Fixes proceed by endpoint family with failing tests first.

**Tech Stack:** Go/chi/httptest, Node.js test runner, Next.js route source extraction, SQLite temp databases, SSE, WebSocket.

---

### Task 1: Build route contract inventory and comparison

**Files:**
- Create: `scripts/parity/route-contract-lib.mjs`
- Create: `scripts/parity/inventory-main-routes.mjs`
- Create: `scripts/parity/compare-route-contracts.mjs`
- Create: `tests/parity/route-contract-lib.test.mjs`
- Refactor: `cmd/omniroute/router.go`
- Create: `cmd/omniroute/router_contract.go`
- Test: `cmd/omniroute/router_contract_test.go`

- [ ] Test normalization of Next `[id]`, catch-all, and chi `{id}` paths plus exported HTTP methods.
- [ ] Test diff of method/path, auth classification, and stream type.
- [ ] Extract Go router construction from `main.go` without behavior changes.
- [ ] Traverse `chi.Routes()` and emit stable JSON inventory; fail on duplicate or missing metadata.
- [ ] Generate main inventory from `src/app/api/**/route.ts` at main ref/worktree.
- [ ] Run comparison; persist no hand-maintained route snapshot.
- [ ] Commit infrastructure in small route-core and Go-router commits.

### Task 2: Detect placeholders and fabricated success

**Files:**
- Create: `scripts/parity/check-go-placeholders.mjs`
- Create: `tests/parity/placeholder-detection.test.mjs`

- [ ] Test detection of `placeholderHandler`, HTTP 501, “not implemented”, and fixed `success:true` with no request parsing/service/DB/state transition.
- [ ] Permit real operations that return success after side effects.
- [ ] Scan all `cmd/omniroute/*.go`, excluding tests; unresolved handler links fail.
- [ ] Run scanner and group findings by auth/settings, providers/combos/keys, usage, CLI/services, v1 inference, streams, miscellaneous.
- [ ] Commit detector before behavior fixes.

### Task 3: Gate provider contracts

**Files:**
- Create: `internal/provider/registry/parity.go`
- Create: `internal/provider/registry/parity_test.go`
- Create: `scripts/parity/compare-provider-parity.mjs`
- Create: `tests/parity/provider-parity.test.mjs`

- [ ] Emit stable canonical ID, aliases, format, auth type, executor, passthrough, deprecated, and system-only fields.
- [ ] Extract main provider data from actual registry exports, not a copied list.
- [ ] Fail on missing/extra providers, aliases, or behavioral field mismatches.
- [ ] Commit provider gate.

### Task 4: Close confirmed route/placeholder gaps by family

**Files:** affected `cmd/omniroute/routes_*.go`, `internal/handler/*.go`, `internal/db/*.go`, focused tests.

For each family:

- [ ] Select one confirmed mismatch from route/placeholder reports.
- [ ] Write an exact failing handler or integration test for method, auth, validation, status, response shape, DB side effect, SSE, or WebSocket contract.
- [ ] Run the target test and confirm failure matches the gap.
- [ ] Implement minimal real behavior using existing DB/service helpers; do not return fabricated success.
- [ ] Re-run target test, relevant parity report, then `go test ./...`, `go vet ./...`, `go build ./...`.
- [ ] Commit one family per commit.

Required order: authentication/settings; providers/combos/API keys; v1 inference; usage/analytics; CLI/services; miscellaneous management; streaming.

### Task 5: Add differential E2E

**Files:**
- Create: `scripts/parity/run-differential-e2e.mjs`
- Create: `tests/parity/fixtures.mjs`
- Create: `tests/parity/differential-e2e.test.mjs`

- [ ] Start main and Go servers with separate temp data dirs and dynamic ports; wait on health, no fixed sleeps.
- [ ] Cover health, auth, protected providers, settings persistence, API key CRUD, combo CRUD, models, chat JSON, Responses API, malformed input, wrong method, unknown route.
- [ ] Use local upstream stub; never call real providers.
- [ ] Compare status, headers, normalized dynamic fields, response shape, SQLite side effects, and restart persistence.
- [ ] Add SSE event order/termination/error tests and WebSocket upgrade/frame/close tests.
- [ ] Commit E2E harness and streaming fixes separately.

### Task 6: CI and final acceptance

**Files:**
- Modify: `.github/workflows/go-ci.yml`

- [ ] Add independent route, placeholder, provider, and differential E2E steps.
- [ ] Run `go test ./...`, `go vet ./...`, `go build ./...`.
- [ ] Require zero supported-scope route/auth/stream mismatches.
- [ ] Require zero placeholders/fabricated success in supported scope.
- [ ] Require provider contracts equal.
- [ ] Require differential E2E pass including persistence, SSE, and WebSocket.
- [ ] Commit: `ci(go): enforce behavioral parity gates`.
