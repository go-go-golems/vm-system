---
Title: Diary
Ticket: VM-006-REVIEW-VM-SYSTEM
Status: complete
Topics:
    - backend
DocType: reference
Intent: long-term
Owners: []
RelatedFiles:
    - Path: vm-system/vm-system/pkg/vmcontrol/execution_service.go
      Note: Path normalization and post-execution limit enforcement checks analyzed and dynamically validated
    - Path: vm-system/vm-system/pkg/vmexec/executor.go
      Note: Execution event/status persistence behavior analyzed for duplication and contract drift
    - Path: vm-system/vm-system/pkg/vmsession/session.go
      Note: Startup flow and session runtime map behavior analyzed and dynamically validated
    - Path: vm-system/vm-system/pkg/vmstore/vmstore.go
      Note: Error typing and JSON decoding behavior analyzed
    - Path: vm-system/vm-system/pkg/vmtransport/http/server.go
      Note: API validation/error mapping surface analyzed and exercised
    - Path: vm-system/vm-system/smoke-test.sh
      Note: |-
        Daemon-first smoke workflow executed and validated
        Executed daemon-first smoke workflow
    - Path: vm-system/vm-system/test-e2e.sh
      Note: |-
        Daemon-first e2e workflow executed and validated
        Executed daemon-first end-to-end workflow
    - Path: vm-system/vm-system/test-goja-library-execution.sh
      Note: |-
        Legacy script executed and failed; stale command contract evidence captured
        Executed and captured stale command failure
    - Path: vm-system/vm-system/test-library-loading.sh
      Note: |-
        Legacy script executed and failed; stale command contract evidence captured
        Executed and captured stale command failure
    - Path: vm-system/vm-system/test-library-requirements.sh
      Note: |-
        Legacy script executed and failed; stale command contract evidence captured
        Executed and captured stale command failure
    - Path: vm-system/vm-system/ttmp/2026/02/08/VM-006-REVIEW-VM-SYSTEM--review-vm-system-implementation-quality-testing-and-setup/design-doc/01-comprehensive-vm-system-implementation-quality-review.md
      Note: |-
        Final detailed review report
        Final report authored from diary evidence
ExternalSources: []
Summary: Detailed implementation diary for VM-006 covering discovery, dynamic validation, deep code review, report authoring, and upload workflow.
LastUpdated: 2026-02-08T11:15:28.600040763-05:00
WhatFor: Preserve exact commands, findings, failures, and review procedure for VM-006.
WhenToUse: Use when auditing VM-006 findings, reproducing discovered defects, or reviewing review-method quality.
---


# Diary

## Goal

This diary captures the complete execution trail for VM-006: how I reviewed vm-system quality in the context of VM-001, VM-004, and VM-005, which dynamic checks I ran, what failed, what worked, and how the final report was produced.

The intent is reproducibility and auditability. Every major finding in the report has a command or source reference here so another maintainer can independently verify it.

## Step 1: Scope anchoring and ticket setup

I began by loading the skill instructions relevant to this request (`code-quality-review-cleanup`, `docmgr`, `diary`, `remarkable-upload`) and then enumerated repo/docmgr context so the review stayed anchored to the right ticket line.

I verified that VM-001, VM-004, and VM-005 existed in docmgr with full ticket IDs and collected their design/task/changelog context before touching implementation behavior.

### Prompt Context

**User prompt (verbatim):** "Go over the code we create in vm-system in the context of docmgr VM-004 and VM-001 and VM-005 and do a deep detailed review and assessment of the quality of the implementation and its testig and its setup, especially wrt duplicated code, code structure, deprecated code, confusing APIs, missing abstractions, unidiomatic code, etc...

Write a detailed report as part of VM-006-REVIEW-VM-SYSTEM and keep a detailed diary. Feel free to compile and run and explore things dynamically, not just on the basis of code.

Upload the document to remarkable when done"

**Assistant interpretation:** Build a deep, evidence-backed implementation review tied to VM-001/004/005, document it under a new VM-006 ticket, keep a detailed diary, and upload the final document to reMarkable.

