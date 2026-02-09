---
Title: Implementation Guide
Ticket: VM-015-PANIC-LOGGING-ALIGNMENT
Status: active
Topics:
    - backend
    - architecture
DocType: design
Intent: long-term
Owners: []
RelatedFiles:
    - /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/cmd/vm-system/main.go
    - /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/cmd/vm-system/glazed_helpers.go
    - /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/cmd/vm-system/cmd_serve.go
    - /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/vmmodels/ids.go
    - /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/vmsession/session.go
    - /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/libloader/loader.go
    - /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/vmdaemon/app.go
    - /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/vmtransport/http/server.go
ExternalSources: []
Summary: >
    Systematic source-code audit and execution plan for removing panic usage from
    production paths and standardizing runtime logging on zerolog configured by
    glazed, without backwards-compatibility shims.
LastUpdated: 2026-02-09T19:20:00-05:00
WhatFor: Execute VM-015 with precise file-level refactors and validation expectations.
WhenToUse: Use when implementing panic removal and zerolog standardization in vm-system.
---

# VM-015 Implementation Guide (Deep Source Audit)

## Objective

Replace all production `panic(...)` usage with explicit error-returning paths and unify operational logging on `zerolog` (configured by glazed logging flags).

This guide is intentionally strict:
- no backwards-compatibility helper APIs
- no panic in production execution paths
- no runtime/service `fmt.Print*` logging
- preserve CLI command output semantics where output is user-facing data

## Scope and Non-Goals

### In scope

- `cmd/vm-system` command construction and root command error handling
- runtime services and adapters in `pkg/` (`vmsession`, `libloader`, `vmdaemon`, `vmtransport/http`)
- replacing unstructured runtime prints with structured zerolog events
- removing `Must*` ID helpers from production API surface

### Out of scope

- test-only panic assertions (`*_test.go`) except where production API changes require test updates
- changing API/CLI domain behavior unrelated to error/logging mechanics

## Baseline: How Logging Is Configured Today

The root command already initializes glazed logging:
- `cmd/vm-system/main.go:24` uses `logging.InitLoggerFromCobra(cmd)`
- `cmd/vm-system/main.go:29` registers logging flags via glazed

This means zerolog global config is already present for command execution paths. The remaining work is to route runtime/service logs through zerolog and remove ad-hoc printing/panics.

## Systematic Inventory

### A. Panic call sites (production code)

1. `cmd/vm-system/glazed_helpers.go:47`
- `panic(err)` in `mustBuildCobraCommand`
- Classification: startup path, but still production panic
- Required action: replace with `error` return and propagation

2. `cmd/vm-system/glazed_helpers.go:57`
- `panic(err)` in `mustCommandDescription` for glazed output schema creation
- Classification: startup path, but still production panic
- Required action: replace with `error` return and propagation

3. `cmd/vm-system/glazed_helpers.go:64`
- `panic(err)` in `mustCommandDescription` for command settings section
- Classification: startup path, but still production panic
- Required action: replace with `error` return and propagation

4. `pkg/vmmodels/ids.go:50`
- `panic(err)` in `MustTemplateID`
- Classification: generic library API; currently unused in production, still unsafe surface
- Required action: remove `Must*` family from production API

5. `pkg/vmmodels/ids.go:58`
- `panic(err)` in `MustSessionID`
- Classification: generic library API; currently unused in production, still unsafe surface
- Required action: remove `Must*` family from production API

6. `pkg/vmmodels/ids.go:66`
- `panic(err)` in `MustExecutionID`
- Classification: generic library API; currently unused in production, still unsafe surface
- Required action: remove `Must*` family from production API

### B. Runtime/daemon unstructured logging (must become zerolog)

1. `pkg/libloader/loader.go:39`
- `fmt.Printf("Downloading %d libraries...\n", ...)`
- Classification: runtime/infra progress logging
- Required action: structured info log

