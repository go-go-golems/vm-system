---
Title: Comprehensive vm-system implementation quality review
Ticket: VM-006-REVIEW-VM-SYSTEM
Status: complete
Topics:
    - backend
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: vm-system/vm-system/cmd/vm-system/cmd_modules.go
      Note: legacy direct-DB command surface
    - Path: vm-system/vm-system/pkg/vmcontrol/execution_service.go
      Note: |-
        run-file path normalization and post-execution limit enforcement
        Path normalization and limit enforcement contract analysis
    - Path: vm-system/vm-system/pkg/vmexec/executor.go
      Note: |-
        duplicated execution pipeline logic and persistence/error-handling behavior
        Execution pipeline duplication and persistence semantics
    - Path: vm-system/vm-system/pkg/vmsession/session.go
      Note: |-
        startup execution path handling, runtime map lifecycle, library loading behavior
        Session startup and runtime registry lifecycle analysis
    - Path: vm-system/vm-system/pkg/vmstore/vmstore.go
      Note: |-
        persistence contracts, error typing, and JSON decoding behavior
        Error mapping and storage-contract analysis
    - Path: vm-system/vm-system/pkg/vmtransport/http/server.go
      Note: |-
        API validation, error mapping, and envelope contracts
        API validation and error envelope behavior
    - Path: vm-system/vm-system/smoke-test.sh
      Note: current daemon-first happy-path setup validation
    - Path: vm-system/vm-system/test-e2e.sh
      Note: current daemon-first end-to-end workflow validation
    - Path: vm-system/vm-system/test-goja-library-execution.sh
      Note: |-
        stale script behavior and repository state mutation
        Legacy script breakage evidence
    - Path: vm-system/vm-system/test-library-loading.sh
      Note: |-
        stale test surface using removed command contracts
        Legacy script breakage evidence
    - Path: vm-system/vm-system/test-library-requirements.sh
      Note: |-
        stale script and unsupported command assumptions
        Legacy script breakage evidence
ExternalSources: []
Summary: Deep implementation, testing, and setup review of vm-system in the context of VM-001, VM-004, and VM-005 with runtime-verified defects, structural quality findings, and prioritized remediation plan.
LastUpdated: 2026-02-08T11:15:24.652798838-05:00
WhatFor: Assess implementation quality and test/setup reliability after VM-001/VM-004/VM-005 and define concrete cleanup priorities.
WhenToUse: Use when planning vm-system hardening work, cleanup/refactor tickets, and test strategy upgrades.
---


# Comprehensive vm-system implementation quality review

## Executive Summary

This review was executed against current `main` workspace code and ticket context from:

- `VM-001-ANALYZE-VM`
- `VM-004-EXPAND-E2E-COVERAGE`
- `VM-005-DEVELOPER-GETTING-STARTED`

The daemon-first architecture introduced in VM-001 is directionally correct, and VM-004 improved HTTP integration coverage substantially. However, the implementation still contains several high-impact correctness and safety gaps, substantial legacy surface drift, and test coverage concentration that leaves core runtime/storage behavior largely unprotected.

Current quality posture:

- Architecture direction: good
- API integration coverage depth: good
- Runtime/path safety: weak
- Legacy surface cleanup: weak
- Core package test balance: weak

Top risks found:

1. Worktree isolation can be bypassed (startup traversal + run-file symlink escape).
2. Failed session startup leaks/retains crashed runtime entries as "active".
3. Execution-limit API behavior is contract-inconsistent (422 returned while execution is persisted as `ok`).
4. Legacy script/test surface is stale and actively broken against current CLI.
5. Missing error typing causes avoidable `500 INTERNAL` on expected not-found paths.

## Problem Statement

VM-001 established the daemon-first split and explicitly called out legacy drift and runtime hardening needs. VM-004 focused on route-level API coverage. VM-005 published a strong contributor guide that documents known weak areas.

The problem now is not missing roadmap clarity. The problem is that high-priority hardening tasks remain incomplete in code while legacy/partial behavior still ships in the same repository surface.

Symptoms:

- Core runtime safety guarantees are incomplete despite strong API envelope testing.
- Legacy command/scripts remain present and failing, creating false confidence and setup confusion.
- Critical behavior contracts are inconsistent across handler response vs persisted state.

## Review Scope And Method

### Context Anchoring

- Read VM-001/VM-004/VM-005 ticket docs (index/tasks/design/changelog).
- Evaluated claims in those tickets against current implementation behavior.

### Static Inspection Scope

- `cmd/vm-system`
- `pkg/vmcontrol`
- `pkg/vmdaemon`
- `pkg/vmtransport/http`
- `pkg/vmclient`
- `pkg/vmsession`
- `pkg/vmexec`
- `pkg/vmstore`
- test scripts and docs linked from VM-004 and VM-005

### Dynamic Validation Executed

- `GOWORK=off go test ./... -count=1`
- `GOWORK=off go test -race ./pkg/vmtransport/http -count=1`
- `GOWORK=off go vet ./...`
- `GOWORK=off go test ./... -cover -count=1`
- `./smoke-test.sh`
- `./test-e2e.sh`
- `./test-library-loading.sh` (fails)
- `./test-library-requirements.sh` (fails)
- `./test-goja-library-execution.sh` (fails)
- Additional targeted runtime probes for path safety, close semantics, startup failure lifecycle, and limit-enforcement contract consistency.

## Findings (Ordered By Severity)

### 1) Path-safety model is bypassable (high)

Problem:

`run-file` path normalization blocks obvious traversal but can be bypassed using symlinks inside the worktree. Startup-file execution has no traversal guard and can run files outside worktree directly.

Where to look:

- `pkg/vmcontrol/execution_service.go:70`
- `pkg/vmexec/executor.go:198`
- `pkg/vmsession/session.go:198`
- `pkg/vmtransport/http/server.go:217`

Example:

```go
fullPath := filepath.Join(worktreePath, cleanRelative)
relativeToRoot, err := filepath.Rel(worktreePath, fullPath)
if strings.HasPrefix(relativeToRoot, "..") {
    return "", vmmodels.ErrPathTraversal
}
```

Why it matters:

- Symlink in worktree can point outside root and still pass normalization.
- Startup file `../outside.js` is accepted and executed at session startup.
- This is a direct boundary/sandbox violation risk.

Dynamic evidence:

- `run-file` on symlinked outside file returned `201` and executed successfully.
- startup file `../outside-startup.js` was accepted (`201`) and set runtime global from outside file.

Cleanup sketch:

```pseudo
func resolveWithinWorktree(root, rel string) (string, error) {
  if abs(rel) || rel == "." || rel starts with ".." => reject
  candidate := join(root, clean(rel))
  realRoot := EvalSymlinks(root)
  realPath := EvalSymlinks(candidate)
  if !isWithin(realRoot, realPath) => reject
  return realPath, nil
}

apply to:
- ExecutionService.normalizeRunFilePath
- SessionManager.runStartupFiles path resolution
- startup-file add validation in HTTP handler (and/or core)
```

### 2) Session startup failure leaks active runtime entry (high)

Problem:

Session is inserted into `SessionManager.sessions` before startup files execute. If startup fails, function returns error without removing map entry.

Where to look:

- `pkg/vmsession/session.go:121`
- `pkg/vmsession/session.go:127`
- `pkg/vmcontrol/runtime_registry.go:18`

Example:

```go
sm.sessions[sessionID] = session
...
if err := sm.runStartupFiles(session); err != nil {
    session.Status = vmmodels.SessionCrashed
    ...
    return nil, fmt.Errorf("startup failed: %w", err)
}
```

Why it matters:

- API returns failure (`500`) while runtime summary still reports an active session.
- Crashed objects can accumulate in daemon memory and distort ops visibility.

Dynamic evidence:

- Invalid startup JS returned `500` on session create.
- immediately after, `/api/v1/runtime/summary` reported `active_sessions:1`.

Cleanup sketch:

```pseudo
create session object
run startup
if startup fails:
  mark crashed in store
  do NOT register in active map OR remove if pre-registered
  return typed startup error
if startup succeeds:
  register in active map
  mark ready
```

