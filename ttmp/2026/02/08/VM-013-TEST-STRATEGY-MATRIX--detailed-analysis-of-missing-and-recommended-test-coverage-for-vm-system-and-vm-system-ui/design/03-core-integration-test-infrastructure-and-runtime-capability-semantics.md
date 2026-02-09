---
Title: Core Integration Test Infrastructure and Runtime Capability Semantics
Ticket: VM-013-TEST-STRATEGY-MATRIX
Status: active
Topics:
    - backend
    - frontend
    - architecture
DocType: design
Intent: long-term
Owners: []
RelatedFiles:
    - Path: vm-system/vm-system-ui/client/src/lib/api.ts
      Note: Frontend API orchestration coverage targets
    - Path: vm-system/vm-system-ui/client/src/lib/normalize.ts
      Note: Frontend normalization coverage targets
    - Path: vm-system/vm-system/pkg/vmcontrol/template_service.go
      Note: Template module and library metadata mutation semantics
    - Path: vm-system/vm-system/pkg/vmmodels/libraries.go
      Note: Builtin module and library catalog definitions
    - Path: vm-system/vm-system/pkg/vmsession/session.go
      Note: Runtime initialization and library load behavior evidence
    - Path: vm-system/vm-system/pkg/vmtransport/http/server_sessions_integration_test.go
      Note: Session lifecycle integration coverage baseline
    - Path: vm-system/vm-system/pkg/vmtransport/http/server_templates_integration_test.go
      Note: Existing API integration coverage baseline
ExternalSources: []
Summary: |
    Detailed design for high-value integration and architecture-level tests, with explicit capability semantics for JSON and lodash behavior and a phased plan to build durable, low-flake infrastructure.
LastUpdated: 2026-02-08T23:56:54-05:00
WhatFor: Build useful, architecture-aligned tests that catch real regressions in runtime/template behavior.
WhenToUse: Use when implementing VM-013 follow-up test tickets and deciding harness direction.
---



# Core Integration Test Infrastructure and Runtime Capability Semantics

## Goal

Define a test infrastructure that produces useful integration and architecture coverage for `vm-system` and `vm-system-ui`, with explicit handling of runtime capability semantics (especially JSON and lodash) so we test actual behavior instead of assumptions.

## Executive answer to the capability questions

### 1) If JSON is not loaded from template, should `JSON.stringify` fail?

With current implementation, **no**.

`JSON` is part of the base JavaScript environment provided by goja. Session creation initializes `goja.New()` and does not gate built-in objects by template module list. The template `ExposedModules` list is stored and retrieved by `TemplateService`, but runtime creation in `SessionManager` does not enforce module filtering.

Relevant implementation evidence:
- `vmsession.CreateSession` initializes goja directly and only toggles `console` plus external libraries (`pkg/vmsession/session.go`, runtime init path around lines 95-121).
- `TemplateService.AddModule/ListModules/RemoveModule` only mutates stored metadata (`pkg/vmcontrol/template_service.go`, around lines 115-162).
- `BuiltinModules` includes `json`, but this list is descriptive today, not an enforcement boundary (`pkg/vmmodels/libraries.go`, around lines 84-110).

Implication: today, "JSON disabled" is not representable through template settings.

### 2) Is it even possible to disable JSON?

Not with current architecture and API shape.

To make JSON disable-able, runtime assembly must enforce capability policy. That would require one of:
- hard runtime sandboxing logic that removes/blocks globals before user code runs,
- capability-aware wrappers/proxies for global objects,
- strict allowlist execution context where only explicit capabilities are injected.

Until that is implemented, tests should assert current truth: JSON remains available even when template modules do not include `json`.

### 3) Should lodash work only when configured?

Mostly yes, with an extra condition.

Lodash availability currently depends on:
1. Template contains `"lodash"` in `vm.Libraries`.
2. Cache file `.vm-cache/libraries/lodash.js` exists and executes successfully.

Library loading behavior is in `loadLibraries` (`pkg/vmsession/session.go`, around lines 260-292). If configured library file is missing, session creation fails before session becomes ready.

So test expectations should be:
- If template includes lodash and cache is present -> `_` should work.
- If template does not include lodash -> `_` should throw at execution time.
- If template includes lodash but cache is missing -> session creation should fail with a library-not-found startup error path.

