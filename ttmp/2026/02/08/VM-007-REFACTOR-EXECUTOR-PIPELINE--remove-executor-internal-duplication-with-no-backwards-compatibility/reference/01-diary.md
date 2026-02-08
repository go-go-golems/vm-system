---
Title: Diary
Ticket: VM-007-REFACTOR-EXECUTOR-PIPELINE
Status: active
Topics:
    - backend
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: pkg/vmcontrol/execution_service.go
      Note: Task 12 direct vmmodels.LimitsConfig usage
    - Path: pkg/vmcontrol/template_service.go
      Note: Task 11 vmcontrol template settings migration to shared helper
    - Path: pkg/vmcontrol/template_service_test.go
      Note: Task 13 config JSON marshalling expectation tests
    - Path: pkg/vmcontrol/types.go
      Note: |-
        Task 11 removed duplicated local helper
        Task 12 removed residual config aliases
    - Path: pkg/vmexec/executor.go
      Note: |-
        Task 3 shared session preparation helper extraction
        Task 4 shared execution record constructor
        Task 5 shared event recorder and AddEvent error propagation
        Task 6 finalize helpers and explicit UpdateExecution error handling
        Task 7 pipeline helper chain and ExecuteREPL migration
        Task 8 ExecuteRunFile pipeline migration and shared helper extraction
        Task 9 run-file value/result parity behavior change
        Task 14 internal executionStore interface for deterministic failure injection
    - Path: pkg/vmexec/executor_persistence_failures_test.go
      Note: Task 14 deterministic persistence failure-path tests
    - Path: pkg/vmexec/executor_test.go
      Note: |-
        Task 2 vmexec regression test coverage
        Task 9 updated vmexec regression expectations
    - Path: pkg/vmmodels/json_helpers.go
      Note: |-
        Task 10 shared marshal-with-fallback helper
        Task 11 removed wrapper layer; single helper API
    - Path: pkg/vmmodels/json_helpers_test.go
      Note: Task 13 shared helper fallback tests
    - Path: pkg/vmstore/vmstore.go
      Note: Task 11 vmstore migration to single shared JSON helper
    - Path: pkg/vmtransport/http/server_execution_contracts_integration_test.go
      Note: |-
        Task 1 baseline contract coverage for execution endpoints and event envelopes
        Task 1 external contract baseline test
        Task 9 updated execution API behavior contract assertions
    - Path: ttmp/2026/02/08/VM-007-REFACTOR-EXECUTOR-PIPELINE--remove-executor-internal-duplication-with-no-backwards-compatibility/changelog.md
      Note: |-
        Task-level changelog entries
        Task 1 changelog entry
        Task 2 changelog entry
        Task 3 changelog entry
        Task 4 changelog entry
        Task 5 changelog entry
        Task 6 changelog entry
        Task 7 changelog entry
        Task 8 changelog entry
        Task 9 changelog entry
        Task 10 changelog entry
        Task 11 changelog entry
        Task 12 changelog entry
        Task 13 changelog entry
        Task 14 changelog entry
    - Path: ttmp/2026/02/08/VM-007-REFACTOR-EXECUTOR-PIPELINE--remove-executor-internal-duplication-with-no-backwards-compatibility/design-doc/01-executor-internal-duplication-inspection-and-implementation-plan.md
      Note: Task 9 decision log for run-file value/result contract
    - Path: ttmp/2026/02/08/VM-007-REFACTOR-EXECUTOR-PIPELINE--remove-executor-internal-duplication-with-no-backwards-compatibility/tasks.md
      Note: |-
        Task checklist updated after Task 1 completion
        Task 1 checklist update
        Task 2 checklist update
        Task 3 checklist update
        Task 4 checklist update
        Task 5 checklist update
        Task 6 checklist update
        Task 7 checklist update
        Task 8 checklist update
        Task 9 checklist update
        Task 10 checklist update
        Task 11 checklist update
        Task 12 checklist update
        Task 13 checklist update
        Task 14 checklist update
ExternalSources: []
Summary: Implementation diary for VM-007 executor/core dedup refactor, recorded per completed task.
LastUpdated: 2026-02-08T12:31:00-05:00
WhatFor: Preserve exact implementation steps, tests, decisions, and follow-ups for VM-007.
WhenToUse: Use when reviewing VM-007 task execution and validating refactor decisions.
---















# Diary

## Goal

This diary captures VM-007 implementation progress task-by-task, including concrete code changes, validation commands, failures, and reviewer checkpoints.

## Step 1: Freeze external execution API contracts before internal refactor

I started by establishing explicit baseline tests for execution endpoint contracts so the internal executor refactor can proceed without accidental API drift. The focus was status codes, execution response envelopes, event envelope shape, and list/get semantics.

The key intent in this step was to lock behavior before changing executor internals. That gives a stable reference point for later tasks where behavior may intentionally change (and must then be documented explicitly).

### Prompt Context

