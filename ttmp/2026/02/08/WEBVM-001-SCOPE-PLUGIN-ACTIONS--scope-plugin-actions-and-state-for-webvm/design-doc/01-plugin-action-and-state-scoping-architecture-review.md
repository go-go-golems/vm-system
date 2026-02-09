---
Title: Plugin Action and State Scoping Architecture Review
Ticket: WEBVM-001-SCOPE-PLUGIN-ACTIONS
Status: active
Topics:
    - architecture
    - plugin
    - state-management
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: client/src/components/WidgetRenderer.tsx
      Note: Canonical renderer contract for UINode kind-based trees
    - Path: client/src/lib/pluginManager.ts
      Note: In-process plugin execution and ID storage behavior
    - Path: client/src/lib/pluginSandboxClient.ts
      Note: Alternate sandbox path and dispatch-guard hook gap
    - Path: client/src/lib/presetPlugins.ts
      Note: Active preset plugin API samples and action semantics
    - Path: client/src/pages/Playground.tsx
      Note: Primary active runtime and plugin load/render/event orchestration
    - Path: client/src/store/store.ts
      Note: Current reducer model and plugin action matcher behavior
    - Path: client/src/workers/pluginSandbox.worker.ts
      Note: Worker contract drift and dispatch/event path
ExternalSources: []
Summary: "Deep architecture analysis of plugin identity, action scoping, and state scoping in plugin-playground using a simplified v1 model with plugin/global selectors and plugin/global actions." 
LastUpdated: 2026-02-08T18:42:00Z
WhatFor: "Define a practical state/action scoping model without capabilities, focused on immediate implementation." 
WhenToUse: "Use when implementing v1 plugin runtime selectors, action wrappers, dispatch tracing, and plugin lifecycle behavior." 
---

# Plugin Action and State Scoping Architecture Review

## Executive Summary

This document defines a simplified v1 architecture for plugin state and action scoping.

The core model is intentionally small:

1. Two selectors:
- `selectPluginState(pluginId)`
- `selectGlobalState()`

2. Two dispatch APIs:
- `dispatchPluginAction(pluginId, type, payload)`
- `dispatchGlobalAction(type, payload)`

3. One global tracing field:
- `dispatchId` on every dispatched action.

This removes capability modeling from v1 and focuses on correctness, clarity, and migration speed.

## Problem Statement

Current behavior has three structural issues:

1. Plugin identity is inconsistent between preset IDs, plugin-declared IDs, and loader IDs.
2. State exposure is broad: plugin render/handlers receive full root state.
3. Action dispatch is unconstrained: plugins can emit arbitrary action types.

These issues make plugin scoping brittle and increase coupling risk.

## Scope of This Design

In scope:

- plugin-local state access model
- global shared state access model
- plugin/global action shape
- dispatch tracing with `dispatchId`
- migration from current code paths

Out of scope (explicitly):

- capability model
- fine-grained permission framework
- third-party trust/security marketplace model

## Current Architecture Snapshot

### Active Runtime Path

`Playground` uses `pluginManager` directly:

- load plugin with `new Function`
- render widget with full app state
- invoke handler with full app state + raw dispatch

### Alternate Runtime Path

`pluginSandboxClient` and `pluginSandbox.worker.ts` exist but are not the primary path. Their contracts drift from the active path.

### State Model Today

`store.ts` mixes plugin metadata and plugin domain states in one slice, plus a duplicate top-level `counter` reducer.

### Dispatch Model Today

`action.type.startsWith("plugin.")` matcher routes by hardcoded string branches.

## V1 Design: State Scoping

## Selector API

```ts
function selectPluginState(pluginId: string): unknown;
function selectGlobalState(): unknown;
```

### Selector Rules

1. Every plugin render/handler receives `pluginState` from `selectPluginState(pluginId)`.
2. Plugins may also receive `globalState` from `selectGlobalState()`.
3. `globalState` must be a curated projection, not raw root state.
4. Unknown plugin IDs return empty/default plugin state.

### Why This Works

- Plugin code has a clear local state entrypoint.
- Shared behavior still works through one explicit global projection.
- Coupling is reduced without adding heavy machinery.

## Suggested State Shape

```ts
interface PluginRuntimeSlice {
  registry: Record<string, {
    pluginId: string;
    packageId: string;
    status: "idle" | "loading" | "loaded" | "error";
    enabled: boolean;
    widgets: string[];
  }>;
  pluginStateById: Record<string, unknown>;
  globalState: {
    counterSummary?: { total: number };
    workspace?: { title: string; activePluginIds: string[] };
  };
}
```

Use `pluginId` as authoritative runtime key.

## V1 Design: Action Scoping

## Dispatch API

```ts
function dispatchPluginAction(
  pluginId: string,
  type: string,
  payload?: unknown,
): void;

function dispatchGlobalAction(
  type: string,
  payload?: unknown,
): void;
```

## Action Envelope

Both dispatch functions produce a common envelope:

```ts
interface ScopedAction {
  type: string;
  payload?: unknown;
  meta: {
    dispatchId: string;
    scope: "plugin" | "global";
    pluginId?: string;
    timestamp: number;
    source: "plugin-runtime";
  };
}
```