## Why this ticket needs integration-first infrastructure

Current tests are strongest at HTTP endpoint contract level and specific execution flows. That is good baseline coverage, but gaps remain in architecture-level invariants:
- template metadata vs runtime behavior consistency,
- capability semantics (what is truly enforced),
- session bootstrap side effects and environment dependencies,
- regression protection for runtime assembly.

Useful tests must validate behavior at boundaries where bugs are expensive:
- template to runtime binding,
- session creation and startup,
- execution semantics under real configured state,
- API and UI normalization contracts.

## Quality criteria for "useful" tests

A test is useful if it:
1. Fails on meaningful regressions, not formatting or implementation detail churn.
2. Explains system behavior/invariants in names and assertions.
3. Is deterministic locally and in CI.
4. Maps to user-visible or operator-visible outcomes.
5. Has low maintenance overhead through reusable harness primitives.

Tests that are less useful:
- asserting private struct internals without behavior impact,
- over-reliance on snapshots for complex runtime payloads,
- broad end-to-end tests with weak failure localization,
- flaky timing-sensitive assertions with no event/state synchronization.

## Proposed architecture test pyramid for this codebase

### Layer A: Semantic contract tests (highest signal)

Scope:
- capability semantics (JSON/lodash/etc),
- startup behavior,
- runtime session state invariants.

Style:
- integration test through public API, plus targeted core-level tests where API is noisy.

Outcome:
- catches regressions in behavior users and templates rely on.

### Layer B: API contract tests

Scope:
- status codes, error envelopes, field shape, list/get semantics, pagination/filtering.

Style:
- existing `vmtransport/http` integration style with strict envelope assertions.

Outcome:
- protects clients (UI/CLI) from breaking changes.

### Layer C: Persistence contract tests

Scope:
- CRUD, ordering, idempotency, JSON serialization/deserialization invariants in `vmstore`.

Style:
- store-focused tests against temp SQLite DB.

Outcome:
- prevents silent data drift and replay bugs.

### Layer D: Frontend boundary tests

Scope:
- normalization (`normalize.ts`), API orchestration (`api.ts`), route-level smoke, critical flow UI behavior.

Style:
- Vitest + Testing Library + MSW.

Outcome:
- protects user workflows while keeping fast feedback.

## Core backend infrastructure design

## 1) Introduce a reusable integration harness package

Create a focused helper package for integration tests (proposal: `pkg/testkit/integration` or `pkg/vmtestkit`) with this API surface:

- `NewHarness(t *testing.T, opts ...Option) *Harness`
- `Harness.ServerURL() string`
- `Harness.Client() *http.Client`
- `Harness.Store() *vmstore.VMStore`
- `Harness.CreateTemplate(name string) TemplateRef`
- `Harness.UpdateTemplateModules(templateID string, modules []string)`
- `Harness.UpdateTemplateLibraries(templateID string, libraries []string)`
- `Harness.CreateSession(templateID, worktree, workspaceID string) SessionRef`
- `Harness.ExecREPL(sessionID, input string) ExecutionRef`
- `Harness.RequireErrorCode(resp, code string)`

Why:
- eliminate duplicate setup in current test files,
- standardize status/error assertions,
- centralize temporary DB + worktree + cache setup,
- reduce copy/paste drift.

## 2) Provide explicit library-cache controls

Current library loader resolves from relative `.vm-cache/libraries`. That is fragile if tests run under different working directories.

Harness responsibilities:
- isolate cwd per test (or per package test suite),
- create `.vm-cache/libraries` under that isolated cwd,
- provide helper `SeedLibraryJS(name, source string)` for deterministic library content,
- provide helper `ClearLibraryCache()` for missing-cache behavior tests.

This directly enables robust lodash/no-lodash scenarios.

## 3) Add semantic assertion helpers

Helpers should decode execution payloads into typed wrappers:
- `RequireExecOK`, `RequireExecError`
- `RequireValuePreview(exec, expected)`
- `RequireExceptionContains(exec, substring)`
- `RequireEventsContainTypes(execID, []string{...})`

Benefit:
- avoid ad-hoc JSON parsing in each test,
- clearer failure messages,
- stable assertions even if fields expand later.

