---
Title: 'Post Mortem Review: VM-007 Executor Pipeline Refactor'
Ticket: VM-007-REFACTOR-EXECUTOR-PIPELINE
Status: active
Topics:
    - backend
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: pkg/vmcontrol/execution_service.go
      Note: Runtime orchestration boundary and limits loading model source cleanup
    - Path: pkg/vmcontrol/template_service.go
      Note: Template default config marshalling path migrated to shared helper
    - Path: pkg/vmcontrol/template_service_test.go
      Note: Config JSON marshalling expectation tests for template defaults
    - Path: pkg/vmexec/executor.go
      Note: |-
        Primary implementation surface for executor dedup and pipeline refactor
        Primary executor pipeline implementation under review
    - Path: pkg/vmexec/executor_persistence_failures_test.go
      Note: |-
        Deterministic persistence failure-path coverage for executor write stages
        Persistence write failure-path deterministic tests
    - Path: pkg/vmexec/executor_test.go
      Note: |-
        Behavioral regression baseline for REPL and run-file execution semantics
        Core vmexec behavioral regression coverage
    - Path: pkg/vmmodels/json_helpers.go
      Note: |-
        Shared marshal fallback helper introduced for helper deduplication
        Shared fallback helper contract
    - Path: pkg/vmmodels/json_helpers_test.go
      Note: Tests for shared helper fallback semantics
    - Path: pkg/vmstore/vmstore.go
      Note: |-
        Store callsite migration to shared helper semantics
        Store callsites migrated to shared helper
    - Path: pkg/vmtransport/http/server_execution_contracts_integration_test.go
      Note: |-
        External API execution contract baseline and parity change coverage
        Execution API contract baseline and parity assertions
    - Path: ttmp/2026/02/08/VM-007-REFACTOR-EXECUTOR-PIPELINE--remove-executor-internal-duplication-with-no-backwards-compatibility/reference/01-diary.md
      Note: |-
        Step-by-step implementation diary used as factual timeline source
        Factual step-by-step implementation timeline
ExternalSources: []
Summary: Detailed engineering post-mortem and full implementation reference map for VM-007 executor pipeline and helper/model dedup refactor.
LastUpdated: 2026-02-08T13:16:00-05:00
WhatFor: Provide a deep retrospective and reference-grade understanding of what changed, why, how risks were managed, and how to navigate the final design.
WhenToUse: Use for engineering review, onboarding, handoff, audit, or future refactor planning around execution flow and core helper/model contracts.
---


# Post Mortem Review: VM-007 Executor Pipeline Refactor

## Executive Summary

VM-007 completed a deep internal refactor that removed high duplication in executor control flow and resolved remaining helper/model duplication around JSON fallback behavior, without backwards-compatibility shim layers. The final result is a shared execution pipeline for both REPL and run-file execution kinds, explicit handling of persistence write failures, unified helper semantics, and stronger regression coverage that locks both external contract and internal failure behavior.

The work was executed in 15 scoped tasks, each with a focused commit and immediate test loop. This reduced refactor risk by never moving too far without green validation and ticket updates. Two important behavior-level outcomes came out of the ticket:

1. Internal robustness improved materially: silent persistence failures for event and execution updates were removed from executor write paths.
2. External behavior was intentionally standardized: successful run-file execution now emits a terminal `value` event and persists `result` payload, aligning with REPL semantics.

The requested validation matrix passed in full after all tasks:

- `GOWORK=off go test ./... -count=1`
- `GOWORK=off go test ./pkg/vmtransport/http -count=1`
- `./smoke-test.sh` (10/10)
- `./test-e2e.sh` (daemon-first end-to-end pass)

The ticket now contains full implementation traceability in `tasks.md`, `changelog.md`, `reference/01-diary.md`, and this post-mortem.

---

## Problem Statement

VM-007 addressed two linked findings:

1. **High executor internal duplication** in `pkg/vmexec/executor.go`:
   - Session retrieval/status/lock acquisition duplicated between REPL and run-file paths.
   - Execution record creation duplicated.
   - Event emission logic duplicated.
   - Success/error finalization duplicated.
   - Duplication already caused semantic drift and made persistence error handling inconsistent.

2. **Core helper/model duplication**, specifically duplicated marshal-with-fallback semantics in `vmstore` and `vmcontrol`:
   - Separate local `mustMarshalJSON` functions with differing signatures and implicit behavior.
   - Drift risk increased because helper behavior was not single-source.
   - Model-source boundaries were blurred by residual alias declarations.

Secondary but critical correctness concern: prior executor code ignored persistence write errors in hot paths (`AddEvent`, `UpdateExecution`). In a system where execution history is part of API surface and auditability, ignored writes create hard-to-diagnose partial-state failure modes.

---

## Proposed Solution (Implemented)

The implemented solution had two tracks that converged:

### Track A: Executor pipeline deduplication

The executor was incrementally refactored into composable lifecycle primitives and then collapsed into a shared pipeline orchestration.

Key internal components:

- `prepareSession(...)` for session get/status/lock
- `newExecutionRecord(...)` for single-source record defaults
- `eventRecorder` for sequence + `AddEvent` writes with explicit failure propagation
- `finalizeExecutionSuccess(...)` / `finalizeExecutionError(...)` for terminal state persistence
- `runExecutionPipeline(...)` for shared orchestration skeleton

Both `ExecuteREPL` and `ExecuteRunFile` now configure this shared pipeline with execution-kind-specific hooks.

### Track B: Helper/model deduplication clean cut

A shared helper was introduced in `vmmodels`:

- `MarshalJSONWithFallback(v interface{}, fallback json.RawMessage) json.RawMessage`

Then callsites were migrated and duplicates removed:

- `vmstore` now calls shared helper directly (string conversion at callsite)
- `vmcontrol/template_service.go` now calls shared helper directly
- local helper duplication removed from `vmcontrol/types.go` and `vmstore`
- residual config alias duplication removed; `vmmodels` is direct source type

No compatibility wrapper layer was retained after explicit user direction.

---

## Part I: Engineering-Focused Retrospective

## Delivery Model and Execution Discipline

The work used strict task slicing:

- One task at a time from ticket `tasks.md`
- Relevant tests after each task
- Ticket updates after each task
- Focused commit per task
- Diary update per task

This approach traded throughput for predictability and auditability. In this ticket, that tradeoff was correct because the change set touched execution semantics, persistence paths, and helper contracts simultaneously.

### What this model prevented

- Hidden coupled changes across unrelated concerns
- Untracked behavior drift during refactor
- Large unreviewable commits where regressions are hard to isolate

### What this model cost

- More doc/commit overhead
- More test cycles and command churn
- More context-switching between code and ticket maintenance

The overhead was acceptable and justified by the risk profile.

---

## Chronological Task Deep-Dive

## Task 1: External contract freeze (HTTP execution surfaces)

### Goal
Establish baseline tests for external behavior before internal refactor.

### Implementation
Added `pkg/vmtransport/http/server_execution_contracts_integration_test.go` to lock:

- response status codes for execution create/get/list/events happy paths
- execution envelope shape
- event envelope shape and sequence behavior
- list/get semantics
- `after_seq` filtering behavior

### Why this mattered
Internal refactors were expected to touch all execution internals. Without a contract baseline, accidental API behavior drift could pass unnoticed until late.

### Outcome
Successful baseline lock. This file later became the explicit place where Task 9 behavior change was intentionally updated.

---

## Task 2: vmexec regression safety net

### Goal
Freeze current executor behavior at package level.

### Implementation
Added `pkg/vmexec/executor_test.go` with focused tests:

- REPL success event order and result persistence
- REPL error exception persistence
- run-file current parity behavior (at that time console-only + empty result)

### Why this mattered
HTTP tests cover contract edges; vmexec tests cover internal behavior details and ordering expectations directly.

### Outcome
Provided a precise internal baseline that was updated intentionally in Task 9 when run-file parity changed.

---

## Task 3: Shared session-preparation helper

### Goal
Remove duplicate `GetSession + status + lock` code.

### Implementation
Added `prepareSession(...)` and migrated both execution entrypoints to use it.

### Outcome
Small but high-leverage simplification; reduced repeated failure-path code and clarified lock lifecycle with returned unlock closure.

---

## Task 4: Shared execution-record constructor

### Goal
Single-source execution record defaults and creation fields.

### Implementation
Introduced `executionRecordInput` + `newExecutionRecord(...)`:

- shared ID generation
- status defaults
- timestamps
- metrics defaults
- args/env default normalization

### Outcome
Removed constructor duplication and made future behavior adjustments easier to centralize.

---

## Task 5: Shared event recorder + explicit AddEvent errors

### Goal
Stop silently ignoring event write failures; remove duplicated event emit code.

### Implementation
Introduced `eventRecorder` with:

- sequence management
- typed/raw emit methods
- first-error capture for callback-driven console writes

All direct `AddEvent` calls in REPL/run-file paths were replaced.

### Notable engineering detail
Console callbacks cannot return errors into calling control flow directly. The recorder captured first callback-side error and the main path checked `recorder.Err()` after runtime execution.

### Outcome
Event persistence failures are now explicit instead of silently ignored.

---

## Task 6: Shared finalize helpers + explicit UpdateExecution errors

### Goal
Remove duplicated finalization logic and stop ignored update writes.

### Implementation
Added:

- `finalizeExecutionSuccess(...)`
- `finalizeExecutionError(...)`

Both REPL and run-file paths now finalize through these helpers.

### Outcome
No ignored `UpdateExecution` in executor success/error completion paths.

---

## Task 7: Pipeline migration for ExecuteREPL

### Goal
Move REPL entrypoint to shared pipeline chain.

### Implementation
Added:

- `executionPipelineConfig`
- `runExecutionPipeline(...)`

Migrated `ExecuteREPL` to configure hooks:

- runtime setup
- run closure
- success handler
- error handler