**User prompt (verbatim):** "Ticket:
  - VM-007-REFACTOR-EXECUTOR-PIPELINE
  - Path: ttmp/2026/02/08/VM-007-REFACTOR-EXECUTOR-PIPELINE--remove-executor-internal-duplication-with-no-backwards-compatibility

  Read first:
  1) index.md
  2) design-doc/01-executor-internal-duplication-inspection-and-implementation-plan.md
  3) tasks.md
  4) changelog.md

  Goal:
  Implement VM-007 task-by-task, with NO backwards-compatibility shims.
  Scope includes BOTH:
  - finding 9: high executor internal duplication
  - finding 8: core model/helper duplication (including mustMarshalJSON duplication)

  Execution rules:
  - Implement one task at a time from tasks.md.
  - After each task:
    - run relevant tests
    - update tasks.md (check completed task)
    - append changelog.md entry
    - append diary step in VM-006 diary style (or create VM-007 diary if needed)
    - make a focused commit
  - Keep API behavior intentional; if changing behavior, document explicitly in changelog + diary.
  - Do not leave ignored persistence errors in executor paths.
  - Consolidate duplicated helper semantics into a single shared helper with explicit fallback behavior.
  - Keep code idiomatic Go and remove dead/duplicated blocks as you refactor.

  Validation baseline per meaningful step:
  - GOWORK=off go test ./... -count=1
  - GOWORK=off go test ./pkg/vmtransport/http -count=1
  - ./smoke-test.sh
  - ./test-e2e.sh
  (Use narrower test loops during development, full matrix before closing ticket.)

  Deliverables:
  - completed VM-007 tasks with commits
  - updated ticket docs (tasks/changelog/design notes)
  - final summary of what changed, what remains, and test evidence
  - call out any decisions needed from reviewer in Open Questions

Use docmgr, keep a frequent diary. As you go, check off task, commit, update diary, etc..."

**Assistant interpretation:** Execute VM-007 sequentially with one scoped task per commit, keep ticket docs current after every task, and enforce no compatibility shims while refactoring executor and helper duplication.

**Inferred user intent:** Safely land a deep refactor with auditable progress and explicit behavioral intent, rather than a large opaque rewrite.

**Commit (code):** 760b3a9 — "vm007: freeze execution API contracts baseline (task 1)"

### What I did

- Read ticket context docs (`index.md`, design doc, `tasks.md`, `changelog.md`).
- Added new HTTP integration baseline test:
  - `pkg/vmtransport/http/server_execution_contracts_integration_test.go`
- Baseline test freezes:
  - `POST /api/v1/executions/repl` returns `201` and expected execution envelope fields
  - `POST /api/v1/executions/run-file` returns `201` and expected execution envelope fields
  - `GET /api/v1/executions/{id}` returns `200` and stable execution record shape
  - `GET /api/v1/executions?session_id=...&limit=2` enforces list/get semantics
  - `GET /api/v1/executions/{id}/events?after_seq=...` returns expected event envelope and filtering semantics
- Ran task-relevant validation:
  - `GOWORK=off go test ./pkg/vmtransport/http -count=1`
- Updated task checklist:
  - checked Task 1 via `docmgr task check --ticket VM-007-REFACTOR-EXECUTOR-PIPELINE --id 1`
- Updated changelog entry for Task 1 via `docmgr changelog update`.
- Created VM-007 diary document via `docmgr doc add --ticket VM-007-REFACTOR-EXECUTOR-PIPELINE --doc-type reference --title "Diary"`.

### Why

The executor refactor will heavily touch internal control flow. Freezing external contract behavior first prevents accidental regressions during extraction and deduplication.

### What worked

- New baseline integration test compiled and passed.
- Existing test helpers were sufficient for setup; only a local request helper was needed to assert exact status codes for successful responses.
- `docmgr` workflow cleanly updated task/changelog records.

### What didn't work

- N/A for this step.

### What I learned

- Existing integration tests already covered many negative contracts; this step mainly needed explicit positive-contract freezing around status codes + envelope shape.

### What was tricky to build

The main sharp edge was balancing “freeze behavior now” with future intended behavior changes (Task 9 run-file result/value-event contract decision). I limited Task 1 assertions to currently intentional external contracts and left cross-kind parity decisions for the dedicated task.

### What warrants a second pair of eyes

- Confirm the new contract baseline test includes exactly the contract surface reviewers want frozen before executor internals change.

### What should be done in the future

- Keep this baseline test updated only when behavior changes are explicitly approved and documented in ticket changelog/diary.

### Code review instructions

- Start with `pkg/vmtransport/http/server_execution_contracts_integration_test.go`.
- Validate with `GOWORK=off go test ./pkg/vmtransport/http -count=1`.
- Confirm Task 1 checklist/changelog updates under ticket path.

### Technical details

- New helper `reqJSONStatus(...)` validates exact HTTP status and JSON envelope decoding in success flows.
- Event envelope checks verify contiguous `seq`, non-empty `type`, non-zero `ts`, and JSON-decodable `payload`.

## Step 2: Add vmexec regression safety net for behavior and parity

I added focused `pkg/vmexec` regression tests to lock current execution behavior before internal helper extraction starts. The tests cover both success and error persistence and pin the current REPL vs run-file event/result parity.

This step intentionally captures current asymmetry: REPL emits input/value events and result payload, while run-file currently persists console events without a result payload.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Add package-level executor regression coverage so upcoming refactors can be done with confidence and explicit behavior checks.

**Inferred user intent:** Prevent regressions while deduplicating executor internals by freezing concrete event and persistence behavior now.

**Commit (code):** b9db8c4 — "vm007: add vmexec regression baseline tests (task 2)"

### What I did

- Added `pkg/vmexec/executor_test.go` with three focused tests:
  - `TestExecuteREPLSuccessPersistsEventOrderAndResult`
  - `TestExecuteREPLErrorPersistsExceptionAndExecutionError`
  - `TestExecuteRunFileCurrentParityPersistsConsoleWithoutValueResult`
- Added shared fixture setup for:
  - temp SQLite store
  - template creation
  - session creation with goja runtime
  - executor wiring
- Verified:
  - event ordering/sequence correctness
  - success/error status persistence fields
  - current run-file parity behavior (console-only event + empty result)
  - persisted args/env payload presence for run-file execution
- Ran task-relevant validation:
  - `GOWORK=off go test ./pkg/vmexec -count=1`
- Checked Task 2 with `docmgr task check --ticket VM-007-REFACTOR-EXECUTOR-PIPELINE --id 2`.
- Appended Task 2 changelog entry with `docmgr changelog update`.

### Why

