---
Title: 'Daemonized vm-system architecture: backend runtime host, REST API, and CLI'
Ticket: VM-001-ANALYZE-VM
Status: active
Topics:
    - backend
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: vm-system/vm-system/cmd/vm-system/cmd_exec.go
      Note: Execution command semantics to remap to REST client behavior
    - Path: vm-system/vm-system/cmd/vm-system/cmd_session.go
      Note: Session command semantics and flags to align with post-cutover CLI/API surface
    - Path: vm-system/vm-system/cmd/vm-system/main.go
      Note: Root command tree for adding serve/daemon command
    - Path: vm-system/vm-system/pkg/libloader/loader.go
      Note: Library cache contract to normalize in daemon plan
    - Path: vm-system/vm-system/pkg/vmexec/executor.go
      Note: Execution/event pipeline to host inside daemon
    - Path: vm-system/vm-system/pkg/vmsession/session.go
      Note: Current in-memory runtime ownership model and lifecycle behavior
    - Path: vm-system/vm-system/pkg/vmstore/vmstore.go
      Note: Persistent schema baseline and versioning extension points
    - Path: vm-system/vm-system/vm-system/ttmp/2026/02/07/VM-001-ANALYZE-VM--analyze-vm-system-architecture-behavior-and-quality/design-doc/01-comprehensive-vm-system-analysis-report.md
      Note: Source findings motivating daemon redesign
    - Path: vm-system/vm-system/vm-system/ttmp/2026/02/07/VM-003-MAKE-WEB-UI-REAL--make-vm-system-ui-real-by-integrating-with-vm-system-backend/design-doc/01-make-web-ui-real-backend-integration-analysis-and-implementation-plan.md
      Note: Cross-ticket API/adapter alignment reference
ExternalSources: []
Summary: Detailed design for evolving vm-system from one-shot CLI runtime to daemon-managed VM sessions and templates, with REST API, CLI client design, and internal architecture.
LastUpdated: 2026-02-08T18:30:00-05:00
WhatFor: Plan and guide implementation of a long-lived vm-system daemon that manages runtime sessions and VM templates/capabilities.
WhenToUse: Use when implementing backend server mode, API endpoints, CLI client behavior, and hard cutover from process-local runtime.
---


# Daemonized vm-system architecture: backend runtime host, REST API, and CLI

## Executive Summary

This document proposes a concrete redesign of `vm-system` from a process-local CLI prototype into a layered runtime platform. The core change is to separate reusable session/template/execution orchestration from daemon hosting concerns.

The design introduces a transport-agnostic control core (`pkg/vmcontrol`) that owns runtime lifecycle and policy enforcement, while daemon logic (`pkg/vmdaemon`) becomes a thin process/lifecycle host. HTTP handlers and CLI client behavior become adapters on top of the same core services.

This separation solves two problems at once:

1. It fixes session continuity by keeping live runtimes in one long-lived process in daemon mode.
2. It makes session management reusable by other packages that can embed `vmcontrol` directly without running daemon HTTP transport.

This design includes:

1. Reusable core architecture (`vmcontrol`) with clear ports/adapters boundaries.
2. Daemon process model as host-only wrapper around the core.
3. REST API contract for templates, sessions, executions, and events.
4. CLI command model that works as API client with in-process core reuse for other packages.
5. Cutover execution plan from current process-local runtime to daemon/core architecture.

## Problem Statement

Current `vm-system` behavior has four structural gaps that prevent dependable backend usage by the UI, CLI, and future package integrations.

### Gap 1: Runtime ownership mismatch

- Durable records (`vm_session`) exist in SQLite.
- Live runtimes exist in `SessionManager.sessions` in memory.
- Each CLI process creates a fresh manager.
- Result: session continuity breaks across commands.

### Gap 2: Policy declaration vs policy enforcement

`VMSettings`, capabilities, and startup files are modeled, but runtime enforcement is partial/inconsistent. This creates ambiguity: users see controls that imply guarantees, but runtime behavior may not enforce all declared constraints.

### Gap 3: Control-plane and runtime-plane coupled to CLI process lifecycle

Control operations (create/list/update/delete) should be durable and process-independent. Runtime operations (execute, stream events, session lock ownership) should be process-stable and centralized. Today both are interleaved inside one-shot CLI processes.

### Gap 4: Runtime orchestration coupled to daemon/transport implementation

Without explicit core/domain separation, any future package that wants session orchestration is forced to depend on daemon internals or duplicate logic. This blocks reuse and increases drift risk.