### Outcome
REPL lifecycle now runs through one orchestration skeleton.

---

## Task 8: Pipeline migration for ExecuteRunFile

### Goal
Complete lifecycle dedup by moving run-file to same pipeline.

### Implementation
Migrated `ExecuteRunFile` to pipeline hooks. Extracted helper functions:

- `installConsoleRecorder(...)`
- `exceptionPayloadJSON(...)`

### Outcome
Primary duplication finding was resolved: both entrypoints share one lifecycle spine.

---

## Task 9: Explicit contract decision (run-file value/result parity)

### Goal
Resolve intentional behavior for run-file terminal outputs.

### Decision
Successful run-file executions now:

- emit terminal `value` event
- persist `result` payload

Aligned to REPL value semantics.

### Implementation
- Added `valuePayloadJSON(...)` helper
- Updated run-file success handler
- Updated vmexec and HTTP contract tests
- Recorded decision in design doc decision log

### Outcome
Behavior divergence removed intentionally and documented.

---

## Task 10: Shared helper introduced

### Goal
Create one shared marshal-fallback helper contract.

### Implementation
Added `vmmodels.MarshalJSONWithFallback(...)`.

### Outcome
Single helper baseline established before callsite migration.

---

## Task 11: Clean-cut callsite migration (no wrappers)

### Goal
Remove duplicate local helper semantics and migrate callsites.

### Implementation
- migrated vmstore + vmcontrol callsites to shared helper
- removed local helpers
- removed wrapper-style extra helper after explicit user direction

### Outcome
One helper API remains; no compatibility wrapper layer.

---

## Task 12: Model-source boundary cleanup

### Goal
Ensure `vmmodels` is direct source of config model types.

### Implementation
- removed residual aliases in `vmcontrol/types.go`
- switched `ExecutionService.loadSessionLimits` to direct `*vmmodels.LimitsConfig`

### Outcome
Residual model-boundary duplication removed.

---

## Task 13: Focused helper/model behavior tests

### Goal
Lock shared helper fallback behavior and config marshalling expectations.

### Implementation
- `pkg/vmmodels/json_helpers_test.go`
- `pkg/vmcontrol/template_service_test.go`

### Outcome
Helper fallback and template default config serialization are now explicitly tested.

---

## Task 14: Persistence failure-path tests

### Goal
Guarantee deterministic outcomes for create/add-event/update failure stages.

### Implementation
- Added `pkg/vmexec/executor_persistence_failures_test.go`
- Introduced internal `executionStore` interface in executor for deterministic failure injection

### Engineering issue encountered
Initial test fixture imported `vmcontrol`, which created package import cycle for internal vmexec tests. Resolved by direct VM/settings setup against store in test fixture.

### Outcome
Failure-path behavior is now test-locked and deterministic.

---

## Task 15: Full validation matrix

### Goal
Run requested final validation set and record evidence.

### Validation results
- `go test ./...`: pass
- `go test ./pkg/vmtransport/http`: pass
- `./smoke-test.sh`: pass (10/10)
- `./test-e2e.sh`: pass

### Outcome
End-to-end evidence confirms refactor stability against requested matrix.

---

## What Went Well

## 1) Refactor decomposition quality

Breaking executor refactor into lifecycle primitives before pipeline migration worked well. It reduced cognitive load and made each task reviewable.

## 2) Contract-first sequencing

Freezing external and package-level behavior early gave a stable baseline and prevented accidental drift.

## 3) Task and commit hygiene

One task per commit, with tests and docs each time, made the change history coherent and easy to audit.

## 4) Explicit behavior decision handling

Task 9 was handled as a deliberate contract change with updated tests/docs, not as accidental side-effect.

## 5) Clean-cut helper migration

After explicit user direction, wrappers were removed and callsites migrated directly. This produced a simpler final state.

## 6) Failure-path rigor

Task 14 converted persistence write error handling from best effort to deterministic and test-backed behavior.

---

## What Was Hard / Where Friction Happened

## 1) Context overhead from strict process

The requirement to update tasks/changelog/diary and commit at each step is correct for auditability but adds significant process overhead. The team needs discipline to maintain momentum.

## 2) Runtime callback error propagation

Console callback writes happen outside direct function return channels. Capturing and propagating these errors in deterministic order required careful recorder design.

## 3) Test fixture dependency traps

While adding failure-injection tests, using higher-level service imports inside package tests caused an import cycle. The solution was to use lower-level direct setup.

## 4) Mid-flight user-direction change on wrappers

The user clarified a strict clean-cut stance while Task 11 was in progress. The implementation had to remove wrapper direction immediately and consolidate to one API.

## 5) Working tree noise

Script/test artifacts and unrelated paths appeared during execution. Scoped staging had to remain disciplined to avoid leaking unrelated files into ticket commits.

---

## Mistakes and Corrections

## Mistake: Wrapper helper temporarily introduced

A string-return helper wrapper was added briefly when introducing shared helper logic. User explicitly requested no wrappers/compat layer.

