---
Title: Analysis and Implementation Guide
Ticket: VM-017-NATIVE-MODULE-REGISTRY
Status: active
Topics:
    - backend
    - frontend
    - architecture
DocType: design
Intent: long-term
Owners: []
RelatedFiles:
    - /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/vmsession/session.go
    - /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/vmcontrol/template_service.go
    - /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/vmmodels/libraries.go
    - /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/vmtransport/http/server.go
    - /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system-ui/client/src/lib/types.ts
    - /home/manuel/code/wesen/corporate-headquarters/go-go-goja/modules/common.go
    - /home/manuel/code/wesen/corporate-headquarters/go-go-goja/engine/runtime.go
ExternalSources: []
Summary: >
    Migrate template module semantics away from JavaScript built-ins and onto
    go-go-goja registrable native modules, with API, runtime, CLI, UI, and test
    alignment.
LastUpdated: 2026-02-09T00:09:43-05:00
WhatFor: Deliver a coherent module policy where built-ins are always available and only native modules are configurable.
WhenToUse: Use while implementing VM-017 tasks and reviewing behavior changes.
---

# VM-017 Analysis and Implementation Guide

## Problem Statement

Current template module configuration implies that built-ins such as `JSON`, `Math`, `Date`, etc. can be enabled/disabled per template. In practice, the runtime does not enforce that policy, so built-ins remain available even when not listed.

This creates contract drift:
- API/UI suggest built-ins are configurable.
- Runtime behavior proves built-ins are always present.
- Tests and docs can encode false assumptions.

## Product Decision

1. JavaScript built-ins are **not template-configurable**.
2. Template module configuration applies only to **registrable native modules** from `github.com/go-go-golems/go-go-goja/modules`.
3. Runtime session initialization enables configured native modules via go-go-goja module loaders and `require()` integration.

## Desired End State

- `template add-module json` is rejected as `MODULE_NOT_ALLOWED`.
- `template add-module fs` (or other registered native module) succeeds.
- Session runtime exposes `require()` and can load configured native modules.
- UI no longer advertises built-ins as toggleable template modules.
- Tests assert this behavior explicitly.

## Scope

In scope:
- Backend validation + runtime integration for native modules.
- API error contract updates for invalid module additions.
- CLI module catalog and text alignment.
- UI module catalog and preset alignment.
- Integration tests covering allow/deny/runtime behavior.

Out of scope:
- Redesigning template data model naming (`exposed_modules` field remains for now).
- Removing module endpoints.
- Broad permission/capability redesign outside module registration.

## Implementation Plan

## Task 1: Ticket scaffolding, analysis, and execution plan

Deliverables:
- Ticket created in docmgr.
- Design guide and diary initialized.
- Task breakdown established with commit boundaries.

Acceptance:
- Ticket has actionable task list and reviewable design doc.

## Task 2: Backend module policy + runtime integration

Deliverables:
- New adapter package around go-go-goja module registry.
- Built-in module names rejected in template add-module.
- Unknown module names rejected.
- Runtime session creation enables configured native modules using registry + `require`.
- Focused tests for acceptance/rejection/runtime usage.

Acceptance:
- Built-ins cannot be configured.
- Registered native modules can be configured and loaded.

## Task 3: API/CLI contract alignment

Deliverables:
- `ErrModuleNotAllowed` mapped to stable API envelope code.
- CLI available-modules output reflects native configurable modules.
- Docs/copy updated to avoid built-in configurability claim.

Acceptance:
- API and CLI behavior match runtime policy.

## Task 4: UI alignment and regression checks

Deliverables:
- UI module list/presets updated for native module semantics.
- Any bootstrapping behavior updated so template initialization does not attempt built-ins as modules.
- Validation pass on backend and frontend checks.

Acceptance:
- UI no longer suggests `json` as a toggleable module.
- Default template bootstrap remains healthy.

## Technical Notes

### go-go-goja integration strategy

Use `go-go-goja/modules` registry directly so template-configured module names map to real `NativeModule` registrations.

Planned backend helper responsibilities:
- list registered module names/docs,
- validate module IDs,
- reject legacy built-in IDs (`json`, `math`, etc.),
- enable selected modules on a per-session runtime through `require.Registry`.

### Backward compatibility intent

No compatibility shim for built-in toggles. This is an explicit policy correction.

Existing templates containing built-in module names will surface errors when sessions attempt to use invalid configured modules unless cleaned. A follow-up migration path can be added if needed, but this ticket intentionally enforces consistency now.

## Risk Register

1. Stored templates already contain built-in module names.
- Mitigation: fail fast with clear error and `MODULE_NOT_ALLOWED` message; optionally add migration follow-up.

2. Module registry initialization depends on module package init registration.
- Mitigation: keep explicit blank imports in adapter package for supported modules.

3. UI auto-bootstrap may fail if presets still include built-in module names.
- Mitigation: update presets in the same ticket before completion.

## Validation Strategy

Backend:
- `go test ./...`
- Assert module add/remove endpoints and session runtime behavior via integration tests.

Frontend:
- `pnpm check`
- ensure template bootstrap path no longer attempts forbidden built-in module additions.

## Review Checklist

- [ ] Built-ins are always runtime-available and never template-configurable.
- [ ] Registered native modules are template-configurable.
- [ ] Error contract for invalid modules is deterministic and documented.
- [ ] UI and docs no longer contradict runtime semantics.
