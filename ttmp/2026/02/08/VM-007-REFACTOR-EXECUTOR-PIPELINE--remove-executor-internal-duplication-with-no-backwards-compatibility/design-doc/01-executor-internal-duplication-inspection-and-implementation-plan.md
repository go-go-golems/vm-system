---
Title: Executor internal duplication inspection and implementation plan
Ticket: VM-007-REFACTOR-EXECUTOR-PIPELINE
Status: active
Topics:
    - backend
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: pkg/vmcontrol/execution_service.go
      Note: |-
        Upstream caller and post-execution limit enforcement behavior
        Caller contract and post-execution policy context
    - Path: pkg/vmexec/executor.go
      Note: |-
        REPL and run-file duplicated lifecycle logic
        Primary duplication hotspot and refactor target
    - Path: pkg/vmstore/vmstore.go
      Note: |-
        Persistence operations used by executor flow and duplicated helper semantics
        Persistence semantics used by execution lifecycle
    - Path: pkg/vmcontrol/types.go
      Note: Shared config aliases and JSON helper behavior
    - Path: pkg/vmmodels/models.go
      Note: Core config model source types
    - Path: pkg/vmtransport/http/server.go
      Note: |-
        External API contract for execution endpoints
        Execution API behavior contracts to preserve intentionally
ExternalSources: []
Summary: Deep inspection of finding 9 (executor duplication) and finding 8 (core model/helper duplication), with a concrete no-backward-compatibility implementation plan.
LastUpdated: 2026-02-08T12:15:40-05:00
WhatFor: Enable a delegated contributor to execute the combined refactor with clear scope, risks, and step-by-step implementation tasks.
WhenToUse: Use as the primary implementation brief for VM-007 refactor work (executor + core helper/model dedup).
---


# Executor internal duplication inspection and implementation plan

## Executive Summary

This ticket combines two VM-006 findings in one implementation stream:

1. finding 9: high duplication in `pkg/vmexec/executor.go`
2. finding 8: remaining core model/helper duplication (especially duplicated `mustMarshalJSON` behavior in `vmstore` and `vmcontrol`)

`pkg/vmexec/executor.go` currently duplicates most execution lifecycle logic across `ExecuteREPL` and `ExecuteRunFile`: session gating, execution record creation, console event handling, error finalization, and persistence updates. This duplication already caused behavioral drift (for example REPL emits value events/result payloads while run-file does not) and hides persistence error handling gaps.

The ticket proposes a no-backward-compatibility internal refactor that:

- replaces duplicated executor branches with a single pipeline-oriented lifecycle
- consolidates duplicated core helper behavior into shared single-source semantics

The goal is simpler code, explicit semantics, and robust failure handling.

## Problem Statement

High-impact duplication exists in the executor:

- Session acquisition/status/lock duplication:
  - `pkg/vmexec/executor.go:35`
  - `pkg/vmexec/executor.go:182`
- Execution record creation duplication:
  - `pkg/vmexec/executor.go:53`
  - `pkg/vmexec/executor.go:211`
- Console event-capture duplication:
  - `pkg/vmexec/executor.go:75`
  - `pkg/vmexec/executor.go:234`
- Error finalization duplication:
  - `pkg/vmexec/executor.go:117`
  - `pkg/vmexec/executor.go:265`

Observed robustness gaps:

- Store-write errors are ignored in multiple places (`AddEvent`, `UpdateExecution`) and can silently corrupt execution history.
- REPL and run-file semantics diverge in ways that are not clearly intentional (value event + result payload behavior).
- The functions are too monolithic for focused unit testing.
- Core helper/model duplication remains:
  - `pkg/vmstore/vmstore.go:19` and `pkg/vmcontrol/types.go:48` both implement `mustMarshalJSON` with different signatures/fallback behavior.
  - This increases drift risk and obscures expected JSON fallback semantics.

## Proposed Solution

Introduce an explicit internal execution pipeline abstraction in `pkg/vmexec`, refactor both execution entry points to use it, and unify duplicated helper behavior in shared model/control utilities.

Proposed internal components:

1. `prepareSession(sessionID string) (*vmsession.Session, unlock func(), error)`
2. `newExecutionRecord(kind, sessionID, input/path, args/env)`
3. `eventRecorder` helper:
   - tracks sequence numbers
   - emits typed events
   - returns errors on persistence failure
4. `finalizeExecutionSuccess(...)` and `finalizeExecutionError(...)` helpers:
   - single source of status/result/error update behavior
   - no ignored persistence failures
5. `runExecutionPipeline(...)`:
   - start record
   - wire console capture
   - execute closure
   - finalize outcome