Task 2 establishes regression coverage where most refactor churn will happen. It reduces risk before extracting common helpers and pipeline flow.

### What worked

- New tests were deterministic and fast.
- Fixture setup through real store/session/runtime integration provided meaningful coverage without heavy HTTP-layer scaffolding.

### What didn't work

- N/A for this step.

### What I learned

- Existing executor behavior is internally inconsistent by execution kind, and this is now encoded in tests so future contract changes can be intentional and explicit.

### What was tricky to build

The tricky part was choosing which behavior to freeze vs defer. I froze current parity differences as "current contract" but kept wording explicit so Task 9 can intentionally redefine run-file value/result behavior without ambiguity.

### What warrants a second pair of eyes

- Review whether current run-file parity assertions should remain strict until Task 9 lands, or be loosened if reviewer prefers earlier contract unification.

### What should be done in the future

- Extend these tests with failure-injection coverage when persistence error handling is refactored (Task 14).

### Code review instructions

- Start with `pkg/vmexec/executor_test.go`.
- Run `GOWORK=off go test ./pkg/vmexec -count=1`.
- Confirm Task 2 state in `tasks.md` and changelog entry for this step.

### Technical details

- `newExecutorFixture` builds real store/session/executor plumbing to validate persisted artifacts, not just in-memory outputs.
- Event-order checks assert exact sequence numbers and event-type ordering for REPL/error paths.

## Step 3: Deduplicate session preparation lifecycle in executor entrypoints

I extracted a shared `prepareSession` helper in `pkg/vmexec/executor.go` to centralize session lookup, readiness checks, and execution lock acquisition. Both `ExecuteREPL` and `ExecuteRunFile` now call this helper.

This step only removes duplicated gatekeeping logic; behavior is intentionally unchanged. It narrows each entrypoint and prepares the next helper extractions.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Implement the task-by-task executor dedup plan, starting with shared session prep and explicit lock/status reuse.

**Inferred user intent:** Reduce duplication incrementally while preserving intentional behavior and validating each step.

**Commit (code):** 8f4f39a — "vm007: share executor session preparation (task 3)"

### What I did

- Added `prepareSession(sessionID string) (*vmsession.Session, func(), error)` in `pkg/vmexec/executor.go`.
- Moved duplicated logic into helper:
  - session retrieval
  - session status readiness check
  - execution lock acquisition
- Replaced duplicated inline blocks in:
  - `ExecuteREPL`
  - `ExecuteRunFile`
- Used returned unlock function (`defer unlock()`) in both execution paths.
- Ran task-relevant validation:
  - `GOWORK=off go test ./pkg/vmexec -count=1`
  - `GOWORK=off go test ./pkg/vmtransport/http -count=1`
- Marked Task 3 complete with `docmgr task check`.
- Appended Task 3 changelog entry with `docmgr changelog update`.

### Why

Session gating duplication was one of the most obvious and repeated executor blocks. Centralizing it first lowers risk for later pipeline extraction steps.

### What worked

- Behavior remained stable under existing vmexec and HTTP integration tests.
- Helper signature cleanly supports current call sites without additional compatibility shims.

### What didn't work

- N/A for this step.

### What I learned

- Returning an unlock closure keeps call sites concise and makes lock lifecycle explicit in one place.

### What was tricky to build

The main constraint was avoiding accidental behavior changes while moving lock/status logic. I preserved existing error semantics (`ErrSessionNotFound`, `ErrSessionNotReady`, `ErrSessionBusy`) and only changed structure.

### What warrants a second pair of eyes

- Confirm helper naming/signature is appropriate before further pipeline helper extraction.

### What should be done in the future

- Continue extraction with execution-record construction and event recording helpers (Tasks 4-5).

### Code review instructions

- Start with `pkg/vmexec/executor.go` and compare old vs new session-prep flow.
- Run `GOWORK=off go test ./pkg/vmexec -count=1`.
- Confirm no HTTP contract regression via `GOWORK=off go test ./pkg/vmtransport/http -count=1`.

### Technical details

- `prepareSession` returns a lock release function bound to the acquired session lock, reducing repeated `TryLock`/`Unlock` boilerplate.

## Step 4: Consolidate execution record construction into one helper

I introduced a shared execution-record builder used by both REPL and run-file flows. This removed duplicate setup of ID/session/kind/status/timestamps/metrics and moved defaults to a single place.

This keeps behavior stable while preparing the next steps (event recorder and finalize helpers). The task is structural deduplication, not behavior redesign.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Continue task-by-task executor dedup by extracting a common execution record constructor for both entrypoints.

**Inferred user intent:** Shrink duplicated internal logic and make execution lifecycle state initialization explicit and centralized.

**Commit (code):** ea3cca9 — "vm007: share execution record construction (task 4)"

### What I did

- Added `executionRecordInput` and `newExecutionRecord(...)` in `pkg/vmexec/executor.go`.
- Centralized shared record defaults:
  - generated execution ID
  - `status=running`
  - `started_at=now`
  - `metrics={}`
  - fallback defaults for empty args/env payloads
- Replaced duplicated record-construction blocks in:
  - `ExecuteREPL`
  - `ExecuteRunFile`
- Preserved existing call-site behavior by still marshaling run-file args/env at call sites before helper invocation.
- Ran task-relevant validation:
  - `GOWORK=off go test ./pkg/vmexec -count=1`
  - `GOWORK=off go test ./pkg/vmtransport/http -count=1`
- Marked Task 4 complete with `docmgr task check`.
- Appended Task 4 changelog entry with `docmgr changelog update`.

### Why

Execution record creation was repeated and error-prone. Centralizing defaults reduces drift risk and sets up later pipeline extraction.

### What worked

- Tests passed without contract drift.
- Helper input object kept call-site changes minimal and readable.

### What didn't work

