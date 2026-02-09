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
    - /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/vmmodels/ids.go
    - /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/vmsession/session.go
    - /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/libloader/loader.go
    - /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/cmd/vm-system/glazed_helpers.go
ExternalSources: []
Summary: >
    Detailed implementation guide for panic-boundary tightening and structured logging
    alignment across runtime and CLI layers.
LastUpdated: 2026-02-09T10:42:00-05:00
WhatFor: Execute P2-2 safely with explicit panic/logging policy enforcement.
WhenToUse: Use while implementing VM-015.
---

# VM-015 Implementation Guide

## Policy to Enforce

1. Panic usage allowed only for:
- startup-time invariant failures
- test-only helpers

2. Runtime/session/daemon/request paths must:
- return typed errors
- log through one structured logger path

## Step 1: Panic Inventory and Classification

Commands:
- `rg -n "panic\(" pkg cmd`
- classify each call site as `invariant-startup`, `runtime-risk`, or `test-only`.

Output artifact:
- small table in ticket notes mapping each panic to planned action.

## Step 2: Introduce Runtime Logger Interface

1. Define minimal logger interface (Debug/Info/Warn/Error) in runtime packages.
2. Inject logger into:
- session manager
- library loader
- related orchestration layers

3. Provide default logger wiring from CLI/daemon bootstrap.

## Step 3: Replace Direct `fmt.Print*` Runtime Output

Targets:
- `pkg/vmsession/session.go`
- `pkg/libloader/loader.go`

Action:
- replace direct printing with structured logger calls including contextual fields (`session_id`, `template_id`, `library`, `path`).

## Step 4: Constrain/Remove Panics in Non-startup Paths

Targets:
- `pkg/vmmodels/ids.go` and any panic usage reachable from request/runtime logic.

Action:
- replace panic calls with parse/return error pattern where runtime-reachable.
- keep explicit panic only where invariant violations are impossible during normal runtime.

## Step 5: Validation and Regression Checks

Commands:
- `go test ./...`
- run representative CLI flows (`template/session/exec`) and verify log output quality.

Assertions:
- no panic in runtime request path under malformed/invalid inputs.
- logs still provide actionable diagnostics.

## Rollout Strategy

- Land changes in two commits:
1. logger interface + runtime print replacement
2. panic boundary tightening

- Keep behavior-neutral except error-safety improvements.