## V1 Guardrails

1. Global action type allowlist.
- `dispatchGlobalAction` may only emit approved action types.

2. Unknown plugin rejection.
- `dispatchPluginAction` throws or logs-and-drops when `pluginId` is not in registry.

3. Curated global selector output.
- Prevent accidental dependence on internal reducer structure.

4. Always stamp `dispatchId`.
- Enables deterministic trace logs and debugging.

## Pros and Cons

## Pros

1. Very simple API surface.
2. Fast to implement from current code.
3. Explicit plugin-vs-global separation.
4. Better observability with `dispatchId`.
5. Easy migration path for existing preset plugins.

## Cons

1. `global` scope remains broad unless allowlist is strict.
2. `selectGlobalState()` can become coupling hotspot.
3. No fine-grained permission semantics.
4. Not enough for untrusted third-party plugin ecosystem.

## Practical Assessment

For current project stage, this tradeoff is good:

- immediate clarity and consistency
- low refactor risk
- clear path to later hardening only if needed

## Detailed Implementation Plan

## Phase 1: Identity Cleanup

1. Make host-generated `pluginId` authoritative.
2. Store plugins only under runtime `pluginId`.
3. Keep plugin-declared IDs as metadata only.

## Phase 2: Selector Introduction

1. Add `selectPluginState(pluginId)`.
2. Add `selectGlobalState()` as curated projection.
3. Replace direct root-state passing in render handlers.

## Phase 3: Dispatch Wrappers

1. Add `dispatchPluginAction`.
2. Add `dispatchGlobalAction`.
3. Ensure `dispatchId` is always stamped.

## Phase 4: Reducer Routing

Route by `meta.scope`:

- `plugin` scope -> plugin state reducer keyed by `pluginId`
- `global` scope -> global reducer paths

## Phase 5: Preset Migration

Update preset plugin code to use wrapper APIs instead of raw action strings where possible.

## Phase 6: Dead Path Cleanup

Remove unused/duplicate runtime pieces once new path is active.

## Example Runtime Flow

### Render

```text
UI -> render(pluginId, widgetId)
   -> pluginState = selectPluginState(pluginId)
   -> globalState = selectGlobalState()
   -> plugin.render({ pluginState, globalState })
   -> UINode tree -> WidgetRenderer
```

### Event

```text
UI event -> plugin handler
        -> dispatchPluginAction(...) or dispatchGlobalAction(...)
        -> wrapper stamps dispatchId + scope
        -> Redux reducer routes by scope
        -> next render sees updated state
```

## Example API Usage

```ts
handlers: {
  increment({ dispatchPluginAction, pluginId }) {
    dispatchPluginAction(pluginId, "counter/increment");
  },
  publish({ dispatchGlobalAction }) {
    dispatchGlobalAction("workspace/publishCounter", { value: 42 });
  }
}
```

## Testing Strategy

## Unit Tests

1. `selectPluginState` returns only plugin-local state.
2. `selectGlobalState` returns curated global projection.
3. plugin dispatch stamps `dispatchId`, `scope=plugin`, and `pluginId`.
4. global dispatch stamps `dispatchId` and `scope=global`.
5. unknown plugin dispatch is rejected.

## Integration Tests

1. two plugin instances do not cross-mutate plugin-local state.
2. global action updates appear in both plugins via `selectGlobalState`.
3. every action in logs has unique `dispatchId`.

## Regression Tests

1. existing preset plugins still render.
2. existing UI interactions still work after wrapper introduction.

## Migration Guidance for Existing Files

### `client/src/pages/Playground.tsx`

- Replace raw `state` pass-through with selector outputs.
- Replace direct handler dispatch with wrapper APIs.

### `client/src/lib/pluginManager.ts`

- Keep temporary, but add wrapper injection surface.
- Stop passing raw root state and raw dispatch directly.

### `client/src/store/store.ts`

- add reducer routing by `meta.scope`
- move plugin-local state into `pluginStateById`
- define curated `globalState` projection

### `client/src/lib/presetPlugins.ts`

- migrate handlers to scoped wrapper usage
- avoid direct string dispatch when not necessary

## Open Questions

1. Should `selectGlobalState()` be one object or multiple domain selectors later?
2. Should unknown global action types throw or warn-and-drop in dev?
3. Should `dispatchId` be UUID v4 or monotonic sequence?
4. Should plugin-local state be reset on code reload by default?

## Decision Summary

Chosen v1 model:

- `selectPluginState(pluginId)`
- `selectGlobalState()`
- `dispatchPluginAction(...)`
- `dispatchGlobalAction(...)`
- global `dispatchId` on all actions

No capability model in this design.

## References

- `client/src/pages/Playground.tsx`
- `client/src/lib/pluginManager.ts`
- `client/src/store/store.ts`
- `client/src/lib/presetPlugins.ts`
- `client/src/lib/pluginSandboxClient.ts`
- `client/src/workers/pluginSandbox.worker.ts`
- `client/src/components/WidgetRenderer.tsx`
- `client/src/lib/uiTypes.ts`

