---
Title: QuickJS Worker Replacement Detailed Analysis and Design
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
    - Path: client/src/lib/pluginManager.ts
      Note: Primary mock runtime path using new Function targeted for removal
    - Path: client/src/lib/pluginSandboxClient.ts
      Note: Bridge/client layer to replace with QuickJS worker RPC client
    - Path: client/src/lib/uiTypes.ts
      Note: Canonical UINode schema required for worker output validation
    - Path: client/src/pages/Playground.tsx
      Note: Main runtime entrypoint and migration integration point
    - Path: client/src/workers/pluginSandbox.worker.ts
      Note: Current worker path with mock execution and contract drift
    - Path: package.json
      Note: Declares quickjs-emscripten dependency for real VM isolation
ExternalSources: []
Summary: Detailed technical design for replacing the mock/new-Function plugin VM paths with a real QuickJS worker runtime, including contracts, resource controls, migration phases, and rollback strategy.
LastUpdated: 2026-02-08T18:47:00Z
WhatFor: Provide an implementation-ready blueprint for QuickJS worker isolation and controlled deprecation of mock runtime code.
WhenToUse: Use when implementing worker runtime, RPC protocols, plugin load/render/event execution, or deleting legacy VM paths.
---


# QuickJS Worker Replacement Detailed Analysis and Design

## Executive Summary

This document specifies how to replace mock plugin execution (`new Function` on host JS engine) with a real QuickJS worker implementation.

Current status:

- active path executes plugin code in host JS engine
- alternate worker path exists but still uses mock execution (`new Function`)
- runtime contracts drift across paths

Target status:

- plugin code executes only in QuickJS VM inside a dedicated worker
- host and VM communicate through explicit RPC contracts
- plugin actions and state use the v1 simplified model:
  - `selectPluginState(pluginId)`
  - `selectGlobalState()`
  - `dispatchPluginAction(...)`
  - `dispatchGlobalAction(...)`
  - global `dispatchId`
- mock execution paths are removed

## Problem Statement

### Current Gaps

1. Isolation is not real.
- `new Function` runs plugin code in host engine.
- plugin code can interact with host globals and can freeze UI thread.

2. Runtime paths diverge.
- `pluginManager` path is active.
- `pluginSandboxClient` + worker path is partial and drifted.

3. Contract mismatch.
- worker code emits `type` nodes, renderer expects `kind`.
- handler/event contracts differ between files.

4. Operational blind spots.
- no strict runtime limits on CPU/memory for active path.
- limited deterministic telemetry for VM-level failures.

### Desired Properties

1. Strong runtime boundary between host app and plugin code.
2. Deterministic, testable contracts for load/render/event.
3. Bounded resource usage with explicit failure modes.
4. Single execution path in production code.

## Current Architecture Analysis

## Active Path (Today)

```text
Playground
  -> pluginManager.loadPlugin(code)
    -> new Function(...)
  -> widget.render({state})
  -> handler({dispatch,state})
```

Characteristics:

- minimal overhead
- zero real isolation
- easy to drift because no strict interface boundary

## Alternate Worker Path (Today)

```text
PluginWidget
  -> PluginSandboxClient
    -> pluginSandbox.worker.ts
      -> new Function(...)
```

Characteristics:

- has RPC framing
- still mock execution
- shape drift from renderer contract

### Key Observation

The codebase already has pieces needed for worker messaging, but not the actual VM isolation layer.

## Target Architecture

## Component Overview

```text
Main Thread
  - QuickJSWorkerClient
  - Redux store (v1 scoped wrappers)
  - UI renderer

Web Worker
  - QuickJSRuntimeManager
  - PluginVMRegistry
  - RPC handler
  - Dispatch intent emitter
```

## Main Thread Responsibilities

1. Hold authoritative plugin registry and state selectors.
2. Send load/render/event/unload RPC requests.
3. Receive UINode trees and dispatch intents.
4. Stamp dispatch metadata including `dispatchId`.
5. Route actions via plugin/global wrappers.

## Worker Responsibilities

1. Create/own/dispose QuickJS runtimes and contexts.
2. Evaluate plugin code inside QuickJS only.
3. Execute render/event functions with provided state snapshots.
4. Emit only serializable results and dispatch intents.
5. Enforce time/memory/stack limits.

