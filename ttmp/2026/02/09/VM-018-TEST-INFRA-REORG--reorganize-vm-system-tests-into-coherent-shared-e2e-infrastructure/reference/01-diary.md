---
Title: Diary
Ticket: VM-018-TEST-INFRA-REORG
Status: active
Topics:
    - backend
    - architecture
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ttmp/2026/02/09/VM-018-TEST-INFRA-REORG--reorganize-vm-system-tests-into-coherent-shared-e2e-infrastructure/design-doc/01-test-infrastructure-reorganization-analysis-and-implementation-guide.md
    - /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/smoke-test.sh
    - /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/test-e2e.sh
    - /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/test-goja-library-execution.sh
    - /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/test-library-loading.sh
    - /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/test-library-requirements.sh
ExternalSources: []
Summary: >
    VM-018 implementation diary capturing task-by-task changes, failures,
    verification commands, and review guidance.
LastUpdated: 2026-02-09T08:16:30-05:00
WhatFor: Maintain an auditable implementation narrative during test-suite reorganization.
WhenToUse: Use while implementing or reviewing VM-018.
---

# Diary

## Goal

Track VM-018 end-to-end while following strict execution hygiene: define tasks, implement one task at a time, commit per task boundary, check task status, and capture validation evidence and failures.

## Step 1: Create VM-018 scaffold, analysis guide, and executable tasks

This step created the working container for VM-018 and locked the scope before editing scripts. The key outcome is a concrete implementation plan that maps specific script changes to acceptance criteria and commit boundaries.

The analysis documents where the current overlap is harmful and defines a coherent target architecture: shared setup harness, explicit scenario scripts, and a single top-level runner.

### Prompt Context

**User prompt (verbatim):** "create a ticket to reorganize the tests into something more coherent, with a detailed analysis and implementation guide. keep a frequent diary. add detailed tasks to this new ticket, then start implementing each task one by one, committing after each step and checking off the task, keeping your diary."

**Assistant interpretation:** Open a new ticket for test reorganization, write a detailed design/implementation guide, add actionable tasks, and then execute tasks incrementally with per-task commits and diary updates.

**Inferred user intent:** Improve test maintainability and trust by replacing overlapping scripts with a coherent, architecture-aligned test system and full implementation traceability.

**Commit (code):** pending

### What I did

- Created ticket workspace `VM-018-TEST-INFRA-REORG` with docmgr.
- Added and filled:
  - `design-doc/01-test-infrastructure-reorganization-analysis-and-implementation-guide.md`
  - `reference/01-diary.md`
  - `tasks.md`
  - `index.md`
- Defined five explicit tasks with acceptance criteria and commit boundaries.

### Why

- Needed a stable execution contract before touching test scripts, so each commit has a clear success condition.

### What worked

- Ticket creation and document scaffolding completed cleanly.
- Current overlap and stale-assertion issues were captured into an actionable plan.

### What didn't work

- `docmgr task list` summary initially reported fewer tasks than expected because the default scaffold task (`Add tasks here`) remained; resolved by replacing tasks file content explicitly.

### What I learned

- The fastest path to coherence is to consolidate around setup primitives first; scenario cleanups become straightforward once shared harness exists.

### What was tricky to build

- Choosing task boundaries that are both reviewable and independently runnable required balancing architecture goals against shell-script practicality.

### What warrants a second pair of eyes

- Whether to keep legacy script names as wrappers or fully remove them in Task 3.

### What should be done in the future

- Revisit a Go-native e2e harness after VM-018 if shell-based orchestration becomes a long-term maintenance burden.

### Code review instructions

- Read the design guide first, then verify task list aligns with proposed sequencing.
- Confirm diary scope/format supports per-task implementation evidence.

### Technical details

- Ticket directory:
  - `/home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ttmp/2026/02/09/VM-018-TEST-INFRA-REORG--reorganize-vm-system-tests-into-coherent-shared-e2e-infrastructure`

## Step 2: Build shared daemon-first harness and migrate smoke/e2e scripts

This step removed the highest-impact duplication first by extracting common setup and lifecycle code from shell integration scripts. The migration immediately reduced drift risk across dynamic port allocation, temp DB/worktree setup, daemon startup, health checks, and ID extraction.

I also corrected smoke semantics so it no longer depends on a legacy module assumption (`console` in module catalog). Smoke now checks stable core behavior and the current module contract.

### Prompt Context

**User prompt (verbatim):** (see Step 1)

**Assistant interpretation:** Implement the first code task by introducing reusable test infrastructure and applying it to the two core daemon-first scripts.

