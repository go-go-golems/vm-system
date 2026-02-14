---
Title: Diary
Ticket: VM-023-IMPROVE-FRONTEND-DOCS
Status: active
Topics:
    - frontend
    - architecture
    - integration
DocType: reference
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: Implementation diary for executing VM-023 P0/P1 documentation remediation with task-by-task commits and validation.
LastUpdated: 2026-02-14T17:00:16.247452417-05:00
WhatFor: Track concrete implementation progress, tests, commits, and decisions while improving frontend plugin docs.
WhenToUse: Read when reviewing VM-023 execution history or continuing follow-up documentation work.
---

# Diary

## Goal

Track implementation steps for VM-023 while fixing high-priority frontend plugin documentation issues, with explicit tests and commit boundaries.

## Step 1: Convert Findings into Executable P0/P1 Task Plan

I translated the assessment findings into an implementation task queue and created a dedicated VM-023 diary so execution can be audited step-by-step.

This step intentionally scoped out P2 work after your follow-up instruction, so the implementation pass remains focused on contract correctness and high-impact clarity fixes.

### Prompt Context

**User prompt (verbatim):** "Ok, create tasks to address all these issues, including implementing ui.column and creating an example for it. 

Then implement the tasks one by one, testing as you go, committing, and then updating your diary."

**Assistant interpretation:** Build an actionable remediation backlog, implement each item sequentially, run tests per step, commit incrementally, and log progress in the diary.

**Inferred user intent:** Move from analysis to execution with reliable traceability and low-risk iterative delivery.

**Commit (code):** N/A

### What I did
- Added a VM-023 diary document.
- Added detailed remediation tasks for P0/P1 issues, including:
  - runtime `ui.column` implementation
  - docs contract corrections
  - embedding guide adapter correction
  - lifecycle/capability accuracy updates
  - README/changelog clarifications
  - full frontend validation run
- Recorded user instruction to skip P2 in this pass.

### Why
- A structured queue is needed to execute one task at a time with clean commit boundaries.

### What worked
- Task list now maps directly to implementable steps (`tasks.md` items 6-12).

### What didn't work
- N/A

### What I learned
- P0/P1 can be implemented independently of P2 additions, which keeps this pass focused and faster to verify.

### What was tricky to build
- The trick was preserving prior VM-023 completion tasks while adding a new implementation phase. I kept the old tasks checked and appended a new execution block instead of rewriting history.

### What warrants a second pair of eyes
- Verify that task granularity (6-12) is the right balance for review vs implementation overhead.

### What should be done in the future
- After P0/P1 merge, revisit deferred P2 docs (`troubleshooting`, `current-host-loop`, `contract-cheatsheet`) in a follow-up ticket.

### Code review instructions
- Start with VM-023 `tasks.md` and this diary to see planned execution order.
- Confirm scope excludes P2 as requested.

### Technical details
- Active implementation tasks:
  - 6: runtime `ui.column`
  - 7/10: docs contract + examples
  - 8/9/11: lifecycle/capability/embedding/README/changelog accuracy
  - 12: validation commands

## Step 2: Implement `ui.column` in Runtime Bootstrap

I implemented `ui.column` in the QuickJS bootstrap API and added integration coverage to ensure plugin render output using `ui.column(...)` is accepted and returned as `kind: "column"`.

This closes the highest-priority runtime-side mismatch and lets docs safely include a canonical column-based layout example in the next step.

### Prompt Context

**User prompt (verbatim):** "don't do P2 for now, btw"

**Assistant interpretation:** Continue implementation but restrict scope to P0/P1 only.

**Inferred user intent:** Prioritize correctness-critical fixes without broadening into additional documentation expansion.

**Commit (code):** `7697125` — "feat(runtime): add ui.column helper to QuickJS bootstrap"

### What I did
- Updated runtime bootstrap in:
  - `frontend/packages/plugin-runtime/src/runtimeService.ts`
  - Added `column(children = [])` helper returning `{ kind: "column", children: ... }`.
- Added integration test:
  - `frontend/packages/plugin-runtime/src/runtimeService.integration.test.ts`
  - New `COLUMN_PLUGIN` fixture and assertion that `render(...).kind === "column"`.
- Ran tests:
  - `pnpm -C frontend test:integration`
  - `pnpm -C frontend test:unit`

