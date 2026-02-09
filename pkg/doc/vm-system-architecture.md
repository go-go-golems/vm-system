---
Title: "vm-system Architecture"
Slug: architecture
Short: "Package layout, layered design, request lifecycle, and key design tradeoffs."
Topics:
- vm-system
- architecture
- design
IsTopLevel: true
IsTemplate: false
ShowPerDefault: true
SectionType: GeneralTopic
---

vm-system is a daemon-first JavaScript runtime service. This page explains the
package structure, how a request flows from CLI to runtime, the data model, and
the tradeoffs behind the current design.

## Layered architecture

The codebase follows a ports-and-adapters style with four layers:

```
┌──────────────────────────────────────────────┐
│  cmd/vm-system    CLI commands (Cobra)        │
│  pkg/vmclient     REST client adapter         │
├──────────────────────────────────────────────┤
│  pkg/vmtransport/http   REST handler layer    │
│  pkg/vmdaemon           Process host          │
├──────────────────────────────────────────────┤
│  pkg/vmcontrol    Core orchestration          │
│    TemplateService / SessionService /          │
│    ExecutionService / RuntimeRegistry          │
├──────────────────────────────────────────────┤
│  pkg/vmsession    In-memory runtime ownership │
│  pkg/vmexec       Execution pipeline + events │
│  pkg/vmstore      SQLite persistence          │
│  pkg/vmmodels     Domain types + errors       │
│  pkg/vmmodules    Module registry             │
│  pkg/vmpath       Path normalization          │
│  pkg/libloader    Library cache loader        │
└──────────────────────────────────────────────┘
```

The critical design rule is that **vmcontrol is transport-agnostic**. No HTTP
request objects should appear there. The HTTP layer translates between REST DTOs
and core input/output types.

## Package responsibilities

### cmd/vm-system — CLI entry point

- `main.go` — root command, global `--db` and `--server-url` flags, Glazed help system
- `cmd_serve.go` — daemon bootstrap via `vmdaemon`
- `cmd_template.go`, `cmd_template_core.go`, `cmd_template_modules.go` — template CRUD, modules, libraries, capabilities, startup files
- `cmd_session.go` — session create/list/get/close
- `cmd_ops.go` — health and runtime-summary

All client-mode commands use `pkg/vmclient` to call the daemon REST API.

### pkg/vmcontrol — core orchestration

- `core.go` — composition root wiring services together
- `template_service.go` — template lifecycle and default settings initialization
- `session_service.go` — session lifecycle orchestration
- `execution_service.go` — run-file path safety and limit enforcement
- `runtime_registry.go` — active session summary
- `ports.go` — interfaces separating domain from adapters
- `types.go` — input/output DTOs for service methods

### pkg/vmtransport/http — REST adapter

- Route registration on `net/http.ServeMux`
- JSON decode with `DisallowUnknownFields`
- Error envelope mapping via `writeCoreError`
- Request-ID middleware (`X-Request-Id` header)

### pkg/vmsession — runtime ownership

Manages the in-memory map of active goja runtimes. Each session has an
`ExecutionLock` ensuring one execution at a time.

### pkg/vmexec — execution pipeline

Runs JavaScript in the goja runtime, captures console/value/exception events,
and persists execution rows and events to the store.

### pkg/vmstore — SQLite persistence

Direct `database/sql` adapter. Schema lives in `initSchema()`. Tables:

| Table | Purpose |
|-------|---------|
| `vm` | Template identity and profile |
| `vm_settings` | Limits/resolver/runtime JSON config |
| `vm_capability` | Capability metadata rows |
| `vm_startup_file` | Ordered startup file entries |
| `vm_session` | Durable session records |
| `execution` | Execution summaries |
| `execution_event` | Event stream per execution |

### pkg/vmmodels — domain types

All shared types (`VM`, `VMSession`, `Execution`, `ExecutionEvent`, etc.) and
domain errors (`ErrVMNotFound`, `ErrSessionNotReady`, etc.).

## Request lifecycle

### Template create

```
CLI flags → vmclient.CreateTemplate → POST /api/v1/templates
  → handler validates name → core.Templates.Create
  → store.CreateVM + store.SetVMSettings (defaults)
  → 201 JSON response → CLI prints ID
```

### Session create

```
CLI flags → vmclient.CreateSession → POST /api/v1/sessions
  → handler validates fields → core.Sessions.Create
  → runtime manager allocates goja runtime
  → console shim + libraries + startup files executed
  → status: starting → ready (or crashed)
  → 201 JSON response
```

### Execution (REPL)

```
CLI args → vmclient.ExecuteREPL → POST /api/v1/executions/repl
  → handler validates → core.Executions.ExecuteREPL
  → executor acquires session lock (or SESSION_BUSY)
  → execution row created (status: running)
  → goja RunString + event capture
  → execution finalized (ok / error)
  → 201 JSON response with events
```

## Data model

Sessions have four states: `starting → ready → closed` (or `crashed`).

Executions have states: `running → ok | error | timeout | cancelled`.

Events are typed: `input_echo`, `console`, `value`, `exception`, `system`,
`stdout`, `stderr`. Each event carries a `seq` number for cursor-based retrieval.

In-memory runtime state is **not** reconstructed from the database after daemon
restart. Persisted session rows survive but active runtimes are lost.

## Design tradeoffs

### In-memory runtimes (no persistence of live state)

**Benefit:** Low complexity, fast execution, straightforward lock semantics.
**Cost:** Daemon restart loses active sessions. Recovery semantics are a known
future work item.

### JSON blobs for settings

Template settings (limits, resolver, runtime) are stored as JSON text columns
rather than normalized tables. This makes schema evolution easy but weakens
static queryability.

### Coarse execution lock

One lock per session guarantees deterministic single-threaded execution.
Concurrent requests get `SESSION_BUSY` (409). This is intentionally simple —
there is no execution queue.

### Explicit SQL, no ORM

The store uses plain `database/sql` with hand-written SQL. This keeps
dependencies minimal and behavior transparent, at the cost of some repetition.

## Concurrency model

- Each session has a `sync.Mutex`-based `ExecutionLock`
- `TryLock` is used — the executor does not block, it returns `SESSION_BUSY`
- There is no global execution lock; different sessions can execute concurrently
- Console shim and event recording happen under the lock

## Reading order for new contributors

Read these files in sequence for the fastest understanding:

1. `cmd/vm-system/main.go`
2. `cmd/vm-system/cmd_serve.go`
3. `pkg/vmdaemon/app.go`
4. `pkg/vmcontrol/core.go`
5. `pkg/vmtransport/http/server.go`
6. `pkg/vmcontrol/template_service.go`
7. `pkg/vmcontrol/session_service.go`
8. `pkg/vmcontrol/execution_service.go`
9. `pkg/vmsession/session.go`
10. `pkg/vmexec/executor.go`
11. `pkg/vmstore/vmstore.go`

## Troubleshooting

| Problem | Cause | Solution |
|---------|-------|----------|
| Transport types leak into vmcontrol | Layer violation | Move HTTP-specific logic to handler; keep core input types in `types.go` |
| Error returns 500 INTERNAL | Missing error mapping | Add sentinel error + mapping in `writeCoreError` |
| Session state inconsistency | Status update ordering | Ensure store writes happen in deterministic order; test transitions |

## See Also

- `vm-system help getting-started`
- `vm-system help api-reference`
- `vm-system help contributing`
