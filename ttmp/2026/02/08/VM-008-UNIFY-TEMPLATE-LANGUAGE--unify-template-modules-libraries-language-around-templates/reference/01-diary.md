---
Title: Diary
Ticket: VM-008-UNIFY-TEMPLATE-LANGUAGE
Status: active
Topics:
    - backend
    - docs
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: pkg/vmclient/templates_client.go
      Note: Task 4 vmclient template module/library methods
    - Path: pkg/vmcontrol/ports.go
      Note: Task 2 template store port extension
    - Path: pkg/vmcontrol/template_service.go
      Note: Task 2 template module/library domain methods
    - Path: pkg/vmcontrol/template_service_test.go
      Note: Task 2 test stub update for port change
    - Path: pkg/vmtransport/http/server.go
      Note: Task 2 template module/library API endpoints
    - Path: ttmp/2026/02/08/VM-008-UNIFY-TEMPLATE-LANGUAGE--unify-template-modules-libraries-language-around-templates/changelog.md
      Note: |-
        Task-level change record
        Task 1 changelog entry
        Task 2 changelog entry
        Task 4 changelog entry
    - Path: ttmp/2026/02/08/VM-008-UNIFY-TEMPLATE-LANGUAGE--unify-template-modules-libraries-language-around-templates/design-doc/01-template-language-unification-review-and-implementation-plan.md
      Note: |-
        Task 1 terminology contract finalized in design doc
        Task 1 finalized terminology contract
    - Path: ttmp/2026/02/08/VM-008-UNIFY-TEMPLATE-LANGUAGE--unify-template-modules-libraries-language-around-templates/tasks.md
      Note: |-
        Task checklist progress tracking
        Task 1 checklist update
        Task 2 checklist update
        Task 4 checklist update
ExternalSources: []
Summary: Implementation diary for VM-008 template language unification work.
LastUpdated: 2026-02-08T13:25:00-05:00
WhatFor: Preserve exact VM-008 implementation sequence, decisions, issues, and validation evidence.
WhenToUse: Use when reviewing VM-008 implementation details or reproducing task-by-task outcomes.
---




# Diary

## Goal

This diary captures VM-008 implementation task-by-task so language/API/CLI/doc cleanup decisions are traceable, reviewable, and reproducible.

## Step 1: Finalize template-only terminology contract before code migration

I started VM-008 by formalizing the terminology contract in the design document before touching code. This locks what language is allowed on user-facing surfaces and what remains acceptable internal wording.

Establishing this early prevents mixed-language regressions while implementing API and CLI changes in later tasks.

### Prompt Context

**User prompt (verbatim):** "ok, now tackle VM-008"

**Assistant interpretation:** Begin VM-008 implementation immediately and execute the ticket task-by-task.

**Inferred user intent:** Move directly into implementation mode for VM-008 with the same disciplined delivery quality used in VM-007.

**Commit (code):** c3c46c6 — "vm008: finalize template terminology contract (task 1)"

### What I did

- Created VM-008 diary document (`reference/01-diary.md`).
- Updated design doc with `Terminology Contract (Finalized)` section defining template-centric naming rules and explicit no-alias/no-compatibility stance for legacy modules command language.
- Clarified the out-of-scope boundary for internal type renames (retain internal model naming where not user-facing).

### Why

The first task in `tasks.md` is terminology-contract finalization. Doing this first gives unambiguous acceptance criteria for all following API/CLI/doc edits.

### What worked

- The ticket already had a strong design baseline; adding a finalized contract section made the implementation constraints explicit and testable.

### What didn't work

- N/A for this step.

### What I learned

- The key ambiguity in this ticket is not technical capability but naming boundary: user-facing language must be strict even if internal models still use legacy names.

### What was tricky to build

The subtle part was balancing strict template-only user language with practical internal constraints so the contract is enforceable without forcing a large internal rename scope in this ticket.

### What warrants a second pair of eyes

- Confirm that the finalized contract wording is strict enough for reviewer expectations on no-compatibility cleanup.

### What should be done in the future

- Proceed with template API/service/client/CLI changes under this contract; reject any `vm-id`/`modules` user-facing reintroductions.

### Code review instructions

- Start with:
  - `ttmp/2026/02/08/VM-008-UNIFY-TEMPLATE-LANGUAGE--unify-template-modules-libraries-language-around-templates/design-doc/01-template-language-unification-review-and-implementation-plan.md`
- Validate contract presence and wording under `Terminology Contract (Finalized)`.

### Technical details

- Contract explicitly distinguishes user-facing naming rules from internal runtime/model naming scope.

## Step 2: Add template module/library API endpoints backed by domain service methods

I implemented template-native module/library endpoint support in the daemon API and wired the domain layer to persist these operations through the template service/store path. This removes the architectural need for command-side direct DB mutation for module/library operations.

I intentionally placed module/library mutation operations under template routes to align with the finalized terminology contract and to establish a single control plane for these mutations.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Continue VM-008 implementation with concrete API foundation so module/library changes are template-owned daemon operations.

**Inferred user intent:** Replace legacy mixed command/control behavior with template-first API semantics before CLI migration.