### 3) Limit-enforcement contract mismatch (high)

Problem:

Limit checks run after executor persists execution as `ok`. API can return `422 OUTPUT_LIMIT_EXCEEDED` while persisted execution remains successful with result payload.

Where to look:

- `pkg/vmcontrol/execution_service.go:27`
- `pkg/vmcontrol/execution_service.go:95`
- `pkg/vmexec/executor.go:147`
- `pkg/vmexec/executor.go:174`

Example:

```go
execution, err := s.runtime.ExecuteREPL(...)
...
if err := s.enforceLimits(input.SessionID, execution.ID); err != nil {
    return nil, err
}
```

Why it matters:

- Client sees failure status.
- later list/get APIs show the same execution as `ok`.
- Creates audit and UX inconsistency.

Dynamic evidence:

- Repl call returned `422 OUTPUT_LIMIT_EXCEEDED`.
- `GET /api/v1/executions?session_id=...` showed the execution as `status:"ok"` with result.

Cleanup sketch:

```pseudo
option A (preferred): enforce in executor before final status write
option B: keep post-check but mutate persisted execution to error/limit_exceeded
option C: return success with warning event instead of HTTP error (explicit contract)

choose one contract and align handler + persistence + tests
```

### 4) Legacy command/script surface is stale and broken (high)

Problem:

Multiple script/test artifacts still use removed `vm` command and old flags (`--vm-id`, `--worktree`, `--file` forms) and fail immediately.

Where to look:

- `test-library-loading.sh:24`
- `test-library-requirements.sh:66`
- `test-goja-library-execution.sh:74`
- `cmd/vm-system/main.go:26`

Example:

```bash
CREATE_OUTPUT=$(./vm-system vm create --name "test-vm-$TIMESTAMP")
```

Why it matters:

- New contributors run scripts that fail instantly.
- Quality signal is noisy: green core scripts coexist with red legacy scripts.
- Direct contradiction to VM-005 onboarding reliability goals.

Dynamic evidence:

- All three library scripts failed with:
  - `Error: unknown command "vm" for "vm-system"`

Cleanup sketch:

```pseudo
choose policy per script:
- migrate to daemon-first template/session/exec commands, OR
- move to archive/legacy and stop advertising as runnable tests

enforce in CI:
- only supported scripts are executed
- legacy scripts fail build if still under active test entrypoints
```

### 5) Expected not-found path returns 500 (medium-high)

Problem:

Missing execution ID returns `500 INTERNAL` instead of `404` typed contract.

Where to look:

- `pkg/vmstore/vmstore.go:487`
- `pkg/vmtransport/http/server.go:460`
- `pkg/vmtransport/http/server_executions_integration_test.go:80`

Example:

```go
if err == sql.ErrNoRows {
    return nil, fmt.Errorf("execution not found")
}
```

Why it matters:

- Inconsistent error envelope semantics vs template/session behavior.
- Harder client-side handling and retry policy.

Dynamic evidence:

- `GET /api/v1/executions/does-not-exist` returned `500` with `code:"INTERNAL"`.

Cleanup sketch:

```pseudo
add vmmodels.ErrExecutionNotFound
store returns typed error
writeCoreError maps to 404 EXECUTION_NOT_FOUND
update integration tests accordingly
```

### 6) Security/validation parity gap for startup paths (medium-high)

Problem:

Run-file path gets normalization checks; startup-file path does not. API accepts traversal path at config time.

Where to look:

- `pkg/vmtransport/http/server.go:225`
- `pkg/vmsession/session.go:198`

Example:

```go
if req.Path == "" {
    writeError(...)
    return
}
```

Why it matters:

- Uneven policy: startup path is currently less constrained than run-file path.
- Configuration-time acceptance of unsafe path increases latent risk.

Cleanup sketch:

```pseudo
introduce shared path policy helper in vmcontrol
validate startup path on add
re-validate before execute for defense in depth
```

### 7) Soft-fail error handling masks policy failures (medium)

Problem:

Limit-loading and event-loading failures are silently ignored in `enforceLimits`, turning enforcement into best-effort.

Where to look:

- `pkg/vmcontrol/execution_service.go:96`
- `pkg/vmcontrol/execution_service.go:103`

Example:

```go
limits, err := s.loadSessionLimits(sessionID)
if err != nil {
    return nil
}
```

Why it matters:

- Broken settings decode or store errors can disable protection silently.
- Operational incidents become harder to detect.

Cleanup sketch:

```pseudo
on policy load failure:
- emit system event and fail closed for strict mode, OR
- mark execution as policy_error with explicit warning

never silently swallow in safety-critical branch
```

### 8) Core model and helper duplication creates drift risk (medium)

Problem:

`LimitsConfig/ResolverConfig/RuntimeConfig` are duplicated in `vmmodels` and `vmcontrol`; `mustMarshalJSON` also duplicated with different defaults.

Where to look:

- `pkg/vmmodels/models.go:45`
- `pkg/vmcontrol/types.go:39`
- `pkg/vmstore/vmstore.go:18`
- `pkg/vmcontrol/types.go:62`

Why it matters:

- Field drift can compile cleanly but break behavior.
- Semantics of fallback defaults are inconsistent (`[]` vs caller-provided JSON).

Cleanup sketch:

```pseudo
single-source config types in vmmodels (or vmcontrol/config)
remove duplicate structs
replace mustMarshalJSON variants with one utility + explicit fallback constants
```

### 9) Executor has high internal duplication and unchecked persistence errors (medium)

Problem:

REPL and run-file paths duplicate event and status logic; many store writes ignore returned errors.

Where to look:

- `pkg/vmexec/executor.go:34`
- `pkg/vmexec/executor.go:180`
- `pkg/vmexec/executor.go:96`
- `pkg/vmexec/executor.go:141`

Why it matters:

- Increases maintenance overhead and bug divergence risk.
- Persistence failures can produce partial/invisible execution history.

Cleanup sketch:

```pseudo
extract shared execution pipeline:
- beginExecution(kind)
- withSessionLock(session)
- captureConsoleAndEvents()
- finishExecution(status, result, err)

propagate write failures explicitly
```

### 10) API/CLI vocabulary and ownership boundaries remain confusing (medium)

Problem:

Primary API uses `template` language, but modules/libraries commands still mutate `vm` rows directly via DB path and bypass daemon.

Where to look:

- `cmd/vm-system/cmd_modules.go:45`
- `cmd/vm-system/cmd_template.go:14`
- `docs/getting-started-from-first-vm-to-contributor-guide.md:530`

Why it matters:

- Two control planes (daemon API vs direct DB) with different validation behavior.
- Users can unknowingly target different DBs (`--server-url` + unrelated `--db`).

Cleanup sketch:

```pseudo
pick one:
1) expose modules/libraries via template API and client
2) mark modules command legacy/read-only and remove mutating ops

avoid mixed runtime-control paths in normal workflows
```

### 11) Test strategy is deep in HTTP layer but shallow elsewhere (medium)

Problem:

Coverage is concentrated in `pkg/vmtransport/http` while runtime/store/control packages have no direct tests.

Measured result:

- `pkg/vmtransport/http`: 72.6%
- all other packages: 0.0%

Where to look:

- test output from `go test ./... -cover`
- `pkg/vmtransport/http/*_integration_test.go`

Why it matters:

- Handler-level tests can miss core invariants and storage edge cases.
- Bugs in `vmexec`, `vmsession`, `vmstore`, `vmcontrol` can pass despite API tests.

Cleanup sketch:

```pseudo
add focused tests:
- vmcontrol: normalizeRunFilePath + limit contract behavior
- vmsession: startup failure lifecycle + path security
- vmexec: event ordering and persistence error paths
- vmstore: typed not-found errors + JSON decode error behavior
```

### 12) Setup script portability and repo hygiene issues (medium-low)

Problem:

Some active scripts are robust (`smoke-test.sh`, `test-e2e.sh`), but others mutate tracked directories and rely on stale contracts.

Where to look:

- `test-goja-library-execution.sh:11`
- `smoke-test.sh:110`
- `test-e2e.sh:67`

Why it matters:

- Running certain scripts dirties repository state (`test-goja-workspace` rewrite).
- `realpath --relative-to` may be non-portable across environments.

