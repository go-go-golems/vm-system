---
Title: Implementation Guide
Ticket: VM-013-TEST-STRATEGY-MATRIX
Status: active
Topics:
    - backend
    - frontend
    - architecture
DocType: design
Intent: long-term
Owners: []
RelatedFiles:
    - /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/vmsession
    - /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/vmstore
    - /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/vmclient
    - /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/libloader
    - /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system-ui/client/src/lib
    - /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system-ui/client/src/pages
ExternalSources: []
Summary: >
    Step-by-step implementation guide for executing the VM-013 test strategy in
    prioritized slices with commands, acceptance checks, and sequencing.
LastUpdated: 2026-02-09T10:40:00-05:00
WhatFor: Implement missing test coverage safely and incrementally.
WhenToUse: Use this guide when executing VM-013 work.
---

# VM-013 Implementation Guide

## Execution Model

Use four delivery slices with commit boundaries:
1. VM-013A backend runtime/session/store tests
2. VM-013B backend client/loader/daemon tests
3. VM-013C frontend normalization + route smoke tests
4. VM-013D frontend component/store behavior tests

## Phase 0: Harness Setup

1. Backend
- Add shared test helpers under `pkg/testutil/` for store/runtime/session setup.
- Standardize temporary DB/workspace setup (`t.TempDir()` based).

2. Frontend
- Install/use `vitest`, `@testing-library/react`, `@testing-library/jest-dom`, `msw`.
- Add `client/src/test/setup.ts` and configure Vitest globals.

Validation:
- `cd vm-system && go test ./...`
- `cd vm-system-ui && pnpm check`
- `cd vm-system-ui && pnpm test --run` (or defined equivalent)

## Phase 1: Backend High-risk Coverage (VM-013A)

Targets:
- `pkg/vmsession`
- `pkg/vmstore`
- `pkg/vmcontrol/execution_service.go`

Test cases:
- Session status transitions: starting->ready, startup failure->crashed, close->closed.
- Startup file ordering and path safety behavior.
- Store-level CRUD and ordering assertions for templates/sessions/executions/events.
- Output/event limit enforcement edge cases.

Acceptance:
- New tests deterministic and isolated from local artifacts.
- No flaky time-dependent assertions.

## Phase 2: Backend Support Coverage (VM-013B)

Targets:
- `pkg/vmclient`
- `pkg/libloader`
- `pkg/vmdaemon`

Test cases:
- REST client error envelope mapping and malformed response handling.
- Loader cache hit/miss, partial failure behavior, and filename determinism.
- Daemon config defaults/overrides and startup sanity checks.

Acceptance:
- Explicit coverage of failure-path branches, not only happy paths.

## Phase 3: Frontend Data + Route Coverage (VM-013C)

Targets:
- `client/src/lib/normalize.ts`
- `client/src/lib/api.ts`
- `client/src/lib/vmService.ts`
- `client/src/App.tsx`

Test cases:
- Date/string normalization invariants.
- API envelope->UI model conversion for success and error payloads.
- Route smoke checks for all active paths.

Acceptance:
- Tests pin current API contracts.
- Route tests verify not-found fallback.

## Phase 4: Frontend UI Behavior Coverage (VM-013D)

Targets:
- `CreateSessionDialog`, `ExecutionConsole`, `ExecutionLogViewer`, `uiSlice`.

Test cases:
- Form validation and disabled states.
- Event payload rendering branches.
- Execution log filtering/expansion behavior.
- Slice reducer transitions.

Acceptance:
- Tests avoid snapshot-only strategy; assert user-observable behavior.

## CI/Quality Gates

Add/confirm gates:
- backend: `go test ./...` (optionally `-race` on PR/merge)
- frontend: `pnpm check`, `pnpm test --run`, `pnpm build`

Merge criteria:
- no failing/ignored tests
- no new flaky tests
- deterministic local and CI runs

## Rollout and Risk Controls

- Land phases independently to keep PR size bounded.
- Keep test fixtures minimal and explicit.
- Avoid introducing global mutable state in test setup.