**Inferred user intent:** Obtain a trustworthy quality assessment that combines static review and live behavior checks, not a superficial code scan.

**Commit (code):** N/A (documentation/review workflow only)

### What I did

- Loaded and read applicable skills from:
  - `/home/manuel/.codex/skills/code-quality-review-cleanup/SKILL.md`
  - `/home/manuel/.codex/skills/docmgr/SKILL.md`
  - `/home/manuel/.codex/skills/diary/SKILL.md`
  - `/home/manuel/.codex/skills/remarkable-upload/SKILL.md`
- Enumerated files and ticket/doc state with:
  - `rg --files`
  - `docmgr status --summary-only`
  - `docmgr ticket list`
  - `docmgr doc list`
- Read VM-001/VM-004/VM-005 ticket artifacts (index/tasks/design/changelog/tutorial) to cross-check review targets.
- Created ticket workspace:
  - `docmgr ticket create-ticket --ticket VM-006-REVIEW-VM-SYSTEM --title "Review vm-system implementation quality, testing, and setup" --topics backend`
- Added report and diary documents:
  - `docmgr doc add --ticket VM-006-REVIEW-VM-SYSTEM --doc-type design-doc --title "Comprehensive vm-system implementation quality review"`
  - `docmgr doc add --ticket VM-006-REVIEW-VM-SYSTEM --doc-type reference --title "Diary"`

### Why

Without ticket/context anchoring first, review conclusions could drift from the explicit commitments and caveats already recorded in VM-001/004/005.

### What worked

- Docmgr ticket/doc creation workflow worked cleanly.
- Existing VM-001/004/005 docs provided enough baseline to evaluate implementation promises against current behavior.

### What didn't work

- Short-form ticket queries like `docmgr ticket list --ticket VM-001` returned `No tickets found.` because full ticket IDs are required.

### What I learned

- The workspace stores VM ticket IDs with full slugs, not short prefixes, so scriptable review tooling should prefer `docmgr ticket list` and then match full IDs.

### What was tricky to build

The ticket paths are long and similarly named; the main risk was accidentally writing VM-006 documents into the wrong workspace. I mitigated this by collecting explicit `Path:` outputs from docmgr creation commands before editing files.

### What warrants a second pair of eyes

- Verify VM-006 ticket naming and location conventions are acceptable for your docmgr taxonomy.

### What should be done in the future

- Add a small helper script under `ttmp/.../scripts` for selecting tickets by prefix safely if this is a recurring workflow.

### Code review instructions

- Start with VM-006 ticket root:
  - `ttmp/2026/02/08/VM-006-REVIEW-VM-SYSTEM--review-vm-system-implementation-quality-testing-and-setup`
- Validate ticket setup commands by replaying `docmgr ticket list` and `docmgr doc list`.

### Technical details

- Key command for orientation:
  - `docmgr ticket list`
- Created ticket path:
  - `ttmp/2026/02/08/VM-006-REVIEW-VM-SYSTEM--review-vm-system-implementation-quality-testing-and-setup`

## Step 2: Baseline tests and script-surface validation

After context setup, I ran baseline test and script surfaces to classify where reliability currently exists and where setup drift appears. I intentionally ran both supported daemon-first scripts and legacy library scripts.

This step quickly showed an important split: daemon-first path is healthy, while library-focused scripts are stale/broken.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Validate behavior dynamically, not only statically, and include test/setup quality in the review.

**Inferred user intent:** Ensure quality claims are grounded in actual command execution outcomes.

**Commit (code):** N/A

### What I did

Executed:

- `GOWORK=off go test ./... -count=1`
- `./smoke-test.sh`
- `./test-e2e.sh`
- `./test-library-loading.sh`
- `./test-library-requirements.sh`
- `./test-goja-library-execution.sh`
- `GOWORK=off go test -race ./pkg/vmtransport/http -count=1`
- `GOWORK=off go vet ./...`
- `GOWORK=off go test ./... -cover -count=1`

