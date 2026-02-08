---
Title: Comprehensive vm-system-ui analysis report
Ticket: VM-002-ANALYZE-VM-SYSTEM-UI
Status: active
Topics:
    - frontend
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: vm-system/vm-system-ui/client/index.html
      Note: Analytics placeholder injection and build warnings
    - Path: vm-system/vm-system-ui/client/src/components/SessionManager.tsx
      Note: Session lifecycle UX controls and reload flow
    - Path: vm-system/vm-system-ui/client/src/components/VMConfig.tsx
      Note: Module/library selection UX and update path
    - Path: vm-system/vm-system-ui/client/src/lib/libraryLoader.ts
      Note: Dynamic CDN library loading and global checks
    - Path: vm-system/vm-system-ui/client/src/lib/vmService.ts
      Note: Core runtime simulation and mock execution semantics
    - Path: vm-system/vm-system-ui/client/src/pages/Home.tsx
      Note: Primary UI orchestration and session/execution workflow
    - Path: vm-system/vm-system-ui/server/index.ts
      Note: Static delivery runtime
    - Path: vm-system/vm-system-ui/vite.config.ts
      Note: Build/server configuration and debug collector plugin
ExternalSources: []
Summary: Deep analysis of vm-system-ui architecture, behavior, build/runtime characteristics, and improvement roadmap.
LastUpdated: 2026-02-07T20:05:00-05:00
WhatFor: Document what vm-system-ui does today, how it is built and used, and how to evolve it safely toward backend integration.
WhenToUse: Use when planning vm-system-ui stabilization, security hardening, and backend API integration.
---


# Comprehensive vm-system-ui analysis report

## Executive Summary

vm-system-ui is a polished, developer-facing React interface that simulates a JavaScript VM workflow in-browser. It has a clear visual direction, a cohesive component model, and a usable interaction loop: create/select sessions, execute code snippets, inspect event logs, and toggle VM module/library configuration. The application builds and typechecks successfully, and the architecture is approachable for contributors.

The central architectural fact is that vm-system-ui is currently a simulation, not an integrated frontend for the Go backend. Most execution semantics are implemented inside `client/src/lib/vmService.ts` as an in-memory mock. Code is executed using `new Function`, sessions are stored in in-memory maps, and library loading is performed dynamically via script tags from CDNs. This has value for demos and UX iteration, but it creates correctness, security, and contract-drift risks if interpreted as production behavior.

Several high-impact findings emerge from code and build experiments:

- Execution path uses `new Function` over user-provided code in the browser context.
- Session creation API in service ignores provided VM ID and always selects default VM.
- Configuration updates in UI are not consistently propagated into service state.
- Library metadata is duplicated across service and loader with version/source drift.
- Build emits warnings for undefined analytics env placeholders and large bundle size.

The project is therefore best understood as a strong UI prototype with a partial runtime simulator. The path forward is not a wholesale rewrite; it is explicit mode separation (`mock` vs `real`), contract alignment with backend APIs, and targeted hardening around execution and configuration flow.

## Problem Statement

vm-system-ui needs to answer one strategic question clearly: is it a demo harness, or is it the user interface for a real VM backend? Right now it occupies both roles ambiguously.

- As a demo harness, it is effective: fast feedback, rich UI, no backend dependency.
- As a production frontend, it is currently insufficient:
- Runtime semantics differ from backend.
- Security posture is weak for arbitrary execution (`new Function`).
- API contracts are implicit/internal rather than explicit/networked.

Without explicit boundaries, teams can misinterpret behavior validated in UI as backend truth, causing integration surprises and reliability issues later.

## Proposed Solution

Adopt a dual-mode frontend architecture with a strict adapter boundary:

1. `MockAdapter` (current behavior) for UI development and demos.
2. `BackendAdapter` for real API integration with vm-system runtime host.

UI components should consume an interface, not concrete mock internals. This preserves current velocity while enabling progressive migration.