## Design Goals

1. **Session continuity**: executions remain valid across separate client calls.
2. **Single runtime authority**: exactly one daemon instance owns active runtimes.
3. **Template-driven sessions**: session creation always derives from explicit template policy.
4. **Stable external contract**: REST API usable by web UI and CLI.
5. **Operational clarity**: health, metrics, logs, and error contracts are standardized.
6. **Hard cutover**: replace process-local CLI runtime with daemon/core architecture with no legacy dual mode.
7. **Reusable runtime core**: session/template/execution orchestration must be consumable by other packages without HTTP/daemon coupling.

## Non-Goals

1. Full multi-tenant authz/rbac in first phase.
2. Distributed runtime clustering in first phase.
3. Cross-node session transfer in first phase.
4. Rewriting storage layer away from SQLite in first phase.

## Terminology and Model Alignment

Current code uses `VM` as profile/config template. This design standardizes naming with a hard cutover to template terminology.

- **Template** (new public name): reusable policy/config object.
- **VM profile** (historical internal term): renamed to template as part of cutover.
- **Session**: live runtime instance created from a template.
- **Execution**: one REPL snippet or run-file invocation within a session.

Cutover naming policy:

- Public API exposes `/templates` only.
- CLI exposes `template` commands only.
- `vm` aliases are removed in the cutover implementation branch.

## Proposed Solution

## High-level architecture

```text
+-------------------+      +-------------------------------------+
| CLI/UI/Other Pkgs | ---> | Transport Adapters                  |
|                   |      | - HTTP (daemon)                     |
|                   |      | - In-process adapter (optional)     |
+-------------------+      +------------------+------------------+
                                            |
                                            v
                             +-------------------------------------+
                             | pkg/vmcontrol (reusable core)       |
                             | - TemplateService                   |
                             | - SessionService                    |
                             | - ExecutionService                  |
                             | - RuntimeRegistry                   |
                             | - Policy/Limits enforcement         |
                             +------------------+------------------+
                                            |
                                            v
                          +-----------------+----------------------+
                          | Infrastructure Ports                   |
                          | - VMStore (SQLite)                     |
                          | - Library loader / module metadata     |
                          | - Clock / ID generator / logger        |
                          +----------------------------------------+

Daemon-specific wrapper:
- `pkg/vmdaemon` hosts process lifecycle, startup recovery, signal handling, bind/listen config.
- `pkg/vmtransport/http` exposes REST API backed by `pkg/vmcontrol`.
```

## Process model

- `vm-system serve` (or `vm-system daemon`) starts daemon.
- Daemon wires one `vmcontrol.Core` instance and one HTTP adapter.
- All session/execution operations in daemon mode run through the core.
- CLI commands in client mode call daemon endpoints instead of direct runtime operations.
- Non-daemon packages may construct `vmcontrol.Core` directly and call services in-process.

## Internal Architecture

## Package proposal

```text
vm-system/
  cmd/vm-system/
    cmd_serve.go
    cmd_template.go      # new user-facing template commands
    cmd_client_*.go      # API client wrappers
  pkg/vmcontrol/
    core.go              # constructor + dependency wiring
    ports.go             # interfaces for store/runtime/clock/idgen
    template_service.go
    session_service.go
    execution_service.go
    runtime_registry.go
    recovery_policy.go
    policy_enforcement.go
  pkg/vmtransport/http/
    server.go
    middleware.go
    handlers_templates.go
    handlers_sessions.go
    handlers_executions.go
    dto.go
  pkg/vmclient/
    rest_client.go       # shared client for CLI and other consumers
    templates_client.go
    sessions_client.go
    executions_client.go
  pkg/vmdaemon/
    app.go               # process host around vmcontrol + transport
    config.go
    lifecycle.go
    recovery.go
  pkg/vmstore/
    vmstore.go           # concrete adapter for vmcontrol store ports
  pkg/libloader/
    loader.go            # concrete adapter for module/library metadata
```

## Core services and responsibilities

### vmcontrol.Core

`vmcontrol.Core` exposes explicit services that are independent of HTTP and CLI transport:

1. `TemplateService`
2. `SessionService`
3. `ExecutionService`
4. `RecoveryPolicy`
5. `RuntimeRegistry`

### TemplateService

- CRUD templates (existing VM profile mapping initially).
- Validate settings/capabilities/startup file declarations.
- Produce immutable creation snapshot for each new session.

### SessionService

- Create/close/list/get sessions.
- Own session state transitions.
- Coordinate startup file execution and crash semantics.

