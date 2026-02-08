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
    - Path: pkg/vmexec/executor.go
      Note: Task 3 shared session preparation helper extraction
    - Path: pkg/vmexec/executor_test.go
      Note: Task 2 vmexec regression test coverage
    - Path: pkg/vmtransport/http/server_execution_contracts_integration_test.go
      Note: |-
        Task 1 baseline contract coverage for execution endpoints and event envelopes
        Task 1 external contract baseline test
    - Path: ttmp/2026/02/08/VM-007-REFACTOR-EXECUTOR-PIPELINE--remove-executor-internal-duplication-with-no-backwards-compatibility/changelog.md
      Note: |-
        Task-level changelog entries
        Task 1 changelog entry
        Task 2 changelog entry
        Task 3 changelog entry
    - Path: ttmp/2026/02/08/VM-007-REFACTOR-EXECUTOR-PIPELINE--remove-executor-internal-duplication-with-no-backwards-compatibility/tasks.md
      Note: |-
        Task checklist updated after Task 1 completion
        Task 1 checklist update
        Task 2 checklist update
        Task 3 checklist update
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

**Commit (code):** Pending for Step 3 commit creation.

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
