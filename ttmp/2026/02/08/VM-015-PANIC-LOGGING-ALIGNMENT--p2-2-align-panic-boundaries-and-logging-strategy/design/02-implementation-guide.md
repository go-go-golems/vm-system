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
    - /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/cmd/vm-system/cmd_serve.go
    - /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/cmd/vm-system/cmd_template_core.go
    - /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/cmd/vm-system/cmd_template_modules.go
    - /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/cmd/vm-system/cmd_template_libraries.go
    - /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/cmd/vm-system/cmd_template_startup.go
    - /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/cmd/vm-system/cmd_session.go
    - /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/cmd/vm-system/cmd_exec.go
    - /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/vmmodels/ids.go
    - /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/vmsession/session.go
    - /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/libloader/loader.go
    - /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/vmdaemon/app.go
    - /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/vmtransport/http/server.go
ExternalSources: []
Summary: >
    Updated VM-015 execution guide after removing glazed closure helpers; now focused
    on remaining panic removal and zerolog standardization work.
LastUpdated: 2026-02-09T20:05:00-05:00
WhatFor: Execute VM-015 with precise, current-state file-level refactors.
WhenToUse: Use when implementing panic removal and zerolog standardization in vm-system.
---

# VM-015 Implementation Guide (Current State)

## Objective

Enforce two rules across production code:
- no panic-based control flow in runtime/production paths
- operational logging goes through `zerolog` (configured by glazed logging flags)

No backwards-compatibility shims.

## Important Update (Already Done)

The CLI closure-wrapper layer has been removed.

Completed change:
- `cmd/vm-system/glazed_helpers.go` deleted
- all command groups now use plain `cobra.Command` directly:
  - `cmd/vm-system/cmd_serve.go`
  - `cmd/vm-system/cmd_libs.go`
  - `cmd/vm-system/cmd_session.go`
  - `cmd/vm-system/cmd_exec.go`
  - `cmd/vm-system/cmd_template*.go`

Implication for VM-015:
- no helper-panics remain in CLI command wiring
- panic inventory must now target remaining production sites only

## Baseline: Logging bootstrap

Root logging bootstrap remains correct:
- `cmd/vm-system/main.go:24` uses `logging.InitLoggerFromCobra(cmd)`
- glazed logging flags are registered at root and configure global zerolog

## Systematic Inventory (After helper removal)

## A. Remaining panic call sites (production code)

1. `pkg/vmmodels/ids.go:50`
- `panic(err)` in `MustTemplateID`

2. `pkg/vmmodels/ids.go:58`
- `panic(err)` in `MustSessionID`

3. `pkg/vmmodels/ids.go:66`
- `panic(err)` in `MustExecutionID`

Classification:
- utility-library panic APIs, currently unused in production call paths but still exported unsafe surface

Required action:
- delete all `Must*` functions
- update tests to assert parse errors instead of panic behavior

## B. Remaining runtime/daemon unstructured logging (must become zerolog)

1. `pkg/libloader/loader.go:39`
- `fmt.Printf("Downloading %d libraries...\n", ...)`

2. `pkg/libloader/loader.go:52`
- `fmt.Printf("✓ Downloaded %s v%s\n", ...)`

3. `pkg/libloader/loader.go:70`
- `fmt.Printf("All libraries downloaded successfully!\n")`

4. `pkg/vmsession/session.go:116`
- startup console hook uses `fmt.Println(args...)`

5. `pkg/vmsession/session.go:294`
- `fmt.Printf("[Session] Loaded library: %s\n", libName)`

6. `cmd/vm-system/cmd_serve.go:38`
- daemon startup message uses `fmt.Printf`

## C. Print sites that should remain output rendering

These are command data outputs, not operational logs, and should stay on stdout:
- `cmd/vm-system/cmd_exec.go`
- `cmd/vm-system/cmd_session.go`
- `cmd/vm-system/cmd_template_*.go`
- `cmd/vm-system/cmd_ops.go`
- `cmd/vm-system/cmd_libs.go`

Rule:
- command result payloads/tables -> stdout
- service/runtime/daemon lifecycle -> zerolog

## Design for Remaining VM-015 Work

## 1) Remove panic-based ID helpers

Target file:
- `pkg/vmmodels/ids.go`

Action:
- remove `MustTemplateID`, `MustSessionID`, `MustExecutionID`

Tests:
- update `pkg/vmmodels/ids_test.go`
- remove panic assertions and assert parse errors instead

## 2) Standardize runtime logging on zerolog

Preferred approach:
- inject logger dependency into runtime services (`zerolog.Logger`)
- avoid package-global ad-hoc logging in runtime code

Required injection points:
1. `pkg/vmsession/session.go`
- `SessionManager` gets logger field
- constructor updated to accept logger

2. `pkg/libloader/loader.go`
- `LibraryCache` gets logger field
- constructor updated to accept logger

3. `pkg/vmcontrol/core.go`
- plumb logger into `NewSessionManager`

4. `pkg/vmdaemon/app.go`
- add lifecycle logging (`listen`, `shutdown`, server errors)

5. `cmd/vm-system/cmd_serve.go`
- use zerolog for daemon listening event

## 3) Add request/daemon observability baseline

1. `pkg/vmdaemon/app.go`
- log startup and shutdown lifecycle transitions

2. `pkg/vmtransport/http/server.go`
- add request summary logging with `request_id`, method, route, status, duration

## Field naming conventions

Use stable structured keys:
- `component`: `daemon`, `session_manager`, `library_cache`, `http_server`
- `listen_addr`, `session_id`, `template_id`, `execution_id`, `library`, `version`
- `request_id`, `status_code`, `duration_ms`

## File-by-file change map (remaining)

1. `pkg/vmmodels/ids.go`
- delete `Must*` panic helpers

2. `pkg/vmmodels/ids_test.go`
- replace panic expectations with parse-error assertions

3. `pkg/libloader/loader.go`
- replace all runtime `fmt.Printf` logging with structured logger calls
- inject logger dependency

4. `pkg/vmsession/session.go`
- replace runtime `fmt.Println/Printf` logging with structured logger calls
- inject logger dependency

5. `pkg/vmcontrol/core.go`
- logger plumbing to runtime constructor wiring

6. `pkg/vmdaemon/app.go`
- daemon lifecycle logs

7. `cmd/vm-system/cmd_serve.go`
- replace daemon startup `fmt.Printf` with logger call

8. `pkg/vmtransport/http/server.go`
- add request summary logging middleware

## Validation checklist

- `rg -n "\bpanic\(" --glob '*.go'` returns no production panic sites
- `rg -n "fmt\.Print|fmt\.Printf|fmt\.Println" pkg cmd/vm-system/cmd_serve.go` returns no operational/runtime prints
- `GOWORK=off go test ./... -count=1` passes
- smoke check daemon startup and request path logs include structured fields

## Commands used to produce this inventory

```bash
rg -n "\bpanic\(" --glob '*.go'
rg -n "fmt\.Print|fmt\.Printf|fmt\.Println" pkg cmd/vm-system/cmd_serve.go
nl -ba pkg/vmmodels/ids.go
nl -ba pkg/libloader/loader.go
nl -ba pkg/vmsession/session.go
nl -ba cmd/vm-system/cmd_serve.go
```
