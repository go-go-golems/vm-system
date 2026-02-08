# Changelog

## 2026-02-08

- Initial workspace created

## 2026-02-08

Created detailed VM-007 inspection/design document and concrete implementation task list for removing high executor internal duplication with no backward-compatibility constraints.

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ttmp/2026/02/08/VM-007-REFACTOR-EXECUTOR-PIPELINE--remove-executor-internal-duplication-with-no-backwards-compatibility/design-doc/01-executor-internal-duplication-inspection-and-implementation-plan.md - Deep review and implementation blueprint
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ttmp/2026/02/08/VM-007-REFACTOR-EXECUTOR-PIPELINE--remove-executor-internal-duplication-with-no-backwards-compatibility/tasks.md - Concrete task breakdown for delegated implementation

## 2026-02-08

Expanded VM-007 scope to explicitly include finding 8 (core model/helper duplication) alongside finding 9 (executor internal duplication), with updated design plan and concrete tasks for shared helper consolidation.

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ttmp/2026/02/08/VM-007-REFACTOR-EXECUTOR-PIPELINE--remove-executor-internal-duplication-with-no-backwards-compatibility/index.md - Updated ticket title/summary/scope to include finding 8 + 9
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ttmp/2026/02/08/VM-007-REFACTOR-EXECUTOR-PIPELINE--remove-executor-internal-duplication-with-no-backwards-compatibility/design-doc/01-executor-internal-duplication-inspection-and-implementation-plan.md - Added core helper/model duplication analysis and implementation phases
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ttmp/2026/02/08/VM-007-REFACTOR-EXECUTOR-PIPELINE--remove-executor-internal-duplication-with-no-backwards-compatibility/tasks.md - Added concrete tasks for shared JSON helper/model deduplication

## 2026-02-08

Task 1: Added execution API contract baseline integration coverage (status codes, envelopes, list/get semantics, after_seq filtering) to freeze intentional external behavior before executor internals refactor.

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/vmtransport/http/server_execution_contracts_integration_test.go — New baseline contract test for execution endpoints and event envelopes
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ttmp/2026/02/08/VM-007-REFACTOR-EXECUTOR-PIPELINE--remove-executor-internal-duplication-with-no-backwards-compatibility/tasks.md — Marked Task 1 complete


## 2026-02-08

Task 2: Added focused vmexec regression tests covering REPL success/error event ordering and persistence fields, plus current run-file parity behavior (console-only events and empty result payload).

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/vmexec/executor_test.go — Regression tests freezing vmexec behavior before internal pipeline refactor
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ttmp/2026/02/08/VM-007-REFACTOR-EXECUTOR-PIPELINE--remove-executor-internal-duplication-with-no-backwards-compatibility/tasks.md — Marked Task 2 complete


## 2026-02-08

Task 3: Extracted shared session-preparation helper (get/status-check/lock) and switched both ExecuteREPL and ExecuteRunFile to the shared path.

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/vmexec/executor.go — Added prepareSession helper and removed duplicated session lock/status blocks
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ttmp/2026/02/08/VM-007-REFACTOR-EXECUTOR-PIPELINE--remove-executor-internal-duplication-with-no-backwards-compatibility/tasks.md — Marked Task 3 complete


## 2026-02-08

Task 4: Introduced shared execution-record constructor used by REPL and run-file flows, centralizing status/started_at/metrics defaults and record ID creation.

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/vmexec/executor.go — Added executionRecordInput and newExecutionRecord helper to remove duplicated record construction
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ttmp/2026/02/08/VM-007-REFACTOR-EXECUTOR-PIPELINE--remove-executor-internal-duplication-with-no-backwards-compatibility/tasks.md — Marked Task 4 complete


## 2026-02-08

Task 5: Added shared event recorder helper and routed REPL/run-file event emission through it so AddEvent persistence failures are explicitly surfaced instead of ignored.

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/vmexec/executor.go — Introduced eventRecorder helper with explicit AddEvent error propagation
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ttmp/2026/02/08/VM-007-REFACTOR-EXECUTOR-PIPELINE--remove-executor-internal-duplication-with-no-backwards-compatibility/tasks.md — Marked Task 5 complete


## 2026-02-08

Task 6: Added shared finalize helpers for success/error execution completion and made UpdateExecution failures explicit across REPL and run-file paths.

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/vmexec/executor.go — Introduced finalizeExecutionSuccess/finalizeExecutionError and removed ignored UpdateExecution calls
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ttmp/2026/02/08/VM-007-REFACTOR-EXECUTOR-PIPELINE--remove-executor-internal-duplication-with-no-backwards-compatibility/tasks.md — Marked Task 6 complete


## 2026-02-08

Task 7: Added internal runExecutionPipeline helper chain and refactored ExecuteREPL to the pipeline flow while preserving current REPL behavior.

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/vmexec/executor.go — Introduced executionPipelineConfig/runExecutionPipeline and migrated ExecuteREPL
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ttmp/2026/02/08/VM-007-REFACTOR-EXECUTOR-PIPELINE--remove-executor-internal-duplication-with-no-backwards-compatibility/tasks.md — Marked Task 7 complete


