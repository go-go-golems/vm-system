---
Title: Monolithic File Decomposition Plan
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
    - /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system-ui/client/src/lib/vmService.ts
ExternalSources: []
Summary: >
    P2-1 planning ticket for breaking up large multi-concern files into maintainable modules
    without changing external behavior.
LastUpdated: 2026-02-09T10:18:00-05:00
WhatFor: Define decomposition boundaries and migration sequence for large files.
WhenToUse: Use when executing monolith decomposition work.
---

# Monolithic File Decomposition Plan

## Scope

Target files:
- `pkg/vmtransport/http/server.go`
- `pkg/vmstore/vmstore.go`
- `cmd/vm-system/cmd_template.go`
- `client/src/lib/vmService.ts`

## Objectives

- Reduce merge conflict pressure and cognitive load.
- Create explicit module boundaries per domain.
- Preserve behavior and CLI/API contracts.

## Proposed Decomposition

### Backend transport (`server.go`)

Split by route domain:
- `server_templates.go`
- `server_sessions.go`
- `server_executions.go`
- `server_errors.go`
- keep `server.go` as router registration and shared helpers.

### Backend store (`vmstore.go`)

Split by persistence aggregate:
- `vmstore_templates.go`
- `vmstore_sessions.go`
- `vmstore_executions.go`
- `vmstore_migrations.go`

### CLI template command (`cmd_template.go`)

Split by command family:
- `cmd_template_core.go`
- `cmd_template_modules.go`
- `cmd_template_libraries.go`
- `cmd_template_startup.go`

### Frontend service (`vmService.ts`)

Split by concern:
- transport request helper
- raw API types
- normalization/mappers
- template operations
- session/execution operations

## Constraints

- No behavior regressions.
- Keep public command/API signatures stable.
- Add/adjust tests around moved logic before large moves.

## Execution Sequence

1. Extract pure helpers first.
2. Move handlers/ops in small slices with compile+test after each slice.
3. Update imports only after new modules are stable.

## Acceptance Criteria

- Target files reduced substantially.
- All checks pass (`go test ./...`, `pnpm check`, `pnpm build`).
- Route/CLI/service behavior unchanged.
