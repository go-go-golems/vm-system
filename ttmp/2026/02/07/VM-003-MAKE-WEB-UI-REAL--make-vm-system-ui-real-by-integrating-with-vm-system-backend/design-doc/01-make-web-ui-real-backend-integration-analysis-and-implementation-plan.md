---
Title: 'Make Web UI Real: backend integration analysis and implementation plan'
Ticket: VM-003-MAKE-WEB-UI-REAL
Status: active
Topics:
    - backend
    - frontend
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: vm-system/vm-system-ui/client/src/components/SessionManager.tsx
      Note: Session UX actions to map to backend calls
    - Path: vm-system/vm-system-ui/client/src/components/VMConfig.tsx
      Note: VM settings/library update integration points
    - Path: vm-system/vm-system-ui/client/src/lib/libraryLoader.ts
      Note: Library metadata/loading behavior to realign
    - Path: vm-system/vm-system-ui/client/src/lib/vmService.ts
      Note: Current mock service behavior to replace
    - Path: vm-system/vm-system-ui/client/src/pages/Home.tsx
      Note: UI orchestration points for adapter migration
    - Path: vm-system/vm-system-ui/vite.config.ts
      Note: Environment/build constraints relevant to rollout
    - Path: vm-system/vm-system/cmd/vm-system/cmd_exec.go
      Note: Execution contracts informing API endpoint design
    - Path: vm-system/vm-system/cmd/vm-system/main.go
      Note: Root command integration point for adding serve mode
    - Path: vm-system/vm-system/pkg/vmexec/executor.go
      Note: Execution and event model driving UI integration semantics
    - Path: vm-system/vm-system/pkg/vmsession/session.go
      Note: Session lifecycle and runtime ownership constraints
    - Path: vm-system/vm-system/pkg/vmstore/vmstore.go
      Note: Persistent schema and API mapping baseline
    - Path: vm-system/vm-system/ttmp/2026/02/07/VM-001-ANALYZE-VM--analyze-vm-system-architecture-behavior-and-quality/design-doc/01-comprehensive-vm-system-analysis-report.md
      Note: Backend runtime constraints and defects used as integration prerequisites
    - Path: vm-system/vm-system/ttmp/2026/02/07/VM-002-ANALYZE-VM-SYSTEM-UI--analyze-vm-system-ui-architecture-behavior-and-quality/design-doc/01-comprehensive-vm-system-ui-analysis-report.md
      Note: UI mock-runtime constraints and adapter migration requirements
ExternalSources: []
Summary: Comprehensive migration plan to replace vm-system-ui mock runtime with real vm-system backend integration, with phased implementation, API contracts, risks, and rollout strategy.
LastUpdated: 2026-02-08T00:40:00-05:00
WhatFor: Guide end-to-end implementation of real backend integration for vm-system-ui.
WhenToUse: Use when implementing, reviewing, or sequencing the migration from browser-mock execution to backend-backed execution.
---


# Make Web UI Real: backend integration analysis and implementation plan

## Executive Summary

This document defines the full path to make `vm-system-ui` a real client for `vm-system` instead of a browser-local simulator. The goal is not simply to “wire HTTP calls”; the goal is to establish a coherent runtime architecture where UI, backend, and persistence share one truth model for sessions, executions, libraries, and events.

Today, `vm-system-ui` uses `client/src/lib/vmService.ts` as an in-memory runtime. It creates sessions in browser memory, executes snippets through `new Function`, and loads external libraries directly via dynamic script tags. This is useful for demo speed, but it is fundamentally disconnected from backend runtime guarantees. In parallel, `vm-system` has strong persistent modeling (`vm`, `session`, `execution`, `execution_event`) but currently lacks a long-lived runtime host process, which prevents cross-command session execution continuity.

The integration must therefore solve both sides together:

1. Backend runtime host and API contract.
2. Frontend adapter migration from mock to real backend.
3. Shared semantics and rollout safety.

The plan below is intentionally detailed and phased. It references the prior findings in:

- `VM-001-ANALYZE-VM` report (`.../design-doc/01-comprehensive-vm-system-analysis-report.md`)
- `VM-002-ANALYZE-VM-SYSTEM-UI` report (`.../design-doc/01-comprehensive-vm-system-ui-analysis-report.md`)

## Implementation Correction (2026-02-08)

The original planning section below used a transitional `/api/v1/vms` naming sketch. The implemented backend contract and UI migration use template-first naming only:

- `GET/POST /api/v1/templates`
- `GET/DELETE /api/v1/templates/{template_id}`
- `GET/POST/DELETE /api/v1/templates/{template_id}/modules`
- `GET/POST/DELETE /api/v1/templates/{template_id}/libraries`
- `GET/POST /api/v1/templates/{template_id}/capabilities`
- `GET/POST /api/v1/templates/{template_id}/startup-files`
- `GET/POST/DELETE /api/v1/sessions...`
- `GET/POST /api/v1/executions...`

VM-003 implementation followed this clean-cut contract with no compatibility wrappers.

## Problem Statement

### What “real UI” means

A real UI is one where user actions in the browser mutate and observe backend-owned state, not local simulation state.

- Session creation in UI creates backend session runtime.
- REPL/run-file in UI executes in backend runtime (goja), not `new Function`.
- Execution events shown in UI are backend events (`execution_event`), not locally fabricated approximations.
- Library/module/setting changes are enforced by backend policy, not only reflected in UI checkboxes.

### Why current architecture cannot satisfy this

From existing reports:

- VM-002: `vmService.ts` is a mock singleton with in-memory maps and browser execution.
- VM-001: backend CLI cannot provide stable session execution across processes because live session map is process-local.

Therefore, integrating the UI requires introducing a long-lived backend runtime host and a stable API contract. Without that, UI would still be forced into partial simulation.

## Existing Evidence Baseline (From Prior Tickets)

## Evidence from VM-001 (backend)

1. Durable control-plane exists in SQLite schema (`vm`, `vm_settings`, `vm_session`, `execution`, `execution_event`).
2. Live runtime continuity is missing across CLI invocations.
3. Library cache path contract mismatch exists (`<id>-<version>.js` download vs `<id>.js` load).
4. Command scripts have drift from current CLI contracts.

Implication for VM-003: build the integration on a backend server mode that owns runtime lifecycle; do not attempt to bind UI directly to one-shot CLI invocations.

## Evidence from VM-002 (frontend)

1. UI architecture is modular and usable.
2. Runtime semantics are mock-local (`new Function`, in-memory sessions/executions).
3. Service API has correctness gaps (`createSession(vmId)` ignores vmId, context resolution uses current session).
4. Build is healthy, but there are analytics placeholder warnings and bundle-size warnings.

Implication for VM-003: preserve UI shell, replace runtime layer via adapter boundary, and make mode explicit.

## Target End State

### Functional end state

- UI can list/create/update VMs from backend.
- UI can list/create/select/close sessions from backend.
- UI can execute REPL and run-file against backend sessions.
- UI can stream/poll execution events from backend.
- UI logs and console panels show backend truth.
- Mock mode remains optional for local UX development, but “real mode” is production default.

### Architectural end state

```text
+--------------------------+       +------------------------------------+
| vm-system-ui (React)     |       | vm-system runtime host (Go server) |
|                          | HTTP  |                                    |
| BackendAdapter ----------+------>+ VM/session/exec APIs               |
| Event stream client      | <-----+ Event stream (SSE/WS or poll)      |
+--------------------------+       |                                    |
                                   | SessionManager (long-lived)        |
                                   | VMStore (SQLite)                   |
                                   +------------------------------------+
```

## Design Principles

1. One source of execution truth: backend runtime.
2. Explicit mode boundary: mock and real paths must be distinguishable.
3. Contract first: frontend integration should follow typed API schema, not ad hoc payloads.
4. Observable behavior: every execution transition should be inspectable from UI and backend logs.
5. Incremental migration: keep UI operable while replacing internals phase by phase.

## Proposed Backend API Contract (v1)

All payloads should align with `vmmodels` naming semantics where possible.

### VM endpoints

- `GET /api/v1/vms`
- `POST /api/v1/vms`
- `GET /api/v1/vms/:vm_id`
- `PATCH /api/v1/vms/:vm_id`
- `DELETE /api/v1/vms/:vm_id`

### Session endpoints

- `GET /api/v1/sessions?status=ready`
- `POST /api/v1/sessions`
- `GET /api/v1/sessions/:session_id`
- `POST /api/v1/sessions/:session_id/close`

### Execution endpoints

- `POST /api/v1/executions/repl`
- `POST /api/v1/executions/run-file`
- `GET /api/v1/executions/:execution_id`
- `GET /api/v1/executions?session_id=<id>&limit=50`
- `GET /api/v1/executions/:execution_id/events?after_seq=0`

### Optional event stream endpoint

- `GET /api/v1/sessions/:session_id/events/stream` (SSE) OR
- `GET /api/v1/executions/:execution_id/events/stream` (SSE)