### Correction
Wrapper removed; callsites converted explicitly using `string(vmmodels.MarshalJSONWithFallback(...))`.

## Mistake: Import cycle in Task 14 test fixture

Test imported `vmcontrol` from `pkg/vmexec` package tests and hit `import cycle not allowed`.

### Correction
Replaced with direct VM/settings creation using store model calls in fixture.

## Mistake: Process interruption handling

Turn interruption occurred mid-progress.

### Correction
State re-verified (`git status`, task list), then resumed from last consistent checkpoint.

---

## Risk Management Review

## Primary risks at ticket start

1. Behavior drift while deduplicating executor control flow
2. Hidden persistence failure regressions
3. Incomplete helper migration leaving semantic duality
4. Contract ambiguity on run-file terminal output

## Mitigations applied

- Baseline contract tests early (Task 1/2)
- Narrow, frequent test loops per task
- Behavior decision captured explicitly (Task 9)
- Failure-path tests added (Task 14)
- Full matrix run at end (Task 15)

## Remaining residual risks

- Consumer assumptions around newly standardized run-file `result` payload may need extra rollout communication (behavior intentionally changed).
- Future feature work may bypass pipeline hooks if not disciplined; architectural conventions should be documented in contributor guidelines.

---

## Part II: Full Implementation Reference Overview

This section is intended as a deep navigation guide for maintainers.

## System-Level Flow (Execution Request to Persistence)

### Before VM-007 (conceptually)

- REPL and run-file had mostly separate lifecycle code.
- Persistence write failures could be silently dropped.
- Terminal success semantics diverged by execution kind.

### After VM-007

1. Caller enters through `vmcontrol.ExecutionService`
2. Service enforces path policies and limit checks
3. Runtime delegation calls `vmexec.Executor`
4. Executor runs one shared pipeline skeleton
5. Pipeline emits events via recorder + finalizes execution with explicit write checks
6. Store persists execution/event artifacts
7. API layer returns envelopes from persisted runtime/store state

---

## File-by-File Reference (Core Runtime)

## `pkg/vmexec/executor.go`

### `type Executor`

Responsibility: execution runtime orchestration against active sessions and persistence.

Fields:

- `store executionStore`: internal persistence surface used by pipeline
- `sessionManager *vmsession.SessionManager`: active runtime session access

### `type executionStore` (internal)

Purpose: internal interface capturing executor-required persistence methods.

Why it exists:

- Enables deterministic failure injection in Task 14 tests.
- Avoids altering public constructor/API surface.

### `type executionRecordInput`

Purpose: single parameter object for execution record construction.

Key fields:

- session ID
- execution kind
- optional input/path
- args/env JSON payloads

### `type eventRecorder`

Purpose: centralized event emission primitive.

Behavior:

- tracks sequence (`nextSeq`)
- builds event rows
- writes via `AddEvent`
- records first callback-side error for later propagation

Methods:

- `emit(...)`: typed payload marshal + write
- `emitRaw(...)`: pre-marshaled payload write
- `recordError(...)`: callback error capture
- `Err()`: retrieve first captured error

### `type executionPipelineConfig`

Purpose: callback-based configuration for execution-kind-specific logic.

Hooks:

- `setupRuntime`
- `run`
- `handleSuccess`
- `handleError`

### `prepareSession(...)`

Purpose: common get/status/lock path.

Returns:

- active session
- unlock closure
- error

### `newExecutionRecord(...)`

Purpose: single-source execution row default creation.

Defaults include:

- generated ID
- `status=running`
- `started_at`
- `metrics={}`
- args/env fallback defaults

### `finalizeExecutionSuccess(...)` / `finalizeExecutionError(...)`

Purpose: terminal status/payload persistence with explicit `UpdateExecution` error propagation.

### `installConsoleRecorder(...)`

Purpose: installs console shim that emits console events through recorder.

### `exceptionPayloadJSON(...)`

Purpose: maps runtime error into standardized exception payload JSON.

### `valuePayloadJSON(...)`

Purpose: maps runtime value into standardized value payload JSON.

### `runExecutionPipeline(...)`

Purpose: shared lifecycle skeleton across execution kinds.

Sequence:

1. prepare session + lock
2. create execution record
3. setup runtime hook
4. run hook
5. propagate callback-side event errors
6. invoke success/error finalize hooks

### `ExecuteREPL(...)`

Usage of pipeline:

- setup: install console + input echo event
- run: `RunString(input)`
- success: emit value event + persist result
- error: emit exception event + persist error

### `ExecuteRunFile(...)`

Usage of pipeline:

- setup: resolve/read file, install console, set `__ARGS__`
- run: `RunString(fileContent)`
- success: emit terminal value event + persist result (Task 9)
- error: emit exception event + persist error

### `GetExecution`, `GetEvents`, `ListExecutions`

Thin forwarding methods to store read surfaces.

---

## File-by-File Reference (Control Layer)

## `pkg/vmcontrol/execution_service.go`

### `type ExecutionService`

