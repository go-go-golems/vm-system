# Changelog

## 2026-02-08

- Initial workspace created.

## 2026-02-09

- Added exhaustive design doc for glazed migration with command-by-command inventory and root flattening plan.
- Rewrote design doc to be strictly new-API-first (`schema/fields/values/sources`) and removed legacy vocabulary/mapping content.
- Added detailed diary with command traces, failure logs, decisions, and review instructions.
- Replaced placeholder tasks with phased implementation checklist covering root wiring, porting, tests, docs, and upload.
- Replaced broad ticket tasks with implementation-ordered checklist (`T01`-`T20`) for commit-by-commit execution.
- Uploaded bundled plan + diary PDFs to reMarkable at `/ai/2026/02/09/VM-011-PORT-VM-CLI-TO-GLAZED` (`VM-011 Glazed Migration Plan` and `VM-011 Glazed Migration Plan Final`).
- Implemented root flattening slice: added Glazed dependency, wired help/logging in root, removed `http` from root registration, and added new `ops` command + vmclient ops methods.
- Added shared Glazed helper layer and ported `serve` + `libs` command groups to new API command implementations.
- Ported `session` command group to Glazed command implementations and switched canonical lifecycle CLI verb to `session close` (no `session delete` registration).
- Ported the full `exec` command group to Glazed command implementations while preserving JSON flag parsing and output formatting behavior.
- Ported the full `template` command group to Glazed command implementations (all existing template verbs).
- Removed legacy `http` command artifacts and added root/session topology tests enforcing flattened command tree semantics.
- Updated `README.md` and getting-started guide CLI examples to root taxonomy (`template/session/exec/ops`) and canonicalized session lifecycle usage to `session close`.
- Updated `smoke-test.sh` and `test-e2e.sh` to invoke flattened root command groups (removed `http` parent usage).
- Validation pass completed with no regressions:
  - `GOWORK=off go test ./cmd/vm-system -count=1`
  - `GOWORK=off go test ./... -count=1`
- Recorded implementation commit sequence for VM-011 execution:
  - `c93eeb9` docs(vm-011): add detailed execution checklist and diary step
  - `5189e66` feat(vm-011): flatten root and wire help/logging plus ops commands
  - `f1e084a` feat(vm-011): add glazed helpers and port serve/libs groups
  - `2d78c17` feat(vm-011): port session commands to glazed and expose close verb
  - `fb191d9` feat(vm-011): port exec command group to glazed
  - `0f52208` feat(vm-011): port template command group to glazed
  - `d15fc55` test(vm-011): remove http artifacts and add root/session topology coverage
  - `6cb58ef` docs(vm-011): align root CLI docs and smoke/e2e scripts
  - `95b1c6f` docs(vm-011): record validation pass for glazed CLI migration

## 2026-02-08

Closed ticket after completing glazed migration, flattened root command taxonomy, docs/script alignment, and validation.

