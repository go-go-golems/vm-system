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
    - Path: cmd/vm-system/cmd_modules.go
      Note: |-
        Task 3 removed command-side store mutation logic
        Task 6 legacy command deleted
    - Path: cmd/vm-system/cmd_session.go
      Note: Task 7 template wording cleanup for session output labels
    - Path: cmd/vm-system/cmd_template.go
      Note: Task 5 template module/library and catalog subcommands
    - Path: cmd/vm-system/main.go
      Note: Task 6 removed modules command registration
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
        Task 3 changelog entry
        Task 5 changelog entry
        Task 6 changelog entry
        Task 7 changelog entry
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
        Task 3 checklist status
        Task 5 checklist update
        Task 6 checklist update
        Task 7 checklist update
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

**Commit (code):** a6febe2 — "vm008: extend vmclient template module/library operations (task 4)"

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

## Step 4: Remove modules command mutation internals and route through template client APIs

I completed removal of command-local persistence mutation logic in `cmd_modules.go` by replacing add-module and add-library direct datastore updates with vmclient template API calls. This keeps mutation behavior centralized in daemon service/store code paths.

This change keeps architecture consistent with VM-008 goals while preserving command behavior shape until later tasks remove the legacy `modules` command surface entirely.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Implement Task 3 by eliminating ad-hoc command-side mutation logic and rely on template service paths.

**Inferred user intent:** Ensure no direct persistence mutation remains in CLI command code for template module/library operations.

**Commit (code):** e57ce0e — "vm008: remove modules command-side mutation logic (task 3)"

### What I did

- Updated `cmd/vm-system/cmd_modules.go`:
  - removed command-local `vmstore` update flow for `add-module`
  - removed command-local `vmstore` update flow for `add-library`
  - now calls:
    - `client.AddTemplateModule(...)`
    - `client.AddTemplateLibrary(...)`
  - aligned success text to template wording for these command outputs
- Validated with:
  - `GOWORK=off go test ./cmd/vm-system ./pkg/vmcontrol ./pkg/vmclient ./pkg/vmtransport/http -count=1`
  - `GOWORK=off go test ./... -count=1`

### Why

Task 3 explicitly requires removing ad-hoc command-side mutation logic so template module/library persistence flows through central service/store pathways.

### What worked

- Conversion to vmclient calls was straightforward because Task 4 client methods were already in place.
- Compile and full test sweep remained green after removal.

### What didn't work

- N/A in this step.

### What I learned

- Performing API and client foundations first made command-side cleanup a small, low-risk diff.

### What was tricky to build

The only tricky part was maintaining the intended no-compatibility direction while staging tasks incrementally. I kept legacy flags untouched for now because explicit flag renaming is scoped to Task 7, avoiding cross-task drift.

### What warrants a second pair of eyes

- Confirm that maintaining temporary legacy flag names in `cmd_modules.go` until Task 7 is acceptable sequencing for review.

### What should be done in the future

- Proceed with Task 5/6 to introduce template commands and remove `modules` command registration entirely.

### Code review instructions

- Start with:
  - `cmd/vm-system/cmd_modules.go`
- Validate with:
  - `GOWORK=off go test ./cmd/vm-system ./pkg/vmcontrol ./pkg/vmclient ./pkg/vmtransport/http -count=1`
  - `GOWORK=off go test ./... -count=1`

### Technical details

- The command no longer imports or uses `vmstore`; mutation requests now transit through daemon template endpoints only.

## Step 5: Add template-native CLI subcommands for module/library operations and catalog listing

I implemented the template command surface required by Task 5 in `cmd_template.go`, adding module/library add/remove/list commands and template-owned available catalog listing commands.

This introduces a complete template-first command path for module/library operations so the legacy `modules` surface can be removed in the next task without losing functionality.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Add the missing template module/library CLI command set and available catalog listing under `vm-system template`.

**Inferred user intent:** Make template command surface authoritative before deleting legacy command entrypoints.

**Commit (code):** 2105ea1 — "vm008: add template module/library CLI subcommands (task 5)"

### What I did

- Updated `cmd/vm-system/cmd_template.go`:
  - Added template module commands:
    - `add-module [template-id] --name ...`
    - `remove-module [template-id] --name ...`
    - `list-modules [template-id]`
  - Added template library commands:
    - `add-library [template-id] --name ...`
    - `remove-library [template-id] --name ...`
    - `list-libraries [template-id]`
  - Added template catalog commands:
    - `list-available-modules`
    - `list-available-libraries`
  - Wired all new subcommands into `newTemplateCommand()`
  - Used vmclient template route methods for mutations/query and `vmmodels` builtin catalogs for available lists
- Ran formatting:
  - `gofmt -w cmd/vm-system/cmd_template.go`
- Validated with:
  - `GOWORK=off go test ./cmd/vm-system ./pkg/vmclient ./pkg/vmtransport/http -count=1`
  - `GOWORK=off go test ./... -count=1`

### Why

Task 5 requires template-native command parity for module/library management and available catalog listing so the project can remove legacy command surfaces cleanly.

### What worked

- Existing vmclient template methods allowed direct and consistent command wiring.
- The expanded command set integrated cleanly into `newTemplateCommand` without side effects.

### What didn't work

- N/A in this step.

### What I learned

- Establishing API and client changes first made CLI surface migration mostly mechanical and low-risk.

### What was tricky to build

The main constraint was preserving strict template terminology while matching expected command ergonomics. I standardized on `[template-id]` args and `--name` flags for all add/remove operations to keep language and UX consistent.

