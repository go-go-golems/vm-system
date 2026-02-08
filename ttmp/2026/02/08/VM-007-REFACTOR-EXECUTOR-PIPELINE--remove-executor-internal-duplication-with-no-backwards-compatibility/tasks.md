# Tasks

## TODO

- [x] Baseline and freeze current external contracts (status codes, event envelopes, execution list/get semantics) that must stay intentional during internal refactor
- [x] Add focused vmexec regression tests for current expected behavior (event ordering, success/error persistence fields, run-file and repl parity checks)
- [x] Introduce shared internal session-preparation helper (`get+status-check+lock`) and remove duplicated lock/status blocks from both execution entrypoints
- [x] Introduce shared execution-record constructor/factory for REPL and run-file kinds (single source of defaults for args/env/metrics/status/timestamps)
- [x] Introduce shared event recorder helper with explicit error propagation from store writes (no silent `AddEvent` failures)
- [x] Introduce shared finalize helpers for success/error paths with explicit persistence error handling (no ignored `UpdateExecution` failures)
- [x] Refactor `ExecuteREPL` to the pipeline helper chain and delete duplicated logic blocks
- [x] Refactor `ExecuteRunFile` to the pipeline helper chain and delete duplicated logic blocks
- [ ] Decide and implement explicit contract for run-file result/value events vs REPL value events (documented behavior, no compatibility shims)
- [ ] Consolidate duplicated JSON helper behavior (`mustMarshalJSON`) into one shared utility with explicit fallback semantics
- [ ] Remove helper duplication between `vmstore` and `vmcontrol` by migrating call sites to shared helper API
- [ ] Verify config model single-source boundary (`vmmodels` as source) and remove any residual duplicated declarations/helpers
- [ ] Add focused tests for shared helper fallback behavior and config JSON marshalling expectations
- [ ] Add tests for persistence-failure paths (CreateExecution/AddEvent/UpdateExecution failures) to ensure deterministic error outcomes
- [ ] Run validation matrix: `go test ./... -count=1`, HTTP integration suite, smoke/e2e scripts, and update ticket docs with results