### Why
- Docs and examples should not claim `ui.column` exists unless runtime actually supports it.

### What worked
- Integration suite passed with the new test (`6 tests` in integration file).
- Unit tests stayed green (`7 tests`).

### What didn't work
- N/A

### What I learned
- `uiSchema` and `WidgetRenderer` were already column-compatible; the missing piece was runtime DSL bootstrap exposure.

### What was tricky to build
- The core subtlety was ensuring this change is contract-level (bootstrap API) rather than only renderer-level. The symptom before fix would be runtime `ui.column is not a function` even though UI types allowed it.

### What warrants a second pair of eyes
- Confirm whether additional DSL helpers should be standardized similarly (for example ensuring docs only include helpers defined in bootstrap).

### What should be done in the future
- Add a small DSL contract snapshot test so future docs/API drift is caught automatically.

### Code review instructions
- Start at `frontend/packages/plugin-runtime/src/runtimeService.ts` bootstrap block.
- Validate via `pnpm -C frontend test:integration`.

### Technical details
- Added helper near existing `row` and `panel` builders.
- New test fixture id/title: `column-demo` / "Column Demo".

## Step 3: Fix UI DSL Reference Contract Signatures (Task 7)

I corrected the UI DSL reference so signature examples match the runtime bootstrap API. The highest-risk mismatch points (`ui.input`, `ui.table`) are now documented with the real call shapes used by QuickJS runtime helpers.

I also added an explicit note in the `ui.column` section that this helper is implemented in the runtime bootstrap.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Implement each remediation task with tests and commits, including contract-level doc fixes.

**Inferred user intent:** Ensure docs are trustworthy enough that copy/paste examples reflect the actual runtime API.

**Commit (code):** `bbffd47` — "docs(ui-dsl): align input/table signatures with runtime contract"

### What I did
- Updated `frontend/docs/architecture/ui-dsl.md`:
  - `ui.input(options)` -> `ui.input(value, options)`
  - `ui.table(options)` -> `ui.table(rows, options)`
  - corrected examples and reference tables accordingly
  - added `ui.column` availability clarification.
- Ran build validation:
  - `pnpm -C frontend build`

### Why
- This doc is the contract reference; if it is wrong, all downstream authoring docs become wrong.

### What worked
- Build succeeded after doc edits.
- DSL reference now matches runtime helper shapes for input/table.

### What didn't work
- N/A (build emitted known non-blocking warnings only).

### What I learned
- Most downstream doc mismatches originate from this one reference file; fixing it first reduces cascading confusion.

### What was tricky to build
- The subtlety was preserving intent while changing signatures. For `table`, docs previously implied a single object argument, but runtime helper splits rows and props; example rewrites needed to show the two-argument pattern clearly.

### What warrants a second pair of eyes
- Verify no remaining stale `ui.input({...})` / `ui.table({...})` calls remain in other docs after upcoming task 10.

### What should be done in the future
- Consider adding an automated contract snippet generation step from runtime helper definitions.

### Code review instructions
- Review `frontend/docs/architecture/ui-dsl.md` signature sections and examples.
- Validate with `pnpm -C frontend build`.

### Technical details
- Build warnings observed were the pre-existing CSS import order and chunk-size advisories.

## Step 4: Update Quickstart and Examples with Real DSL Calls (Task 10)

I updated plugin authoring docs so copy/paste snippets use runtime-accurate DSL signatures and include a dedicated `ui.column` starter example.

This addresses the highest-friction authoring confusion path: a user following examples should no longer hit signature mismatches for `ui.input` and `ui.table`.

### Prompt Context

**User prompt (verbatim):** "don't do P2 for now, btw"

**Assistant interpretation:** Continue implementation for P0/P1 tasks only, with per-task testing and commits.

**Inferred user intent:** Improve practical onboarding docs first, without widening scope into deferred P2 sections.

**Commit (code):** `94aabec` — "docs(authoring): add ui.column starter and fix ui.input/ui.table examples"

### What I did
- Updated `frontend/docs/plugin-authoring/quickstart.md`:
  - added an explicit `ui.column` quick example section
  - fixed form example to `ui.input(pluginState.name, { ... })`