### Why

I needed runtime evidence for implementation/test/setup quality, especially to validate VM-004 claims and detect setup regressions affecting day-to-day contributors.

### What worked

- `go test ./...` passed.
- `smoke-test.sh` passed all 10 checks.
- `test-e2e.sh` passed full daemon-first workflow.
- `go test -race` and `go vet` passed for current exercised paths.

### What didn't work

All three library scripts failed immediately on removed commands:

- `test-library-loading.sh`:
  - `Error: unknown command "vm" for "vm-system"`
- `test-library-requirements.sh`:
  - `Error: unknown command "vm" for "vm-system"`
- `test-goja-library-execution.sh`:
  - `Error: unknown command "vm" for "vm-system"`

Coverage output revealed major skew:

- `pkg/vmtransport/http`: `72.6%`
- all other packages: `0.0%`

### What I learned

- VM-004 significantly strengthened HTTP integration tests.
- Core packages (`vmcontrol`, `vmsession`, `vmexec`, `vmstore`) remain largely untested directly.
- Legacy script surface is now a source of breakage rather than confidence.

### What was tricky to build

Balancing signal from passing integration tests with failing legacy scripts required explicit separation: "supported path healthy" vs "repo test/setup surface inconsistent." Without that distinction, conclusions would be misleading.

### What warrants a second pair of eyes

- Decide whether failing library scripts should be migrated immediately or archived as legacy artifacts.

### What should be done in the future

- Add CI policy that only maintained scripts are treated as test gates.

### Code review instructions

- Re-run script matrix exactly:
  - `smoke-test.sh`, `test-e2e.sh`, `test-library-loading.sh`, `test-library-requirements.sh`, `test-goja-library-execution.sh`
- Confirm coverage skew with:
  - `GOWORK=off go test ./... -cover -count=1`

### Technical details

- Coverage output:
  - `pkg/vmtransport/http 72.6%`
  - others `0.0%`

## Step 3: Dynamic edge-case probing for correctness and safety

With baseline established, I ran targeted behavior probes to verify potential high-risk concerns discovered during code reading. I focused on session lifecycle consistency, error contracts, path boundaries, and limits behavior.

This step produced the strongest defect evidence for the final report: boundary bypasses and contract mismatches were reproducible with minimal setups.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Go beyond tests and intentionally explore dynamic edge conditions to uncover hidden quality issues.

**Inferred user intent:** Catch issues that static review or happy-path tests may miss.

**Commit (code):** N/A

### What I did

Ran custom command sequences to probe:

1. close semantics and missing execution status mapping
- first close -> `200`
- second close -> `404 SESSION_NOT_FOUND`
- missing execution get -> `500 INTERNAL`

2. run-file symlink escape
- created symlink inside worktree pointing to JS file outside worktree
- `/api/v1/executions/run-file` returned `201` and executed outside file

3. startup file traversal
- added startup file path `../outside-startup.js`
- startup accepted (`201`), session created, and external startup value visible in REPL (`99`)

4. startup failure lifecycle
- invalid startup script caused `500` on session creation
- runtime summary still showed `active_sessions:1`

5. limit-enforcement persistence mismatch
- forced tight limits via DB update
- repl request returned `422 OUTPUT_LIMIT_EXCEEDED`
- executions list showed persisted execution `status:"ok"` with result

### Why

These probes directly test security and contract boundaries that were not comprehensively covered by current integration tests.

### What worked

- The custom probes reliably reproduced the same behavior multiple times.
- The observed outputs mapped cleanly back to specific code paths.

### What didn't work

One intermediate limits test failed due DB storage-type mismatch after manual SQL update:

- `failed to get VM settings: sql: Scan error on column index 1, name "limits_json": unsupported Scan, storing driver.Value type string into type *json.RawMessage`

I corrected this by casting the JSON literal to `BLOB` in sqlite update.

### What I learned

- Current run-file protection is string/relative-path based, not canonical-path safe.
- Startup path policy is less strict than run-file policy.
- Session startup error handling has map-lifecycle leak side effects.