### Current Architecture (Observed)

```text
+-------------------------------+
| React Pages + Components      |
| Home, SessionManager, VMConfig|
+---------------+---------------+
                |
                v
+-------------------------------+
| vmService.ts (in-memory model)|
| sessions, executions, eval    |
+---------------+---------------+
                |
                v
+-------------------------------+
| libraryLoader.ts              |
| dynamic CDN script injection  |
+-------------------------------+
```

### Target Architecture (Proposed)

```text
+----------------------------------------------+
| React UI (pure presentation + interaction)   |
+----------------------+-----------------------+
                       |
                       v
         +-------------------------------+
         | VMAdapter interface           |
         | createSession, executeREPL... |
         +---------------+---------------+
                         |
             +-----------+-----------+
             |                       |
             v                       v
+------------------------+   +--------------------------+
| MockAdapter            |   | BackendAdapter           |
| local maps + fake exec |   | HTTP/WebSocket to vm host|
+------------------------+   +--------------------------+
```

### Adapter Interface Sketch

```ts
interface VMAdapter {
  listSessions(): Promise<VMSession[]>;
  createSession(vmId: string, name?: string): Promise<VMSession>;
  setCurrentSession(id: string): Promise<void>;
  executeREPL(sessionId: string, code: string): Promise<Execution>;
  executeRunFile(sessionId: string, path: string, args?: Record<string, unknown>): Promise<Execution>;
  listExecutions(sessionId: string): Promise<Execution[]>;
  updateVMConfig(vmId: string, patch: VMConfigPatch): Promise<VMProfile>;
}
```

## Design Decisions

### Decision 1: Keep the current UI composition and visual language

The component structure and visual system are already effective. Most improvements should target service/contracts rather than large presentation rewrites.

### Decision 2: Explicitly mark mock execution mode

Mock mode is useful and should remain available, but UI must clearly indicate when users are not interacting with a real backend runtime.

### Decision 3: Remove dangerous defaults from production path

Direct browser execution of arbitrary code through `new Function` must not be part of production mode.

### Decision 4: Centralize library metadata contract

Library definitions should exist in one source of truth (or be provided by backend API), avoiding drift across service/loader constants.

## Architecture Walkthrough

### 1. App Shell and Routing

- `vm-system-ui/client/src/App.tsx` defines route shell and providers.
- Routes include `/`, `/docs`, `/overview`, and fallback 404.
- Theme is hardcoded to dark by default through `ThemeProvider`.

Strengths:

- Minimal routing complexity.
- Error boundary and global UI providers are correctly centralized.

### 2. Main UX Flow (`Home.tsx`)

`Home.tsx` orchestrates core state:

- `code`, `executions`, `sessions`, `currentSession`, `activeTab`, `vmProfile`.
- Executes actions via `vmService` calls.

Strengths:

- Simple and readable control flow.
- Clear separation across tabs (editor, sessions, logs, config).

Weaknesses:

- Session creation call uses `vmService.createSession('', name)` and does not use selected VM profile.
- VM configuration updates (`setVMProfile`) do not consistently update service-level VM state.

### 3. Runtime Simulation (`vmService.ts`)

This file is the behavioral core and also the main risk center.

Strengths:

- Domain vocabulary mirrors backend concepts (sessions, executions, events).
- Event model is expressive enough for UI debugging.
- Built-in module/library metadata offers good discoverability.

Weaknesses:

- Execution uses `new Function` with code interpolation.
- Session store and execution store are in-memory only.
- `createSession(vmId, ...)` ignores provided vmId and always uses first VM.
- `safeEval` pulls VM config from `getCurrentSession()` rather than explicit session argument.
- Library metadata duplicated and divergent from loader constants.

### 4. Dynamic Library Loader (`libraryLoader.ts`)

Strengths:

- Has de-duplication via `loadedLibraries` and `loadingPromises`.
- Uses simple async loading contract.

Weaknesses:

- Relies on global variables injected by script tag; this is brittle for ESM modules.
- `zustand` URL points to ESM entry while check expects `window.zustand` global.
- Version drift with `vmService` library constants.

### 5. Build + Delivery Layer

- Vite builds client assets to `dist/public`.
- esbuild bundles `server/index.ts` for static serving.
- `server/index.ts` provides catch-all static delivery.

Strengths:

- Simple deploy shape for static SPA.
- Fast build loop.

Weaknesses:

- Analytics placeholders in `index.html` generate warnings when env vars absent.
- Bundle warning indicates large initial payload.

## How vm-system-ui Is Built And Used

### Build and run commands

```bash
pnpm -C vm-system-ui run check
pnpm -C vm-system-ui run build
pnpm -C vm-system-ui run dev
```

### Usage model (today)

1. Open app.
2. Use default in-memory session.
3. Execute snippets in editor.
4. Inspect events and results in execution log.
5. Toggle modules/libraries in config.

Important: these actions currently simulate backend behavior; they are not backed by vm-system Go runtime APIs.

## Experimental Findings

### Experiment A: Typecheck

Command:

```bash
pnpm -C vm-system-ui run check
```

Result:

- Passed (`tsc --noEmit`).

### Experiment B: Production build

Command:

```bash
pnpm -C vm-system-ui run build
```

Result:

- Build successful.
- Warnings:
- Undefined analytics env placeholders (`%VITE_ANALYTICS_ENDPOINT%`, `%VITE_ANALYTICS_WEBSITE_ID%`).
- Non-module script bundling warning for analytics script.
- Large chunk warning (~536.75 kB JS asset before gzip).

Interpretation:

- Build pipeline works but production hardening tasks remain.

## What Is Good

- UI is coherent and purpose-driven.
- Component boundaries are understandable.
- Session/execution/event mental model aligns with backend domain language.
- Developer experience for trying examples is strong.
- Build and typecheck are healthy.

## Problems And Cleanup Opportunities

### Issue 1: Unsafe browser-side code execution path

Problem: user code is executed with `new Function` inside client runtime simulation.

Where to look:

- `vm-system-ui/client/src/lib/vmService.ts:1097`

Example:

```ts
const func = new Function(...Object.keys(context), `"use strict";\n${code}\n`);
return func(...Object.values(context));
```

Why it matters:

- This is unsafe for untrusted code and not equivalent to backend sandbox semantics.

Cleanup sketch:

```pseudo
if mode == "mock":
  gate with explicit dev-only warning banner
  disable for production builds
if mode == "real":
  send code to backend execute endpoint
```

### Issue 2: createSession ignores vmId argument

Problem: service API suggests caller can choose VM, but implementation always picks first VM.

Where to look:

- `vm-system-ui/client/src/lib/vmService.ts:876`
- `vm-system-ui/client/src/lib/vmService.ts:883`

Why it matters:

- Breaks expected behavior for multi-VM support and confuses UI-state semantics.

Cleanup sketch:

```pseudo
createSession(vmId, name):
  if vmId provided:
    target = vms.get(vmId)
  else:
    target = default
  assert target exists
  return createSessionSync(target.id,...)
```

### Issue 3: safeEval uses current session, not explicit target session

Problem: execution context derives from `getCurrentSession()` during eval instead of the target execution session.

Where to look:

- `vm-system-ui/client/src/lib/vmService.ts:1036`

Why it matters:

- Can produce wrong library/module context if execution is requested for non-current session.

Cleanup sketch:

```pseudo
safeEval(code, execution, session):
  vm = session.vm
  context = buildContext(vm)
```

### Issue 4: VMConfig updates are not integrated into service contract

Problem: UI config toggles update local state but do not call service-level update path consistently.

Where to look:

- `vm-system-ui/client/src/pages/Home.tsx:290`
- `vm-system-ui/client/src/lib/vmService.ts:1133`

Why it matters:

- User can change UI config with unclear runtime effect across sessions.

Cleanup sketch:

```pseudo
onVMConfigChange(newVM):
  setVMProfile(newVM)
  vmAdapter.updateVMConfig(newVM.id, patch)
  refresh sessions that depend on vm
```

### Issue 5: Library metadata duplication and drift

Problem: library constants are declared in both `vmService` and `libraryLoader`, and versions/sources diverge.

Where to look:

- `vm-system-ui/client/src/lib/vmService.ts:153`
- `vm-system-ui/client/src/lib/libraryLoader.ts:11`

Examples:

- Axios: `1.6.0` in service vs `1.6.2` in loader.
- Ramda: `0.29.0` vs `0.29.1`.
- Zustand loaded as ESM URL but checked via window global.

Why it matters:

- Behavior and docs can diverge silently.

Cleanup sketch:

```pseudo
sourceOfTruth = backend /shared/libraries.json
vmService + libraryLoader consume same schema
```

### Issue 6: Analytics placeholder and bundle warnings in build

Problem: analytics script placeholders are unresolved in many environments; main bundle is large.

Where to look:

- `vm-system-ui/client/index.html:20`
- build output warnings from Vite

Why it matters:

- Noise in CI, potential runtime breakage for analytics tag, and performance regressions.

Cleanup sketch:

```pseudo
if analytics env present:
  inject script tag via plugin transform
else:
  omit analytics tag
apply manual chunk strategy for heavy feature groups
```

## Alternatives Considered

### Alternative A: Keep everything mock-only indefinitely

Rejected for long-term product trajectory. Good for demos, poor for real integration confidence.

### Alternative B: Immediate full backend integration and remove mock mode

Rejected for short term. Too disruptive; would slow UI iteration and onboarding.

### Alternative C: Adapter-based dual mode (recommended)

Accepted. Preserves current velocity and enables progressive correctness.

## Implementation Plan

### Phase 1: Contract clarity and safety flags

1. Introduce `VMAdapter` interface and isolate mock implementation.
2. Add explicit mode indicator in UI header (`Mock Runtime` / `Backend Runtime`).
3. Gate `new Function` path behind mock-only build/runtime guard.

### Phase 2: Correctness fixes in mock mode

1. Respect `vmId` in `createSession`.
2. Pass explicit session context into `safeEval`.
3. Wire VMConfig updates into service update methods.
4. Consolidate library metadata source.

### Phase 3: Backend integration path

1. Implement `BackendAdapter` with endpoints matching vm-system server contracts.
2. Add event polling/streaming from real execution endpoints.
3. Add integration test suite for adapter behavior parity.

### Phase 4: Build/performance hardening

1. Env-conditional analytics injection.
2. Bundle splitting by route/feature.
3. Add CI checks for chunk budget and unresolved env placeholders.

## Open Questions

- Should mock mode remain available in production builds for demo environments?
- What exact API contract should backend expose for event streaming (poll vs WebSocket)?
- Should library loading in real mode be entirely backend-controlled (recommended) or partially client-assisted?

## References

- `vm-system-ui/client/src/App.tsx`
- `vm-system-ui/client/src/pages/Home.tsx`
- `vm-system-ui/client/src/lib/vmService.ts`
- `vm-system-ui/client/src/lib/libraryLoader.ts`
- `vm-system-ui/client/src/components/SessionManager.tsx`
- `vm-system-ui/client/src/components/VMConfig.tsx`
- `vm-system-ui/client/index.html`
- `vm-system-ui/vite.config.ts`
- `vm-system-ui/server/index.ts`

## Chapter: Conceptual Model Of The UI Runtime

A precise mental model helps separate UI concerns from runtime concerns.

- UI concerns:
- state presentation
- user interaction
- navigation and discoverability
- error/success messaging

- Runtime concerns:
- session lifecycle
- code execution semantics
- policy enforcement
- event capture/persistence