## Contracts

## RPC Request Types

```ts
type WorkerRequest =
  | { id: number; type: "load"; pluginId: string; code: string }
  | { id: number; type: "render"; pluginId: string; widgetId: string; pluginState: unknown; globalState: unknown }
  | { id: number; type: "event"; pluginId: string; widgetId: string; handler: string; event: unknown; pluginState: unknown; globalState: unknown }
  | { id: number; type: "unload"; pluginId: string }
  | { id: number; type: "health" };
```

## RPC Response Types

```ts
type WorkerResponse = {
  id: number;
  ok: boolean;
  result?: unknown;
  error?: {
    code:
      | "PLUGIN_LOAD_ERROR"
      | "PLUGIN_NOT_FOUND"
      | "WIDGET_NOT_FOUND"
      | "HANDLER_NOT_FOUND"
      | "VM_TIMEOUT"
      | "VM_MEMORY_LIMIT"
      | "CONTRACT_VIOLATION"
      | "INTERNAL_ERROR";
    message: string;
    details?: unknown;
  };
};
```

## Dispatch Intent Event

```ts
type DispatchIntentEvent = {
  type: "dispatch-intent";
  pluginId: string;
  scope: "plugin" | "global";
  actionType: string;
  payload?: unknown;
};
```

Host stamps final action:

```ts
meta: {
  dispatchId: string;
  scope: "plugin" | "global";
  pluginId?: string;
  timestamp: number;
  source: "quickjs-worker";
}
```

## UINode Contract

Worker must return canonical `kind`-based node structure used by `WidgetRenderer`.

Reject any output that does not conform.

## VM Bootstrapping Design

## Runtime Creation

On plugin load:

1. `const QuickJS = await getQuickJS()`
2. `const runtime = QuickJS.newRuntime()`
3. apply limits:
- `runtime.setMemoryLimit(...)`
- `runtime.setMaxStackSize(...)`
- `runtime.setInterruptHandler(...)`
4. `const ctx = runtime.newContext()`

## Bridge Installation

Define host bridge functions in context:

- `__hostDispatchPlugin(type, payload)`
- `__hostDispatchGlobal(type, payload)`
- `__hostNow()`

These functions only serialize data and emit worker events; they do not expose host objects.

## Plugin Loader Bootstrap

Install a bootstrap script once per context:

- internal `definePlugin` capture
- registry object for plugin definition
- helper wrappers for render/event invocation
- error normalization

Then evaluate plugin code in VM.

## VM Execution Design

## Load

- eval plugin code
- assert `definePlugin` called
- validate schema (id/widgets/handlers)
- store VM handles in registry keyed by `pluginId`

## Render

- set per-call deadline
- call VM helper with `widgetId`, `{ pluginState, globalState }`
- validate returned tree as canonical UINode
- return tree

## Event

- set per-call deadline
- call handler with `event`, `pluginState`, `globalState`
- bridge emits dispatch intents
- return success/failure

## Unload

- dispose context handles
- dispose context
- dispose runtime
- remove plugin from worker registry

## Resource and Failure Controls

## CPU Timeout

Use interrupt handler with absolute deadline per call.

Recommended initial values:

- render timeout: 50ms
- event timeout: 50ms
- load timeout: 500ms

## Memory and Stack Limits

Recommended initial limits per plugin runtime:

- memory: 16MB
- stack: 512KB

Tune after profiling with real plugins.

## Structured Error Policy

Map VM failures to stable error codes.

Examples:

- interrupted execution -> `VM_TIMEOUT`
- allocation failure -> `VM_MEMORY_LIMIT`
- malformed tree -> `CONTRACT_VIOLATION`

## Telemetry

For every worker request:

- pluginId
- op type
- start/end timestamp
- elapsed ms
- success/failure code
- interrupted flag

For every dispatch intent:

- dispatchId (host side)
- pluginId
- scope
- actionType

## Migration Plan (Detailed)

## Phase 0: Freeze and Guard

1. Add repo note: no new `new Function` paths.
2. Add test asserting production runtime path does not use `new Function`.

## Phase 1: Build QuickJS Worker Runtime

