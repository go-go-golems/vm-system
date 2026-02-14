---
Title: Pre-doc architecture analysis of frontend plugin system
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
    - Path: go-go-labs/ttmp/2026/02/09/WEBVM-003-DEVX-UI-PACKAGE-DOCS-OVERHAUL--developer-ui-overhaul-reusable-vm-package-and-documentation/design-doc/02-deep-pass-refresh-current-codebase-audit-and-ui-runtime-docs-roadmap.md
      Note: Historical architecture context from predecessor ticket
    - Path: vm-system/frontend/client/src/pages/WorkbenchPage.tsx
      Note: Host orchestration source for load/render/event lifecycle analysis
    - Path: vm-system/frontend/packages/plugin-runtime/src/redux-adapter/store.ts
      Note: Capability model
    - Path: vm-system/frontend/packages/plugin-runtime/src/runtimeService.ts
      Note: Canonical runtime bootstrap API and QuickJS isolation behavior
ExternalSources: []
Summary: Code-first architecture explainer of the frontend plugin system created before reviewing frontend docs.
LastUpdated: 2026-02-14T16:43:41.296755309-05:00
WhatFor: Establish an implementation-grounded baseline to identify why the system is confusing and what docs must cover.
WhenToUse: Read before documentation updates to understand the current runtime/host architecture and lifecycle.
---


# VM-023 Pre-Doc Analysis: Frontend Plugin System (Code-First, No Docs Read)

## Method and Constraints

This analysis intentionally avoids reading `vm-system/frontend/docs/*` content.  
It is based on:

- Current implementation in `vm-system/frontend` code and tests.
- Historical context from February tickets in `go-go-labs/ttmp/2026/02` (WEBVM-001/002/003).

Goal: capture what the system is now, where confusion naturally appears, and what documentation must explain before we evaluate current docs.

## Executive Summary

The frontend plugin system is a browser-hosted, QuickJS-isolated runtime with a React workbench shell and Redux policy layer.

Current shape:

- Runtime engine + contracts are in `packages/plugin-runtime`.
- Host UI orchestration is in `client/src/pages/WorkbenchPage.tsx`.
- Policy/state/capabilities are in `packages/plugin-runtime/src/redux-adapter/store.ts`.
- Built-in plugin examples are code strings in `client/src/lib/presetPlugins.ts`.

Why it feels confusing:

- There are three mental models at once: VM runtime, Redux policy engine, and Workbench UI.
- Historical naming drift (older tickets mention `Playground.tsx` and `quickjsSandboxClient.ts`; current code uses `WorkbenchPage.tsx` and `packages/plugin-runtime/src/worker/sandboxClient.ts`) obscures continuity.
- Capability enforcement and action application are split across async event flow and reducer logic, making cause/effect hard to trace without an explicit lifecycle document.

## System Topology (As Implemented)

## 1) Runtime Package (`packages/plugin-runtime`)

Core files:

- `packages/plugin-runtime/src/contracts.ts`
- `packages/plugin-runtime/src/runtimeService.ts`
- `packages/plugin-runtime/src/worker/runtime.worker.ts`
- `packages/plugin-runtime/src/worker/sandboxClient.ts`
- `packages/plugin-runtime/src/dispatchIntent.ts`
- `packages/plugin-runtime/src/uiSchema.ts`
- `packages/plugin-runtime/src/redux-adapter/store.ts`

Responsibilities:

- Define worker request/response contract.
- Execute plugin code inside isolated QuickJS runtimes with memory/stack/time limits.
- Validate output UI trees and dispatch intents before host consumption.
- Own runtime state model and capability-gated shared-domain policy.

## 2) Host App (`client/src`)

Core files:

- `client/src/pages/WorkbenchPage.tsx`
- `client/src/store/index.ts`
- `client/src/store/workbenchSlice.ts`
- `client/src/components/WidgetRenderer.tsx`
- `client/src/lib/presetPlugins.ts`

Responsibilities:

- Provide developer workbench UI (catalog/editor/preview/devtools).
- Load/unload plugin instances through sandbox client.
- Re-render widgets when state changes.
- Route validated intents into runtime reducers.
- Present timeline/state/capabilities/errors/shared panels.

## 3) Embedded Documentation Surface in UI

Core files:

- `client/src/lib/docsManifest.ts`
- `client/src/features/workbench/components/DocsPanel.tsx`

Observation:

- Docs are bundled into the client via `?raw` imports and exposed in an in-app docs tab.
- This means documentation quality directly affects runtime learnability inside the tool itself.

## Runtime Lifecycle (What Actually Happens)

## Plugin Load