### RuntimeRegistry

- Map `session_id -> runtime handle`.
- Enforce single active owner per session.
- Provide concurrency guardrails and cleanup semantics.

### ExecutionService

- Validate execution request against session state/template policy.
- Run REPL or run-file.
- Persist execution and event stream.
- Return structured execution summary.

### RecoveryPolicy (core) vs RecoveryOrchestrator (daemon)

- `RecoveryPolicy` in `vmcontrol` defines restart behavior and state transitions.
- `RecoveryOrchestrator` in `vmdaemon` decides when to invoke recovery (startup hooks, admin commands).

On daemon startup:

1. Load durable sessions with `status in (starting, ready)`.
2. Apply restart policy:
- v1 default: mark as `closed` with restart reason.
3. Ensure no stale “ready but no runtime” entries remain.

### Transport responsibilities (explicitly outside core)

- HTTP request/response mapping, status codes, JSON DTO validation.
- Auth/cors/rate-limit middleware.
- Streaming protocol details (polling vs SSE vs websocket) without changing core service interfaces.

## Data Model and Schema Changes

Existing schema is strong, so schema evolution should remain additive and low risk.

## Proposed schema extensions

1. `vm` table may remain as-is; expose as templates in API.
2. Add `template_version` field (optional) for immutable snapshots.
3. Add `session_origin_template_version` in `vm_session` to track exact policy snapshot.
4. Add `restart_reason` and `closed_by` fields in `vm_session` for operations clarity.
5. Add `execution_request_id` (idempotency key) optional in `execution`.
6. Add schema version log table:

```sql
CREATE TABLE IF NOT EXISTS schema_version_log (
  version INTEGER PRIMARY KEY,
  applied_at INTEGER NOT NULL,
  description TEXT NOT NULL
);
```

## Template snapshot principle

At session creation, copy the effective template policy into session runtime metadata. This prevents ambiguous behavior if template is edited while session is active.

```pseudo
session.runtime_meta = {
  template_id,
  template_version,
  limits_snapshot,
  capabilities_snapshot,
  libraries_snapshot,
  startup_snapshot
}
```

## REST API Design

All endpoints versioned under `/api/v1`.

The REST layer is an adapter over `vmcontrol` service interfaces. Handler methods should remain thin and follow this shape:

```pseudo
func (h *SessionHandler) Create(c):
  req = decodeAndValidate(c)
  out, err = h.core.Sessions.Create(ctx, req.toCoreInput())
  return mapCoreResultToHTTP(out, err)
```

## Error contract

```json
{
  "error": {
    "code": "SESSION_NOT_READY",
    "message": "Session is not ready",
    "details": {"session_id":"..."}
  }
}
```

Status code guidance:

- `200` success retrieval/update
- `201` created
- `400` validation error
- `404` not found
- `409` conflict (busy/not-ready/duplicate)
- `422` semantic violation
- `500` internal failure

## Template endpoints

- `GET /api/v1/templates`
- `POST /api/v1/templates`
- `GET /api/v1/templates/:template_id`
- `PATCH /api/v1/templates/:template_id`
- `DELETE /api/v1/templates/:template_id`
- `POST /api/v1/templates/:template_id/capabilities`
- `GET /api/v1/templates/:template_id/capabilities`
- `POST /api/v1/templates/:template_id/startup-files`
- `GET /api/v1/templates/:template_id/startup-files`

## Session endpoints

- `GET /api/v1/sessions?status=ready&template_id=...`
- `POST /api/v1/sessions`
- `GET /api/v1/sessions/:session_id`
- `POST /api/v1/sessions/:session_id/close`
- `DELETE /api/v1/sessions/:session_id` (optional destructive cleanup)

Request example:

```json
{
  "template_id": "tpl-123",
  "workspace_id": "ws-abc",
  "base_commit_oid": "deadbeef",
  "worktree_path": "/tmp/ws"
}
```

## Execution endpoints

- `POST /api/v1/executions/repl`
- `POST /api/v1/executions/run-file`
- `GET /api/v1/executions/:execution_id`
- `GET /api/v1/executions?session_id=<id>&limit=50`
- `GET /api/v1/executions/:execution_id/events?after_seq=0`

REPL request example:

```json
{
  "session_id": "sess-123",
  "input": "1+2"
}
```

run-file request example:

```json
{
  "session_id": "sess-123",
  "path": "scripts/main.js",
  "args": {"name":"test"},
  "env": {"MODE":"demo"}
}
```

## Metadata endpoints (optional but recommended)