2. `pkg/libloader/loader.go:52`
- `fmt.Printf("✓ Downloaded ...")` inside goroutine
- Classification: concurrent progress logging
- Required action: structured info log with `library`, `version`

3. `pkg/libloader/loader.go:70`
- `fmt.Printf("All libraries downloaded successfully!\n")`
- Classification: summary logging
- Required action: structured info log with count and duration

4. `pkg/vmsession/session.go:116`
- startup console function writes `fmt.Println(args...)`
- Classification: runtime logging bypassing structured logging
- Required action: emit structured event instead of raw stdout

5. `pkg/vmsession/session.go:294`
- `fmt.Printf("[Session] Loaded library: %s\n", libName)`
- Classification: runtime service log
- Required action: structured info log with `session_id`/`vm_id`/`library`

6. `cmd/vm-system/cmd_serve.go:51`
- daemon startup status via `fmt.Printf`
- Classification: daemon operational log (not command data)
- Required action: structured info log with `listen_addr`

### C. Ambiguous print sites (do NOT convert to logs)

These are command response payloads/data rendering and should remain stdout writer output:
- `cmd/vm-system/cmd_exec.go` table/detail rendering
- `cmd/vm-system/cmd_session.go` table/detail rendering
- `cmd/vm-system/cmd_template_*.go` payload rendering
- `cmd/vm-system/cmd_ops.go` health/runtime-summary JSON output
- `cmd/vm-system/cmd_libs.go` user-facing output writer rendering

Rule: if the text is command result content intended for users/scripts, keep `fmt.Fprintf(w, ...)` or equivalent writer rendering.

## Target Refactor Design (No Compatibility Shims)

## 1) Remove panic-based command helpers

### Current issue

`mustBuildCobraCommand` and `mustCommandDescription` panic on wiring errors.

### Required signature changes

In `cmd/vm-system/glazed_helpers.go`:
- `mustBuildCobraCommand(c cmds.Command) *cobra.Command`
  -> `buildCobraCommand(c cmds.Command) (*cobra.Command, error)`
- `mustCommandDescription(...) *cmds.CommandDescription`
  -> `newCommandDescription(...) (*cmds.CommandDescription, error)`

### Propagation strategy

All command constructors that currently return `*cobra.Command` and rely on must-helpers should return `(*cobra.Command, error)` and bubble failure up to root assembly.

Minimum impacted constructors:
- `cmd/vm-system/cmd_serve.go:newServeCommand`
- `cmd/vm-system/cmd_template.go:newTemplateCommand` and nested builders
- `cmd/vm-system/cmd_session.go:newSessionCommand` and nested builders
- `cmd/vm-system/cmd_exec.go:newExecCommand` and nested builders
- `cmd/vm-system/cmd_libs.go:newLibs*` builders

Root assembly changes:
- `cmd/vm-system/main.go:newRootCommand(...)` returns `(*cobra.Command, error)`
- `main()` handles and logs root-build errors before exit

### Error handling pattern

Pseudocode:

```go
func newServeCommand() (*cobra.Command, error) {
    desc, err := newCommandDescription(...)
    if err != nil {
        return nil, fmt.Errorf("build serve command description: %w", err)
    }
    cmd := &bareCommand{CommandDescription: desc, run: ...}
    cobraCmd, err := buildCobraCommand(cmd)
    if err != nil {
        return nil, fmt.Errorf("build serve cobra command: %w", err)
    }
    return cobraCmd, nil
}
```

## 2) Remove panic-based model APIs

### Current issue

`pkg/vmmodels/ids.go` exposes panic helpers:
- `MustTemplateID`
- `MustSessionID`
- `MustExecutionID`

### Required action

Delete all three `Must*` functions.

### Why delete (instead of deprecate)

User requirement is explicit: no backwards compatibility. Keeping panic helpers, even deprecated, retains unsafe production API surface.

### Required test updates

