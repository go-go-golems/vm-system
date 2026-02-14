---
Title: vm-system + vm-system-ui merge integration design and analysis
Ticket: VM-017-MERGE-UI-VM
Status: active
Topics:
    - architecture
    - frontend
    - backend
    - integration
    - monorepo
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: vm-system/vm-system-ui/client/src/lib/types.ts
      Note: Environment defaults and UI domain model assumptions.
    - Path: vm-system/vm-system-ui/client/src/lib/vm/endpoints/executions.ts
      Note: |-
        Execution/event retrieval behavior and N+1 request pattern.
        Execution/event retrieval behavior
    - Path: vm-system/vm-system-ui/client/src/lib/vm/endpoints/templates.ts
      Note: |-
        Template bootstrap behavior and API choreography.
        Template endpoint behavior
    - Path: vm-system/vm-system-ui/client/src/lib/vm/transport.ts
      Note: UI API base path and error handling
    - Path: vm-system/vm-system-ui/package.json
      Note: Build/start/check scripts and runtime packaging model.
    - Path: vm-system/vm-system-ui/server/index.ts
      Note: Current production static hosting approach (Node/Express).
    - Path: vm-system/vm-system-ui/vite.config.ts
      Note: |-
        Frontend dev proxy and build output settings.
        Vite proxy/output settings
    - Path: vm-system/vm-system/cmd/vm-system/cmd_serve.go
      Note: |-
        Serve command composition where SPA hosting can be wired.
        Serve command wiring for future SPA hosting
    - Path: vm-system/vm-system/pkg/vmdaemon/app.go
      Note: |-
        Daemon/server lifecycle and handler injection point.
        Daemon lifecycle and handler injection
    - Path: vm-system/vm-system/pkg/vmtransport/http/server.go
      Note: |-
        Backend API route contract consumed by the UI.
        Backend API route contract consumed by the UI
ExternalSources: []
Summary: Deep integration analysis for merging vm-system-ui into vm-system as a subdirectory, with a recommended git history-preserving import strategy and phased backend+frontend consolidation plan.
LastUpdated: 2026-02-14T16:35:00-05:00
WhatFor: Choose the least risky path to combine backend and frontend into one repository and one production runtime surface.
WhenToUse: Use when preparing the actual merge PR, sequencing migration tasks, or validating tradeoffs between subtree/submodule/filter-repo approaches.
---


# vm-system + vm-system-ui merge integration design and analysis

## Executive Summary

`vm-system` and `vm-system-ui` should be merged into one repo, with the UI imported under a stable subdirectory (`ui/`) and packaged by the Go daemon for production. The best path is:

1. History-preserving import into `vm-system` via `git subtree` (recommended operationally) or `git filter-repo` (recommended only for a one-time clean rewrite workflow).
2. Keep the existing two-process dev loop (`pnpm dev` + `vm-system serve`) using Vite proxy.
3. Add single-binary production serving from Go (`go:embed`) so `/` serves the SPA and `/api/v1/*` remains daemon-owned.

This aligns with existing contracts already in place:
- Backend route surface is centralized and stable in `vm-system/pkg/vmtransport/http/server.go:20`.
- UI already targets `/api/v1/*` and uses Vite proxy indirection in `vm-system-ui/vite.config.ts:157`.
- UI can already work with relative API paths when no absolute base URL is provided (`vm-system-ui/client/src/lib/types.ts:215`, `vm-system-ui/client/src/lib/vm/transport.ts:18`).

## Problem Statement

Today the backend and frontend are separate repositories with separate histories:

- `vm-system` (Go daemon + CLI + HTTP API) has 130 commits.
- `vm-system-ui` (React/Vite UI) has 14 commits.

They are conceptually one product, but split repos create friction:

1. Cross-repo change coordination is slower (API changes and UI adaptation cannot land atomically).
2. Release artifacts are fragmented (backend binary and UI static server are built independently).
3. Documentation and ticket references already cross repo boundaries frequently, which increases drift risk.
4. Operational topology is inconsistent: backend is one process, UI production currently assumes a second Node process (`vm-system-ui/package.json:8`, `vm-system-ui/server/index.ts:9`).

The merge must preserve development speed, preserve useful git history, and avoid breaking API semantics.

## Proposed Solution

### Target Repository Shape

Import `vm-system-ui` into `vm-system` as:

```text
vm-system/
  cmd/
  pkg/
  test/
  ui/                  # imported vm-system-ui history/content
    client/
    server/            # transitional, can be removed later
    package.json
    vite.config.ts
  internal/web/        # new Go embed bridge (phase 2)
    embed/public/
```

### Runtime Topology

1. Dev mode (unchanged behavior):
- UI dev server on `:3000` with proxy to Go daemon (already implemented in `vm-system-ui/vite.config.ts:188`).
- Go daemon on `:3210` serving API only during early phase.

2. Production mode (new target behavior):
- Go daemon serves:
  - `/api/v1/*` from existing handlers (`vm-system/pkg/vmtransport/http/server.go:20`)
  - `/` + `/assets/*` from embedded UI assets
- Node `server/index.ts` becomes optional/deprecated.

### API and Contract Continuity

No contract rewrite is required for the merge itself. UI endpoint usage already maps 1:1 to backend routes:
- templates: `vm-system-ui/client/src/lib/vm/endpoints/templates.ts:10`
- sessions: `vm-system-ui/client/src/lib/vm/endpoints/sessions.ts:13`
- executions: `vm-system-ui/client/src/lib/vm/endpoints/executions.ts:14`

### Recommended Git Import Strategy

Use `git subtree` import into `ui/` to preserve history with lowest operational overhead for this repository size.

Recommended command sequence:

```bash
# from vm-system repo root
git checkout -b vm-017-merge-ui
git subtree add --prefix ui git@github.com:wesen/vm-system-ui.git main
```

Notes:
- This keeps UI history, unlike a plain copy.
- If a smaller import is preferred, add `--squash` (not recommended for archaeology/debugging).
- If future temporary sync from old UI repo is needed:
  `git subtree pull --prefix ui git@github.com:wesen/vm-system-ui.git main`.

## Design Decisions

### Decision 1: Merge into one repo, not a submodule

Rationale:
- Product-level coupling is high and growing.
- Submodules add user friction (`clone --recurse-submodules`, detached HEAD workflows, split PR lifecycle).
- Atomic backend+frontend API migrations are easier in one repo.

### Decision 2: Keep `/api/v1` unchanged

Rationale:
- Backend routing is explicit and battle-tested in integration tests (`vm-system/pkg/vmtransport/http/server.go:20`).
- UI client already centralizes URL construction and handles relative vs absolute base cleanly (`vm-system-ui/client/src/lib/vm/transport.ts:18`).

### Decision 3: Adopt Go-served SPA for production

Rationale:
- Reduces production process count from 2 to 1.
- Aligns with existing daemon ownership model in `cmd_serve` and `vmdaemon` (`vm-system/cmd/vm-system/cmd_serve.go:30`, `vm-system/pkg/vmdaemon/app.go:26`).
- Fits the established Go+Vite embed pattern for fast dev and stable prod packaging.

### Decision 4: Keep UI `server/index.ts` only as a transitional fallback

Rationale:
- Current behavior is simple static hosting (`vm-system-ui/server/index.ts:19`) and can remain while embed integration stabilizes.
- Final architecture should not require Node for prod availability.

### Decision 5: Defer endpoint optimization until after physical merge

Rationale:
- Merge objective is repository/runtime consolidation, not behavior redesign.
- Known inefficiencies (template bootstrap side-effects and execution N+1 event fetches) should be tracked as post-merge tickets:
  - `vm-system-ui/client/src/lib/vm/endpoints/templates.ts:15`
  - `vm-system-ui/client/src/lib/vm/endpoints/executions.ts:22`

## Alternatives Considered

### Option A: Git submodule (`vm-system/ui` points to vm-system-ui repo)

Pros:
- Keeps repos independently releasable.
- No history rewrite.

Cons:
- Highest day-to-day friction.
- Still enforces split lifecycle for tightly-coupled changes.
- Easy to end up with mismatched backend/UI commits.

Verdict: Rejected.

### Option B: Direct copy into subdirectory (no history)

Pros:
- Fastest one-time action.
- No git tooling prerequisites.

Cons:
- Loses UI commit archaeology and blame utility.
- Harder rollback analysis for regressions.

Verdict: Rejected.

### Option C: `git filter-repo` + merge unrelated histories

Pros:
- Clean imported history rewritten under `ui/`.
- Excellent for long-term repository cleanliness.

Cons:
- More tooling steps and operator error surface.
- Best performed with temporary clones and careful validation.