- N/A for this step.

### What I learned

- Pulling constructor defaults into one helper simplifies future behavior decisions (for example, if args/env normalization changes later).

### What was tricky to build

The subtle part was preserving existing run-file serialization behavior while introducing shared defaults. I kept marshaling at call sites to avoid accidental early behavior changes.

### What warrants a second pair of eyes

- Confirm fallback/default semantics in `newExecutionRecord` are acceptable before we wire deeper pipeline helpers.

### What should be done in the future

- Extract event emission into a dedicated recorder next (Task 5) and route persistence errors explicitly.

### Code review instructions

- Start with `pkg/vmexec/executor.go` (`executionRecordInput`, `newExecutionRecord`).
- Validate with:
  - `GOWORK=off go test ./pkg/vmexec -count=1`
  - `GOWORK=off go test ./pkg/vmtransport/http -count=1`

### Technical details

- `newExecutionRecord` now owns shared initialization values and returns a fully initialized `*vmmodels.Execution`.

## Step 5: Introduce shared event recorder with explicit write-error surfacing

I added a dedicated `eventRecorder` helper and moved REPL/run-file event emission through it. This eliminates direct `store.AddEvent` duplication and ensures event persistence failures are surfaced instead of silently ignored.

The goal here was to improve write-path rigor before moving to finalize helpers and full pipeline consolidation.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Implement shared event recording internals and remove silent `AddEvent` failure behavior in executor paths.

**Inferred user intent:** Ensure persistence reliability by making event-write failures explicit during refactor.

**Commit (code):** 835de10 — "vm007: share event recorder with explicit write errors (task 5)"

### What I did

- Added `eventRecorder` in `pkg/vmexec/executor.go` with:
  - shared sequence tracking
  - typed payload marshaling helper
  - raw payload emission helper
  - first-error capture and retrieval (`recordError`, `Err`)
- Replaced direct `AddEvent` calls in REPL/run-file flows with recorder usage.
- Routed console event capture through recorder and stored console-side write errors.
- Added explicit `recorder.Err()` checks after runtime execution to surface asynchronous console write failures.
- Replaced exception/value/input event writes with explicit error-returning recorder calls.
- Ran task-relevant validation:
  - `GOWORK=off go test ./pkg/vmexec -count=1`
  - `GOWORK=off go test ./pkg/vmtransport/http -count=1`
- Marked Task 5 complete and appended changelog entry via `docmgr`.

### Why

Task 5 removes silent persistence failure paths for event writes and creates one authoritative event-emission primitive for both execution kinds.

### What worked

- Existing behavior tests still pass for success/error flows.
- Event sequence handling is now centralized rather than duplicated per entrypoint.

### What didn't work

- N/A for this step.

### What I learned

- Capturing console callback write failures requires delayed propagation because runtime callbacks cannot directly return errors to caller control flow.

### What was tricky to build

The tricky part was propagating console write failures that happen inside callback closures. I used first-error capture in the recorder and explicit post-run checks so failures are still surfaced deterministically.

### What warrants a second pair of eyes

- Verify post-run recorder error handling order is acceptable before finalize-helper extraction in Task 6.

### What should be done in the future

- Implement shared finalize helpers to ensure execution status updates also fail explicitly (Task 6).

### Code review instructions

- Start with `pkg/vmexec/executor.go` (`eventRecorder` and REPL/run-file event call sites).
- Validate with:
  - `GOWORK=off go test ./pkg/vmexec -count=1`
  - `GOWORK=off go test ./pkg/vmtransport/http -count=1`

### Technical details

- `emit(...)` marshals typed payloads; `emitRaw(...)` handles pre-encoded payloads used for value/exception records.

## Step 6: Share success/error finalization and stop ignoring UpdateExecution failures

I extracted dedicated finalization helpers for successful and failed executions and replaced direct `UpdateExecution` calls with explicit error handling. This removes the remaining silent persistence-update writes in executor completion paths.

This step keeps external behavior stable for successful persistence while making store-write failures fail fast and explicit.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Continue dedup by centralizing execution finalization and enforcing explicit persistence error handling for updates.

**Inferred user intent:** Ensure no ignored persistence failures remain in executor completion paths.

**Commit (code):** cb9a10c — "vm007: share execution finalization helpers (task 6)"

### What I did

- Added in `pkg/vmexec/executor.go`:
  - `finalizeExecutionSuccess(...)`
  - `finalizeExecutionError(...)`
- Moved shared finalize behavior into helpers:
  - set terminal status
  - set `ended_at`
  - set result/error payloads
  - persist via `UpdateExecution`
- Replaced all prior direct `UpdateExecution` calls in REPL/run-file paths with helper calls.
- Changed write-error handling to return explicit errors when persistence update fails.
- Ran task-relevant validation:
  - `GOWORK=off go test ./pkg/vmexec -count=1`
  - `GOWORK=off go test ./pkg/vmtransport/http -count=1`
- Marked Task 6 complete and updated changelog via `docmgr`.

### Why

Task 6 closes the second major persistence write gap after Task 5: update writes are now explicitly checked and surfaced.

### What worked

- Existing behavior coverage passed after helper extraction.
- Finalization semantics are now centralized and easier to reason about.

### What didn't work

- N/A for this step.

### What I learned

- Splitting finalize behavior by terminal state keeps execution paths short and clarifies write-failure handling boundaries.

### What was tricky to build

The tricky part was preserving return semantics for runtime errors (`exec` with status `error`, nil transport error) while still introducing explicit persistence-update error returns. I preserved existing runtime-error contract and only changed behavior when persistence itself fails.

### What warrants a second pair of eyes

- Confirm that failing fast on `UpdateExecution` write failures is acceptable for callers expecting legacy silent behavior.

### What should be done in the future