- Updated `frontend/docs/plugin-authoring/examples.md`:
  - added "Example 0: Column Layout Starter"
  - changed stale `ui.input({ value: ... })` usages to `ui.input(value, { ... })`
  - changed stale `ui.table({ headers, rows })` usages to `ui.table(rows, { headers })`
- Ran validation:
  - stale-pattern scan on both docs with `rg`
  - `pnpm -C frontend build`

### Why
- The quickstart and examples are the highest-leverage docs for new plugin authors. If they drift from runtime signatures, users fail immediately.

### What worked
- Build succeeded after doc changes.
- Pattern scan found no remaining stale `ui.input({` / `ui.table({` calls in these two authoring docs.

### What didn't work
- N/A (only known existing non-blocking build warnings remained).

### What I learned
- A compact "Example 0" helps establish `ui.column` as the baseline layout primitive before readers hit larger examples.

### What was tricky to build
- Keeping examples readable while adjusting signatures required small expression rewrites, especially where rows were built inline for `ui.table`.

### What warrants a second pair of eyes
- Confirm the new example ordering (with a new Example 0) still matches any external references into this page.

### What should be done in the future
- Add a doc-lint pass that checks common DSL signature anti-patterns directly in markdown code fences.

### Code review instructions
- Review:
  - `frontend/docs/plugin-authoring/quickstart.md`
  - `frontend/docs/plugin-authoring/examples.md`
- Validate with:
  - `pnpm -C frontend build`

### Technical details
- Task 10 is now marked complete in VM-023 tasks.

## Step 5: Align Lifecycle and Capability Architecture Docs (Task 8)

I aligned architectural docs with the implementation details in `WorkbenchPage` and the Redux runtime adapter, focusing on re-render behavior and shared-state projection semantics.

This closes the core P1 architecture drift called out in the assessment and makes the docs safer for debugging and performance reasoning.

### Prompt Context

**User prompt (verbatim):** "don't do P2 for now, btw"

**Assistant interpretation:** Continue only P0/P1 fixes; execute task-by-task with tests and commits.

**Inferred user intent:** Fix correctness and mental-model drift without expanding into extra doc surface.

**Commit (code):** `1f3aca1` — "docs(architecture): align lifecycle and capability docs with host behavior"

### What I did
- Updated `frontend/docs/architecture/dispatch-lifecycle.md`:
  - clarified current host behavior re-renders all loaded widgets after runtime changes
  - clarified `globalState` as projected (`shared` by read grants + `system`)
  - documented that widget trees/errors are React local state in `WorkbenchPage`, not runtime slice state
- Updated `frontend/docs/architecture/capability-model.md`:
  - added "Internal vs Projected Shared State" section
  - clarified projection for `counter-summary` visible fields
  - clarified `globalState.system` is host-provided and not gated by `readShared`
  - documented `runtime-registry` `id` alias and write-ignored caveat for read-only domains
- Ran validation:
  - `pnpm -C frontend build`

### Why
- These files define how users reason about state flow and policy behavior. Drift here leads to debugging mistakes and wrong embedding assumptions.

### What worked
- Build succeeded after architecture doc updates.
- Task 8 is now checked in VM-023 `tasks.md`.

### What didn't work
- N/A (known non-blocking frontend build warnings remained unchanged).

### What I learned
- The most important distinction for readers is not just "what domain exists" but "what projection is visible to a plugin instance".

### What was tricky to build
- The re-render wording needed to be precise: current host implementation re-renders all loaded widgets, while future optimized hosts may narrow this.

### What warrants a second pair of eyes
- Confirm that the `globalState.system` clarification matches intended long-term capability policy (currently host-provided, ungated).

### What should be done in the future
- Add a small architecture note linking directly to `WorkbenchPage` render loop as the current reference implementation.

### Code review instructions
- Review:
  - `frontend/docs/architecture/dispatch-lifecycle.md`
  - `frontend/docs/architecture/capability-model.md`
- Validate with:
  - `pnpm -C frontend build`

### Technical details
- Step focuses on doc/runtime parity only; no runtime behavior change.

## Step 6: Correct Embedding Mode C Adapter Contract (Task 9)

I rewrote the embedding guide's Mode C section to reflect the real `RuntimeHostAdapter` contract and added concrete wrapper patterns for both direct-service and worker-client modes.

