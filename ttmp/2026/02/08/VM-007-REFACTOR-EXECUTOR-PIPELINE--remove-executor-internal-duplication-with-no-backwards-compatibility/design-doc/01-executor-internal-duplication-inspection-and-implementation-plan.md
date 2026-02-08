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
        Persistence operations used by executor flow
        Persistence semantics used by execution lifecycle
    - Path: pkg/vmtransport/http/server.go
      Note: |-
        External API contract for execution endpoints
        Execution API behavior contracts to preserve intentionally
ExternalSources: []
Summary: Deep inspection of vmexec duplication with a concrete no-backward-compatibility implementation plan to consolidate execution lifecycle code into a single pipeline.
LastUpdated: 2026-02-08T12:12:26-05:00
WhatFor: Enable a delegated contributor to execute the executor refactor with clear scope, risks, and step-by-step implementation tasks.
WhenToUse: Use as the primary implementation brief for VM-007 refactor work.
---


# Executor internal duplication inspection and implementation plan

## Executive Summary

`pkg/vmexec/executor.go` currently duplicates most of the execution lifecycle across `ExecuteREPL` and `ExecuteRunFile`: session gating, execution record creation, console event handling, error finalization, and persistence updates. This duplication already caused behavioral drift (for example REPL emits value events/result payloads while run-file does not) and hides persistence error handling gaps.

This ticket proposes a no-backward-compatibility internal refactor that replaces duplicated branches with a single pipeline-oriented execution lifecycle. The goal is simpler code, explicit semantics, and robust failure handling.

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

## Proposed Solution

Introduce an explicit internal execution pipeline abstraction in `pkg/vmexec`, then refactor both execution entry points to use it.

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

## Design Decisions

1. Keep API endpoints and payload schema stable unless explicitly changed in this ticket.
Rationale: this ticket targets executor internals; API breakage would increase blast radius.

2. Fail fast on persistence failures.
Rationale: silent failures are worse than explicit errors for execution auditing.

3. Use one sequence/event recorder implementation for all execution kinds.
Rationale: sequence ordering bugs are likely with duplicated event emission logic.

4. Consolidate run-file and REPL completion semantics explicitly.
Rationale: current divergence is likely accidental and increases client confusion.

## Alternatives Considered

1. Keep duplication and add comments/tests only.
Rejected: does not reduce maintenance risk or future drift.

2. Minimal extraction of utility snippets only (without lifecycle pipeline).
Rejected: reduces lines but keeps duplicated control flow and hidden divergence.

3. Rewrite executor from scratch with new interfaces.
Rejected for now: too high-risk relative to incremental pipeline extraction.

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

Phase 3: Migrate execution entry points

1. Refactor `ExecuteREPL` to pipeline.
2. Refactor `ExecuteRunFile` to pipeline.
3. Delete duplicated blocks and dead code.

Phase 4: Hardening and validation

1. Add tests for store-write failure paths.
2. Run full regression matrix:
   - `GOWORK=off go test ./... -count=1`
   - `GOWORK=off go test ./pkg/vmtransport/http -count=1`
   - `./smoke-test.sh`
   - `./test-e2e.sh`

Definition of done:

- No duplicated lifecycle blocks between REPL and run-file.
- All persistence write failures in executor are handled explicitly.
- Behavior contract is documented and tested.

## Open Questions

1. Should run-file emit `value` events and set `result_json` symmetrically with REPL, or remain intentionally different?
2. Should console-capture setup/restoration become a shared session-level utility beyond executor scope?
3. Should limit enforcement eventually move into pipeline finalization (cross-ticket with VM-006 findings)?

## References

- `pkg/vmexec/executor.go`
- `pkg/vmcontrol/execution_service.go`
- `pkg/vmtransport/http/server.go`
- `ttmp/2026/02/08/VM-006-REVIEW-VM-SYSTEM--review-vm-system-implementation-quality-testing-and-setup/design-doc/01-comprehensive-vm-system-implementation-quality-review.md`