Today, vm-system-ui blends these layers in `vmService.ts`. This works for speed, but long-term maintainability improves when runtime semantics are delegated to an adapter boundary.

### State graph (today)

```text
Home.tsx local state
  ├─ code
  ├─ sessions[]
  ├─ currentSession
  ├─ executions[]
  └─ vmProfile

vmService singleton state
  ├─ vms Map
  ├─ sessions Map
  ├─ executions Map
  ├─ executionsBySession Map
  └─ currentSessionId
```

This dual-state model introduces synchronization risk: UI local state can drift from service singleton state when updates are partial.

## Chapter: Trust Boundaries And Threat Model

### Trust boundary diagram

```text
+------------------------- Browser -------------------------+
|                                                           |
|  User input code --> vmService.safeEval --> new Function  |
|                                                           |
|  Dynamic script tags --> third-party CDN JS              |
|                                                           |
+-----------------------------------------------------------+
```

### Threat categories

1. **Execution safety**: arbitrary user code is executed in browser context.
2. **Supply chain**: libraries loaded at runtime from CDN endpoints.
3. **Semantic drift**: mock behavior diverges from backend behavior, producing false confidence.
4. **State integrity**: multiple state sources can disagree.

### Security posture recommendation

- Keep mock eval strictly dev/demo-only.
- In production mode, route all execution through backend sandbox.
- Pin library metadata/checksums from trusted backend manifest instead of direct ad hoc URL constants.

## Chapter: Detailed Behavioral Sequence Analysis

### Sequence 1: App startup

```text
App mount
  -> Home useEffect
      -> vmService.listSessions()
      -> vmService.getCurrentSession()
      -> vmService.getExecutionsBySession(current)
```

If no session exists, vmService constructor creates default VM and session, so startup appears populated.

### Sequence 2: Execute snippet

```text
User clicks Run
  -> Home.handleExecute
     -> vmService.executeREPL(code)
        -> create execution(status=running)
        -> delay(100ms)
        -> safeEval(code)
        -> append events
        -> status ok/error
  -> Home appends execution to local list
```

Strength: easy-to-understand interaction.

Risk: execution semantics are mock-specific and unsafe for untrusted payloads.

### Sequence 3: Configure libraries

```text
User toggles library in VMConfig
  -> VMConfig local selectedLibraries update
  -> onUpdate(vm)
  -> Home.setVMProfile(newVm)
```

Observed gap: no consistent call to service update function for all VM config changes. This means runtime behavior may not fully reflect UI selection state.

### Sequence 4: Session reload in SessionManager

```text
User clicks Reload
  -> vmService.reloadSessionLibraries(session.id)
     -> libraryLoader.loadLibraries(vm.libraries)
```

This path can succeed/fail depending on CDN reachability and global injection assumptions.

## Chapter: Build System And Delivery Mechanics

### Build pipeline

- `vite build` compiles client and outputs to `dist/public`.
- `esbuild` bundles `server/index.ts` to `dist/index.js`.
- Express serves static assets and index fallback.

This is clean for SPA hosting. The primary missing piece is stricter production gating around env placeholders and optional scripts.

### Observed warnings and implications

1. Undefined analytics env placeholders in `index.html`.
2. Non-module analytics script bundling warning.
3. Large JS chunk warning.

The warnings are not fatal today, but each represents hidden operational debt:

- deployment config ambiguity
- analytics script behavior ambiguity
- initial load performance risk

### Hardening pseudocode for analytics injection

```pseudo
transformIndexHtml(html):
  if hasEnv(VITE_ANALYTICS_ENDPOINT) and hasEnv(VITE_ANALYTICS_WEBSITE_ID):
    inject analytics script
  else:
    remove analytics placeholders/scripts entirely
```

## Chapter: Performance And Scalability Considerations

### Frontend bundle profile

The reported main bundle (~536 kB pre-gzip) is workable for internal tooling but large for latency-sensitive environments.

