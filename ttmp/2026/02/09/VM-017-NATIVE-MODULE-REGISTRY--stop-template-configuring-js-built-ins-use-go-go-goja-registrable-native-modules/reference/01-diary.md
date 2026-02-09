---
Title: Diary
Ticket: VM-017-NATIVE-MODULE-REGISTRY
Status: active
Topics:
    - backend
    - frontend
    - architecture
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
LastUpdated: 2026-02-09T00:16:41-05:00
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

**Commit (code):** pending (filled after backend commit)

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