Responsibilities:

- inbound orchestration for REPL/run-file
- path normalization/security checks for run-file
- post-execution limit enforcement scaffolding

Important methods:

- `ExecuteREPL(...)`: delegates to runtime then enforces limits
- `ExecuteRunFile(...)`: path normalization + runtime delegation + limit enforcement
- `normalizeRunFilePath(...)`: traversal-safe relative path normalization
- `loadSessionLimits(...)`: now uses `vmmodels.LimitsConfig` directly (Task 12 cleanup)

Relation to VM-007:

- Not the primary dedup target, but model-boundary cleanup touched this file.

## `pkg/vmcontrol/template_service.go`

### `type TemplateService`

Responsibility: template CRUD/settings policy defaults.

VM-007-relevant changes:

- default settings marshalling now uses shared `vmmodels.MarshalJSONWithFallback`.

## `pkg/vmcontrol/types.go`

Purpose: public input types for control layer.

VM-007 changes:

- removed config alias declarations to eliminate residual model duplication boundary.

---

## File-by-File Reference (Model + Store)

## `pkg/vmmodels/json_helpers.go`

### `MarshalJSONWithFallback(...)`

Single-source helper contract for marshal-with-fallback behavior.

Semantics:

- marshal success -> marshaled bytes
- marshal failure + fallback provided -> fallback bytes
- marshal failure + no fallback -> JSON `null`

Used by:

- template default settings marshalling
- vmstore VM profile JSON columns

## `pkg/vmstore/vmstore.go`

### `type VMStore`

Responsibility: SQLite persistence for templates, sessions, executions, and events.

VM-007-relevant changes:

- callsites for VM profile JSON fields use shared helper directly
- local duplicate helper removed

Key execution persistence methods used by executor pipeline:

- `CreateExecution`
- `UpdateExecution`
- `AddEvent`
- `GetExecution`
- `GetEvents`
- `ListExecutions`

---

## File-by-File Reference (Tests)

## `pkg/vmexec/executor_test.go`

Purpose: core behavior regression coverage.

Tests:

- REPL success ordering/result
- REPL error exception persistence
- run-file success parity (now includes value/result)

## `pkg/vmexec/executor_persistence_failures_test.go`

Purpose: deterministic failure-path tests for write stages.

Coverage:

- create execution failure
- add event failure
- update execution failure

Mechanism:

- injected failing store implementing internal `executionStore`

## `pkg/vmtransport/http/server_execution_contracts_integration_test.go`

Purpose: API execution contract baseline.

Covers:

- status/envelope expectations
- list/get semantics
- event shape/sequence
- run-file value/result contract parity

## `pkg/vmmodels/json_helpers_test.go`

Purpose: lock helper fallback semantics.

Covers:

- success marshal
- fallback on marshal failure
- null behavior on empty fallback

## `pkg/vmcontrol/template_service_test.go`

Purpose: lock default config JSON marshalling expectations.

Covers:

- default limits/resolver/runtime payloads from `TemplateService.Create`

---

## Execution Pipeline Reference (How Things Are Used)

## REPL Flow

1. API receives `POST /executions/repl`
2. `ExecutionService.ExecuteREPL` delegates to runtime executor
3. `Executor.ExecuteREPL` configures pipeline
4. pipeline prepares session + lock
5. execution record persisted (`running`)
6. input echo event emitted
7. runtime executes input
8. success: value event emitted + execution persisted `ok` with result
9. error: exception event emitted + execution persisted `error`

## Run-File Flow

1. API receives `POST /executions/run-file`
2. `ExecutionService.ExecuteRunFile` normalizes path, delegates
3. `Executor.ExecuteRunFile` configures pipeline
4. pipeline prepares session + lock
5. execution record persisted (`running`)
6. setup hook resolves/reads file, installs console, sets `__ARGS__`
7. runtime executes content
8. success: value event emitted + execution persisted `ok` with result
9. error: exception event emitted + execution persisted `error`

## Persistence Error Behavior

At any write stage failure:

- create failure -> returned error (`failed to create execution`)
- add-event failure -> returned error (`failed to persist event ...`)
- update failure -> returned error (`failed to persist ... execution ...`)

This is explicitly tested in Task 14.

---

## Data Contract Summary

## Execution row contract (selected fields)

- `kind`: `repl` or `run_file`
- `status`: `running` -> terminal `ok` or `error`
- `result`: value payload for successful REPL and run-file
- `error`: exception payload for failed execution

## Event contract (selected)

- `seq`: monotonic per execution
- `type`: `input_echo`, `console`, `value`, `exception`
- `payload`: type-specific JSON payload

## Value payload contract

- `type`: exported runtime type name
- `preview`: stringified value
- `json`: optional marshaled exported value

---

## Commit/Change Stream Reference

Task commits (oldest to newest in VM-007 implementation stream):