### Sources of payload growth (likely)

- large UI component surface from generated UI primitives
- rich icon usage
- runtime utility dependencies
- no aggressive route-level code splitting on feature groups

### Strategy

1. Split heavy tabs (`ExecutionLogViewer`, docs pages, optional widgets) with lazy loading.
2. Define `manualChunks` in Rollup output for predictable vendor partitioning.
3. Set a CI bundle budget and fail builds that exceed agreed thresholds.

### Example chunking sketch

```ts
build: {
  rollupOptions: {
    output: {
      manualChunks: {
        ui: ['react', 'react-dom', 'wouter'],
        charts: ['recharts'],
        motion: ['framer-motion'],
      }
    }
  }
}
```

## Chapter: Data Contract Alignment With Backend

The fastest path to integration quality is a shared contract for VM/session/execution types and operations.

### Required contract surfaces

- `POST /vms`, `GET /vms`, `GET /vms/:id`
- `POST /sessions`, `GET /sessions`, `GET /sessions/:id`, `POST /sessions/:id/close`
- `POST /exec/repl`, `POST /exec/run-file`, `GET /exec/:id`, `GET /exec/:id/events`

### UI adapter mapping

```text
Home.handleExecute -> adapter.executeREPL(sessionId, code)
SessionManager.create -> adapter.createSession(vmId, name)
ExecutionLogViewer -> adapter.listExecutions(sessionId)
VMConfig -> adapter.updateVMConfig(vmId, patch)
```

### Event contract suggestion

```json
{
  "execution_id": "...",
  "seq": 12,
  "ts": "2026-02-08T01:02:03Z",
  "type": "console",
  "payload": { "level": "log", "text": "hello" }
}
```

Use this schema identically in mock and real adapters to avoid divergent UI behavior.

## Chapter: Quality Strategy And Test Plan

A practical test stack for vm-system-ui should include:

### 1. Unit tests

- VM adapter behavior (mock mode).
- Session transitions and execution record updates.
- libraryLoader de-duplication behavior.

### 2. Component tests

- Home execute flow with mocked adapter.
- SessionManager controls and state transitions.
- VMConfig toggle propagation.

### 3. Contract tests (when backend adapter exists)

- Ensure mock and real adapters produce shape-compatible outputs.

### Example test pseudocode

```ts
test('createSession respects vmId', async () => {
  const svc = new VMService();
  const [vm1, vm2] = seedTwoVMs(svc);
  const session = await svc.createSession(vm2.id, 'target');
  expect(session.vmId).toBe(vm2.id);
});
```

```ts
test('safeEval uses target session context', async () => {
  const exec = await svc.executeREPL('typeof _', sessionWithLodash.id);
  expect(exec.status).toBe('ok');
});
```

## Chapter: UX Continuity Recommendations

The UI itself is generally strong. Recommended UX adjustments are mostly clarity improvements:

1. Display current runtime mode badge (`Mock Runtime`, `Backend Runtime`).
2. Show explicit warning when executing code in mock eval mode.
3. Clarify when config changes are pending vs applied.
4. Make session-to-VM mapping visible near execution controls.

These changes reduce semantic ambiguity without changing layout fundamentals.

## Chapter: Long-Term Evolution Path

### Stage 0: Honest mock mode

- Keep current simulator.
- Label it clearly.
- Fix obvious correctness bugs (`vmId`, session context).

### Stage 1: Adapter boundary

- Introduce interfaces and move current singleton logic into `MockAdapter`.

### Stage 2: Real backend mode

- Add `BackendAdapter` implementation.
- Preserve same UI components.

### Stage 3: Hybrid verification mode

- Run both adapters in CI against shared scenario fixtures to detect drift.

### Stage 4: Production hardening

- Remove eval path from production build.
- Enforce signed/pinned library manifest strategy.
- Monitor bundle budget and runtime errors.

This staged approach protects current developer flow while steadily improving correctness and safety.