## 4) Separate environment-dependent tests by tag or naming

Proposed suites:
- `integration` (default in `go test ./...`): no network, no external cache dependency.
- `integration_env` (opt-in): depends on predownloaded real third-party library artifacts.

Default behavior should always be hermetic and deterministic.

## Scenario-specific test design (JSON and lodash)

The following are high-value tests to add first.

## A) JSON semantics tests

### Test A1: JSON remains available without module entry

Name proposal:
- `TestRuntime_JSONGlobalAvailableWithoutTemplateJsonModule`

Setup:
1. Create template.
2. Ensure template module list excludes `json` (or leave empty).
3. Create session.
4. Execute `JSON.stringify({x: 1})`.

Expected:
- execution status `ok`, value preview includes JSON output.

Purpose:
- codifies current architecture truth and prevents accidental change without intentional migration.

### Test A2: removing `json` module metadata does not remove runtime global

Name proposal:
- `TestRuntime_RemoveJsonModule_DoesNotDisableJSONGlobal`

Setup:
1. Add then remove `json` via template module endpoints.
2. Create session after removal.
3. Execute `typeof JSON.stringify`.

Expected:
- returns `"function"`.

Purpose:
- makes metadata/enforcement gap explicit.

## B) Lodash semantics tests

### Test B1: lodash works when configured and cached

Name proposal:
- `TestRuntime_LodashConfiguredAndCached_AvailableAsUnderscore`

Setup:
1. Seed `.vm-cache/libraries/lodash.js` in harness with deterministic minimal underscore shim (or real downloaded lodash fixture).
2. Template libraries includes `lodash`.
3. Create session.
4. Execute `_.chunk([1,2,3,4],2).length` (or shim-compatible call).

Expected:
- execution `ok`, expected result.

### Test B2: lodash absent when not configured

Name proposal:
- `TestRuntime_LodashNotConfigured_UnderscoreUndefined`

Setup:
1. Keep cache seeded to avoid false negatives.
2. Template libraries empty.
3. Create session.
4. Execute `_.chunk([1,2],1)`.

Expected:
- execution error, exception contains `_ is not defined` (or equivalent goja reference error).

Purpose:
- proves template library list is behaviorally relevant.

### Test B3: configured lodash but missing cache fails session startup

Name proposal:
- `TestSessionCreate_ConfiguredLibraryMissingCache_Fails`

Setup:
1. Template libraries includes `lodash`.
2. Ensure `.vm-cache/libraries/lodash.js` absent.
3. Create session request.

Expected:
- session create fails with unprocessable/code mapped from startup failure path;
- error text includes library missing cache hint.

Purpose:
- protects operator-facing failure mode and docs accuracy.

## C) Contract-guard tests for future capability enforcement

If/when capability enforcement is implemented, add versioned behavior tests:
- legacy mode assertions (JSON always available),
- strict mode assertions (JSON blocked unless enabled),
- migration contract tests for endpoint behavior and UI representation.

This prevents silent breaking changes for existing templates.

## Recommended file organization for new backend tests

Current tests are in `pkg/vmtransport/http/*_integration_test.go`. Keep endpoint contracts there, and add a dedicated semantic test file group:

- `pkg/vmtransport/http/server_runtime_capabilities_integration_test.go`
- `pkg/vmtransport/http/server_libraries_integration_test.go`
- `pkg/vmtransport/http/testkit_harness_test.go` (or shared helper package)

If helpers become large, move to:
- `pkg/vmtestkit/harness.go`
- `pkg/vmtestkit/http_assertions.go`
- `pkg/vmtestkit/runtime_assertions.go`

Keep helper APIs behavior-oriented, not endpoint-shaped.

## Frontend test infrastructure design

Frontend currently has no automated tests; add a minimal but strong foundation first.

## 1) Vitest baseline

Add:
- `vitest.config.ts` with jsdom environment,
- `client/src/test/setup.ts` to load `@testing-library/jest-dom` and global MSW hooks,
- scripts:
  - `pnpm test`
  - `pnpm test:run`
  - `pnpm test:coverage`.

## 2) MSW API contract fixtures

Create fixtures that mirror backend envelopes and edge cases:
- template detail payload with modules/libraries,
- execution success payload with `value` event,
- execution error payload with exception,
- malformed or missing optional fields to verify normalizers.