1. Add `quickjsRuntime.worker.ts`.
2. Implement runtime manager and plugin VM registry.
3. Implement RPC request/response handling.

## Phase 2: Build Main-thread Client

1. Add `quickjsWorkerClient.ts`.
2. Promise-based request tracking by RPC `id`.
3. Dispatch intent event subscription.

## Phase 3: Integrate with Playground

1. Replace `pluginManager` calls with client RPCs.
2. On render/event use v1 selectors:
- `selectPluginState(pluginId)`
- `selectGlobalState()`
3. Route intents via action wrappers.

## Phase 4: Contract Stabilization

1. enforce `kind` schema in worker
2. align event payload signatures
3. remove drifted code paths

## Phase 5: Remove Mock Runtime

Delete/retire:

- `client/src/lib/pluginManager.ts` runtime execution path
- `window.definePlugin` in `pluginSandboxClient.ts`
- `new Function` inside `pluginSandbox.worker.ts`

## Phase 6: Hardening and Rollout

1. timeout tests
2. memory limit tests
3. malformed output tests
4. stress tests with repeated load/unload cycles

## Backward Compatibility Strategy

For existing preset plugins:

- keep `definePlugin(({ ui, createActions }) => ...)` shape initially
- bridge `createActions` to wrapper-friendly action types
- migrate presets incrementally

Compatibility boundary is runtime interface, not raw eval behavior.

## Risks and Mitigations

## Risk 1: Serialization Overhead

Mitigation:

- keep `globalState` curated and small
- avoid sending full root state
- memoize selector outputs when possible

## Risk 2: Handle Leaks in Worker

Mitigation:

- strict wrapper utilities around handle lifetimes
- explicit disposal audits in tests
- teardown tests with many cycles

## Risk 3: Migration Breakages

Mitigation:

- feature flag the new client
- run both paths in non-prod for shadow comparisons briefly
- switch default only after parity tests pass

## Risk 4: False Confidence from “QuickJS” Label

Mitigation:

- enforce that production path uses worker runtime only
- CI grep/assert for forbidden mock execution patterns

## Testing Plan

## Unit

1. request/response correlation logic
2. dispatch intent translation and metadata stamping
3. schema validators for UINode and RPC payloads

## Integration

1. load + render + event happy path
2. unknown plugin/widget/handler error paths
3. timeout interruption scenario
4. memory pressure scenario

## End-to-End

1. existing preset plugins run on worker path
2. UI remains responsive during heavy plugin loops
3. no `new Function` path used in production bundle

## Rollback Plan

If critical regressions appear after cutover:

1. keep feature flag to revert to previous path temporarily
2. preserve compatibility adapters for one release window
3. collect failing plugin payloads + VM traces
4. patch worker runtime and re-enable

Note: rollback should be time-boxed; objective remains full removal of mock runtime.

## Acceptance Criteria

1. Plugin code executes in QuickJS worker only.
2. Main-thread `new Function` runtime path removed.
3. Worker `new Function` mock execution removed.
4. UI consumes canonical `kind` node contract from worker.
5. All dispatched actions include `dispatchId`.
6. Tests cover timeout, memory, malformed output, and unload leaks.

## Open Questions

1. runtime-per-plugin vs pooled-runtime performance tradeoff in your workload?
2. do we need module imports in plugin scripts at v1 launch?
3. how strict should the global action allowlist be initially?
4. should plugin runtime restart automatically after repeated failures?

## References

- `client/src/lib/pluginManager.ts`
- `client/src/lib/pluginSandboxClient.ts`
- `client/src/workers/pluginSandbox.worker.ts`
- `client/src/pages/Playground.tsx`
- `client/src/lib/uiTypes.ts`
- `client/src/components/WidgetRenderer.tsx`
- `package.json`
- `node_modules/.pnpm/quickjs-emscripten@0.23.0/node_modules/quickjs-emscripten/README.md`
- `node_modules/.pnpm/quickjs-emscripten@0.23.0/node_modules/quickjs-emscripten/dist/runtime.d.ts`
- `node_modules/.pnpm/quickjs-emscripten@0.23.0/node_modules/quickjs-emscripten/dist/context.d.ts`