### What was tricky to build

The limits probe required care around SQLite value typing because direct SQL updates produced `TEXT` values that failed scanning into `json.RawMessage`. The workaround (`CAST(... AS BLOB)`) restored parity with the application write path.

### What warrants a second pair of eyes

- Security review specifically on path-canonicalization and symlink handling.
- Product/API decision on close idempotency and limit-exceeded persistence semantics.

### What should be done in the future

- Add regression tests for:
  - startup traversal rejection
  - symlink escape rejection
  - startup-failure runtime-summary consistency
  - execution-limit result/status contract

### Code review instructions

- Reproduce each probe in isolated temp DB/worktree.
- Compare runtime responses with implementation at:
  - `pkg/vmcontrol/execution_service.go`
  - `pkg/vmsession/session.go`
  - `pkg/vmexec/executor.go`

### Technical details

Observed outputs (representative):

- second close:
  - `{"error":{"code":"SESSION_NOT_FOUND",...}}` with `404`
- missing execution:
  - `{"error":{"code":"INTERNAL","message":"execution not found"...}}` with `500`
- symlink run-file:
  - `status=201`, execution `status:"ok"`
- startup traversal:
  - startup add `201`, REPL result preview `99`
- limit mismatch:
  - request `422 OUTPUT_LIMIT_EXCEEDED`, list shows stored `status:"ok"`

## Step 4: Static deep review synthesis and report authoring

After dynamic probes, I completed package-by-package static inspection and merged code evidence with runtime evidence into the final report document.

The report was structured to prioritize defects by severity and include concrete remediation sketches rather than only critiques.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Deliver a deep detailed quality assessment covering duplication, structure, deprecated code, confusing APIs, missing abstractions, and unidiomatic patterns.

**Inferred user intent:** Get a practical quality blueprint for cleanup and stabilization, not only an issue list.

**Commit (code):** N/A

### What I did

- Inspected all main backend packages and command surfaces.
- Quantified duplication hotspots (for example, repeated `vmclient.New(...)` usage and duplicated executor pipelines).
- Authored the full report in:
  - `ttmp/2026/02/08/VM-006-REVIEW-VM-SYSTEM--review-vm-system-implementation-quality-testing-and-setup/design-doc/01-comprehensive-vm-system-implementation-quality-review.md`

### Why

A useful review should tie each claim to location and runtime implications, then convert findings into a prioritized cleanup sequence.

### What worked

- Static and dynamic evidence aligned on major issues.
- Report format captures both immediate risks and phased cleanup plan.

### What didn't work

- N/A (no tooling blockers in this phase)

### What I learned

- The architecture is structurally promising, but safety and legacy cleanup are the dominant next-leverage tasks.
- VM-005 docs already acknowledge several weak spots; the highest-value next work is implementation alignment, not more broad analysis.

### What was tricky to build

The main challenge was separating "architectural direction is good" from "current behavior still unsafe/confusing in important edges" without flattening nuance. I resolved this by ordering findings by impact and explicitly listing strengths separately.

### What warrants a second pair of eyes

- Severity ranking for close idempotency and legacy modules command treatment.
- Prioritization of phase-1 fixes if only one sprint is available.

### What should be done in the future

- Spin follow-up implementation tickets directly from report phases (safety first, then surface cleanup, then refactor/test-balance).

### Code review instructions

- Read report sections in order:
  1. Findings (severity ordered)
  2. Alignment with VM-001/004/005
  3. Prioritized cleanup plan
- For each high-severity issue, verify runtime evidence in this diary Step 3.

### Technical details

- Duplication count example:
  - `vmclient.New(serverURL, nil)` appears 17 times across command handlers.

## Step 5: Docmgr bookkeeping and reMarkable upload

This step captures final task/checklist updates, file relations, changelog updates, and reMarkable upload outputs.

### Prompt Context

**User prompt (verbatim):** (same as Step 1)

**Assistant interpretation:** Finalize VM-006 deliverables in docmgr and upload to reMarkable.

