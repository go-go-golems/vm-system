---
Title: Implementation plan for configurable custom plugin permissions
Ticket: VM-024-CUSTOM-PLUGIN-PERMISSIONS
Status: active
Topics:
    - frontend
    - security
    - plugin
    - state-management
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: frontend/client/src/pages/WorkbenchPage.tsx
    - Path: frontend/client/src/store/workbenchSlice.ts
    - Path: frontend/client/src/features/workbench/components/EditorTabBar.tsx
    - Path: frontend/client/src/features/workbench/components/CapabilitiesPanel.tsx
    - Path: frontend/packages/plugin-runtime/src/redux-adapter/store.ts
ExternalSources: []
Summary: Phased implementation plan to move custom plugins to default-deny grants and add per-tab permission configuration UX.
LastUpdated: 2026-02-14T18:49:36.285618529-05:00
WhatFor: ""
WhenToUse: ""
---

# Implementation plan for configurable custom plugin permissions

## Executive Summary

Move non-preset plugins from permissive grants to explicit default-deny, then add a per-editor-tab permissions configuration UI so users can opt into shared-domain access before running custom code.

This preserves the capability model in normal workbench usage while keeping plugin experimentation possible with clear, intentional grant controls.

## Problem Statement

Current `runEditorTab` behavior for non-preset plugins grants `readShared` and `writeShared` access to all domains by default. This bypasses least-privilege expectations and lets arbitrary custom code immediately read and mutate shared runtime state.

Consequences:

- Capability model is weakened by default behavior.
- Cross-plugin state corruption is easy and accidental.
- Security semantics differ between presets (explicit) and custom tabs (implicit full access).

## Proposed Solution

Introduce explicit grant configuration for custom editor tabs, backed by workbench state, and apply those grants at run time.

Scope:

1. Default-deny custom plugins:
   - For non-preset runs, use empty grants by default:
     - `readShared: []`
     - `writeShared: []`
     - `systemCommands: []`
2. Per-tab permissions model:
   - Extend `EditorTab` state with configurable grants for custom tabs.
   - Keep preset tabs fixed to preset-defined capabilities.
3. Permissions UI in editor controls:
   - Add a compact "Permissions" control near Run/Reload.
   - Render per-domain read/write toggles for shared domains.
   - Disable write toggles for read-only domains (`runtime-registry`, `runtime-metrics`) to match reducer behavior.
4. Run path integration:
   - `runEditorTab` reads grants from active tab for custom plugins.
   - Keep existing preset grant behavior unchanged.
5. Validation:
   - Typecheck + unit tests.
   - Manual smoke tests for denied and granted shared dispatch behavior.

## Design Decisions

1. Store grants per editor tab, not globally.
   - Rationale: permissions should travel with the code in that tab and avoid cross-tab leakage.
2. Keep presets immutable.
   - Rationale: preset capabilities are part of example contract and should remain deterministic.
3. Default-deny for custom tabs.
   - Rationale: aligns with capability model and least privilege.
4. Surface grant editing in the editor toolbar.
   - Rationale: permission decisions happen at authoring/run time; users should not have to switch to devtools to configure grants.
5. Reflect runtime truth in Capabilities panel (existing).
   - Rationale: one panel remains authoritative for currently loaded instances.

## Alternatives Considered

1. Keep full grants by default and add warning text only.
   - Rejected: does not fix underlying policy weakness.
2. Global grant settings shared across all custom tabs.
   - Rejected: causes surprising coupling between unrelated tab experiments.
3. Parse grants from plugin source annotations/comments.
   - Rejected for initial pass: adds parser complexity and unclear UX for edits.
4. Prompt user with modal on every Run.
   - Rejected: high-friction loop for iterative editing.

## Implementation Plan

Phase 1: Security baseline

1. Change custom fallback grants in `WorkbenchPage` to default-deny.
2. Add tests (or assertions in integration behavior) confirming denied shared writes without explicit grants.

Phase 2: State model

1. Extend `EditorTab` in `workbenchSlice` with grant configuration.
2. Initialize:
   - preset tabs with preset grants metadata (read-only in UI),
   - custom tabs with empty grants.
3. Add reducers/actions to update tab grants.

Phase 3: UI controls

1. Add a permissions control (button/popover/panel) in `EditorTabBar`.
2. Render domain rows with read/write toggles.
3. Disable invalid writes for read-only domains.
4. Provide a clear indicator when custom tab has no grants.

Phase 4: Run path wiring

1. Update `runEditorTab` to consume tab grants for non-preset tabs.
2. Preserve existing preset grant behavior.
3. Keep dirty-state and active-instance behavior unchanged.

Phase 5: Validation + docs

1. Run `pnpm -C frontend check`, `test:unit`, `test:integration`, `build`.
2. Update plugin authoring/runtime docs to mention default-deny custom behavior and where to configure grants.
3. Add changelog/ticket diary entries.

## Open Questions

1. Should custom tab grant settings persist across reloads (localStorage), or remain session-only for now?
2. Should we allow editing preset tab grants behind an "advanced override" mode, or keep immutable permanently?
3. Should write attempts on read-only runtime domains be hidden in UI or shown disabled for educational clarity?

## References

- `frontend/client/src/pages/WorkbenchPage.tsx`
- `frontend/client/src/store/workbenchSlice.ts`
- `frontend/client/src/features/workbench/components/EditorTabBar.tsx`
- `frontend/client/src/features/workbench/components/CapabilitiesPanel.tsx`
- `frontend/packages/plugin-runtime/src/redux-adapter/store.ts`