1. `760b3a9` task 1 contract baseline
2. `b9db8c4` task 2 vmexec regression tests
3. `8f4f39a` task 3 session prep helper
4. `ea3cca9` task 4 execution record helper
5. `835de10` task 5 event recorder + add-event errors
6. `cb9a10c` task 6 finalize helpers + update errors
7. `effc6cc` task 7 REPL pipeline migration
8. `233d194` task 8 run-file pipeline migration
9. `c5b2e11` task 9 run-file value/result parity decision
10. `7efa95a` task 10 shared helper introduction
11. `b927b5d` task 11 callsite migration clean cut
12. `c9ba191` task 12 model alias cleanup
13. `435ca32` task 13 helper/config tests
14. `166ba68` task 14 persistence failure tests
15. `e277521` task 15 full matrix evidence/docs

---

## Full-Picture Mental Model (Concise)

If you want one operational model to keep in your head:

- **Control layer** validates inputs and policy boundaries.
- **Executor layer** now has one lifecycle engine (pipeline), execution-kind-specific hooks, and explicit persistence failure semantics.
- **Store layer** is authoritative persistence backend; writes are no longer silently ignored by executor.
- **Model/helper layer** has one marshal-fallback helper and one source for config types.
- **Tests** lock both behavior and failure semantics at package and HTTP integration levels.

This is the key VM-007 architecture outcome.

---

## Open Questions (Reviewer Decisions)

1. Confirm acceptance of Task 9 intentional behavior change:
   - successful run-file now emits terminal `value` event and returns/persists `result` payload parity with REPL.

2. Confirm whether to close ticket now that all tasks are complete and matrix is green:
   - suggested command: `docmgr ticket close --ticket VM-007-REFACTOR-EXECUTOR-PIPELINE`

3. Confirm whether to add a short contributor guideline note in repo docs describing executor pipeline conventions to keep future changes aligned.

---

## References

- `pkg/vmexec/executor.go`
- `pkg/vmexec/executor_test.go`
- `pkg/vmexec/executor_persistence_failures_test.go`
- `pkg/vmtransport/http/server_execution_contracts_integration_test.go`
- `pkg/vmmodels/json_helpers.go`
- `pkg/vmmodels/json_helpers_test.go`
- `pkg/vmcontrol/template_service.go`
- `pkg/vmcontrol/template_service_test.go`
- `pkg/vmcontrol/execution_service.go`
- `pkg/vmstore/vmstore.go`
- `ttmp/2026/02/08/VM-007-REFACTOR-EXECUTOR-PIPELINE--remove-executor-internal-duplication-with-no-backwards-compatibility/reference/01-diary.md`
- `ttmp/2026/02/08/VM-007-REFACTOR-EXECUTOR-PIPELINE--remove-executor-internal-duplication-with-no-backwards-compatibility/changelog.md`
- `ttmp/2026/02/08/VM-007-REFACTOR-EXECUTOR-PIPELINE--remove-executor-internal-duplication-with-no-backwards-compatibility/tasks.md`
- `ttmp/2026/02/08/VM-007-REFACTOR-EXECUTOR-PIPELINE--remove-executor-internal-duplication-with-no-backwards-compatibility/design-doc/01-executor-internal-duplication-inspection-and-implementation-plan.md`

---

## Deep Technical Addendum: How the Pieces Interact in Practice

This addendum is intentionally dense. It is meant for engineers who need to debug, extend, or review the implementation under real conditions.

## A. Call-Graph Walkthroughs (Concrete)

## A.1 REPL Happy Path (Concrete Function-Level Sequence)

1. HTTP handler `handleExecutionREPL` parses JSON payload and validates `session_id`.
2. Handler calls `core.Executions.ExecuteREPL(...)` in control layer.
3. `ExecutionService.ExecuteREPL(...)` delegates to runtime executor.
4. `Executor.ExecuteREPL(...)` builds `executionPipelineConfig` and calls `runExecutionPipeline(...)`.
5. Pipeline calls `prepareSession(...)`:
   - `sessionManager.GetSession`
   - status check (`ready`)
   - lock acquisition (`TryLock`)
6. Pipeline creates execution row via `newExecutionRecord(...)` and persists with `CreateExecution`.
7. Setup hook:
   - installs console recorder in runtime
   - emits input echo event via recorder
8. Run hook executes `session.Runtime.RunString(input)`.
9. Post-run recorder error check executes (`recorder.Err()`).
10. Success hook:
   - builds value payload (`valuePayloadJSON`)
   - emits value event
   - finalizes success (`finalizeExecutionSuccess` + `UpdateExecution`)
11. Pipeline returns execution object.
12. Control layer runs limit checks (`enforceLimits`) and returns.
13. HTTP handler writes `201` with execution payload.

## A.2 REPL Error Path

Differences from happy path:

- Run hook returns error.
- Error hook builds exception payload (`exceptionPayloadJSON`), emits exception event, finalizes error (`finalizeExecutionError`).
- Execution object returns with terminal `status=error` if persistence succeeded.
- If persistence write fails during error finalization, explicit error is returned and mapped at API layer.

## A.3 Run-File Happy Path

Differences from REPL path:

- Setup hook performs file existence/read check under session worktree.
- Setup hook sets `__ARGS__` global in runtime.
- Run hook executes file content string.
- Success hook now mirrors REPL terminal behavior (Task 9): emits value event and persists result payload.

## A.4 Run-File Error Path

Matches REPL error treatment pattern after runtime execution:

- exception payload event
- explicit finalize error update
- deterministic error if persistence write fails

## A.5 Failure-Path Ordering Guarantees

For every execution kind:

- `CreateExecution` occurs before any event write.
- Event writes happen in deterministic sequence order for that execution.
- Terminal `UpdateExecution` occurs after terminal event emission attempt.
- First callback-side recorder error is surfaced on main path after run closure.

This ordering matters for debugging and API consumer expectations.

---

## B. Invariants Introduced or Strengthened

These invariants should be treated as design-level contracts for future changes:

1. **Single lifecycle skeleton invariant**
   - Both execution kinds use `runExecutionPipeline(...)`.
   - If a future execution kind is added, it should use the same pipeline.

2. **No silent persistence write invariant**
   - `CreateExecution`, `AddEvent`, and `UpdateExecution` failures are explicit.
   - Returning success while dropping writes is not allowed in executor path.

3. **Monotonic event sequence invariant**
   - All persisted events for an execution use recorder-managed sequence progression.

4. **Terminal event + row consistency invariant**
   - On success, terminal value event and execution result payload are consistent.
   - On failure, exception event and execution error payload are consistent.

5. **Model/source-of-truth invariant**
   - Core config model types come directly from `vmmodels`.
   - No boundary alias duplicates in `vmcontrol`.

6. **Single shared helper invariant**
   - Marshal fallback semantics are centralized in `vmmodels.MarshalJSONWithFallback`.

---

## C. Failure Taxonomy and Expected Behavior

## C.1 CreateExecution Fails

Expected behavior:

- Executor returns error with context (`failed to create execution`).
- No events should be emitted/persisted for that attempt.
- Caller sees transport-level error path.

Coverage:

- `TestExecuteREPLCreateExecutionFailureReturnsDeterministicError`

## C.2 AddEvent Fails (Input/Console/Value/Exception)

Expected behavior:

- Executor returns deterministic wrapped error (`failed to persist event ...`).
- Execution may exist in `running` state depending on stage of failure.
- Caller receives explicit failure signal.

Coverage:

- `TestExecuteREPLAddEventFailureReturnsDeterministicError`

## C.3 UpdateExecution Fails in Finalization

Expected behavior:

- Executor returns deterministic wrapped error (`failed to persist successful execution ...` or failed execution variant).
- Terminal row state may remain stale due to write failure.
- Caller receives explicit failure signal.

Coverage:

- `TestExecuteREPLUpdateExecutionFailureReturnsDeterministicError`

## C.4 Runtime Script Error (Not Persistence Error)

Expected behavior:

- Exception event emitted
- execution row finalized with `status=error`
- function returns `exec, nil` when persistence succeeds

This preserves existing high-level runtime error contract while still enforcing write-path explicitness.

---

## D. Interface and Abstraction Decisions

## D.1 Why `executionStore` is Internal (Not Public API)

The internal interface exists only to:

- isolate executor-required persistence behavior
- support deterministic failure-injection tests

It is not exposed as a new public package API or compatibility layer. Constructor surface remains unchanged (`NewExecutor(*vmstore.VMStore, ...)`), so this is an internal testing-enabling design choice, not an outward abstraction promise.

## D.2 Why Pipeline Uses Callback Hooks Instead of Kind Switches

Benefits:

- keeps lifecycle stages linear and explicit
- contains kind-specific behavior to small functions
- avoids giant switch blocks with repeated scaffolding
- easier to add execution kinds later without cloning lifecycle code

Tradeoff:

- callback boundaries can obscure control flow if overused

Mitigation:

- keep config struct small
- maintain one clear call chain in `runExecutionPipeline`

## D.3 Why Fallback Helper Lives in `vmmodels`

Chosen because:

- both control and store layers use it
- helper semantics map directly to model JSON representation concerns
- avoids creating another utility package solely for one helper

---

## E. Review Playbook (If You Need to Re-Audit Quickly)

Use this order for high-signal review:

1. `pkg/vmexec/executor.go`
   - verify no duplicated lifecycle branches
   - verify explicit write error propagation
   - verify run-file success parity behavior

2. `pkg/vmexec/executor_persistence_failures_test.go`
   - verify deterministic failure expectations

3. `pkg/vmtransport/http/server_execution_contracts_integration_test.go`
   - verify external contract coverage and run-file parity assertions

4. `pkg/vmmodels/json_helpers.go` + `pkg/vmmodels/json_helpers_test.go`
   - verify single helper semantics and fallback behavior

5. `pkg/vmcontrol/template_service.go` + `pkg/vmcontrol/template_service_test.go`
   - verify config marshalling default expectations

6. `pkg/vmstore/vmstore.go`
   - verify migrated callsites, no local helper duplicates