- `GET /api/v1/metadata/modules`
- `GET /api/v1/metadata/libraries`

These remove duplication across UI and backend and prevent version drift.

## Health/ops endpoints

- `GET /api/v1/health`
- `GET /api/v1/metrics` (Prometheus/plaintext optional)
- `GET /api/v1/runtime/summary` (sessions active, queue depth, etc.)

## CLI Design (Client + Daemon)

## New daemon command

```bash
vm-system serve \
  --db vm-system.db \
  --listen 127.0.0.1:3210 \
  --cors-origin http://localhost:3000
```

## CLI mode strategy

### Mode A: Client mode (default for runtime commands)

- CLI calls daemon REST endpoints.
- Works across invocations with stable sessions.

### Embedded core usage (for reusable package integrations)

- non-CLI packages can import `pkg/vmcontrol` directly.
- no HTTP bind required; call service interfaces in-process.
- same behavior/contracts as daemon mode because both paths use the same core.

## Proposed command tree

```text
vm-system
  serve
  template
    create
    list
    get
    update
    delete
    add-capability
    list-capabilities
    add-startup
    list-startup
  session
    create
    list
    get
    close
  exec
    repl
    run-file
    list
    get
    events
  meta
    modules
    libraries
```

## Cutover policy

1. Remove `vm` command aliases.
2. Expose only `template` command group in CLI.
3. Expose only `/templates` resources in API.
4. Update scripts and UI integration to the new surface in the same release boundary.

## Runtime Lifecycle and Concurrency

## Session state machine

```text
starting -> ready -> closing -> closed
   |                    ^
   v                    |
 crashed ---------------+
```

Additional states can improve observability (`recovering`, `stale`), but not mandatory in v1.

## Execution concurrency policy

- One active execution per session.
- Use per-session lock + queue policy (reject or queue).
- v1 recommendation: reject with `409 SESSION_BUSY` to keep behavior simple/predictable.

## Timeout and limits enforcement

Mandatory v1 enforcement hooks:

1. wall-clock timeout
2. max events
3. max output bytes

Pseudo-flow:

```pseudo
execute(session, request):
  lock(session)
  startTimer(wall_ms)
  for each event emitted:
    if events > max_events: abort
    if output_bytes > max_output_kb*1024: abort
  persist metrics
  unlock(session)
```

## Template, Capability, and Startup Policy

Templates should define the exact capability envelope for sessions.

### Capability contract

A capability entry should include:

- `kind` (module/global/fs/net/env)
- `name`
- `enabled`
- `config`

Runtime context should be derived from enabled capabilities only.

### Startup file handling

- Resolve startup files relative to `worktree_path`.
- Enforce deterministic order by `order_index`.
- Record startup execution as a first-class execution entry or structured startup event for auditability.

### Template mutation policy

When template changes after session creation:

- Existing sessions continue using snapshot.
- New sessions use latest template version.

## Security and Trust Model

## Threat surfaces

1. Remote execution requests.
2. File path traversal in run-file/startup resolution.
3. Capability overexposure.
4. Resource exhaustion via large outputs/event floods.

## Required safeguards (v1)

1. Normalize and validate file paths against allowed roots.
2. Enforce template capability policies before runtime context build.
3. Add request size limits at HTTP layer.
4. Add optional auth token for non-local usage.
5. Restrict daemon bind address by default (`127.0.0.1`).

## Observability and Operations

## Structured logging

Log fields:

- `request_id`
- `session_id`
- `execution_id`
- `template_id`
- `status`
- `duration_ms`
- `error_code`

## Metrics

Recommended metrics set:

- active_sessions
- session_creates_total
- session_crashes_total
- execution_total{kind,status}
- execution_duration_ms_bucket
- execution_output_bytes_total
- execution_events_total

## Runbook snippets

### Health check

```bash
curl -s http://127.0.0.1:3210/api/v1/health
```

### List ready sessions

```bash
curl -s "http://127.0.0.1:3210/api/v1/sessions?status=ready"
```

## Alternatives Considered

## Alternative 1: Keep current one-shot CLI runtime

Rejected. This preserves existing session continuity defect and cannot support real UI integration.

## Alternative 2: Reconstruct runtimes on every command from DB

Partially viable for stateless run-file, but poor for stateful REPL sessions and expensive for repeated usage. Rejected as default design.

## Alternative 3: Daemon with REST API (selected)

Chosen with one constraint: daemon and REST must be adapters on top of a reusable core package.

## Alternative 4: Daemon with gRPC-only API