**Inferred user intent:** Ensure review output is documented, traceable, and delivered to reading device.

**Commit (code):** N/A

### What I did

- Marked remaining VM-006 task complete in `tasks.md`.
- Updated ticket `index.md` and `changelog.md` with completion status and delivery notes.
- Verified upload prerequisites:
  - `remarquee status` -> `remarquee: ok`
  - `remarquee cloud account --non-interactive` -> `user=wesen@ruinwesen.com sync_version=1.5`
  - `which pandoc && which xelatex` confirmed both present.
- Ran upload dry-run:
  - `remarquee upload bundle --dry-run <report> <diary> --name "VM-006-REVIEW-VM-SYSTEM Review + Diary" --remote-dir "/ai/2026/02/08/VM-006-REVIEW-VM-SYSTEM" --toc-depth 2`
- Ran real upload:
  - `remarquee upload bundle <report> <diary> --name "VM-006-REVIEW-VM-SYSTEM Review + Diary" --remote-dir "/ai/2026/02/08/VM-006-REVIEW-VM-SYSTEM" --toc-depth 2`
- Verified remote listing:
  - `remarquee cloud ls /ai/2026/02/08/VM-006-REVIEW-VM-SYSTEM --long --non-interactive`

### Why

- Ticket artifacts should be complete and discoverable, with explicit delivery confirmation.

### What worked

- Dry-run and real upload succeeded on first attempt.
- Remote verification showed expected uploaded bundle document.

### What didn't work

- Initial `docmgr doc relate --doc ...` call failed with:
  - `expected exactly 1 doc for --doc "...", got 0`
- Root cause was path selection; command succeeded when retried with absolute doc paths.

### What I learned

- For this docmgr setup, absolute `--doc` paths are the safest choice for scripted relation updates.
- Bundle upload gives a cleaner reMarkable reading artifact than separate single-doc uploads for long review material.

### What was tricky to build

- The report and diary are both long and include many code paths; bundling them while preserving a usable table of contents required using `upload bundle` with `--toc-depth 2` and a stable ticket folder path.

### What warrants a second pair of eyes

- Confirm uploaded document naming and remote folder convention.

### What should be done in the future

- Keep a stable ticket-based remote upload path convention for archive consistency.

### Code review instructions

- Verify docmgr task/changelog/related-file updates.
- Verify remote listing after upload.

### Technical details

- Dry-run output confirmed include paths and render target:
  - `DRY: upload VM-006-REVIEW-VM-SYSTEM Review + Diary.pdf -> /ai/2026/02/08/VM-006-REVIEW-VM-SYSTEM`
- Real upload output:
  - `OK: uploaded VM-006-REVIEW-VM-SYSTEM Review + Diary.pdf -> /ai/2026/02/08/VM-006-REVIEW-VM-SYSTEM`
- Remote verification output:
  - `[f] VM-006-REVIEW-VM-SYSTEM Review + Diary`

## Step 6: Type-system follow-up kickoff and typed path primitives

I reopened VM-006 for implementation follow-up and converted the path-hardening direction into concrete ticket tasks. Then I implemented the first task: introducing typed worktree path primitives in a dedicated package.

This is the foundation for replacing string-based path checks in execution and startup flows with canonicalized, reusable, and testable path contracts.

### Prompt Context

**User prompt (verbatim):** "Can we improve path resolution with type safety? Are there other opportunities to leverage the type system more (through types, structs, interfaces?), for simplicity and robustness?"

**Assistant interpretation:** Introduce stronger types around path handling and identify/implement practical type-system upgrades that improve safety and clarity.

**Inferred user intent:** Reduce fragile string-based logic and make safety-critical behavior enforced by construction.

**Commit (code):** Pending in this step (task commit follows implementation + diary update)

### What I did

- Reopened VM-006 execution by adding a new `Type-System Follow-Up` section in:
  - `ttmp/2026/02/08/VM-006-REVIEW-VM-SYSTEM--review-vm-system-implementation-quality-testing-and-setup/tasks.md`