### Canonical response sketch

```json
{
  "id": "execution-uuid",
  "session_id": "session-uuid",
  "kind": "repl",
  "status": "ok",
  "started_at": "2026-02-08T00:00:00Z",
  "ended_at": "2026-02-08T00:00:00Z",
  "result": {"type":"number","preview":"3","json":3}
}
```

## Backend Implementation Plan (Detailed)

## Phase B0: Contract and server scaffolding

### Goals

- Add long-lived backend process mode.
- Establish versioned API base and shared DTO definitions.

### Steps

1. Add new command entrypoint in backend CLI:
- file: `vm-system/cmd/vm-system/cmd_serve.go`
- symbol: `newServeCommand()`
- root registration: add to `main.go`

2. Add server package:
- new dir: `vm-system/pkg/vmapi/`
- files:
- `server.go` (router, startup, graceful shutdown)
- `handlers_vm.go`
- `handlers_session.go`
- `handlers_execution.go`
- `dto.go`

3. Introduce dependency graph:

```text
vmapi.Server
  -> vmstore.VMStore
  -> vmsession.SessionManager (single, process-long)
  -> vmexec.Executor
```

4. Add health endpoint:
- `GET /api/v1/health`

5. Add CORS config for UI dev origin (configurable).

### Deliverables

- `vm-system serve --db ... --listen ...` starts API server.
- Basic VM/session/execution GET endpoints return JSON.

## Phase B1: Runtime continuity and session registry

### Goals

- Ensure session creation + execution works across UI requests.
- Retain one live runtime registry within server process.

### Steps

1. Keep single `SessionManager` instance in server.
2. Ensure execution handlers always resolve session via that singleton.
3. Add session close endpoint wiring to manager close path.
4. Add startup behavior for stale durable sessions from DB:
- policy option A: mark stale sessions `closed` at boot.
- policy option B: reconstruct limited runtime state (advanced; defer).

### Pseudocode

```pseudo
onServerStart:
  sessions = store.listSessions(status in [starting, ready])
  for s in sessions:
    mark s closed with reason "server restart recovery"
```

### Deliverables

- UI can create session then execute multiple times without “session not found”.

## Phase B2: Execution and event retrieval endpoints

### Goals

- Expose stable execution API and event listing endpoints.

### Steps

1. Implement `POST /executions/repl` -> `Executor.ExecuteREPL`.
2. Implement `POST /executions/run-file` -> `Executor.ExecuteRunFile`.
3. Implement `GET /executions/:id` and `GET /executions/:id/events`.
4. Add request/response validation and stable error shape.
5. Surface meaningful status codes:
- `404` session/execution not found
- `409` session busy/not ready
- `400` invalid payload

### Deliverables

- Backend adapter can replace mock execution path.

## Phase B3: Library and policy fixes required for production semantics

### Goals

- Fix known defects before UI relies on backend behavior.

### Steps

1. Fix library filename contract:
- unify downloader and runtime loader path resolution.
- recommended: manifest-based resolution from `id -> versioned filename`.

2. Add minimum limits enforcement:
- wall timeout
- max events
- max output size

3. Capability enforcement baseline:
- ensure disabled modules/globals are not exposed.

### Deliverables

- Backend behavior matches UI expectations for configured VM policies.

## Phase B4: Optional event streaming

### Goals

- Improve UX latency for console/event display.

### Steps

1. Start with polling endpoint support (simpler).
2. Add SSE stream if needed for near-real-time logs.
3. Implement reconnect semantics with `after_seq` resume.

### Deliverables

- UI console updates smoothly without manual refresh loops.

## Frontend Implementation Plan (Detailed)

## Phase F0: Adapter interface extraction

### Goals

- Decouple UI components from mock singleton specifics.

### Steps

1. Create adapter interface:
- new file: `vm-system-ui/client/src/lib/adapters/types.ts`

2. Move current behavior into `MockAdapter`:
- new file: `.../adapters/mockAdapter.ts`

3. Add `BackendAdapter` skeleton:
- new file: `.../adapters/backendAdapter.ts`
- use `axios` as transport.

4. Add adapter factory and runtime mode config:
- `VITE_VM_RUNTIME_MODE=mock|backend`
- file: `.../adapters/index.ts`

### Deliverables

- UI uses `vmClient` interface; no direct dependency on legacy `vmService` internals.

## Phase F1: Replace execution path with backend in real mode

### Goals

- Remove browser eval from real mode.

### Steps