- Proceed with full pipeline extraction in Task 7/8 now that session prep, record creation, event emission, and finalization are all helper-backed.

### Code review instructions

- Start with `pkg/vmexec/executor.go` finalize helpers and call sites.
- Validate with:
  - `GOWORK=off go test ./pkg/vmexec -count=1`
  - `GOWORK=off go test ./pkg/vmtransport/http -count=1`

### Technical details

- Finalize helpers wrap update failures with execution ID context to improve debugging of persistence issues.

## Step 7: Move ExecuteREPL onto the shared pipeline helper chain

I introduced a reusable `runExecutionPipeline` orchestration helper and migrated `ExecuteREPL` to that flow. REPL now supplies setup/run/error/success callbacks while shared lifecycle stages are handled once.

This is the first full entrypoint migration to the pipeline style and removes most remaining REPL-specific lifecycle boilerplate.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Refactor ExecuteREPL to use the new internal pipeline helper chain while preserving behavior.

**Inferred user intent:** Land structural deduplication incrementally, with isolated commits and stable behavior checks after each step.

**Commit (code):** effc6cc — "vm007: move ExecuteREPL to pipeline chain (task 7)"

### What I did

- Added `executionPipelineConfig` and `runExecutionPipeline(...)` in `pkg/vmexec/executor.go`.
- Centralized shared pipeline stages:
  - session preparation/unlock
  - execution record creation/persistence
  - event recorder creation
  - runtime setup hook
  - run hook
  - recorder error check
  - success/error finalize hooks
- Replaced `ExecuteREPL` body with pipeline configuration callbacks:
  - setup console + input echo
  - execute REPL snippet
  - map runtime errors to exception event + finalize error
  - map runtime success to value event + finalize success
- Kept run-file entrypoint untouched for Task 8.
- Ran task-relevant validation:
  - `GOWORK=off go test ./pkg/vmexec -count=1`
  - `GOWORK=off go test ./pkg/vmtransport/http -count=1`
- Marked Task 7 complete and updated changelog via `docmgr`.

### Why

Task 7 is the core executor internal dedup milestone for REPL. It demonstrates the target pipeline structure before applying it to run-file.

### What worked

- REPL behavior remained stable under existing tests.
- Pipeline callback shape cleanly encapsulates execution-kind specifics.

### What didn't work

- N/A for this step.

### What I learned

- The callback-based pipeline reduces entrypoint complexity significantly without forcing behavior decisions prematurely.

### What was tricky to build

The tricky part was threading callback responsibilities without obscuring control flow. I kept a small config struct with explicit hooks to make phase boundaries visible and reviewable.

### What warrants a second pair of eyes

- Validate pipeline callback API ergonomics before run-file migration, especially hook granularity and error handling order.

### What should be done in the future

- Migrate `ExecuteRunFile` to the same pipeline helper chain (Task 8).

### Code review instructions

- Start with `pkg/vmexec/executor.go` (`executionPipelineConfig`, `runExecutionPipeline`, updated `ExecuteREPL`).
- Validate with:
  - `GOWORK=off go test ./pkg/vmexec -count=1`
  - `GOWORK=off go test ./pkg/vmtransport/http -count=1`

### Technical details

- `runExecutionPipeline` enforces a single lifecycle skeleton and keeps execution-kind variation in supplied hooks.

## Step 8: Move ExecuteRunFile onto the shared pipeline chain

I migrated `ExecuteRunFile` to the same `runExecutionPipeline` flow so both execution entrypoints now share lifecycle orchestration. This removed the remaining large duplicated run-file lifecycle block.

I also extracted small shared helpers for console installation and exception payload encoding to reduce callback duplication across REPL and run-file.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Complete pipeline migration by refactoring run-file entrypoint to the shared helper chain and deleting duplicated lifecycle code.

**Inferred user intent:** Fully remove internal REPL/run-file lifecycle duplication before moving to explicit contract decisions.

**Commit (code):** 233d194 — "vm007: move ExecuteRunFile to pipeline chain (task 8)"

### What I did

- Refactored `ExecuteRunFile` to call `runExecutionPipeline(...)` with run-file-specific hooks.
- Moved file path resolution/read into run-file setup hook (still using session worktree path).
- Preserved current run-file behavior:
  - console events captured
  - exception event on runtime error
  - successful completion with empty result payload
- Added shared helper methods in `pkg/vmexec/executor.go`:
  - `installConsoleRecorder(...)`
  - `exceptionPayloadJSON(...)`
- Reused shared helpers in both REPL and run-file pipeline callbacks.
- Ran task-relevant validation:
  - `GOWORK=off go test ./pkg/vmexec -count=1`
  - `GOWORK=off go test ./pkg/vmtransport/http -count=1`
- Marked Task 8 complete and updated changelog via `docmgr`.

### Why

Task 8 closes the primary duplication finding by putting both entrypoints on one execution-lifecycle skeleton.

### What worked

- Tests passed with unchanged contract behavior.
- Run-file callback configuration stayed concise after helper extraction.

### What didn't work

- N/A for this step.

### What I learned

- Once both entrypoints share the pipeline, behavior decisions (Task 9) can be implemented in a single, explicit place without copy/paste risk.

### What was tricky to build

The tricky part was preserving current run-file semantics while moving file-read and runtime wiring into callback hooks. I kept run-file success finalization payload as `nil` to avoid changing behavior before the dedicated contract task.

### What warrants a second pair of eyes

- Confirm helper extraction (`installConsoleRecorder`, `exceptionPayloadJSON`) and hook placement make callback boundaries clear.

### What should be done in the future

- Decide and implement explicit run-file result/value-event contract (Task 9).

### Code review instructions