- Marked Task 1 complete after implementation:
  - Introduced typed path model in new package:
    - `pkg/vmpath/path.go`
    - `pkg/vmpath/path_test.go`
- Added core types and constructors:
  - `WorktreeRoot`
  - `RelWorktreePath`
  - `ResolvedWorktreePath`
- Added strict parsing/canonicalization rules:
  - reject empty/absolute/traversal relative paths
  - canonicalize root with `EvalSymlinks`
  - detect and reject resolved paths escaping root
- Added unit tests for:
  - relative-path parsing constraints
  - symlink escape rejection
  - canonical resolution of in-root symlink targets
  - root directory validation
- Ran:
  - `gofmt -w pkg/vmpath/path.go pkg/vmpath/path_test.go`
  - `GOWORK=off go test ./pkg/vmpath -count=1`

### Why

- Current path checks are duplicated and string-based in safety-critical flows.
- A typed path package creates one authoritative place for validation + canonicalization rules.

### What worked

- New `vmpath` package compiled cleanly.
- Unit tests passed and captured intended invariants.

### What didn't work

- N/A for this step.

### What I learned

- Introducing small opaque path types substantially simplifies downstream call-site semantics and error reasoning.

### What was tricky to build

- The main edge case was deciding behavior when `EvalSymlinks` fails on non-existing targets. I kept resolution permissive for non-existing files (so existing `file not found` behavior can remain at execution layer) while still rejecting canonical escapes when symlinks are resolvable.

### What warrants a second pair of eyes

- Review whether missing-target behavior should fail earlier in resolver or remain delegated to execution-time file checks.

### What should be done in the future

- Wire these types into run-file and startup execution paths (next tasks in this ticket).

### Code review instructions

- Start with:
  - `pkg/vmpath/path.go`
  - `pkg/vmpath/path_test.go`
- Validate by running:
  - `GOWORK=off go test ./pkg/vmpath -count=1`

### Technical details

- New error surface in `vmpath`:
  - `ErrInvalidRoot`
  - `ErrRootNotDirectory`
  - `ErrEmptyRelativePath`
  - `ErrAbsoluteRelativePath`
  - `ErrTraversalRelativePath`
  - `ErrPathEscapesRoot`

## Step 7: Task 2 - run-file typed resolver integration and symlink-escape test

After introducing typed path primitives, I integrated them into run-file normalization so execution uses canonicalized, validated relative paths. I paired this with an integration test extension that verifies symlink escapes are rejected with `INVALID_PATH`.

This step directly addresses one of the highest-severity findings from VM-006 by moving symlink handling into a typed resolver path instead of ad-hoc string checks.

### Prompt Context

**User prompt (verbatim):** (same as Step 6)

**Assistant interpretation:** Apply the new type-safe path model in real runtime logic and prove it with tests.

**Inferred user intent:** Ensure type-system improvements are practical and close real security gaps.

**Commit (code):** Pending in this step (task commit follows implementation + diary update)

### What I did

- Updated `pkg/vmcontrol/execution_service.go`:
  - `normalizeRunFilePath` now uses:
    - `vmpath.NewWorktreeRoot`
    - `vmpath.ParseRelWorktreePath`
    - `WorktreeRoot.Resolve`
  - mapped typed path errors to domain errors:
    - absolute/traversal -> `vmmodels.ErrPathTraversal`
    - empty relative path -> `vmmodels.ErrFileNotFound`
    - resolved escape -> `vmmodels.ErrPathTraversal`
- Extended safety integration coverage in:
  - `pkg/vmtransport/http/server_safety_integration_test.go`
  - added symlink escape case:
    - symlink inside worktree -> file outside worktree
    - expected API response: `422 INVALID_PATH`
- Ran:
  - `gofmt -w pkg/vmcontrol/execution_service.go pkg/vmtransport/http/server_safety_integration_test.go`
  - `GOWORK=off go test ./pkg/vmtransport/http -run TestSafetyPathTraversalAndOutputLimitEnforcement -count=1`
  - `GOWORK=off go test ./... -count=1`