## 2026-02-08

Task 8: Migrated ExecuteRunFile to runExecutionPipeline and removed duplicated lifecycle blocks, reusing shared console and exception helpers.

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/vmexec/executor.go — ExecuteRunFile now uses pipeline hooks plus shared console/exception helpers
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ttmp/2026/02/08/VM-007-REFACTOR-EXECUTOR-PIPELINE--remove-executor-internal-duplication-with-no-backwards-compatibility/tasks.md — Marked Task 8 complete


## 2026-02-08

Task 9: Decided and implemented explicit run-file success contract parity with REPL: run-file now emits terminal value events and persists result payload JSON; tests updated to lock this intentional behavior change.

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/vmexec/executor.go — Run-file success path now emits value event and persists result payload
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/vmexec/executor_test.go — Updated vmexec regression expectations for run-file value/result semantics
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/vmtransport/http/server_execution_contracts_integration_test.go — Updated API contract baseline for run-file result/value event behavior
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ttmp/2026/02/08/VM-007-REFACTOR-EXECUTOR-PIPELINE--remove-executor-internal-duplication-with-no-backwards-compatibility/design-doc/01-executor-internal-duplication-inspection-and-implementation-plan.md — Recorded run-file contract decision in Decision Log
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ttmp/2026/02/08/VM-007-REFACTOR-EXECUTOR-PIPELINE--remove-executor-internal-duplication-with-no-backwards-compatibility/tasks.md — Marked Task 9 complete


## 2026-02-08

Task 10: Introduced shared JSON marshal fallback utility in vmmodels with explicit fallback semantics to serve as single-source helper for vmstore/vmcontrol migration.

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/vmmodels/json_helpers.go — New shared JSON helper with explicit fallback behavior
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ttmp/2026/02/08/VM-007-REFACTOR-EXECUTOR-PIPELINE--remove-executor-internal-duplication-with-no-backwards-compatibility/tasks.md — Marked Task 10 complete


## 2026-02-08

Task 11: Migrated vmstore and vmcontrol call sites to the single shared vmmodels JSON fallback helper and removed duplicated local mustMarshalJSON implementations (clean cut, no compatibility wrapper layer).

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/vmcontrol/template_service.go — Switched template settings marshaling to shared vmmodels helper
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/vmcontrol/types.go — Removed duplicated local mustMarshalJSON helper
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/vmmodels/json_helpers.go — Removed string-wrapper helper to enforce single shared helper API
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/vmstore/vmstore.go — Replaced local helper usage with direct vmmodels.MarshalJSONWithFallback calls
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ttmp/2026/02/08/VM-007-REFACTOR-EXECUTOR-PIPELINE--remove-executor-internal-duplication-with-no-backwards-compatibility/tasks.md — Marked Task 11 complete


## 2026-02-08

Task 12: Removed residual vmcontrol config-type alias duplication so vmmodels is the sole config model source; execution service now consumes vmmodels.LimitsConfig directly.

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/vmcontrol/execution_service.go — Switched loadSessionLimits to vmmodels.LimitsConfig
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/vmcontrol/types.go — Removed config alias declarations to eliminate residual duplication boundary
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ttmp/2026/02/08/VM-007-REFACTOR-EXECUTOR-PIPELINE--remove-executor-internal-duplication-with-no-backwards-compatibility/tasks.md — Marked Task 12 complete


## 2026-02-08

Task 13: Added focused tests for shared helper fallback semantics and template default config JSON marshalling expectations.

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/vmcontrol/template_service_test.go — Tests template default settings JSON marshalling into vmmodels config structs
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/vmmodels/json_helpers_test.go — Tests for marshal fallback success/failure/null behavior
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ttmp/2026/02/08/VM-007-REFACTOR-EXECUTOR-PIPELINE--remove-executor-internal-duplication-with-no-backwards-compatibility/tasks.md — Marked Task 13 complete


## 2026-02-08

Task 14: Added deterministic persistence failure-path tests for ExecuteREPL covering CreateExecution, AddEvent, and UpdateExecution write failures with explicit wrapped error outcomes.

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/vmexec/executor.go — Introduced internal executionStore interface to enable deterministic failure-path testing without compatibility shims
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/vmexec/executor_persistence_failures_test.go — New failure-injection tests for create/add-event/update persistence errors
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ttmp/2026/02/08/VM-007-REFACTOR-EXECUTOR-PIPELINE--remove-executor-internal-duplication-with-no-backwards-compatibility/tasks.md — Marked Task 14 complete


## 2026-02-08

Task 15: Ran full validation matrix after VM-007 refactor. All checks passed: go test ./... -count=1, go test ./pkg/vmtransport/http -count=1, ./smoke-test.sh (10/10), ./test-e2e.sh (daemon-first end-to-end).

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/smoke-test.sh — Validation script executed successfully
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/test-e2e.sh — Validation script executed successfully
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ttmp/2026/02/08/VM-007-REFACTOR-EXECUTOR-PIPELINE--remove-executor-internal-duplication-with-no-backwards-compatibility/tasks.md — Marked Task 15 complete