This removes a high-risk integration trap where embedders might assume `QuickJSRuntimeService`/`QuickJSSandboxClient` can be passed directly as adapters.

### Prompt Context

**User prompt (verbatim):** "don't do P2 for now, btw"

**Assistant interpretation:** Continue remaining P0/P1 task execution only.

**Inferred user intent:** Resolve contract-level doc hazards that could break external embedding work.

**Commit (code):** `f3bc287` — "docs(runtime): fix host adapter mode with explicit wrapper patterns"

### What I did
- Updated `frontend/docs/runtime/embedding.md` Mode C:
  - documented signature mismatch caveat explicitly:
    - adapter uses object inputs and async methods
    - runtime service/client APIs use positional arguments
  - added `createDirectAdapter(...)` wrapper example for `QuickJSRuntimeService`
  - added `createWorkerAdapter(...)` wrapper example for `QuickJSSandboxClient`
  - corrected usage example to `async function createMyApp(...)` with adapter-based calls
- Ran validation:
  - `pnpm -C frontend build`

### Why
- The prior wording implied direct substitutability and would produce avoidable type and integration churn in real embeddings.

### What worked
- Build succeeded after embedding doc updates.
- `tasks.md` marks task 9 as complete.

### What didn't work
- `docmgr task list` briefly showed stale state right after `task check`; direct file inspection confirmed task 9 was checked.

### What I learned
- Adapter abstraction docs must be explicit about call-shape normalization, not just conceptual portability.

### What was tricky to build
- The direct-service wrapper needs to normalize sync runtime methods into async adapter methods while keeping examples concise.

### What warrants a second pair of eyes
- Validate wrapper examples against internal embedding conventions if a shared adapter utility is introduced later.

### What should be done in the future
- Consider publishing these wrappers as actual runtime helpers to reduce copy/paste divergence.

### Code review instructions
- Review:
  - `frontend/docs/runtime/embedding.md`
- Validate with:
  - `pnpm -C frontend build`

### Technical details
- No runtime code changed; this is a documentation contract-correction step.

## Step 7: Add API Truth-Source and Migration Clarifications (Task 11)

I updated top-level frontend docs and migration notes to make contract authority explicit and to call out DSL signature-sensitive migration details.

This reduces recurring confusion by giving contributors one canonical place to verify runtime truth before copying snippets.

### Prompt Context

**User prompt (verbatim):** "don't do P2 for now, btw"

**Assistant interpretation:** Finish remaining P0/P1 ticket tasks only.

**Inferred user intent:** Close the high-value doc reliability gaps without adding new P2 documents.

**Commit (code):** `13bdaa7` — "docs: add API truth-source callouts and DSL migration notes"

### What I did
- Updated `frontend/docs/README.md`:
  - added "API Truth Source (Read This First)" section
  - linked canonical implementation files:
    - runtime DSL bootstrap (`runtimeService.ts`)
    - adapter contract (`hostAdapter.ts`)
    - runtime policy/projection behavior (`redux-adapter/store.ts`)
  - documented current high-signal DSL signatures (`input`, `table`, `column`)
- Updated `frontend/docs/migration/changelog-vm-api.md`:
  - added DSL signature migration notes and anti-pattern warning
  - added adapter-wrapper normalization note for `RuntimeHostAdapter`
  - added source-of-truth references section
- Ran validation:
  - `pnpm -C frontend build`

### Why
- README and migration notes are entry points for contributors and embedders; they need explicit direction on where runtime truth lives.

### What worked
- Build succeeded after these doc updates.
- Task 11 is checked in VM-023 tasks.

### What didn't work
- N/A

### What I learned
- A short "truth source" section in the entry doc prevents downstream drift across multiple pages.

### What was tricky to build
- Balancing concise migration notes with enough specificity to prevent signature mistakes required explicit call-shapes.

### What warrants a second pair of eyes
- Confirm wording around adapter wrappers aligns with desired public API posture if wrappers later become first-class exports.

### What should be done in the future
- Add a lightweight docs consistency check that validates signature text in docs against a small canonical list.

### Code review instructions
- Review:
  - `frontend/docs/README.md`
  - `frontend/docs/migration/changelog-vm-api.md`
- Validate with:
  - `pnpm -C frontend build`

### Technical details
- This step documents contract authority and migration caveats only; no behavior changes.