### Why

- String-prefix traversal checks do not protect against symlink indirection.
- Integrating typed resolver in the run-file path is the minimal high-impact change to close this class of escape.

### What worked

- All tests passed after integration.
- Safety integration now verifies both direct traversal and symlink traversal behavior.

### What didn't work

- N/A for this step.

### What I learned

- Canonicalization plus root-relative validation is straightforward to compose once path parsing is strongly typed.

### What was tricky to build

- Error mapping required care to preserve existing API semantics (`INVALID_PATH` vs `FILE_NOT_FOUND`) while moving to typed resolver errors.

### What warrants a second pair of eyes

- Confirm domain-error mapping policy for empty relative paths (`.`/blank) should remain `FILE_NOT_FOUND` rather than `INVALID_PATH`.

### What should be done in the future

- Apply same typed resolver path to startup file execution and startup-path admission checks (next task).

### Code review instructions

- Review normalization flow in:
  - `pkg/vmcontrol/execution_service.go`
- Review new safety assertion in:
  - `pkg/vmtransport/http/server_safety_integration_test.go`
- Re-run:
  - `GOWORK=off go test ./pkg/vmtransport/http -run TestSafetyPathTraversalAndOutputLimitEnforcement -count=1`

### Technical details

- New test branch validates a symlink named `escape-link.js` in worktree pointing to an external JS file and asserts `code: INVALID_PATH`.

## Step 8: Task 3 - startup path validation/resolution with typed model

I implemented typed startup-path handling at both API ingress and runtime execution. This closes the previous mismatch where startup paths were less constrained than run-file paths and could traverse or escape via symlink.

I also expanded the safety integration test to cover startup traversal rejection and startup symlink-escape rejection at session-create time.

### Prompt Context

**User prompt (verbatim):** (same as Step 6)

**Assistant interpretation:** Extend type-safe path guarantees to startup file lifecycle, not just run-file execution.

**Inferred user intent:** Make path safety consistent and robust across all execution entry points.

**Commit (code):** Pending in this step (task commit follows implementation + diary update)

### What I did

- Updated API validation in:
  - `pkg/vmtransport/http/server.go`
  - `handleTemplateAddStartupFile` now parses startup paths with `vmpath.ParseRelWorktreePath`
  - invalid absolute/traversal/empty path now returns `422 INVALID_PATH`
  - persisted path now uses normalized typed relative path
- Updated startup runtime execution in:
  - `pkg/vmsession/session.go`
  - `runStartupFiles` now:
    - constructs `WorktreeRoot` once
    - parses each startup file path via `ParseRelWorktreePath`
    - resolves each path via `WorktreeRoot.Resolve`
    - executes canonical absolute resolved path
  - path escape errors are mapped to `vmmodels.ErrPathTraversal` for typed transport mapping
- Expanded integration test in:
  - `pkg/vmtransport/http/server_safety_integration_test.go`
  - added checks for:
    - startup path `../outside-startup.js` rejected at template startup add (`422 INVALID_PATH`)
    - startup symlink in worktree pointing outside rejected on session create (`422 INVALID_PATH`)
- Ran:
  - `gofmt -w pkg/vmtransport/http/server.go pkg/vmsession/session.go pkg/vmtransport/http/server_safety_integration_test.go`
  - `GOWORK=off go test ./pkg/vmtransport/http -run TestSafetyPathTraversalAndOutputLimitEnforcement -count=1`
  - `GOWORK=off go test ./... -count=1`

### Why

- Startup files are code execution inputs and need the same boundary guarantees as run-file inputs.
- Typed validation at ingress + typed canonical resolution at runtime gives defense in depth.

### What worked

- Safety integration tests passed with new startup path assertions.
- API now rejects traversal startup paths earlier and more explicitly.

### What didn't work

- N/A for this step.

### What I learned

- Mapping typed path failures back to `vmmodels.ErrPathTraversal` allows existing transport error mapping to remain simple and consistent.

### What was tricky to build

