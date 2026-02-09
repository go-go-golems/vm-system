---
Title: Implementation Guide
Ticket: VM-014-DECOMPOSE-MONOLITHS
Status: active
Topics:
    - backend
    - frontend
    - architecture
DocType: design
Intent: long-term
Owners: []
RelatedFiles:
    - /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/vmtransport/http/server.go
    - /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/vmstore/vmstore.go
    - /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/cmd/vm-system/cmd_template.go
    - /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system-ui/client/src/lib/api.ts
ExternalSources: []
Summary: >
    Detailed implementation guide for decomposing monolithic backend/frontend files
    into module-focused units while preserving behavior.
LastUpdated: 2026-02-09T10:41:00-05:00
WhatFor: Execute P2-1 decomposition in safe, test-backed increments.
WhenToUse: Use while implementing VM-014.
---

# VM-014 Implementation Guide

## Guardrails

- No API/CLI behavior changes in this ticket.
- Refactor in small slices with compile/test after each slice.
- Keep external package paths stable unless explicitly planned.

## Slice A: `pkg/vmtransport/http/server.go`

1. Create files:
- `server_templates.go`
- `server_sessions.go`
- `server_executions.go`
- `server_errors.go`

2. Move logic in order:
- error helpers first (`writeCoreError`, envelope types)
- template handlers
- session handlers
- execution handlers

3. Keep `server.go` responsible for:
- `Server` struct
- route registration
- tiny shared parse/decode helpers

Checks per slice:
- `go test ./pkg/vmtransport/http -run TestTemplate`
- `go test ./pkg/vmtransport/http -run TestSession`
- `go test ./pkg/vmtransport/http -run TestExecution`

## Slice B: `pkg/vmstore/vmstore.go`

1. Split persistence groups:
- `vmstore_templates.go`
- `vmstore_sessions.go`
- `vmstore_executions.go`
- `vmstore_migrations.go`

2. Keep `vmstore.go` for:
- type/constructor
- DB open/close
- shared scan helpers

Checks:
- `go test ./pkg/vmstore` (when tests exist)
- `go test ./...` as full safety net

## Slice C: `cmd/vm-system/cmd_template.go`

1. Split command builders:
- `cmd_template_core.go`
- `cmd_template_modules.go`
- `cmd_template_libraries.go`
- `cmd_template_startup.go`

2. Keep root registration in one file.

Checks:
- `go test ./cmd/vm-system`
- `go run ./cmd/vm-system --help` sanity check

## Slice D: `vm-system-ui/client/src/lib/api.ts`

1. Create modules:
- `lib/vm/transport.ts`
- `lib/vm/endpoints/shared.ts`
- `lib/vm/endpoints/templates.ts`
- `lib/vm/endpoints/sessions.ts`
- `lib/vm/endpoints/executions.ts`

2. Keep a compatibility export surface:
- maintain current imports via a thin facade file during migration.

Checks:
- `cd vm-system-ui && pnpm check`
- `cd vm-system-ui && pnpm build`

## Recommended Commit Strategy

- One commit per slice (A-D), not one giant commit.
- Include only moved/adjusted files in each commit.
- Run checks before each commit.

## Acceptance Checklist

- Largest-file pressure reduced for each target.
- No route/CLI/API regressions.
- Build and tests pass after each slice.