1. Change `Home.handleExecute` to call `vmClient.executeREPL(sessionId, code)`.
2. Ensure execution list refresh uses backend endpoint.
3. Keep mock implementation for local demos.
4. Add UI runtime badge in header (`Mock Runtime`/`Backend Runtime`).

### Deliverables

- Real mode never calls `new Function`.

## Phase F2: Session and VM config integration

### Goals

- Ensure session and VM operations are backend-backed.

### Steps

1. `SessionManager` actions call `create/list/select/close/delete` on adapter.
2. `VMConfig` toggles call `updateVMConfig` API.
3. When VM libraries/modules change, refresh current session metadata and show “pending apply” if backend requires restart/reload.

### Deliverables

- UI controls reflect persisted backend state.

## Phase F3: Event viewer integration

### Goals

- Show backend events as primary source in console and logs.

### Steps

1. Poll `GET /executions/:id/events?after_seq=` after execution start.
2. Update `ExecutionConsole` and `ExecutionLogViewer` to use backend event payloads directly.
3. Normalize timestamps and payload rendering with stable parser.

### Deliverables

- Console/log tabs accurately represent backend execution events.

## Phase F4: Remove semantic drift and duplicate metadata

### Goals

- Ensure library/module metadata source is unified.

### Steps

1. Fetch built-in modules/libraries from backend metadata endpoints OR shared generated schema.
2. Remove duplicated constants in both `vmService` and `libraryLoader`.
3. Keep fallback static metadata only for mock mode.

### Deliverables

- No version/source drift between UI lists and backend runtime.

## Cross-Cutting Tasks

## Schema and contract versioning

- Add `api_version` in responses or use URL version (`/api/v1` already recommended).
- Introduce migration/version policy for backend DB where needed.

## Error contract normalization

Use one shape everywhere:

```json
{
  "error": {
    "code": "SESSION_NOT_FOUND",
    "message": "Session not found",
    "details": {}
  }
}
```

## Security posture

1. Real mode: no browser eval execution.
2. Add CORS allowlist.
3. Add request size limits on execution endpoints.
4. Add optional auth token for non-local usage.

## Observability

- Correlate request/session/execution IDs in logs.
- Expose basic metrics:
- active sessions
- execution rate
- error rate by code
- average execution latency

## Risk Register

| Risk | Probability | Impact | Mitigation |
|---|---|---|---|
| Backend runtime host instability | Medium | High | Phase B0/B1 integration tests + recovery policy |
| API contract drift during migration | High | High | Adapter interface + typed DTOs + contract tests |
| UI behavior regression during mode switch | Medium | Medium | Keep mock mode + feature flags + staged rollout |
| Event stream complexity | Medium | Medium | Start with polling, add SSE later |
| Policy mismatch (capabilities/limits) | High | High | Phase B3 before enabling production real mode |
| Library loading failures | High | Medium | manifest-based cache resolution + clear errors |

## Testing Strategy (Very Detailed)

## Backend tests

### Unit

- `vmstore` CRUD + error conditions.
- session status transitions.
- execution result/error event formatting.
- library manifest resolution.

### Integration

1. create VM -> create session -> execute repl -> list events.
2. add library -> create session -> library available.
3. startup file failure -> session crashed + last_error set.
4. concurrent execution on same session -> one rejected with busy.

### API contract tests

- Request validation.
- Status code mapping.
- Response schema shape.

## Frontend tests

### Unit/component

- Adapter mode selection.
- Home execution flow in backend mode.
- SessionManager action dispatch.
- VMConfig update flow.
- Execution console payload rendering.

### End-to-end

- Start backend server + UI.
- Create session in UI.
- Execute code.
- Observe events in UI and verify against backend endpoint.

## Compatibility tests (mock vs real)

Run selected scenario fixtures against both adapters and compare semantic outputs where expected.

## Rollout Strategy

## Stage 1: Developer preview (real mode hidden)

- Real adapter implemented behind env flag.
- Default remains mock.

## Stage 2: Internal alpha

- Real mode enabled for selected developers.
- Collect execution/session error metrics.

## Stage 3: Default real mode for local environments

- Mock remains available explicitly for UI-only development.

## Stage 4: Production readiness gate

Require:

1. Backend runtime continuity proven.
2. Library contract fixed.
3. Limits/capabilities baseline enforcement active.
4. E2E tests stable in CI.

## Definition of Done

This ticket is done when all of the following are true:

1. `vm-system-ui` in backend mode performs session create/select/execute/log without mock runtime code paths.
2. `vm-system` server mode provides stable runtime session continuity.
3. UI logs are sourced from backend execution events.
4. Library and VM configuration changes propagate through backend APIs.
5. CI includes backend integration tests and frontend e2e real-mode checks.
6. Mock mode is clearly marked and isolated.

## Detailed Work Breakdown Structure (WBS)

### WBS-1 Backend server

- WBS-1.1 add serve command
- WBS-1.2 add vmapi package
- WBS-1.3 wire session/execution handlers
- WBS-1.4 error schema and middleware
- WBS-1.5 health and metrics

### WBS-2 Runtime and policy correctness

- WBS-2.1 session continuity behavior
- WBS-2.2 library path manifest fix
- WBS-2.3 limits enforcement (wall/events/output)
- WBS-2.4 capability gating baseline

### WBS-3 Frontend adapter migration

- WBS-3.1 adapter interfaces
- WBS-3.2 backend adapter implementation
- WBS-3.3 UI refactor to adapter usage
- WBS-3.4 runtime mode indicator and warnings

### WBS-4 Verification and rollout

- WBS-4.1 backend contract tests
- WBS-4.2 frontend real-mode e2e
- WBS-4.3 rollout flags and docs
- WBS-4.4 post-rollout monitoring dashboards

## Sequence Diagram (End-to-End Real Mode)

```text
User clicks Run
  -> Home.tsx
     -> BackendAdapter.executeREPL(sessionId, code)
        -> POST /api/v1/executions/repl
           -> vmapi handler
              -> vmexec.Executor.ExecuteREPL
                 -> goja runtime executes
                 -> execution + events stored
              <- execution response
     <- execution summary
  -> Home refreshes/logs execution
  -> GET /api/v1/executions/:id/events?after_seq=0
     <- ordered events
  -> ExecutionConsole renders backend events
```

## Dependencies and Ordering Constraints

1. Backend server mode must land before real frontend mode can be default.
2. Library filename contract fix should land before enabling library-heavy UI demos in real mode.
3. Adapter extraction should land before broad UI feature additions to avoid compounding migration cost.
4. Event endpoint stabilization should precede console UX tuning.

## Alternatives Considered

## A. Replace frontend runtime all at once

Rejected. High regression risk and poor debuggability.

## B. Keep frontend mock forever and only “preview” backend responses

Rejected. Does not solve core trust issue; still dual semantics with no convergence.

## C. Adapter-based phased migration (selected)

Accepted. Minimal disruption with explicit convergence path.

## Open Questions

1. Should server expose SSE first or polling first? (Recommendation: polling first)
2. Should backend own library metadata endpoint, or should metadata live in shared generated package?
3. Is multi-user/auth required in first real mode milestone, or local trusted usage only?
4. Should stale sessions be auto-closed on server restart or reconstructed?
5. Is run-file expected to execute workspace files only, or also ad hoc uploads from UI?

## References

### Existing ticket docs

- `vm-system/ttmp/2026/02/07/VM-001-ANALYZE-VM--analyze-vm-system-architecture-behavior-and-quality/design-doc/01-comprehensive-vm-system-analysis-report.md`
- `vm-system/ttmp/2026/02/07/VM-002-ANALYZE-VM-SYSTEM-UI--analyze-vm-system-ui-architecture-behavior-and-quality/design-doc/01-comprehensive-vm-system-ui-analysis-report.md`
- `vm-system/ttmp/2026/02/07/VM-001-ANALYZE-VM--analyze-vm-system-architecture-behavior-and-quality/reference/01-diary.md`
- `vm-system/ttmp/2026/02/07/VM-002-ANALYZE-VM-SYSTEM-UI--analyze-vm-system-ui-architecture-behavior-and-quality/reference/01-diary.md`

### Backend code anchors

- `vm-system/cmd/vm-system/main.go`
- `vm-system/cmd/vm-system/cmd_session.go`
- `vm-system/cmd/vm-system/cmd_exec.go`
- `vm-system/pkg/vmsession/session.go`
- `vm-system/pkg/vmexec/executor.go`
- `vm-system/pkg/vmstore/vmstore.go`
- `vm-system/pkg/libloader/loader.go`

### Frontend code anchors

- `vm-system-ui/client/src/pages/Home.tsx`
- `vm-system-ui/client/src/lib/vmService.ts`
- `vm-system-ui/client/src/lib/libraryLoader.ts`
- `vm-system-ui/client/src/components/SessionManager.tsx`
- `vm-system-ui/client/src/components/VMConfig.tsx`
- `vm-system-ui/server/index.ts`
- `vm-system-ui/vite.config.ts`