**Commit (code):** 786d482 — "vm008: add template module/library API endpoints (task 2)"

### What I did

- Extended `TemplateStorePort` with `UpdateVM` to support persisted module/library mutations.
- Added template service methods:
  - `ListModules`, `AddModule`, `RemoveModule`
  - `ListLibraries`, `AddLibrary`, `RemoveLibrary`
- Added HTTP routes:
  - `GET/POST/DELETE /api/v1/templates/{template_id}/modules[/{module_name}]`
  - `GET/POST/DELETE /api/v1/templates/{template_id}/libraries[/{library_name}]`
- Added corresponding server handlers with validation/error mapping:
  - required `name` on POST
  - required path name on DELETE
  - template-id validation and core error mapping
- Updated template service unit test stub to satisfy the expanded store port (`UpdateVM`).
- Ran validation:
  - `GOWORK=off go test ./pkg/vmcontrol ./pkg/vmtransport/http -count=1`
  - `GOWORK=off go test ./... -count=1`

### Why

Task 2 requires template API parity for module/library mutation/query. Without this foundation, CLI migration would either stay direct-DB or need temporary compatibility surfaces.

### What worked

- API route wiring and service methods integrated cleanly with existing core architecture.
- Full repository tests passed after adding service/store contract method.

### What didn't work

- Initial compile failed because `templateStoreStub` in `template_service_test.go` no longer satisfied `TemplateStorePort` after adding `UpdateVM`.
- Fix: implemented `UpdateVM` in the test stub and reran tests.

### What I learned

- Expanding shared ports has immediate downstream effects on existing test stubs; updating stubs early keeps task loops tight.

### What was tricky to build

The main design choice was deciding where mutation semantics should live. I placed them in template service methods (not handler-local logic) to keep domain behavior centralized and reusable for client/CLI layers.

### What warrants a second pair of eyes

- Confirm endpoint response shape for module/library POST/DELETE operations is acceptable (`template_id` + `name` + status on deletes).

### What should be done in the future

- Extend vmclient with these new template module/library routes (Task 4), then move CLI to those client APIs.

### Code review instructions

- Start with:
  - `pkg/vmtransport/http/server.go`
  - `pkg/vmcontrol/template_service.go`
  - `pkg/vmcontrol/ports.go`
- Validate with:
  - `GOWORK=off go test ./pkg/vmcontrol ./pkg/vmtransport/http -count=1`
  - `GOWORK=off go test ./... -count=1`

### Technical details

- Service add/remove operations are idempotent: adding an existing item or removing a missing item is a no-op success.

## Step 3: Extend vmclient template routes for module/library operations

I added template module/library client operations to `pkg/vmclient/templates_client.go` so command layers can call daemon template routes directly rather than writing datastore mutation logic in CLI commands.

This step keeps route contracts aligned between daemon and client and sets up the remaining CLI migration tasks to stay API-first with no legacy mutation wrappers.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Continue VM-008 by adding vmclient coverage for the newly added template module/library endpoints.

**Inferred user intent:** Ensure module/library operations are available through standard client pathways before replacing legacy command code.

**Commit (code):** Pending for Step 3 commit creation.

### What I did

- Added request models:
  - `AddTemplateModuleRequest`
  - `AddTemplateLibraryRequest`
- Added shared response model:
  - `TemplateNamedResourceResponse`
- Added client methods:
  - `ListTemplateModules`, `AddTemplateModule`, `RemoveTemplateModule`
  - `ListTemplateLibraries`, `AddTemplateLibrary`, `RemoveTemplateLibrary`
- Verified task-relevant test loops:
  - `GOWORK=off go test ./pkg/vmclient ./pkg/vmtransport/http -count=1`
  - `GOWORK=off go test ./... -count=1`

### Why

Task 4 requires explicit vmclient coverage so CLI/task workflows can depend on a consistent API abstraction and not introduce direct store writes or route duplication.

### What worked

- The client extension integrated cleanly with existing `Client.do` request plumbing.
- Existing route contracts from Task 2 mapped directly with straightforward request/response structures.

### What didn't work

- N/A in this step.

### What I learned

- A small shared response type (`template_id`, `name`, optional `status`) keeps add/remove flows symmetric and reduces future parsing drift.

### What was tricky to build

The key point was keeping method naming and route composition strictly template-centric while preserving compatibility with existing client conventions (context-first signatures, `Client.do` delegation, and concrete response decoding).

### What warrants a second pair of eyes

- Confirm response payload expectations for DELETE routes remain stable (`status` field optionality).

### What should be done in the future

- Continue with CLI migration so all module/library operations flow through these vmclient methods.

### Code review instructions

- Start with:
  - `pkg/vmclient/templates_client.go`
- Validate with:
  - `GOWORK=off go test ./pkg/vmclient ./pkg/vmtransport/http -count=1`
  - `GOWORK=off go test ./... -count=1`

### Technical details

- Methods target:
  - `/api/v1/templates/{template_id}/modules`
  - `/api/v1/templates/{template_id}/libraries`
- DELETE methods target resource-name path variants and decode into `TemplateNamedResourceResponse`.