- Start with `pkg/vmexec/executor.go` (`ExecuteRunFile` pipeline config and shared helper additions).
- Validate with:
  - `GOWORK=off go test ./pkg/vmexec -count=1`
  - `GOWORK=off go test ./pkg/vmtransport/http -count=1`

### Technical details

- Run-file setup hook now validates path/readability and stores file content for run hook execution.

## Step 9: Decide and implement explicit run-file value/result contract

I made the explicit Task 9 contract decision to align successful `run_file` behavior with REPL terminal success semantics. Run-file now emits a terminal `value` event and persists `result` JSON using the same value payload shape.

This is an intentional behavior change (no compatibility shim) and is now documented in tests, changelog, and the design-doc decision log.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Resolve the run-file vs REPL parity decision explicitly, implement it, and document behavior change intentionally.

**Inferred user intent:** Remove accidental behavioral divergence and make the execution contract explicit and testable.

**Commit (code):** c5b2e11 — "vm007: align run-file success value/result contract (task 9)"

### What I did

- Updated run-file success path in `pkg/vmexec/executor.go`:
  - emit terminal `value` event
  - persist `result` JSON payload
- Added shared `valuePayloadJSON(...)` helper and reused it for REPL + run-file value payload creation.
- Updated vmexec regression test:
  - run-file now expects non-empty result payload
  - run-file now expects `(console, value)` events on success
- Updated HTTP contract integration test:
  - run-file response/get envelopes now assert non-empty `result`
  - run-file events now assert terminal `value` event
- Updated design doc decision log to record the contract decision and rationale.
- Ran task-relevant validation:
  - `GOWORK=off go test ./pkg/vmexec -count=1`
  - `GOWORK=off go test ./pkg/vmtransport/http -count=1`
- Marked Task 9 complete and updated changelog via `docmgr`.

### Why

Standardizing success artifacts across execution kinds removes accidental drift and simplifies downstream consumers.

### What worked

- Behavior change was fully captured in package and API-level tests.
- Shared value payload helper reduced duplicate encoding logic.

### What didn't work

- N/A for this step.

### What I learned

- With both entrypoints on shared pipeline hooks, contract changes become localized and safer to reason about.

### What was tricky to build

The tricky part was making an intentional behavior change while preserving the per-task discipline and clearly separating “contract freeze” (earlier tasks) from “contract change” (this task). I handled this by updating all relevant contract tests and documenting the decision explicitly.

### What warrants a second pair of eyes

- Confirm client expectations are compatible with run-file now returning `result` and a `value` terminal event.

### What should be done in the future

- Keep REPL/run-file value payload semantics synchronized unless a future explicit contract decision says otherwise.

### Code review instructions

- Start with:
  - `pkg/vmexec/executor.go`
  - `pkg/vmexec/executor_test.go`
  - `pkg/vmtransport/http/server_execution_contracts_integration_test.go`
  - design doc decision log update
- Validate with:
  - `GOWORK=off go test ./pkg/vmexec -count=1`
  - `GOWORK=off go test ./pkg/vmtransport/http -count=1`

### Technical details

- Successful run-file executions now produce the same value payload structure as REPL (`type`, `preview`, optional exported `json`).

## Step 10: Add single shared JSON fallback helper

I introduced a shared JSON helper in `vmmodels` to be the single source for marshal-with-fallback behavior. This step establishes the common helper contract before migrating vmstore/vmcontrol call sites.

This is intentionally additive for this step: legacy helpers remain until Task 11 migration is applied.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Implement the shared JSON helper utility with explicit fallback semantics as the foundation for deduping `mustMarshalJSON`.

**Inferred user intent:** Eliminate helper drift by defining one canonical fallback-marshalling implementation.

**Commit (code):** 7efa95a — "vm007: add shared JSON fallback helper (task 10)"

### What I did

- Added `pkg/vmmodels/json_helpers.go` with:
  - `MarshalJSONWithFallback(v interface{}, fallback json.RawMessage) json.RawMessage`
- Defined explicit fallback behavior:
  - return marshaled JSON when marshal succeeds
  - return fallback verbatim when marshal fails
  - return JSON `null` when fallback is empty
- Ran task-relevant validation:
  - `GOWORK=off go test ./pkg/vmmodels -count=1`
  - `GOWORK=off go test ./pkg/vmexec -count=1`
- Marked Task 10 complete and updated changelog via `docmgr`.

### Why

Task 10 creates one authoritative helper contract so vmstore/vmcontrol no longer need duplicate private marshal helpers.

### What worked

- Helper compiled cleanly and package tests passed.
- Shared helper compiled cleanly and provided one central fallback-marshalling behavior.

### What didn't work

- N/A for this step.

### What I learned

- Keeping the helper focused on a single raw-message API keeps semantics explicit and avoids compatibility-style wrapper layering.

### What was tricky to build

The key detail was making fallback behavior explicit for empty fallback input. I set empty fallback to `null` to avoid returning invalid empty JSON.

### What warrants a second pair of eyes

- Confirm the empty-fallback-to-`null` behavior is acceptable before callsite migration.

### What should be done in the future

- Migrate vmstore/vmcontrol callsites to this shared helper and delete old duplicates (Task 11).

### Code review instructions

- Start with `pkg/vmmodels/json_helpers.go`.
- Validate with:
  - `GOWORK=off go test ./pkg/vmmodels -count=1`
  - `GOWORK=off go test ./pkg/vmexec -count=1`

### Technical details

- The shared helper returns fallback verbatim (or `null` if empty fallback) when marshaling fails.

## Step 11: Migrate vmstore/vmcontrol to single shared helper (clean cut, no wrappers)

I migrated all vmstore and vmcontrol callsites to the single `vmmodels.MarshalJSONWithFallback` helper and removed duplicated local marshal helpers. After your explicit direction ("no backwards compatibility / wrappers"), I also removed the extra wrapper layer so only one helper API remains.