6. shared JSON helper utility:
   - one authoritative `mustMarshalJSON`-style helper
   - explicit fallback semantics
   - reused by vmstore/vmcontrol call sites

Illustrative structure:

```go
type pipelineInput struct {
    sessionID string
    kind      vmmodels.ExecutionKind
    input     string
    path      string
    args      map[string]interface{}
    env       map[string]interface{}
}

func (e *Executor) runPipeline(in pipelineInput, run func(*goja.Runtime, *eventRecorder) (goja.Value, error)) (*vmmodels.Execution, error)
```

No-backward-compatibility scope:

- Internal helper signatures and flow are free to change.
- Behavior alignment between REPL/run-file can be standardized without preserving incidental divergence.
- We do not add compatibility shims for old internals.
- We do not preserve duplicate helper APIs merely for transition convenience.

## Design Decisions

1. Keep API endpoints and payload schema stable unless explicitly changed in this ticket.
Rationale: this ticket targets executor internals; API breakage would increase blast radius.

2. Fail fast on persistence failures.
Rationale: silent failures are worse than explicit errors for execution auditing.

3. Use one sequence/event recorder implementation for all execution kinds.
Rationale: sequence ordering bugs are likely with duplicated event emission logic.

4. Consolidate run-file and REPL completion semantics explicitly.
Rationale: current divergence is likely accidental and increases client confusion.

5. Consolidate duplicated helper behavior instead of keeping parallel utilities.
Rationale: one helper contract prevents silent drift in fallback JSON behavior.

## Alternatives Considered

1. Keep duplication and add comments/tests only.
Rejected: does not reduce maintenance risk or future drift.

2. Minimal extraction of utility snippets only (without lifecycle pipeline).
Rejected: reduces lines but keeps duplicated control flow and hidden divergence.

3. Rewrite executor from scratch with new interfaces.
Rejected for now: too high-risk relative to incremental pipeline extraction.

4. Keep duplicated helper implementations and “document the difference”.
Rejected: accepts ongoing drift and does not improve robustness.

## Implementation Plan

Phase 1: Lock expected behavior and add safety nets

1. Add focused vmexec tests that capture current/target behavior:
   - session readiness and busy lock handling
   - event ordering for console/input/value/exception
   - persisted execution status/result/error fields
2. Decide and document run-file result/value-event contract.

Phase 2: Extract shared lifecycle building blocks

1. Extract session-prep helper (`get+ready+lock`).
2. Extract execution-record builder.
3. Extract event recorder with error-returning write path.
4. Extract finalize helpers for success/error.

Phase 3: Consolidate core helper/model duplication

1. Introduce a shared helper for JSON marshalling fallback semantics.
2. Replace duplicated helper usage in:
   - `pkg/vmstore/vmstore.go`
   - `pkg/vmcontrol/template_service.go` / `pkg/vmcontrol/types.go`
3. Add focused tests for fallback semantics to freeze behavior.

Phase 4: Migrate execution entry points

1. Refactor `ExecuteREPL` to pipeline.
2. Refactor `ExecuteRunFile` to pipeline.
3. Delete duplicated blocks and dead code.

Phase 5: Hardening and validation

1. Add tests for store-write failure paths.
2. Run full regression matrix:
   - `GOWORK=off go test ./... -count=1`
   - `GOWORK=off go test ./pkg/vmtransport/http -count=1`
   - `./smoke-test.sh`
   - `./test-e2e.sh`

Definition of done:

- No duplicated lifecycle blocks between REPL and run-file.
- All persistence write failures in executor are handled explicitly.
- No duplicated `mustMarshalJSON` helper semantics across vmstore/vmcontrol.
- Behavior contract is documented and tested.

## Open Questions

1. Should run-file emit `value` events and set `result_json` symmetrically with REPL, or remain intentionally different?
2. Should console-capture setup/restoration become a shared session-level utility beyond executor scope?
3. Should limit enforcement eventually move into pipeline finalization (cross-ticket with VM-006 findings)?
4. Where should shared JSON fallback helper live (`pkg/vmmodels`, `pkg/vmcontrol`, or a tiny new internal util package)?

## References

- `pkg/vmexec/executor.go`
- `pkg/vmcontrol/execution_service.go`
- `pkg/vmcontrol/types.go`
- `pkg/vmmodels/models.go`
- `pkg/vmstore/vmstore.go`
- `pkg/vmtransport/http/server.go`
- `ttmp/2026/02/08/VM-006-REVIEW-VM-SYSTEM--review-vm-system-implementation-quality-testing-and-setup/design-doc/01-comprehensive-vm-system-implementation-quality-review.md`