### What warrants a second pair of eyes

- Confirm the chosen command names and `--name` flag pattern align with reviewer expectations for long-term CLI stability.

### What should be done in the future

- Remove `cmd_modules.go` and root registration next (Task 6), then complete global naming cleanup (Task 7).

### Code review instructions

- Start with:
  - `cmd/vm-system/cmd_template.go`
- Validate with:
  - `GOWORK=off go test ./cmd/vm-system ./pkg/vmclient ./pkg/vmtransport/http -count=1`
  - `GOWORK=off go test ./... -count=1`

### Technical details

- New template command handlers delegate to:
  - `Client.AddTemplateModule`, `Client.RemoveTemplateModule`, `Client.ListTemplateModules`
  - `Client.AddTemplateLibrary`, `Client.RemoveTemplateLibrary`, `Client.ListTemplateLibraries`
- Catalog commands intentionally remain read-only and source built-ins from `vmmodels`.

## Step 6: Remove legacy modules command file and root registration

I removed the legacy modules command surface entirely by deleting `cmd_modules.go` and removing `modulesCmd` from root command registration. This leaves template commands as the only user-facing path for module/library operations.

This is the clean-cut behavior change expected by VM-008: no compatibility wrappers, no alias command, and no duplicate command namespace.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Execute Task 6 by deleting the old modules command and its registration from CLI entrypoints.

**Inferred user intent:** Complete hard removal of legacy command surface so command model is unambiguous.

**Commit (code):** 2fdb181 — "vm008: remove legacy modules command surface (task 6)"

### What I did

- Updated `cmd/vm-system/main.go`:
  - removed `modulesCmd` from `rootCmd.AddCommand(...)`
- Deleted `cmd/vm-system/cmd_modules.go`
- Validated with:
  - `GOWORK=off go test ./cmd/vm-system ./pkg/vmclient ./pkg/vmtransport/http -count=1`
  - `GOWORK=off go test ./... -count=1`

### Why

Task 6 explicitly requires removal of legacy modules command artifacts now that equivalent template commands are in place.

### What worked

- Deletion was mechanically clean because Task 5 already provided replacement template commands.
- Repository tests remained green after hard removal.

### What didn't work

- N/A in this step.

### What I learned

- Sequencing (new template commands first, then delete legacy command) avoided user-facing capability gaps while still honoring no-compatibility cleanup.

### What was tricky to build

The key concern was making sure no stale symbol references remained after file deletion. A quick codebase search for `modulesCmd` before and after removal ensured root wiring was clean.

### What warrants a second pair of eyes

- Confirm there are no external docs/scripts still invoking `vm-system modules ...` prior to script/doc cleanup tasks.

### What should be done in the future

- Proceed to Task 7 to remove remaining `vm-id` flag language where template resources are targeted.

### Code review instructions

- Start with:
  - `cmd/vm-system/main.go`
  - `cmd/vm-system/cmd_modules.go` (deleted)
- Validate with:
  - `GOWORK=off go test ./cmd/vm-system ./pkg/vmclient ./pkg/vmtransport/http -count=1`
  - `GOWORK=off go test ./... -count=1`

### Technical details

- CLI root command now registers `serve`, `template`, `session`, `exec`, and `libs`, with no `modules` command surface.

## Step 7: Complete remaining CLI template-targeted wording cleanup

I completed remaining CLI wording cleanup for template-targeted identifiers by replacing user-facing `VM ID` labels with `Template ID` in session command outputs.

At this point, template command surfaces use template terminology and do not expose `vm-id` flags.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Execute Task 7 and remove residual legacy `vm-id` wording where template resources are referenced in CLI output/flags.

**Inferred user intent:** Eliminate user-facing terminology drift so template resources are consistently named.

**Commit (code):** Pending for Step 7 commit creation.

### What I did

- Updated `cmd/vm-system/cmd_session.go`:
  - changed `VM ID` output labels to `Template ID` in:
    - session create output
    - session list table header
    - session get output
- Verified no template command flags use `vm-id`.
- Validated with:
  - `GOWORK=off go test ./cmd/vm-system ./pkg/vmclient ./pkg/vmtransport/http -count=1`
  - `GOWORK=off go test ./... -count=1`

### Why

Task 7 explicitly targets cleanup of remaining template-resource wording drift. Session output labels were the remaining user-facing CLI spots still showing `VM ID`.

### What worked

- Wording-only changes were low-risk and preserved behavior.
- Full test sweep remained green.

### What didn't work

- N/A in this step.

### What I learned

- Even with major command cleanup complete, small output labels can still carry old terminology and should be included in the consistency pass.

### What was tricky to build

The subtle part was distinguishing true runtime VM wording from template-targeted identifiers. I limited changes to places where `session.VMID` is presented as the template reference in user-facing output.

### What warrants a second pair of eyes

- Confirm reviewer agrees these session labels should be template-oriented and not runtime-oriented.

### What should be done in the future

- Continue with integration tests/scripts/docs cleanup tasks to remove legacy command usages still present outside core CLI code.

### Code review instructions

- Start with:
  - `cmd/vm-system/cmd_session.go`
- Validate with:
  - `GOWORK=off go test ./cmd/vm-system ./pkg/vmclient ./pkg/vmtransport/http -count=1`
  - `GOWORK=off go test ./... -count=1`

### Technical details

- Change scope is label-only; API fields and model names remain unchanged.
