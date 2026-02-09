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
    - /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/cmd/vm-system/glazed_support.go
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
    Updated VM-015 execution guide after removing the glazed closure-wrapper layer
    while keeping commands Glazed-based; includes current panic and logging debt.
LastUpdated: 2026-02-09T20:45:00-05:00
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

Completed change:
- `cmd/vm-system/glazed_helpers.go` deleted
- commands remain Glazed-based (`cmds.WriterCommand` / `cmds.BareCommand`)
- closure wrappers were replaced with explicit command structs and action dispatchers
- shared Glazed wiring now lives in `cmd/vm-system/glazed_support.go`

Implication for VM-015:
- helper-panics still exist in `glazed_support.go` and must be removed as part of this ticket

## Baseline: Logging bootstrap

Root logging bootstrap remains correct:
- `cmd/vm-system/main.go:24` uses `logging.InitLoggerFromCobra(cmd)`
- glazed logging flags are registered at root and configure global zerolog

## Systematic Inventory (Current Status)

## A. Panic call sites (production code)

Current status:
- none (`rg -n "\bpanic\(" --glob '*.go'` returns no matches)

Completed actions:
- replaced `cmd/vm-system/glazed_support.go` panic paths with error-returning helpers + fallback error handling
- deleted `Must*` ID helpers in `pkg/vmmodels/ids.go` and updated `pkg/vmmodels/ids_test.go`

## B. Runtime/daemon unstructured logging

Current status in target files:
- `pkg/libloader/loader.go`: operational lifecycle logging migrated to structured zerolog
- `pkg/vmsession/session.go`: startup/runtime operational logging migrated to structured zerolog
- `cmd/vm-system/cmd_serve.go`: daemon startup banner migrated to structured zerolog

## C. Print sites that should remain output rendering

These are command data outputs and should stay stdout-oriented:
- `cmd/vm-system/cmd_exec.go`
- `cmd/vm-system/cmd_session.go`
- `cmd/vm-system/cmd_template_*.go`
- `cmd/vm-system/cmd_ops.go`
- `cmd/vm-system/cmd_libs.go`

Rule:
- command result payloads/tables -> stdout
- service/runtime/daemon lifecycle -> zerolog

## Design for Remaining VM-015 Work

## 1) Remove panic-based command support functions

Target:
- `cmd/vm-system/glazed_support.go`

Action:
- convert `buildCobraCommand` and `commandDescription` to error-returning APIs
- update command constructors to propagate errors up to root wiring

## 2) Remove panic-based ID helpers

Target:
- `pkg/vmmodels/ids.go`

Action:
- delete `MustTemplateID`, `MustSessionID`, `MustExecutionID`

Tests:
- update `pkg/vmmodels/ids_test.go` to assert parse errors (not panic)

## 3) Standardize runtime logging on zerolog

Injection targets:
1. `pkg/vmsession/session.go`
2. `pkg/libloader/loader.go`
3. `pkg/vmcontrol/core.go`
4. `pkg/vmdaemon/app.go`
5. `cmd/vm-system/cmd_serve.go`

## 4) Add daemon/request lifecycle observability

1. `pkg/vmdaemon/app.go`
- startup/shutdown/error lifecycle logs

2. `pkg/vmtransport/http/server.go`
- request summary logs (`request_id`, method, route, status, duration)

## Field naming conventions

Use stable keys:
- `component`: `daemon`, `session_manager`, `library_cache`, `http_server`
- `listen_addr`, `session_id`, `template_id`, `execution_id`, `library`, `version`
- `request_id`, `status_code`, `duration_ms`

## File-by-file change map (remaining)

1. `cmd/vm-system/glazed_support.go`
- remove command-construction panic paths

2. `pkg/vmmodels/ids.go`
- delete `Must*` panic helpers

3. `pkg/vmmodels/ids_test.go`
- replace panic expectations with parse-error assertions

4. `pkg/libloader/loader.go`
- replace operational `fmt.Printf` with structured logger calls
- inject logger dependency

5. `pkg/vmsession/session.go`
- replace runtime `fmt.Println/Printf` with structured logger calls
- inject logger dependency

6. `pkg/vmcontrol/core.go`
- logger plumbing to runtime constructor wiring

7. `pkg/vmdaemon/app.go`
- daemon lifecycle logs

8. `cmd/vm-system/cmd_serve.go`
- replace daemon startup `fmt.Printf` with logger call

9. `pkg/vmtransport/http/server.go`
- add request summary logging middleware

## Validation checklist

- `rg -n "\bpanic\(" --glob '*.go'` returns no production panic sites
- `rg -n "fmt\.Print|fmt\.Printf|fmt\.Println" pkg cmd/vm-system/cmd_serve.go` returns no operational/runtime prints
- `GOWORK=off go test ./... -count=1` passes

## Commands used for this inventory

```bash
rg -n "\bpanic\(" --glob '*.go'
rg -n "fmt\.Print|fmt\.Printf|fmt\.Println" pkg cmd/vm-system/cmd_serve.go
nl -ba cmd/vm-system/glazed_support.go
nl -ba pkg/vmmodels/ids.go
nl -ba pkg/libloader/loader.go
nl -ba pkg/vmsession/session.go
nl -ba cmd/vm-system/cmd_serve.go
```