**Inferred user intent:** Make the tests maintainable and trustworthy by removing duplicated script logic and stale assertions.

**Commit (code):** pending

### What I did

- Added shared harness:
  - `test/lib/e2e-common.sh`
  - helpers for port allocation, temp resource initialization, daemon lifecycle, health wait loop, binary build, and template/session ID extraction.
- Migrated `smoke-test.sh` to source harness and use shared primitives.
- Migrated `test-e2e.sh` to source harness and use shared primitives.
- Replaced stale smoke `console` assumption with stable module-catalog command output check.
- Updated smoke module configuration path from legacy capability call to `template add-module --name fs`.

### Why

- Duplication across smoke/e2e was the main source of drift and unnecessary maintenance cost.

### What worked

- `./smoke-test.sh` passed (10/10 checks).
- `./test-e2e.sh` passed full daemon-first lifecycle.
- Shared harness reduced repeated setup code in both scripts.

### What didn't work

- There is still noisy debug output from CLI command-map parsing (`unknown section type Reference`) during command execution. This is pre-existing and does not fail tests.

### What I learned

- Once lifecycle primitives are centralized, scenario scripts become much easier to reason about and review.

### What was tricky to build

- Balancing harness abstraction with script readability: over-abstracting would hide scenario intent, so only repeated lifecycle primitives were extracted.

### What warrants a second pair of eyes

- Confirm smoke should continue asserting non-empty module catalog output (instead of specific module names) as the intended stability boundary.

### What should be done in the future

- Remove the pre-existing CLI command-map debug noise to keep script output cleaner.

### Code review instructions

- Start at `test/lib/e2e-common.sh` for shared behavior.
- Review `smoke-test.sh` changes for stale-assertion removal and module-path alignment.
- Review `test-e2e.sh` for migration fidelity.
- Re-run:
  - `./smoke-test.sh`
  - `./test-e2e.sh`

### Technical details

- Health wait now uses a retry loop around `/api/v1/health` instead of one-shot checks.

## Step 3: Consolidate overlapping library scripts into a capability matrix

This step replaced three overlapping library-focused scripts with one authoritative matrix test. The new script asserts the core semantics we needed explicitly: `JSON` is always available as a built-in and cannot be configured as a template module, while lodash is only available when configured on the template.

To avoid abrupt command breakage while removing duplicated logic, legacy script entry points were retained as thin wrappers that delegate to the new matrix script.

### Prompt Context

**User prompt (verbatim):** (see Step 1)

**Assistant interpretation:** Execute Task 3 by consolidating redundant library scripts into one coherent test and preserving a clean migration path.

**Inferred user intent:** Get useful, architecture-aligned capability tests with less script drift and clearer ownership.

**Commit (code):** pending

### What I did

- Added `test-library-matrix.sh` as the canonical library/module capability script.
- Implemented assertions for:
  - `template add-module json` should fail (`MODULE_NOT_ALLOWED` semantics),
  - `JSON.stringify` should work in sessions without library configuration,
  - lodash-dependent execution should fail when lodash is not configured,
  - lodash-dependent execution should succeed when lodash is configured,
  - post-hoc library configuration path should succeed.
- Converted legacy scripts to wrappers:
  - `test-library-loading.sh`
  - `test-goja-library-execution.sh`
  - `test-library-requirements.sh`
- Updated VM-018 design guide open question with resolved wrapper decision.

### Why

- Three nearly-duplicate scripts made maintenance and behavior auditing expensive.

### What worked

- `./test-library-matrix.sh` passed (10/10 checks).
- Wrapper compatibility path verified via `./test-library-loading.sh` (delegates and passes).

### What didn't work

- Same pre-existing CLI debug noise appears during command execution (`unknown section type Reference`), but does not affect pass/fail behavior.

### What I learned

- Consolidation into a matrix format makes policy assertions explicit and easy to extend.

### What was tricky to build

- Ensuring failure-path assertions were robust: the script now treats both non-zero command exits and explicit error outputs as expected failure forms for missing-lodash scenarios.

### What warrants a second pair of eyes

- Confirm wrapper strategy should remain temporary versus long-term; if temporary, schedule explicit removal.

### What should be done in the future

- Add explicit matrix cases for additional native modules (`database`, `exec`) as dedicated runtime scenarios.

### Code review instructions

- Review `test-library-matrix.sh` end-to-end first.
- Confirm wrappers contain no duplicated logic and only delegate.
- Re-run:
  - `./test-library-matrix.sh`
  - `./test-library-loading.sh`

### Technical details

- Matrix script uses the shared harness at `test/lib/e2e-common.sh` and standard `template/session/exec` CLI surfaces.