7. `pkg/vmcontrol/types.go` + `pkg/vmcontrol/execution_service.go`
   - verify model-source cleanup completion

Then run full matrix.

---

## F. “How Do I Extend This?” Guidance

## F.1 Add New Event Type

Suggested process:

1. Add payload struct in `vmmodels` if needed.
2. Emit through recorder (`emit` or `emitRaw`) in appropriate hook.
3. Update vmexec regression tests for event ordering.
4. Update HTTP contract test if externally visible expectations change.
5. Document behavior in changelog/decision log if contract-affecting.

## F.2 Add New Execution Kind

Suggested process:

1. Add kind constant in `vmmodels`.
2. Implement new public executor method or route existing method to new config.
3. Configure `executionPipelineConfig` hooks (setup/run/success/error).
4. Add focused vmexec tests and HTTP contract tests.
5. Ensure limits and path/security boundaries are enforced at control layer as needed.

## F.3 Change Result Payload Semantics

Suggested process:

1. Update value payload helper logic.
2. Update vmexec tests.
3. Update HTTP contract tests.
4. Record decision in design doc decision log.
5. Mention behavior change in changelog + diary.

---

## G. Debugging Runbook

## G.1 Symptom: Missing Events in API

Check:

1. `GetEvents` query with `after_seq=0`
2. executor logs/errors for `failed to persist event`
3. whether failure-path tests mirror observed stage

Likely causes:

- store write failure at add-event stage
- callback-side recorder error surfaced post-run

## G.2 Symptom: Execution Stuck Running

Check:

1. whether `UpdateExecution` failed in finalize helper
2. returned error path in caller/logs
3. DB row for execution status/ended_at fields

Likely cause:

- finalize update persistence failure

## G.3 Symptom: Run-file returns unexpected result behavior

Check:

1. Task 9 contract expectations in HTTP contract test
2. `ExecuteRunFile` success hook in executor pipeline
3. value payload construction helper output

---

## H. Quantitative Change Summary

## H.1 Architectural Simplification

- Two large monolithic entrypoint flows reduced to one shared pipeline skeleton + hooks.
- Common lifecycle logic (session prep, record creation, event writing, finalization) centralized.

## H.2 Behavioral Clarification

- Run-file success terminal semantics now explicitly aligned to REPL.
- Helper fallback behavior now centralized and directly tested.

## H.3 Test Coverage Improvements (Qualitative)

- Added explicit tests for:
  - vmexec event ordering and terminal persistence behavior
  - run-file contract parity behavior
  - deterministic write failure modes
  - helper fallback semantics
  - template default config JSON marshalling expectations

---

## I. Process Retrospective (Meta)

## I.1 What to Keep for Future Tickets

1. Task slicing + commit slicing model for high-risk refactors.
2. Contract freeze before deep internal rewrites.
3. Immediate per-task ticket documentation.
4. Dedicated failure-path test task near end (before full matrix).

## I.2 What to Improve Next Time

1. Add a lightweight checklist template for per-task diary updates to reduce writing overhead.
2. Pre-define behavior-decision windows (like Task 9) to reduce mid-ticket ambiguity.
3. Add one small contributor note in repo docs explaining executor pipeline extension conventions.

---

## J. Complete File Reference Index

For fast lookup, grouped by role:

### Runtime / Execution Core

- `pkg/vmexec/executor.go`
- `pkg/vmexec/executor_test.go`
- `pkg/vmexec/executor_persistence_failures_test.go`

### Control / Orchestration

- `pkg/vmcontrol/execution_service.go`
- `pkg/vmcontrol/template_service.go`
- `pkg/vmcontrol/template_service_test.go`
- `pkg/vmcontrol/types.go`

### Models / Shared Helper

- `pkg/vmmodels/models.go`
- `pkg/vmmodels/json_helpers.go`
- `pkg/vmmodels/json_helpers_test.go`

### Persistence

- `pkg/vmstore/vmstore.go`

### HTTP Contract Layer

- `pkg/vmtransport/http/server.go`
- `pkg/vmtransport/http/server_execution_contracts_integration_test.go`

### Ticket Artifacts

- `ttmp/.../tasks.md`
- `ttmp/.../changelog.md`
- `ttmp/.../reference/01-diary.md`
- `ttmp/.../design-doc/01-executor-internal-duplication-inspection-and-implementation-plan.md`
- `ttmp/.../design-doc/02-post-mortem-review-vm-007-executor-pipeline-refactor.md`

---

## K. Closing Statement

VM-007 succeeded not because the refactor was small, but because it was controlled. The critical engineering move was to separate structural deduplication from behavior decisions, then make behavior changes explicit, test-backed, and documented.

Final architecture characteristics after VM-007:

- one shared executor lifecycle pipeline
- explicit persistence write failure handling
- one shared helper for JSON fallback semantics
- one source of truth for config model types
- multi-layer regression coverage that includes both happy and failure paths

This is a solid base for future execution features and reliability work.