Verdict: Valid alternative when maintainers specifically want rewritten import history.

### Option D: `git subtree` import into `ui/` (recommended)

Pros:
- Preserves history.
- Simple operational model.
- Supports optional follow-up sync from source repo during transition.

Cons:
- Adds subtree merge commits.

Verdict: Recommended.

## Implementation Plan

### Phase 0: Pre-merge safeguards

1. Create merge branch in `vm-system`.
2. Ensure `.gitignore` coverage for `ui/node_modules`, `ui/dist`, and logs (the imported UI ignore file already covers these patterns).
3. Tag current backend state for rollback (`pre-vm-017-ui-merge`).

### Phase 1: Import UI into `ui/`

1. Run subtree import command.
2. Verify repository shape and that no nested `.git` directory was imported.
3. Run sanity checks:
- `go test ./...`
- `pnpm -C ui check`
- `pnpm -C ui build`

Acceptance:
- One repo contains both systems with preserved histories.

### Phase 2: Build/runtime bridge (Go serves UI assets)

1. Add `internal/web` package:
- `embed.go` (`//go:build embed`)
- `embed_none.go` (`//go:build !embed`)
- `spa.go` (file serve + SPA fallback)
- `generate.go` and `generate_build.go` (frontend build+copy)
2. Register SPA handler after API route registration so `/api/v1` cannot be shadowed.
3. Keep current API mux unchanged.

Acceptance:
- `go build -tags embed ./cmd/vm-system` serves UI and API together.

### Phase 3: Developer workflow standardization

1. Add task entry points:
- `make dev-backend` -> `go run ./cmd/vm-system serve --listen 127.0.0.1:3210`
- `make dev-frontend` -> `pnpm -C ui dev`
- `make frontend-check` -> `pnpm -C ui check`
- `make build` -> `go generate ./internal/web && go build -tags embed ./cmd/vm-system`
2. Update contributor docs for two-process dev loop.

### Phase 4: CI and release hardening

1. Install pnpm dependencies before `go generate` in CI.
2. Run frontend typecheck and build in pipeline.
3. Add a regression test that `GET /` returns `index.html` when assets are available.

### Phase 5: Cleanup and deprecation

1. Mark `ui/server/index.ts` deprecated.
2. Remove it once Go-hosted SPA passes soak period.
3. Archive or lock old `vm-system-ui` repo after cutover.

## Risk Analysis and Mitigations

1. Risk: Merge introduces large non-source artifacts.
Mitigation: verify ignored files before import (`node_modules`, `dist`, logs); subtree from remote HEAD avoids local workspace artifacts.

2. Risk: Static SPA handler accidentally shadows API routes.
Mitigation: register API handlers first; add explicit guard for `/api` and fallback ordering tests.

3. Risk: Runtime behavior differences between Node static server and Go SPA fallback.
Mitigation: route smoke tests for `/`, `/assets/*`, client-side paths, and `/api/v1/*`.

4. Risk: UI performance regressions become more visible in integrated deployment.
Mitigation: post-merge tickets for endpoint fan-out optimization (`templates.ts` auto-bootstrap, `executions.ts` event N+1).

## Rollback Plan

1. Keep merge as isolated branch/PR.
2. If critical breakage appears, revert merge commit and keep repos split.
3. Preserve pre-merge tag to restore release baseline quickly.

## Open Questions

1. Do we want to continue syncing from `wesen/vm-system-ui` for a transition window, or freeze it immediately after import?
2. Should UI production serving be fully switched in the same PR as subtree import, or in a follow-up PR (lower risk)?
3. Do we want to rename `ui/` to `web/` for consistency with future non-React frontends?
4. Should template auto-bootstrap remain in UI runtime flow (`templates.ts`) or move to explicit onboarding action?

## References

- `vm-system/pkg/vmtransport/http/server.go:20`
- `vm-system/pkg/vmdaemon/app.go:26`
- `vm-system/cmd/vm-system/cmd_serve.go:30`
- `vm-system-ui/vite.config.ts:157`
- `vm-system-ui/client/src/lib/vm/transport.ts:18`
- `vm-system-ui/client/src/lib/types.ts:215`
- `vm-system-ui/client/src/lib/vm/endpoints/templates.ts:15`
- `vm-system-ui/client/src/lib/vm/endpoints/executions.ts:22`
- `vm-system-ui/package.json:6`
- `vm-system-ui/server/index.ts:9`