- Deciding where to enforce each check required balancing UX and safety:
  - ingress validation catches obvious invalid paths early,
  - runtime canonicalization is still required to catch symlink-based escapes.

### What warrants a second pair of eyes

- Confirm that returning `422 INVALID_PATH` on session-create startup failure is preferred over generic startup failure semantics.

### What should be done in the future

- Add a dedicated integration test asserting startup syntax errors still map to non-path errors (to avoid overfitting safety tests to path-only failures).

### Code review instructions

- Review startup add validation in:
  - `pkg/vmtransport/http/server.go`
- Review startup execution path in:
  - `pkg/vmsession/session.go`
- Review test additions in:
  - `pkg/vmtransport/http/server_safety_integration_test.go`

### Technical details

- Startup symlink safety test path:
  - `startup-link.js` -> symlink to outside file
  - add-startup returns `201` (path is syntactically valid)
  - session create returns `422 INVALID_PATH` due canonical resolution escape detection.

## Step 9: Task 4 - typed execution-not-found error contract

I implemented a typed not-found error for executions and propagated it through store and transport mapping. This removes the prior behavior where missing execution IDs surfaced as `500 INTERNAL`.

The integration contract now aligns with other not-found resources in the API (`template` and `session` already had typed 404 mappings).

### Prompt Context

**User prompt (verbatim):** (same as Step 6)

**Assistant interpretation:** Apply stronger typing to error contracts to improve robustness and API consistency.

**Inferred user intent:** Replace stringly/implicit error behavior with explicit typed domain errors.

**Commit (code):** Pending in this step (task commit follows implementation + diary update)

### What I did

- Added new typed model error:
  - `pkg/vmmodels/models.go`
  - `ErrExecutionNotFound = errors.New(\"execution not found\")`
- Updated persistence mapping:
  - `pkg/vmstore/vmstore.go`
  - `GetExecution` now returns `vmmodels.ErrExecutionNotFound` on `sql.ErrNoRows`
- Updated transport mapping:
  - `pkg/vmtransport/http/server.go`
  - `writeCoreError` now maps `ErrExecutionNotFound` -> `404 EXECUTION_NOT_FOUND`
- Updated integration expectation:
  - `pkg/vmtransport/http/server_executions_integration_test.go`
  - missing execution check now expects `404` + `EXECUTION_NOT_FOUND`
- Ran:
  - `gofmt -w pkg/vmmodels/models.go pkg/vmstore/vmstore.go pkg/vmtransport/http/server.go pkg/vmtransport/http/server_executions_integration_test.go`
  - `GOWORK=off go test ./pkg/vmtransport/http -run TestExecutionEndpointsLifecycle -count=1`
  - `GOWORK=off go test ./... -count=1`

### Why

- Typed domain errors make API behavior explicit and stable.
- Returning `404` for missing execution is semantically correct and simplifies client behavior.

### What worked

- Integration tests passed with updated expected contract.
- Full test suite remained green.

### What didn't work

- N/A for this step.

### What I learned

- Small typed error additions can remove broad error-class ambiguity with minimal code churn.

### What was tricky to build

- Ensuring mapping consistency required touching model, store, transport, and test layers together; partial changes would have produced inconsistent behavior.

### What warrants a second pair of eyes

- Confirm whether API docs/guides should explicitly enumerate the new `EXECUTION_NOT_FOUND` error code.

### What should be done in the future

- Apply same typed-error discipline to any remaining generic `INTERNAL` fallback cases that are really contract-level errors.

### Code review instructions

- Check error declaration in:
  - `pkg/vmmodels/models.go`
- Check store mapping in:
  - `pkg/vmstore/vmstore.go`
- Check HTTP mapping in:
  - `pkg/vmtransport/http/server.go`
- Check test assertion in:
  - `pkg/vmtransport/http/server_executions_integration_test.go`

### Technical details

- Contract change:
  - before: missing execution -> `500 INTERNAL`
  - after: missing execution -> `404 EXECUTION_NOT_FOUND`
