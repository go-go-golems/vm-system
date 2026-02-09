---
Title: Diary
Ticket: VM-017-NATIVE-MODULE-REGISTRY
Status: active
Topics:
    - backend
    - frontend
    - infrastructure
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ttmp/2026/02/09/VM-017-NATIVE-MODULE-REGISTRY--stop-template-configuring-js-built-ins-use-go-go-goja-registrable-native-modules/design/01-analysis-and-implementation-guide.md
    - /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/vmsession/session.go
    - /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/vmcontrol/template_service.go
    - /home/manuel/code/wesen/corporate-headquarters/go-go-goja/modules/common.go
ExternalSources: []
Summary: >
    Implementation diary for VM-017 with task-by-task progress, commits,
    failures, and verification commands.
LastUpdated: 2026-02-09T00:20:37-05:00
WhatFor: Preserve a strict execution trail while implementing VM-017 in incremental commits.
WhenToUse: Use while reviewing or continuing VM-017 work.
---

# Diary

## Goal

Track VM-017 implementation step-by-step while enforcing the requested loop: implement one task at a time, check task status, commit, and keep the reasoning trail explicit.

## Step 1: Create VM-017 ticket, analysis guide, and execution tasks

This step established the ticket workspace and implementation guardrails before touching runtime behavior. The core purpose was to define explicit acceptance criteria and commit boundaries so implementation can proceed in controlled increments.

I captured the product decision in writing: built-ins are never template-configurable, and template modules should map only to go-go-goja registrable native modules.

### Prompt Context

**User prompt (verbatim):** "so just don't make inbuilt modules configurable through templates. If JSON cannot be disabled, then it shouldn't be configurable. However, look at /home/manuel/code/wesen/corporate-headquarters/go-go-goja and use github.com/go-go-golems/go-go-goja as configurable registerable modules.

Create a new ticket for that, and add a detailed analysis guide, then create tasks and implement them one by one, checking off task, then commit, and updating the diary as you go."

**Assistant interpretation:** Create VM-017 as a new implementation ticket, document a concrete migration plan, then execute work task-by-task with checked tasks, commits, and diary updates.

**Inferred user intent:** Resolve module-policy inconsistency across runtime/API/UI and ensure the implementation process is auditable and disciplined.

**Commit (code):** a64aead — "docs(vm-017): create ticket, analysis guide, tasks, and diary scaffold"

### What I did

- Created ticket workspace:
  - `docmgr ticket create-ticket --ticket VM-017-NATIVE-MODULE-REGISTRY --title "Stop template-configuring JS built-ins; use go-go-goja registrable native modules" --topics backend,frontend,architecture`
- Added ticket documents:
  - `design/01-analysis-and-implementation-guide.md`
  - `reference/01-diary.md`
- Drafted the design guide with:
  - policy decision,
  - scoped tasks,
  - risk and validation sections.

### Why

- Needed a stable source of truth before changing runtime/API behavior.

### What worked

- Ticket scaffolding and doc creation completed cleanly.

### What didn't work

- N/A.

### What I learned

- Existing module semantics are drifted enough that a single focused ticket is justified.

### What was tricky to build

- Keeping scope constrained while still addressing backend, API, CLI, and UI alignment in one coherent migration plan.

### What warrants a second pair of eyes

- Whether legacy templates with built-in module entries should be auto-migrated or fail-fast (this plan currently chooses fail-fast).

### What should be done in the future

- Add a data migration follow-up only if real-world template data requires it.

### Code review instructions

- Review `design/01-analysis-and-implementation-guide.md` first for intended behavior and sequencing.
- Confirm task list in `tasks.md` matches commit boundaries.

### Technical details

- Ticket path:
  - `/home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ttmp/2026/02/09/VM-017-NATIVE-MODULE-REGISTRY--stop-template-configuring-js-built-ins-use-go-go-goja-registrable-native-modules`

## Step 2: Implement backend policy and runtime integration for native modules

This step implemented the behavior change itself: template modules now represent only go-go-goja registrable native modules, while JavaScript built-ins are explicitly not configurable. Runtime session startup now installs configured native modules through go-go-goja module loaders and `require()`.

I also added/updated tests to prove the new contract: `json` is rejected as template module config, `fs` is accepted, and `require(\"fs\")` works in execution when configured.

### Prompt Context