This completed the helper deduplication with a clean cut and no compatibility shim functions.

### Prompt Context

**User prompt (verbatim):** "no backwards compatibility / wrappers btw, clean cut"

**Assistant interpretation:** Remove wrapper-style helper layers and complete migration directly to one shared helper API.

**Inferred user intent:** Enforce a strict deduplication boundary with no compatibility abstractions left behind.

**Commit (code):** b927b5d — "vm007: migrate JSON helper callsites clean-cut (task 11)"

### What I did

- Updated `pkg/vmstore/vmstore.go`:
  - switched VM JSON serialization callsites to `string(vmmodels.MarshalJSONWithFallback(...))`
  - removed local helper usage
- Updated `pkg/vmcontrol/template_service.go`:
  - switched template settings marshaling to `vmmodels.MarshalJSONWithFallback(...)`
- Updated `pkg/vmcontrol/types.go`:
  - removed duplicated local `mustMarshalJSON` helper entirely
- Updated `pkg/vmmodels/json_helpers.go`:
  - removed wrapper function so only one shared helper API remains
- Ran task-relevant validation:
  - `GOWORK=off go test ./pkg/vmstore ./pkg/vmcontrol ./pkg/vmmodels -count=1`
  - `GOWORK=off go test ./pkg/vmexec ./pkg/vmtransport/http -count=1`
- Marked Task 11 complete and updated changelog via `docmgr`.

### Why

Task 11 is where helper duplication is actually eliminated at callsites. Removing wrapper layering honors the no-backwards-compatibility requirement and keeps the helper contract singular.

### What worked

- All updated packages compiled and tests passed.
- Callsite migration was straightforward with explicit `string(...)` conversion where needed.

### What didn't work

- N/A for this step.

### What I learned

- One raw-message helper API is enough; string callsites can convert explicitly without dedicated wrapper functions.

### What was tricky to build

The subtle part was preserving storage-layer string behavior while removing helper wrappers. I handled this by converting at callsites (`string(...)`) instead of retaining abstraction indirection.

### What warrants a second pair of eyes

- Confirm there are no remaining private marshal-fallback helpers outside `vmmodels`.

### What should be done in the future

- Continue with Task 12 to verify config model single-source boundaries and remove any remaining duplication artifacts.

### Code review instructions

- Start with:
  - `pkg/vmstore/vmstore.go`
  - `pkg/vmcontrol/template_service.go`
  - `pkg/vmcontrol/types.go`
  - `pkg/vmmodels/json_helpers.go`
- Validate with:
  - `GOWORK=off go test ./pkg/vmstore ./pkg/vmcontrol ./pkg/vmmodels -count=1`
  - `GOWORK=off go test ./pkg/vmexec ./pkg/vmtransport/http -count=1`

### Technical details

- There is now one marshal-fallback helper API: `vmmodels.MarshalJSONWithFallback`.

## Step 12: Remove residual config alias duplication boundary in vmcontrol

I removed the remaining vmcontrol config alias layer so `vmmodels` is the direct and only source of config model types. This closes the residual “core model boundary duplication” item after helper migration.

This keeps the codebase on a clean-cut model contract with no shadow aliases.

### Prompt Context

**User prompt (verbatim):** (same as Step 1 + Step 11 note: "no backwards compatibility / wrappers btw, clean cut")

**Assistant interpretation:** Complete config model boundary cleanup by removing remaining alias indirection and using vmmodels types directly.

**Inferred user intent:** Ensure no residual duplicate model/helper declarations remain after dedup refactor.

**Commit (code):** c9ba191 — "vm007: remove residual vmcontrol config aliases (task 12)"

### What I did

- Updated `pkg/vmcontrol/types.go`:
  - removed `LimitsConfig`, `ResolverConfig`, `RuntimeConfig` alias declarations
- Updated `pkg/vmcontrol/execution_service.go`:
  - changed `loadSessionLimits` return type to `*vmmodels.LimitsConfig`
  - unmarshaled directly into `vmmodels.LimitsConfig`
- Ran task-relevant validation:
  - `GOWORK=off go test ./pkg/vmcontrol ./pkg/vmstore ./pkg/vmmodels -count=1`
  - `GOWORK=off go test ./pkg/vmexec ./pkg/vmtransport/http -count=1`
- Marked Task 12 complete and updated changelog via `docmgr`.

### Why

Residual alias layers can drift and blur ownership of canonical model types. Direct use of `vmmodels` keeps the source-of-truth boundary explicit.

### What worked

- Changes were small and compiled cleanly.
- No behavior changes; only type-boundary cleanup.

### What didn't work

- N/A for this step.

### What I learned

- Alias-based convenience in boundary packages can persist longer than intended and should be removed once migration is complete.

### What was tricky to build

This step was mostly mechanical; the main care point was updating any remaining references so there were no stale alias assumptions.

### What warrants a second pair of eyes

- Verify no remaining code or docs refer to vmcontrol-owned config aliases.

### What should be done in the future

- Proceed with Task 13 tests to lock shared helper fallback behavior and config JSON marshalling expectations.

### Code review instructions

- Start with:
  - `pkg/vmcontrol/types.go`
  - `pkg/vmcontrol/execution_service.go`
- Validate with:
  - `GOWORK=off go test ./pkg/vmcontrol ./pkg/vmstore ./pkg/vmmodels -count=1`
  - `GOWORK=off go test ./pkg/vmexec ./pkg/vmtransport/http -count=1`

### Technical details

- `ExecutionService.loadSessionLimits` now unmarshals limits into `vmmodels.LimitsConfig` directly.

## Step 13: Add focused fallback and config-marshalling tests

I added targeted tests for the shared marshal-fallback helper and for template default config JSON marshalling. This locks the helper semantics and the default config payload contract in executable tests.