- Replace panic assertions in `pkg/vmmodels/ids_test.go` with parse error assertions.
- Ensure coverage validates:
  - valid UUID parse succeeds
  - invalid/empty UUID parse returns typed errors

## 3) Standardize runtime logging on zerolog

### Logging contract decision

Use a package-local narrow logger interface in runtime components, backed by zerolog adapter.

Suggested minimal interface:

```go
type Logger interface {
    Debug(msg string, fields map[string]any)
    Info(msg string, fields map[string]any)
    Warn(msg string, fields map[string]any)
    Error(msg string, err error, fields map[string]any)
}
```

Or, if avoiding custom interface overhead, inject `zerolog.Logger` directly and keep helpers for consistent fields.

Given current codebase size, direct `zerolog.Logger` injection is acceptable and faster.

### Injection points (required)

1. `pkg/vmsession/session.go`
- Add logger field to `SessionManager`
- Constructor change:
  - `NewSessionManager(store *vmstore.VMStore) *SessionManager`
  - -> `NewSessionManager(store *vmstore.VMStore, logger zerolog.Logger) *SessionManager`

2. `pkg/libloader/loader.go`
- Add logger field to `LibraryCache`
- Constructor change:
  - `NewLibraryCache(cacheDir string) (*LibraryCache, error)`
  - -> `NewLibraryCache(cacheDir string, logger zerolog.Logger) (*LibraryCache, error)`

3. `pkg/vmcontrol/core.go`
- Wire logger into session manager creation
- `NewCore` (and optionally `NewCoreWithPorts`) needs logger plumbed through

4. `pkg/vmdaemon/app.go`
- hold logger for lifecycle logs (`listen`, `shutdown`, `server_error`)

5. `cmd/vm-system/cmd_serve.go`
- replace daemon startup `fmt.Printf` with zerolog info event

### Required field conventions

Use stable structured keys:
- `component`: `session_manager` | `library_cache` | `daemon` | `http_server`
- `session_id`, `template_id`, `execution_id`, `library`, `version`, `listen_addr`
- `request_id` for HTTP request logs

## 4) Replace exact runtime print sites

### `pkg/libloader/loader.go`

Replace:
- line 39 `fmt.Printf("Downloading ...")`
- line 52 per-library success print
- line 70 completion print

With structured logs:
- `Info` start event with `library_count`
- `Info` success event with `library`, `version`
- `Info` completion event with `success_count`, `duration_ms`
- `Error` aggregate failure event with first failing library metadata

Concurrency note:
- keep non-blocking logging per goroutine
- avoid mutating shared logger state

### `pkg/vmsession/session.go`

Replace:
- line 116 `fmt.Println(args...)` inside startup console
- line 294 loaded-library print

With:
- console forwarding through structured event:
  - message: `startup_console_log`
  - fields: `session_id`, `vm_id`, `args_count`, `text`
- library load event:
  - message: `library_loaded`
  - fields: `session_id`, `vm_id`, `library`

Important: startup execution currently occurs before executor-specific event recorder is installed. Logging is the correct sink for startup console output unless startup-event persistence is added in a separate ticket.

### `cmd/vm-system/cmd_serve.go`

Replace line 51 with structured info log:
- message: `daemon_listening`
- field: `listen_addr`

## 5) Strengthen top-level error handling

### Current

`cmd/vm-system/main.go:52` prints plain text to stderr on execute failure.

### Target

Emit zerolog error event and keep non-zero exit status.

Pattern:

```go
if err != nil {
    log.Error().Err(err).Msg("command failed")
    os.Exit(1)
}
```

Edge case note:
- if failure occurs before logging initialization, fallback stderr print is acceptable as a final guard.

## 6) Add request/daemon lifecycle logging coverage

Even without existing `fmt.Print*` lines in HTTP transport, VM-015 should establish baseline operational visibility:

1. `pkg/vmdaemon/app.go`
- log daemon start attempt
- log graceful shutdown begin/end
- log listen-and-serve fatal error path