Cleanup sketch:

```pseudo
all scripts must:
- use mktemp dirs only
- avoid editing tracked test fixtures unless explicitly intended
- avoid GNU-only flags or guard with compatibility checks
```

## Strengths Observed

1. Daemon-first architecture and layering from VM-001 is coherent and readable.
2. HTTP integration suite from VM-004 is meaningful and catches key contract regressions.
3. Contributor documentation from VM-005 is unusually explicit and accurate about known limitations.
4. `smoke-test.sh` and `test-e2e.sh` are good operational checks for supported workflow.

## Duplicated/Deprecated Surface Summary

### Notable Duplication

- Config structs duplicated (`vmmodels` and `vmcontrol`).
- JSON marshal helper duplicated with differing fallback behavior.
- REPL and run-file execution pipelines duplicated in executor.
- Test bootstrap helpers spread across multiple integration files.

### Deprecated/Confusing Surface

- Old `vm` CLI vocabulary still present in scripts though command removed.
- `modules add-*` mutators preserve legacy DB-direct behavior that bypasses daemon API.
- docs already label this as legacy, but repository still presents scripts as runnable tests.

## Alignment With VM-001 / VM-004 / VM-005

### VM-001 Alignment

- Achieved: daemonized host and API-centric command model.
- Not fully achieved: runtime hardening and legacy-surface cleanup.

### VM-004 Alignment

- Achieved: broad route-level integration and error/safety contract testing.
- Gap: safety model not fully covered (startup traversal, symlink escape) and core package tests still absent.

### VM-005 Alignment

- Achieved: high-quality onboarding document with realistic caveats.
- Gap: repository still ships broken legacy scripts that conflict with first-run expectations.

## Prioritized Cleanup Plan

### Phase 1 (Immediate safety/correctness)

1. Fix startup path validation and run-file symlink canonicalization.
2. Fix session startup-failure map lifecycle leak.
3. Fix execution-limit persistence/API contract mismatch.
4. Add typed `ErrExecutionNotFound` and 404 mapping.

### Phase 2 (Surface cleanup)

1. Migrate or archive stale library scripts.
2. Decide authoritative module/library control path (API vs direct DB).
3. Reduce terminology drift (`template`/`vm` naming consistency).

### Phase 3 (Refactor + test balance)

1. Consolidate duplicated config types/helpers.
2. Refactor executor into shared pipeline functions.
3. Add unit tests for `vmcontrol`, `vmsession`, `vmexec`, and `vmstore`.

## Implementation Plan

1. Create follow-up hardening ticket for path safety + startup lifecycle fixes.
2. Add regression tests reproducing findings from this review.
3. Create legacy-surface cleanup ticket for script migration/archive policy.
4. Create targeted refactor ticket for model/helper duplication and executor extraction.

## Open Questions

1. Should `session close` be idempotent (`200`) or strict (`404` after first close)?
2. For limit breaches, should runtime truncate and return `ok` with warning events, or fail execution state persistently?
3. Should startup-file path restrictions be relative-only with explicit disallow of symlinks, or allow controlled symlink roots?
4. Should `modules` mutating commands be removed in next major cut?

## References

- `ttmp/2026/02/07/VM-001-ANALYZE-VM--analyze-vm-system-architecture-behavior-and-quality/design-doc/01-comprehensive-vm-system-analysis-report.md`
- `ttmp/2026/02/08/VM-004-EXPAND-E2E-COVERAGE--expand-vm-system-daemon-api-e2e-and-integration-coverage/design-doc/01-daemon-api-test-coverage-matrix-and-expansion-plan.md`
- `docs/getting-started-from-first-vm-to-contributor-guide.md`
- `pkg/vmtransport/http/server.go`
- `pkg/vmcontrol/execution_service.go`
- `pkg/vmsession/session.go`
- `pkg/vmexec/executor.go`
- `pkg/vmstore/vmstore.go`
- `smoke-test.sh`
- `test-e2e.sh`
- `test-library-loading.sh`
- `test-library-requirements.sh`
- `test-goja-library-execution.sh`