1. User opens preset/custom code in workbench.
2. `WorkbenchPage.runEditorTab()` creates instance id via `createInstanceId()`.
3. `quickjsSandboxClient.loadPlugin(packageId, instanceId, code)` sends `loadPlugin` to worker.
4. Worker calls `QuickJSRuntimeService.loadPlugin()`.
5. Runtime bootstraps DSL and plugin host bindings, executes code, returns plugin metadata.
6. Host dispatches `pluginRegistered(...)` with capability grants and initial state.

## Render

1. React effect in `WorkbenchPage` watches runtime-related dependencies.
2. For each loaded instance and widget, host computes:
   - `pluginState` from runtime slice.
   - `globalState` via `selectGlobalStateForInstance(...)`.
3. Host calls `quickjsSandboxClient.render(...)`.
4. Worker runtime executes widget render; `uiSchema` validates returned `UINode`.
5. Trees are stored in local component state for `WidgetRenderer`.

## Event

1. User interaction in `WidgetRenderer` emits `UIEventRef`.
2. Host calls `quickjsSandboxClient.event(...)` with handler and payload.
3. Runtime executes widget handler and collects intents:
   - `scope: "plugin"` intents
   - `scope: "shared"` intents with domain
4. `dispatchIntent` validation normalizes/guards intents.
5. Host applies intents:
   - plugin intents -> `dispatchPluginAction(...)`
   - shared intents -> `dispatchSharedAction(...)`
6. Runtime reducer enforces grants and records dispatch timeline outcomes (`applied`, `denied`, `ignored`).

## State and Capability Model (Current)

Runtime state includes:

- `plugins` registry by instance.
- `pluginStateById`.
- `grantsByInstance`.
- shared domains:
  - `counter-summary`
  - `greeter-profile`
  - derived: `runtime-registry`, `runtime-metrics` (projected)
- dispatch trace + timeline.

Key behavior:

- Shared writes require per-instance write grant.
- Shared reads are filtered in per-instance `globalState.shared`.
- Unsupported actions/domains become ignored outcomes, not hard runtime crashes.

## Evidence From Tests

Unit/integration/e2e coverage demonstrates key invariants:

- Runtime can load, render, event, and dispose instances.
- Infinite render loops are interrupted by timeout.
- Multiple instances of same package can run independently.
- Shared writes are blocked without grants.
- Greeter shared/local state propagation works end-to-end.

Files:

- `packages/plugin-runtime/src/runtimeService.integration.test.ts`
- `packages/plugin-runtime/src/dispatchIntent.test.ts`
- `tests/e2e/quickjs-runtime.spec.ts`

## Sources of Confusion (Code-Level)

## 1) Mixed ownership boundaries

- Runtime policy lives in package Redux adapter.
- UI orchestration lives in workbench page.
- Preset plugin source is app-local embedded code.

A newcomer must understand all three to debug one behavior.

## 2) Naming/history drift

Older ticket artifacts describe earlier paths (`Playground.tsx`, legacy file names).  
Current implementation has evolved, but docs may still reference old names unless maintained aggressively.

## 3) Hidden data transformations

- `globalState` projection is grant-filtered and per-instance.
- Shared domains include both mutable domains and derived metrics/registry.

Without a lifecycle diagram, users can misinterpret why a plugin “cannot see” or “cannot write” data.

## 4) Async render state separate from Redux

Widget trees/errors are local React state (not runtime slice), while runtime state is Redux.  
This split is legitimate, but it is non-obvious and must be documented explicitly.

## 5) DSL + reducer coupling is implicit

Plugin handlers emit generic intents; reducer decides semantics by package id and action type.  
This creates a contract that is more policy-driven than plugin-self-contained.

## Documentation Requirements Identified (Before Reading Existing Docs)

Minimum docs needed for clarity:

1. One-page architecture map (runtime package vs host app vs docs panel).
2. Sequence diagram for load/render/event/dispose flow.
3. Capability model with concrete read/write allow/deny examples.
4. Shared-domain catalog with action contracts and reducer behavior.
5. Plugin authoring contract:
   - required `definePlugin` shape
   - widget/handler signatures
   - intent semantics
6. Debugging guide:
   - where errors surface (load/render/event)
   - how timeline outcomes map to policy decisions.
7. Migration note mapping historical names to current file/module names.

## Initial Risks If Docs Are Incomplete

- High onboarding friction and oral-knowledge dependency.
- Misuse of shared domains or grants leading to “invisible” denied behavior.
- Incorrect assumptions that runtime package is fully host-agnostic when policy/UI coupling still exists.
- Confusion between app shell state (`workbenchSlice`) and runtime state (`runtimeReducer`).

## Next Step

Now that the code-first baseline is documented, the next report will inspect `vm-system/frontend/docs` itself, validate coverage against these requirements, and provide a concrete improvement plan.