Focus first on:
- `client/src/lib/normalize.ts` invariants,
- `client/src/lib/api.ts` template/module/library mutation orchestration.

## 3) Route smoke tests

For `App.tsx` routes:
- `/templates`, `/templates/:id`, `/sessions`, `/sessions/:id`, `/system`, `/reference`, not-found.

Assertions:
- expected page shell content appears,
- critical fallback/error UI states appear when API fails,
- no uncaught rendering errors for empty datasets.

## 4) Template detail behavior tests for libraries

High-value UI tests:
- toggling lodash calls correct mutation endpoint,
- success path updates state,
- failure path shows toast and reverts optimistic state as needed,
- info banner that changes apply to new sessions remains visible.

This ties directly to runtime semantics and user expectation.

## CI and execution strategy

### Mandatory checks (initial)

Backend:
- `go test ./...`

Frontend:
- `pnpm check`
- `pnpm test:run`

### Stage-2 checks (after stabilization)

Backend:
- `go test -race ./...` (or scoped package set if runtime is heavy)

Frontend:
- coverage threshold gating on critical files (`normalize.ts`, `api.ts`, selected pages/components)

### Flake controls

- no sleep-based synchronization; assert via explicit API state/events,
- hermetic temp dirs and cache seeding,
- avoid dependence on host network for library availability tests,
- deterministic fixture payloads with static timestamps where possible.

## Proposed rollout plan for VM-013 follow-up tickets

1. Build backend harness primitives and migrate one existing integration test file to prove API.
2. Add JSON/lodash semantic tests (A1/A2/B1/B2/B3).
3. Add vmstore contract tests for template/session/execution ordering/idempotency gaps.
4. Stand up frontend Vitest + MSW baseline and cover normalizers/routes.
5. Add targeted UI behavior tests around template library toggling and session creation flows.
6. Enable CI gates incrementally with flake burn-in window.

## Risk register and mitigations

### Risk: tests encode accidental behavior (JSON always available)

Mitigation:
- name tests explicitly as `current behavior` where needed,
- add migration notes in test comments and docs,
- prepare dual-mode assertions if strict capability enforcement lands.

### Risk: library cache path coupling causes intermittent failures

Mitigation:
- harness controls cwd and cache setup,
- avoid reliance on developer machine global cache.

### Risk: duplication across backend integration tests

Mitigation:
- central harness + assertion helpers,
- one authoritative request/response utility layer.

### Risk: frontend tests become snapshot-heavy and low-signal

Mitigation:
- behavior/assertion-first tests,
- minimize snapshot use to tiny stable components only.

## Definition of done for this design work

- Capability semantics for JSON and lodash are explicitly documented with expected test assertions.
- A concrete harness architecture is defined for backend and frontend.
- Initial high-value scenario tests are specified and prioritized.
- CI rollout and flake mitigation strategy are documented.

## Implementation checklist (copy into VM-013A..D)

- [ ] Add backend test harness package and migrate one existing test file.
- [ ] Add JSON/lodash semantic integration tests (A1/A2/B1/B2/B3).
- [ ] Add vmstore contract tests for ordering and idempotency.
- [ ] Add frontend Vitest + setup + MSW baseline.
- [ ] Add frontend normalizer + route smoke tests.
- [ ] Add template-detail library-toggle behavior tests.
- [ ] Enable CI test scripts and enforce non-flaky defaults.

## Appendix: candidate assertion snippets

```go
// JSON is currently always available in goja runtime.
exec := harness.ExecREPL(sessionID, `JSON.stringify({x: 1})`)
harness.RequireExecOK(exec)
harness.RequireValuePreview(exec, `{"x":1}`)
```

```go
// lodash should only be available when template libraries includes "lodash".
exec := harness.ExecREPL(sessionID, `_.chunk([1,2,3,4], 2).length`)
harness.RequireExecError(exec)
harness.RequireExceptionContains(exec, "is not defined")
```

```ts
// Frontend normalization contract for template detail payload.
const vm = normalizeTemplateDetail(rawPayload);
expect(vm.libraries).toEqual(["lodash"]);
expect(vm.exposedModules).toEqual([]);
```
