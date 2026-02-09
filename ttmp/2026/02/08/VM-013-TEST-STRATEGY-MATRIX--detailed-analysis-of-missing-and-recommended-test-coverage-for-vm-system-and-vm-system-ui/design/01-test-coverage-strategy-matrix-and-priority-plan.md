---
Title: Test Coverage Strategy Matrix and Priority Plan
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
    - /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/vmcontrol
    - /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/vmsession
    - /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/vmstore
    - /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/vmtransport/http
    - /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system-ui/client/src/lib
    - /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system-ui/client/src/pages
ExternalSources: []
Summary: >
    Detailed analysis of tests that should exist across vm-system and vm-system-ui,
    prioritized by risk, confidence impact, and implementation effort.
LastUpdated: 2026-02-09T10:15:00-05:00
WhatFor: Plan and sequence missing test coverage without implementing tests yet.
WhenToUse: Use when planning future coverage tickets or deciding what to test next.
---

# Test Coverage Strategy Matrix and Priority Plan

## Context

This ticket captures the test suite that could and should be written across backend and frontend. The user explicitly asked for analysis only (no implementation yet).

Current baseline:
- Backend has strong HTTP integration coverage and some core unit coverage.
- Backend has gaps in runtime/session/store/client/library loader packages.
- Frontend currently has no automated tests.

## Current Coverage Snapshot

### Existing backend tests

- `pkg/vmtransport/http/*_integration_test.go`
- `pkg/vmexec/executor_test.go`
- `pkg/vmexec/executor_persistence_failures_test.go`
- `pkg/vmcontrol/template_service_test.go`
- `pkg/vmmodels/ids_test.go`
- `pkg/vmmodels/json_helpers_test.go`
- `pkg/vmpath/path_test.go`

### Backend packages with no tests

- `pkg/vmsession`
- `pkg/vmstore`
- `pkg/vmdaemon`
- `pkg/vmclient`
- `pkg/libloader`

### Frontend

- No `*.test.*` / `*.spec.*` coverage in `vm-system-ui/client`, `server`, `shared`.

## Test Matrix (What Should Be Added)

## Backend Priority P0 (high risk)

1. Session lifecycle invariants (`pkg/vmsession`)
- create/ready/crashed/closed transitions
- startup file order behavior
- unsupported startup mode rejection behavior
- runtime map consistency under repeated create/close

2. Store contract tests (`pkg/vmstore`)
- CRUD and listing for templates/sessions/executions/events
- foreign key and delete behavior
- ordering semantics (startup file order, execution event seq)
- idempotency behavior for module/library add/remove patterns

3. Execution limit behavior (`pkg/vmcontrol/execution_service.go`)
- output limit triggered by event payload size
- max events limit triggered correctly
- error paths around limit loading and event retrieval

## Backend Priority P1

1. Client contract tests (`pkg/vmclient`)
- successful response decoding
- typed error mapping from structured error envelopes
- malformed response and transport error handling

2. Library loader tests (`pkg/libloader`)
- cache hit/miss behavior
- partial download failures and aggregate error behavior
- deterministic filename mapping from library ID/version

3. Daemon bootstrap tests (`pkg/vmdaemon`)
- config parsing defaults/overrides
- startup wiring smoke checks

## Frontend Priority P0

1. Data normalization boundary tests (`client/src/lib/normalize.ts`, `client/src/lib/api.ts`, `client/src/lib/vmService.ts`)
- Date/string conversion invariants
- missing optional fields and null payload normalization
- API error envelope normalization to UI error model

2. Route smoke tests (`client/src/App.tsx`, key pages)
- `/templates`, `/templates/:id`, `/sessions`, `/sessions/:id`, `/system`, `/reference`
- not-found behavior

3. Session and template critical flows
- create template
- create session
- execute code flow happy path/error path

## Frontend Priority P1

1. Component behavior tests
- `CreateSessionDialog` validation
- `ExecutionConsole` payload rendering branches
- `ExecutionLogViewer` filtering + expansion

2. Store/reducer tests
- `uiSlice` and derived selectors

3. System page state tests
- status-card summaries and empty states

## Cross-cutting test tooling recommendations

1. Backend
- continue `go test ./...` with race tests in CI (`go test -race ./...`)
- add per-package fixture builders under `pkg/.../testutil`

2. Frontend
- use `vitest` + `@testing-library/react` + `msw`
- isolate HTTP layer with mock server for deterministic API tests

3. Contract layer
- add API contract fixtures for error envelopes and startup mode validation

## Proposed Ticket Split for Implementation

1. `VM-013A` backend session/store coverage
2. `VM-013B` backend client/loader/daemon coverage
3. `VM-013C` frontend normalization and route smoke coverage
4. `VM-013D` frontend component/store behavior coverage

## Suggested Sequencing

1. Start with backend runtime/state invariants (highest risk).
2. Add frontend normalization tests before component tests.
3. Add CI enforcement only after initial flakiness pass.

## Acceptance Criteria for This Analysis Ticket

- Complete map of missing test areas exists.
- Each recommended test area has risk rationale.
- Follow-up implementation tickets are ready to be created/executed.