**User prompt (verbatim):** (see Step 1)

**Assistant interpretation:** Implement backend behavior first, with tests, and commit as its own task slice.

**Inferred user intent:** Make runtime/API behavior truthful and enforceable before touching UI polish.

**Commit (code):** 5fc8d65 — "feat(vm-017): enforce native module policy with go-go-goja registry"

### What I did

- Added `pkg/vmmodules/registry.go`:
  - module-name normalization,
  - built-in JS module deny-list (`json`, `math`, etc.),
  - validation against `go-go-goja/modules` registry,
  - per-session runtime module installation using `goja_nodejs/require`.
- Updated `pkg/vmcontrol/template_service.go`:
  - `AddModule` now validates module names through the registry helper.
- Updated `pkg/vmsession/session.go`:
  - session runtime now enables configured native modules at startup.
- Updated `pkg/vmtransport/http/server.go`:
  - `ErrModuleNotAllowed` now maps to `MODULE_NOT_ALLOWED` (422).
- Updated module catalog in `pkg/vmmodels/libraries.go`:
  - only native configurable modules are listed (`database`, `exec`, `fs`).
- Added/updated tests:
  - `pkg/vmcontrol/template_service_test.go` (allow/deny validation tests),
  - `pkg/vmtransport/http/server_templates_integration_test.go` (reject `json`, accept/list `fs`),
  - `pkg/vmtransport/http/server_native_modules_integration_test.go` (runtime `require(\"fs\")` + JSON builtin semantics).
- Added dependencies and tidy:
  - `github.com/go-go-golems/go-go-goja`,
  - updated `go.mod` / `go.sum`.

### Why

- The policy needed enforceable backend semantics before API/CLI/UI could be considered aligned.

### What worked

- `GOWORK=off go test ./... -count=1` passed in `vm-system/vm-system`.
- New integration test confirmed `require(\"fs\")` behavior only when module is configured.

### What didn't work

- Initial test runs failed due unrelated parent `go.work` problems and missing module dependencies.
- Exact failures:
  - parent workspace loading errors (`go.work`),
  - missing module requirements / go.sum entries,
  - attempted import of `glazehelp` module not present in `go-go-goja v0.0.4`.
- Resolution:
  - run tests with `GOWORK=off`,
  - add required module dependency and run `go mod tidy`,
  - remove unsupported `glazehelp` reference and keep registry-backed set to available modules.

### What I learned

- In this environment, explicit `GOWORK=off` is required to isolate validation from a broken parent workspace.
- go-go-goja module availability is version-dependent; catalog must match actually registered modules.

### What was tricky to build

- The hard part was integrating new registry semantics without overhauling the existing template model.
- The chosen approach keeps `exposed_modules` as storage but changes meaning to \"native configurable modules only.\"

### What warrants a second pair of eyes

- Whether fail-fast on existing templates containing built-in module names should be replaced by an auto-clean migration.

### What should be done in the future

- Add a one-time template cleanup command/migration if production data includes legacy built-in module entries.

### Code review instructions

- Start with `pkg/vmmodules/registry.go`.
- Then review call sites in `pkg/vmcontrol/template_service.go` and `pkg/vmsession/session.go`.
- Validate error mapping in `pkg/vmtransport/http/server.go`.
- Run and inspect:
  - `GOWORK=off go test ./pkg/vmtransport/http -count=1`
  - `GOWORK=off go test ./... -count=1`

### Technical details

- New runtime module behavior test:
  - `pkg/vmtransport/http/server_native_modules_integration_test.go`

## Step 3: Align API/CLI contracts and module catalog messaging

After enforcing backend behavior, I aligned API/CLI-facing contracts so user-visible language and contract tests match the new policy. This removes ambiguity around what template modules mean.

This step focused on consistency, not new runtime capability: it updated CLI wording and hardened API contract coverage for built-in module rejection.

### Prompt Context

**User prompt (verbatim):** (see Step 1)

**Assistant interpretation:** Complete API/CLI alignment as its own task and commit.

**Inferred user intent:** Ensure consumers do not see contradictory language or contract behavior after backend changes.

**Commit (code):** c16f740 — "chore(vm-017): align API and CLI module contracts"

### What I did

- Updated CLI command descriptions in:
  - `cmd/vm-system/cmd_template.go`
  - `list-available-modules` now explicitly describes configurable native modules and clarifies built-ins are always available.