Not selected for v1. gRPC is viable but raises integration complexity for browser clients without additional gateway layer.

## Alternative 5: Separate template service and execution service processes in v1

Rejected for initial rollout; premature operational complexity. Keep single daemon process with clear internal service boundaries first.

## Alternative 6: Daemon-centric implementation without reusable core

Rejected. This would fix session continuity but would keep orchestration locked inside daemon internals and prevent reuse by other packages.

## Implementation Plan

## Phase 0: Extract reusable control core

1. Create `pkg/vmcontrol` with explicit interfaces and constructor wiring.
2. Move/refactor `vmsession` and `vmexec` logic behind core service interfaces.
3. Define store/runtime/module loader ports and concrete adapters.
4. Add in-process integration tests that do not require HTTP.

## Phase 1: Add daemon host + HTTP adapter

1. Add `pkg/vmdaemon` process host with lifecycle/signal/recovery wiring.
2. Add `pkg/vmtransport/http` with health endpoint and middleware baseline.
3. Implement session/execution handlers as thin mappings to core services.
4. Add integration tests for continuity across multiple API requests.

## Phase 2: Template-first API and CLI cutover

1. Introduce `/templates` endpoints (backed by existing vm store initially).
2. Add `template` CLI command group and `pkg/vmclient` shared REST client.
3. Remove `vm` aliases and obsolete command paths.
4. Add metadata endpoints for modules/libraries.

## Phase 3: Policy and runtime hardening

1. Implement capability-enforced runtime context build in core.
2. Enforce limits (wall/events/output) in core execution path.
3. Fix library filename/manifest contract end-to-end.
4. Add startup execution audit entries.

## Phase 4: Reuse enablement and adapter expansion

1. Add in-process adapter examples/tests for non-daemon package usage.
2. Publish stable `vmcontrol` service interfaces and API contract guarantees.
3. Add architecture tests to prevent transport logic from leaking into core.
4. Document embedding guidelines for other packages.

## Phase 5: Operations and rollout

1. Add metrics endpoint and dashboards.
2. Update scripts (`smoke-test.sh`, etc.) to daemon client mode.
3. Provide cutover playbook for CLI/UI/operator teams.
4. Stabilize for `vm-system-ui` BackendAdapter integration.

## Post-cutover API/CLI Surface

| CLI command path | Backend route |
|---|---|
| `template create` | `POST /templates` |
| `template list` | `GET /templates` |
| `session create` | `POST /sessions` |
| `exec repl` | `POST /executions/repl` |
| `exec run-file` | `POST /executions/run-file` |
| `exec events` | `GET /executions/:id/events` |

## Definition of Done

1. Daemon mode can host long-lived sessions and execute across independent client calls.
2. REST endpoints for template/session/execution/event workflows are available and tested.
3. CLI runtime commands use daemon API by default.
4. `pkg/vmcontrol` is reusable from another package without importing daemon/http code.
5. Library loading contract fixed and verified.
6. Capability and limits baseline enforced.
7. Updated smoke/e2e scripts pass in daemon client mode.
8. vm-system-ui can run in real backend mode without browser eval path.

## Open Questions

1. Should daemon support persistent auth in v1 or remain localhost-only trusted mode?
2. Should event transport in v1 stay polling or introduce SSE immediately?
3. Should stale sessions on restart be always auto-closed or optionally rehydrated?
4. Should template versioning be explicit user-managed or auto-incremented on every mutation?
5. Should `pkg/vmcontrol` expose concrete structs, interfaces only, or both for external consumers?

## References

- `vm-system/vm-system/ttmp/2026/02/07/VM-001-ANALYZE-VM--analyze-vm-system-architecture-behavior-and-quality/design-doc/01-comprehensive-vm-system-analysis-report.md`
- `vm-system/vm-system/ttmp/2026/02/07/VM-003-MAKE-WEB-UI-REAL--make-vm-system-ui-real-by-integrating-with-vm-system-backend/design-doc/01-make-web-ui-real-backend-integration-analysis-and-implementation-plan.md`
- `vm-system/vm-system/cmd/vm-system/main.go`
- `vm-system/vm-system/cmd/vm-system/cmd_exec.go`
- `vm-system/vm-system/cmd/vm-system/cmd_session.go`
- `vm-system/vm-system/pkg/vmsession/session.go`
- `vm-system/vm-system/pkg/vmexec/executor.go`
- `vm-system/vm-system/pkg/vmstore/vmstore.go`
- `vm-system/vm-system/pkg/libloader/loader.go`
