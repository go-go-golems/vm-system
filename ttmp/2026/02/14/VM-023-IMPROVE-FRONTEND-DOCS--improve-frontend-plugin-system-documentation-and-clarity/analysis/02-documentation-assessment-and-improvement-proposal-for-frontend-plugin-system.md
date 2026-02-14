---
Title: Documentation assessment and improvement proposal for frontend plugin system
Ticket: VM-023-IMPROVE-FRONTEND-DOCS
Status: active
Topics:
    - frontend
    - architecture
    - integration
DocType: analysis
Intent: long-term
Owners: []
RelatedFiles:
    - Path: vm-system/frontend/docs/architecture/ui-dsl.md
      Note: Primary contract reference with signature mismatches
    - Path: vm-system/frontend/docs/plugin-authoring/examples.md
      Note: Copy-paste examples assessed against runtime API
    - Path: vm-system/frontend/docs/plugin-authoring/quickstart.md
      Note: Author onboarding guide assessed for correctness
    - Path: vm-system/frontend/docs/runtime/embedding.md
      Note: Embedding guidance assessed for adapter contract accuracy
    - Path: vm-system/frontend/packages/plugin-runtime/src/hostAdapter.ts
      Note: Ground truth for runtime host adapter interface
    - Path: vm-system/frontend/packages/plugin-runtime/src/runtimeService.ts
      Note: Ground truth for ui helpers and runtime behavior
ExternalSources: []
Summary: Severity-ranked assessment of vm-system frontend docs with concrete contract fixes and improvement roadmap.
LastUpdated: 2026-02-14T16:43:41.399960091-05:00
WhatFor: Identify correctness gaps and provide a practical plan to make frontend plugin docs reliable for onboarding and embedding.
WhenToUse: Use when planning or executing documentation fixes for frontend plugin/runtime contracts.
---


# VM-023 Documentation Assessment and Improvement Proposal

## Scope

This report evaluates `vm-system/frontend/docs/*` against the implementation currently in:

- `packages/plugin-runtime/*`
- `client/src/*`

It follows the pre-doc baseline in:

- `analysis/01-pre-doc-architecture-analysis-of-frontend-plugin-system.md`

## Overall Assessment

Current docs are strong in intent and structure, but they are not yet reliable as executable source-of-truth for plugin authors.

Summary grade: `B-` (structure good, correctness issues significant).

Key outcome:

- Several high-severity API mismatches would cause copy/paste plugin examples to fail or behave incorrectly.
- Runtime embedding guidance includes a contract mismatch around `RuntimeHostAdapter`.
- Core architectural concepts are present, but some behavior details are currently inaccurate or underspecified.

## File-by-File Review

## `docs/README.md`

Strengths:

- Good audience framing.
- Clear doc map.
- Strong conceptual introduction.

Gaps:

- No explicit warning that docs must match runtime bootstrap DSL exactly.
- No “quick reality check” snippet pointing to authoritative API locations (`runtimeService.ts` bootstrap and contracts).

## `docs/plugin-authoring/quickstart.md`

Strengths:

- Good narrative onboarding.
- Useful handler/context explanation.
- Useful devtools guidance.

Critical issues:

- Uses `ui.input({ value, ... })` examples, but runtime DSL implements `ui.input(value, props)`.
- Uses patterns relying on `ui.column(...)`, but runtime DSL currently has no `ui.column` helper.

Impact:

- New users copying examples can get runtime errors (`ui.column is not a function`) or invalid widgets.

## `docs/plugin-authoring/examples.md`

Strengths:

- Broad scenario coverage.
- Teaches common state patterns.

Critical issues:

- Same signature mismatch issues as quickstart:
  - `ui.input({...})` instead of `ui.input(value, props)`
  - `ui.table({...})` instead of `ui.table(rows, { headers })`
  - heavy use of `ui.column(...)` with no matching runtime helper.

Impact:

- “Ready-to-paste” promise is currently unreliable.

## `docs/architecture/ui-dsl.md`

Strengths:

- Correct conceptual explanation of data-only UI trees.
- Good node-by-node reference format.

Critical issues:

- Documents DSL surface not matching runtime bootstrap implementation:
  - documents `ui.column(children)` but bootstrap lacks it
  - documents `ui.input(options)` but bootstrap expects `(value, props)`
  - documents `ui.table(options)` but bootstrap expects `(rows, props)`.

Impact:

- This is intended as reference; mismatch here propagates errors into all other docs.

## `docs/architecture/dispatch-lifecycle.md`

Strengths:

- Good sequence narrative.
- Good outcomes terminology (`applied/denied/ignored`).

Gaps:

- Says host re-renders “affected widgets”; current `WorkbenchPage` implementation re-renders all loaded plugin widgets on dependency changes.
- Should explicitly note that widget trees/errors are React local state, not runtime slice state.

## `docs/architecture/capability-model.md`

Strengths:

- Clear policy explanation.
- Good grant/outcome terminology.

Gaps:

- Domain “state shape” can be interpreted as full internal storage exposure, while `selectGlobalStateForInstance` provides projected subsets for certain domains (for example, `counter-summary` omits `valuesByInstance`).
- Should distinguish internal domain model vs per-instance projected model.

## `docs/runtime/embedding.md`

Strengths:

- Best architecture map in current docs.
- Good host loop explanation.

Critical issues:

- “Mode C” implies `QuickJSRuntimeService` and `QuickJSSandboxClient` are drop-in `RuntimeHostAdapter` implementations, but signatures do not match the adapter interface directly.

Gaps:

- Lacks explicit adapter-wrapper examples for both direct and worker modes.

## `docs/migration/changelog-vm-api.md`

Strengths:

- Useful import migration map.
- Clear compatibility statement.

Gaps:

- No mention of current DSL signature realities (`input`, `table`, absence of `column`) as migration-sensitive details.

## High-Severity Findings (Ordered)

## P0-1: UI DSL contract mismatch across docs vs runtime

Evidence:

- Runtime bootstrap (`packages/plugin-runtime/src/runtimeService.ts`) exposes:
  - `input(value, props = {})`
  - `table(rows = [], props = {})`
  - no `column` helper
- Docs teach:
  - `ui.input({...})`
  - `ui.table({...})`
  - `ui.column([...])`

Why this matters:

- Breaks trust in docs and causes immediate onboarding failure.

## P0-2: Embedding guide overstates adapter compatibility

Evidence:

- `RuntimeHostAdapter` (`packages/plugin-runtime/src/hostAdapter.ts`) uses object-based method signatures.
- `QuickJSSandboxClient` (`packages/plugin-runtime/src/worker/sandboxClient.ts`) methods use positional arguments.
- `QuickJSRuntimeService` also differs in method shape and sync/async behavior.

Why this matters:

- Embedders may implement against incorrect assumptions and hit integration churn.

## P1-1: Dispatch lifecycle detail mismatch (affected vs all renders)

Why this matters:

- Performance/debug mental model becomes inaccurate.

## P1-2: Capability model lacks explicit projection semantics

Why this matters:

- Plugin authors may expect fields that are not exposed in `globalState.shared`.

## P2-1: Missing troubleshooting and runtime error cookbook

Why this matters:

- Current docs explain concepts, but not fast diagnosis for common runtime failures.

## Proposed Improvements

## Immediate Fixes (P0)

1. Update `docs/architecture/ui-dsl.md` to match runtime bootstrap exactly.
2. Update all plugin snippets in:
   - `docs/plugin-authoring/quickstart.md`
   - `docs/plugin-authoring/examples.md`
3. Decide one of two paths for `column`:
   - Path A (preferred for ergonomics): add `ui.column(children)` helper in runtime bootstrap and keep docs.
   - Path B: remove `ui.column` from docs and use `ui.panel`/`ui.row`/manual node objects.
4. Revise `docs/runtime/embedding.md` Mode C:
   - clarify wrappers are needed
   - provide concrete wrapper examples.

## Clarity Fixes (P1)

1. `dispatch-lifecycle.md`: align re-render description with current host behavior.
2. `capability-model.md`: add “internal vs projected state” section with `counter-summary` example.
3. `README.md`: add “API truth source” callout (bootstrap and contracts paths).

## Completeness Fixes (P2)

Add new docs:

1. `docs/runtime/troubleshooting.md`
   - load/render/event failure taxonomy
   - timeout errors
   - denied/ignored intent debugging checklist.
2. `docs/architecture/current-host-loop.md`
   - explicit map of `WorkbenchPage` + runtime package boundary.
3. `docs/plugin-authoring/contract-cheatsheet.md`
   - canonical signatures and minimal valid plugin template.

## Concrete Edit Sketches

## `ui-dsl.md` corrections

- Replace:
  - `ui.input({ value, placeholder, onChange })`
  with:
  - `ui.input(value, { placeholder, onChange })`
- Replace:
  - `ui.table({ headers, rows })`
  with:
  - `ui.table(rows, { headers })`
- Explicitly state whether `ui.column` is available; if not, remove it from reference/examples.

## `embedding.md` Mode C corrections

Add wrapper example:

```ts
const workerAdapter: RuntimeHostAdapter = {
  loadPlugin: ({ packageId, instanceId, code }) =>
    sandbox.loadPlugin(packageId, instanceId, code),
  render: ({ instanceId, widgetId, pluginState, globalState }) =>
    sandbox.render(instanceId, widgetId, pluginState, globalState),
  event: ({ instanceId, widgetId, handler, args, pluginState, globalState }) =>
    sandbox.event(instanceId, widgetId, handler, args, pluginState, globalState),
  disposePlugin: (instanceId) => sandbox.disposePlugin(instanceId),
  health: () => sandbox.health(),
  terminate: () => sandbox.terminate(),
};
```

## Validation Checklist for Doc Refresh

After updates, validate with:

1. Copy/paste run each example from `quickstart.md` and `examples.md` in playground.
2. Confirm no `ui.* is not a function` runtime errors.
3. Confirm `table` and `input` examples render correctly.
4. Confirm embedding snippets type-check against current exported types.
5. Confirm docs panel can render all updated markdown via `docsManifest`.

## Suggested Execution Plan

1. Patch reference contracts first (`ui-dsl.md`, `embedding.md`).
2. Patch authoring docs (`quickstart.md`, `examples.md`).
3. Add troubleshooting + host-loop docs.
4. Run smoke validation with one custom plugin and one shared-write-denied scenario.
5. Update migration changelog with these doc contract corrections.

## Bottom Line

The documentation set already has good structure and intent, but contract correctness must be fixed first.  
Once P0 issues are resolved, the docs can become a reliable onboarding system rather than a conceptual overview that still requires oral guidance.