These tests are intentionally narrow and map directly to Task 13 acceptance criteria.

### Prompt Context

**User prompt (verbatim):** (same as Step 1 + Step 11 note: "no backwards compatibility / wrappers btw, clean cut")

**Assistant interpretation:** Add focused tests that freeze shared helper fallback behavior and config JSON marshalling expectations after dedup migration.

**Inferred user intent:** Ensure helper/model dedup changes are guarded by explicit regression tests, not only by compile/test smoke.

**Commit (code):** 435ca32 — "vm007: add helper/config marshalling tests (task 13)"

### What I did

- Added `pkg/vmmodels/json_helpers_test.go` with tests for:
  - successful marshal path
  - marshal failure fallback path
  - empty-fallback-to-`null` behavior
- Added `pkg/vmcontrol/template_service_test.go` with focused test:
  - `TemplateService.Create` default settings are valid JSON and unmarshal into expected `vmmodels` config structs/values
- Ran task-relevant validation:
  - `GOWORK=off go test ./pkg/vmmodels ./pkg/vmcontrol -count=1`
  - `GOWORK=off go test ./pkg/vmexec ./pkg/vmtransport/http -count=1`
- Marked Task 13 complete and updated changelog via `docmgr`.

### Why

Task 13 freezes the expected behavior of the newly centralized helper contract and verifies config defaults are still serialized as intended after refactor.

### What worked

- New tests run quickly and are deterministic.
- Template-service test validates real JSON payload contents, not only non-empty blobs.

### What didn't work

- N/A for this step.

### What I learned

- A small store stub is sufficient to verify template-service marshalling behavior without full integration scaffolding.

### What was tricky to build

The subtle part was crafting a reliable marshal-failure case for tests. Using a struct with function fields provided a deterministic unsupported-type failure path.

### What warrants a second pair of eyes

- Confirm expected default values in template-service test match product expectations and not just implementation defaults.

### What should be done in the future

- Implement Task 14 persistence failure-path tests for executor write operations.

### Code review instructions

- Start with:
  - `pkg/vmmodels/json_helpers_test.go`
  - `pkg/vmcontrol/template_service_test.go`
- Validate with:
  - `GOWORK=off go test ./pkg/vmmodels ./pkg/vmcontrol -count=1`
  - `GOWORK=off go test ./pkg/vmexec ./pkg/vmtransport/http -count=1`

### Technical details

- Marshal-failure tests use unsupported function-field payloads to deterministically trigger fallback behavior.

## Step 14: Add deterministic persistence-failure tests for executor writes

I added focused failure-path tests for persistence writes during REPL execution, covering create, event append, and finalize update failures. These tests verify deterministic wrapped error outcomes for each failure mode.

To enable clean failure injection without compatibility shims, I introduced an internal `executionStore` interface in `Executor` and used a test store implementation that can force specific write failures.

### Prompt Context

**User prompt (verbatim):** (same as Step 1 + Step 11 note: "no backwards compatibility / wrappers btw, clean cut")

**Assistant interpretation:** Add targeted persistence failure tests and keep the implementation clean-cut (no compatibility indirection layers).

**Inferred user intent:** Guarantee executor persistence errors are explicit and test-locked, especially after refactor removed previously ignored writes.

**Commit (code):** Pending for Step 14 commit creation.

### What I did

- Added `pkg/vmexec/executor_persistence_failures_test.go` with tests:
  - `TestExecuteREPLCreateExecutionFailureReturnsDeterministicError`
  - `TestExecuteREPLAddEventFailureReturnsDeterministicError`
  - `TestExecuteREPLUpdateExecutionFailureReturnsDeterministicError`
- Added internal test fixture with configurable failing store implementation to force:
  - `CreateExecution` failure
  - `AddEvent` failure
  - `UpdateExecution` failure
- Updated `pkg/vmexec/executor.go`:
  - switched internal store field to private `executionStore` interface for deterministic failure injection in tests
  - kept public constructor unchanged (`NewExecutor(*vmstore.VMStore, ...)`)
- Ran task-relevant validation:
  - `GOWORK=off go test ./pkg/vmexec -count=1`
  - `GOWORK=off go test ./pkg/vmtransport/http ./pkg/vmcontrol ./pkg/vmmodels -count=1`
- Marked Task 14 complete and updated changelog via `docmgr`.

### Why

Task 14 ensures persistence-failure paths are no longer accidental behavior and are now explicitly validated with deterministic tests.

### What worked

- Failure tests are stable and assert wrapped error context.
- Existing package tests remained green after internal store abstraction.

### What didn't work

- Initial test attempt imported `vmcontrol` from `pkg/vmexec` tests, which caused an import cycle (`vmcontrol` imports `vmexec`). I replaced it with direct VM/store setup in test fixture.

### What I learned

- For internal package tests, keeping fixtures independent of higher-level packages avoids cyclic dependency traps.

### What was tricky to build

The tricky part was forcing each write failure independently without altering runtime/session setup. The injected store wrapper solved this cleanly while preserving production constructor behavior.

### What warrants a second pair of eyes

- Confirm internal `executionStore` interface shape is minimal and appropriate for future maintenance.

### What should be done in the future

- Run full validation matrix and finalize ticket docs/test evidence (Task 15).

### Code review instructions

- Start with:
  - `pkg/vmexec/executor_persistence_failures_test.go`
  - `pkg/vmexec/executor.go` (`executionStore` interface usage)
- Validate with:
  - `GOWORK=off go test ./pkg/vmexec -count=1`
  - `GOWORK=off go test ./pkg/vmtransport/http ./pkg/vmcontrol ./pkg/vmmodels -count=1`

### Technical details

- Each failure-path test asserts both deterministic failure type (create/add-event/update stage) and wrapped forced error marker text.