2. `pkg/vmtransport/http/server.go`
- extend middleware to log request summary with `request_id`, method, route, status, duration
- keep API response payloads unchanged

This is needed to satisfy “zerolog across the board” for service operation observability.

## File-by-File Change Map

1. `cmd/vm-system/glazed_helpers.go`
- remove `must*` panic helpers
- introduce error-returning builders

2. `cmd/vm-system/main.go`
- make root assembly return error
- replace plain stderr error print with structured logging

3. `cmd/vm-system/cmd_serve.go`
- convert daemon startup message to structured log
- adapt constructor to propagated build errors

4. `cmd/vm-system/cmd_template*.go`, `cmd/vm-system/cmd_session.go`, `cmd/vm-system/cmd_exec.go`, `cmd/vm-system/cmd_libs.go`
- constructor signatures + error propagation due helper change
- no change to user-facing output rendering

5. `pkg/vmmodels/ids.go`
- delete `Must*` helpers

6. `pkg/vmsession/session.go`
- add logger dependency
- replace runtime prints with structured logs

7. `pkg/libloader/loader.go`
- add logger dependency
- replace runtime prints with structured logs

8. `pkg/vmcontrol/core.go`
- plumb logger into runtime service constructors

9. `pkg/vmdaemon/app.go`
- add lifecycle logging

10. `pkg/vmtransport/http/server.go`
- add request logging middleware tied to request-id

## Tests Required for This Refactor

## A. Panic removal / constructor failure tests

1. `cmd/vm-system` tests
- root command build should return error, never panic
- command constructor errors are surfaced to caller

2. `pkg/vmmodels/ids_test.go`
- remove panic expectations
- assert parse errors instead

## B. Logging behavior tests (lightweight)

1. `pkg/libloader/loader_test.go`
- inject test logger writer
- assert key events emitted for start/success/failure

2. `pkg/vmsession/session*_test.go`
- assert startup console/library load paths emit logs and do not write stdout directly

3. `pkg/vmdaemon` and `pkg/vmtransport/http`
- verify request/lifecycle logs include key correlation fields (`request_id`, `listen_addr`)

## C. Regression tests

- `go test ./...` full suite
- smoke:
  - `vm-system serve --log-level debug`
  - create session and execute snippet
  - verify structured logs only for operational paths

## Risk Notes and Mitigations

1. Signature churn risk
- many command builders impacted
- mitigate with incremental compile checkpoints after each command file conversion

2. Logger injection fan-out risk
- constructor updates across `vmcontrol` and tests
- mitigate by passing one shared logger from daemon/bootstrap wiring

3. Output behavior confusion risk
- converting command data output to logs would break scripts
- mitigate with strict separation:
  - operational events -> logger
  - command result payloads -> stdout writer

## Recommended Execution Sequence

1. Remove panic helpers in command wiring and propagate errors to root build.
2. Delete `Must*` ID functions and update tests.
3. Introduce logger plumbing for `SessionManager` and `LibraryCache`.
4. Replace all runtime `fmt.Print*` sites.
5. Add daemon + HTTP request lifecycle logs.
6. Run full tests and smoke checks.

## Acceptance Criteria

- Production code contains zero `panic(...)` calls.
- Runtime/service paths contain zero `fmt.Print*` logging.
- Zerolog is used for daemon/session/library operational logs.
- CLI result rendering remains stable for user/script consumption.
- `go test ./...` passes after refactor.

## Exact Command Set Used for This Audit

```bash
rg -n "\bpanic\(" --glob '*.go'
rg -n "\bpanic\b|recover\(" --glob '*.go'
rg -n "fmt\.(Print|Printf|Println|Fprint|Fprintf|Fprintln)\(|os\.Stderr|stderr|stdout" --glob '*.go'
rg -n "MustTemplateID|MustSessionID|MustExecutionID" --glob '*.go'
```