- Extended API error contract integration coverage in:
  - `pkg/vmtransport/http/server_error_contracts_integration_test.go`
  - adds assertion that posting module `json` returns `MODULE_NOT_ALLOWED`.
- Re-ran verification:
  - `GOWORK=off go test ./cmd/vm-system ./pkg/vmtransport/http -count=1`
  - `GOWORK=off go test ./... -count=1`

### Why

- Behavior and language needed to converge, otherwise operators/clients would still infer that built-ins are configurable.

### What worked

- Contract tests and full test suite passed after updates.

### What didn't work

- N/A in this step.

### What I learned

- Tightening contract tests early prevents drift from reappearing through copy changes alone.

### What was tricky to build

- The main nuance was keeping this step focused on contract/alignment and not mixing in UI changes.

### What warrants a second pair of eyes

- Confirm CLI wording is clear enough for users migrating from prior built-in module assumptions.

### What should be done in the future

- Keep module policy wording centralized to avoid divergence across docs/UI/CLI.

### Code review instructions

- Review command description updates in `cmd/vm-system/cmd_template.go`.
- Confirm new error contract assertion in `pkg/vmtransport/http/server_error_contracts_integration_test.go`.

### Technical details

- New contract assertion: `POST /api/v1/templates/:id/modules` with `{ \"name\": \"json\" }` -> `422 MODULE_NOT_ALLOWED`.

## Step 4: Align UI module catalog/presets and validate frontend build

This step removed the last user-facing mismatch by updating vm-system-ui to treat template modules as native modules only. Built-in JavaScript globals are now presented as always available rather than checkbox-configurable template options.

I adjusted both data definitions and screen copy so defaults, settings screens, and reference docs all reflect the same behavior now enforced in backend.

### Prompt Context

**User prompt (verbatim):** (see Step 1)

**Assistant interpretation:** Finish UI alignment as the final task, then validate with frontend checks.

**Inferred user intent:** Ensure users cannot be misled by UI terminology after backend policy changes.

**Commit (code):** afd6e3e (vm-system-ui repo) — "chore(vm-017): align UI module semantics to native modules"

### What I did

- Updated module catalog and template presets in:
  - `vm-system-ui/client/src/lib/types.ts`
  - configurable modules now: `database`, `exec`, `fs`.
  - default template presets no longer try to configure built-in module IDs.
- Updated template detail UX copy in:
  - `vm-system-ui/client/src/pages/TemplateDetail.tsx`
  - labels/toasts now explicitly refer to \"native modules\".
  - added note that built-ins (JSON/Math/etc.) are always available.
- Updated reference docs page in:
  - `vm-system-ui/client/src/pages/Reference.tsx`
  - object model/API descriptions updated to native module semantics.
  - runtime note now distinguishes built-ins vs template-configured native modules.
- Validation:
  - `pnpm check` passed.
  - `pnpm build` passed (with existing analytics/chunk-size warnings only).

### Why

- Without UI alignment, users would still see built-in modules as template toggles and infer incorrect behavior.

### What worked

- Frontend typecheck/build passed after updates.

### What didn't work

- N/A; only non-blocking build warnings were emitted for existing analytics env placeholders and chunk size.

### What I learned

- Aligning defaults (`DEFAULT_TEMPLATE_SPECS`) was critical to avoid silent bootstrap regressions when backend rejects built-in module IDs.

### What was tricky to build

- The tricky part was ensuring language, presets, and module catalogs were all updated together so no stale UX path remained.

### What warrants a second pair of eyes

- Confirm whether the UI should fetch native module catalog dynamically from backend in a follow-up, rather than relying on static constants.

### What should be done in the future

- Add a backend endpoint for configurable native module catalog if live sync is required.

### Code review instructions

- Check `client/src/lib/types.ts` for module list and default presets.
- Check `client/src/pages/TemplateDetail.tsx` for user-facing semantics and note text.
- Check `client/src/pages/Reference.tsx` for API/object-model phrasing consistency.
- Re-run:
  - `pnpm check`
  - `pnpm build`

### Technical details

- Build warnings observed (non-blocking):
  - missing `%VITE_ANALYTICS_ENDPOINT%` and `%VITE_ANALYTICS_WEBSITE_ID%` placeholders.
