---
Title: Cleanup Audit Report
Ticket: VM-012-CLEANUP-MESS
Status: active
Topics:
    - backend
    - frontend
    - architecture
DocType: design
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ../../../../../../../vm-system-ui/client/src/App.tsx
      Note: Current live route map
    - Path: ../../../../../../../vm-system-ui/client/src/lib/types.ts
      Note: Competing transport/domain types
    - Path: ../../../../../../../vm-system-ui/client/src/lib/vmService.ts
      Note: Duplicate domain model and service complexity
    - Path: ../../../../../../../vm-system-ui/client/src/pages/SessionDetail.tsx
      Note: Broken import and build failure source
    - Path: ../../../../../../../vm-system-ui/client/src/pages/System.tsx
      Note: Broken import and implicit-any typing issues
    - Path: IMPLEMENTATION_SUMMARY.md
      Note: Potentially stale CLI contract language
    - Path: pkg/vmcontrol/execution_service.go
      Note: Soft-fail limit enforcement behavior
    - Path: pkg/vmsession/session.go
      Note: Startup mode and runtime logging behavior
    - Path: pkg/vmstore/vmstore.go
      Note: Monolithic persistence implementation
    - Path: pkg/vmtransport/http/server.go
      Note: Monolithic HTTP transport and route surface
ExternalSources: []
Summary: |
    Exhaustive code + non-code audit across vm-system and vm-system-ui, with concrete findings, validation failures, and a prioritized cleanup plan.
LastUpdated: 2026-02-09T04:45:00-05:00
WhatFor: |
    Prioritize high-leverage cleanup work and stabilize build, testing, docs, and repository hygiene.
WhenToUse: Use when planning or executing VM-012 cleanup work; this report is the source of truth for priorities.
---


# VM-012 Cleanup Audit Report

## Scope and Method

This audit covered `vm-system/` and `vm-system-ui/` with explicit inclusion of code, markdown, config, scripts, generated artifacts, and ticket docs.

Audit method:
- Full file inventory (`rg --files`) in both repos.
- Pattern sweeps for risky constructs (`any`, `panic`, `fmt.Print*`, TODO/FIXME/deprecated markers).
- Build/test/typing validation (`go test ./...`, `pnpm check`, `pnpm build`).
- Markdown link integrity check across all `*.md` in both repos.
- Deep manual review of high-complexity and runtime-critical files.

Coverage snapshot:
- `vm-system`: 179 files (`113 .md`, `50 .go`, `5 .sh`, `5 .js`, `2 .db`, others).
- `vm-system-ui`: 102 files (`79 .tsx`, `13 .ts`, plus config/html/css/patch/docs).

## Critical Findings (P0)

### P0-1: `vm-system-ui` is not build-green (typecheck + production build failing)

Problem:
- The UI currently fails both typecheck and build, which blocks safe cleanup/refactor work and creates hidden regressions.

Where to look:
- `vm-system-ui/client/src/pages/SessionDetail.tsx:4`
- `vm-system-ui/client/src/pages/System.tsx:2`
- `vm-system-ui/client/src/components/AppShell.tsx:209`
- `vm-system-ui/client/src/lib/vmService.ts:63`
- `vm-system-ui/client/src/lib/types.ts:65`

Example:
```tsx
// SessionDetail.tsx
import { useAppState } from '@/components/AppShell';
```
```tsx
// AppShell.tsx exports AppShell only
export function AppShell({ children }: { children: ReactNode }) {
```

Validation evidence:
- `pnpm check` fails with missing export (`useAppState`) and type incompatibilities (`Date` vs `string` execution fields).
- `pnpm build` fails with missing export from `AppShell`.

Why it matters:
- Cleanup changes cannot be confidently validated until baseline build health is restored.

Cleanup sketch:
```text
1) Remove/replace stale `useAppState` imports in pages using RTK hooks/selectors.
2) Choose one canonical Execution model (Date or string) and remove duplicated competing model.
3) Re-run pnpm check + pnpm build; require green status before broader refactors.
```

### P0-2: Broken repository submodule wiring (`gitlink` without `.gitmodules` mapping)

Problem:
- Repository contains a gitlink entry for `test-goja-workspace` but no submodule mapping, causing git submodule operations to fail.

Where to look:
- `vm-system/.git` (index state)
- `vm-system/test-goja-workspace`

Evidence:
- `git ls-files -s | rg 160000` -> `160000 ... test-goja-workspace`
- `git submodule status` -> `fatal: no submodule mapping found in .gitmodules for path 'test-goja-workspace'`

Why it matters:
- Checkout, cloning, and CI behavior become non-deterministic.

Cleanup sketch:
```text
Either:
A) Convert `test-goja-workspace` to normal tracked directory (remove gitlink, add files), or
B) Add valid `.gitmodules` + initialize submodule intentionally.
Prefer A unless true external-submodule intent exists.
```

## High Findings (P1)

### P1-1: Limit enforcement silently bypasses on load/fetch errors

Problem:
- `ExecutionService.enforceLimits` returns `nil` when limits fail to load or events fail to fetch.

Where to look:
- `vm-system/pkg/vmcontrol/execution_service.go:105`

Example:
```go
limits, err := s.loadSessionLimits(sessionID)
if err != nil {
    // Scaffolding is intentionally soft-fail while limit enforcement matures.
    return nil
}
```

Why it matters:
- Runtime safeguards can be skipped silently under failure conditions.

Cleanup sketch:
```go
if err != nil {
    return fmt.Errorf("limit-enforcement unavailable: %w", err)
}
```
- Fail closed for hard limits, or emit explicit system event + mark execution error with dedicated code.

### P1-2: Startup `import` mode is implemented as `eval` placeholder

Problem:
- Startup file mode `import` is currently executed by `RunString` same as `eval`.

Where to look:
- `vm-system/pkg/vmsession/session.go:240`

Example:
```go
} else if file.Mode == "import" {
    // For now, treat it as eval
    if _, err := session.Runtime.RunString(string(content)); err != nil {
```

Why it matters:
- Behavior diverges from configuration semantics and can mislead users.

Cleanup sketch:
```text
Option A: Implement real module-import semantics.
Option B (faster): reject `import` mode with explicit validation error until implemented.
```

### P1-3: Architectural drift in routing: legacy pages exist but are unreachable / stale

Problem:
- Router only serves `/templates`, `/sessions`, `/system`, `/reference`, while legacy pages still exist and link to `/docs` and `/` flows that are not routed.

Where to look:
- `vm-system-ui/client/src/App.tsx:22`
- `vm-system-ui/client/src/pages/Docs.tsx:36`
- `vm-system-ui/client/src/pages/SystemOverview.tsx:39`
- `vm-system-ui/client/src/pages/Home.tsx:187`

Example:
```tsx
// App.tsx routes
<Route path="/templates" component={Templates} />
<Route path="/sessions" component={Sessions} />
<Route path="/system" component={System} />
<Route path="/reference" component={Reference} />
```

Why it matters:
- Dead pages rot, increase maintenance surface, and reintroduce broken imports/types when touched.

Cleanup sketch:
```text
Inventory + decide for Home/Docs/SystemOverview:
- keep and route intentionally, or
- archive/remove completely.
Then remove stale links and navigation references.
```

### P1-4: Duplicate front-end domain models create type conflicts

Problem:
- `Execution` and related models exist in both `vmService.ts` (Date fields) and `types.ts` (string fields).

Where to look:
- `vm-system-ui/client/src/lib/vmService.ts:63`
- `vm-system-ui/client/src/lib/types.ts:65`

Example:
```ts
// vmService.ts
startedAt: Date;

// types.ts
startedAt: string;
```

Why it matters:
- Causes current typecheck breakage and makes data normalization unclear.

Cleanup sketch:
```text
Create one canonical transport model + one canonical domain model.
Normalization boundary should be centralized in `normalize.ts`.
```

### P1-5: Missing tests in core runtime/persistence/UI paths

Problem:
- No package tests for several runtime-critical Go packages; no UI tests at all.

Where to look:
- Go packages without tests: `pkg/vmsession`, `pkg/vmstore`, `pkg/vmdaemon`, `pkg/vmclient`, `pkg/libloader`.
- UI test files: none (`client`, `server`, `shared` have no `*.test|*.spec`).

Why it matters:
- Refactor risk is high, especially with already-red UI build.

Cleanup sketch:
```text
Go: add focused tests around session lifecycle, store CRUD constraints, and error contracts.
UI: add route smoke tests + data normalization tests + one integration flow for templates/sessions.
```

## Medium Findings (P2)

### P2-1: Monolithic files are refactor bottlenecks

Problem:
- Very large files combine multiple concerns.

Where to look (line counts):
- `vm-system/pkg/vmtransport/http/server.go` (719)
- `vm-system/pkg/vmstore/vmstore.go` (587)
- `vm-system/cmd/vm-system/cmd_template.go` (519)
- `vm-system-ui/client/src/lib/vmService.ts` (1140)
- `vm-system-ui/client/src/pages/Docs.tsx` (632)
- `vm-system-ui/client/src/pages/SystemOverview.tsx` (621)

Why it matters:
- Raises cognitive load and merge conflict probability.

Cleanup sketch:
```text
Split by domain boundaries:
- server handlers by resource (templates/sessions/executions/ops)
- vmstore by aggregate root
- vmService into request client + mappers + feature modules
```

### P2-2: Panic/logging behavior inconsistent with production daemon quality

Problem:
- Panic-based helpers and mixed direct stdout logging in runtime code.

Where to look:
- `vm-system/pkg/vmmodels/ids.go:47`
- `vm-system/pkg/vmsession/session.go:111`
- `vm-system/pkg/vmsession/session.go:294`
- `vm-system/pkg/libloader/loader.go:39`

Example:
```go
func MustTemplateID(raw string) TemplateID {
    id, err := ParseTemplateID(raw)
    if err != nil {
        panic(err)
    }
```

Why it matters:
- Panic paths and ad-hoc logs complicate daemon reliability/observability.

Cleanup sketch:
```text
Keep Must* only for tests/internal invariants.
Use structured logger interface in runtime/loader/session code paths.
```

### P2-3: Repo hygiene issues in `vm-system` (artifacts + ignore policy)

Problem:
- No root `.gitignore`, plus local artifacts are present (`vm-system` binary, `.db` files, test workspaces).

Where to look:
- `vm-system` repo root listing
- `git status --short` output for untracked runtime artifacts.

Why it matters:
- Easy accidental commits and noisy diffs.

Cleanup sketch:
```text
Add root .gitignore for binaries, db, temp workspaces, logs.
Document which fixture dirs are intentionally tracked.
```

### P2-4: Third-party JS cache files are tracked as source

Problem:
- Vendored library blobs in `.vm-cache/libraries/*.js` are tracked.

Where to look:
- `.vm-cache/libraries/axios-1.6.0.js`
- `.vm-cache/libraries/dayjs-1.11.10.js`
- `.vm-cache/libraries/lodash-4.17.21.js`
- `.vm-cache/libraries/moment-2.29.4.js`
- `.vm-cache/libraries/ramda-0.29.0.js`
- `.vm-cache/libraries/zustand-4.4.7.js`

Why it matters:
- Binary-like churn in git, harder dependency governance.

Cleanup sketch:
```text
Move cache to runtime-generated + ignored path; keep only metadata in source.
If offline fixtures are required, place in dedicated fixtures dir with checksum manifest.
```

### P2-5: Markdown/documentation drift and template integrity issues

Problem:
- Historical docs make claims no longer aligned with current CLI shape; template has broken local links.

Where to look:
- `vm-system/IMPLEMENTATION_SUMMARY.md:45` (`vm-system vm` command group)
- `vm-system/cmd/vm-system/main.go:38` (actual root commands: `template`, `session`, `exec`, `ops`, etc.)
- `vm-system/ttmp/_templates/index.md:39` (`./tasks.md` link)
- `vm-system/ttmp/_templates/index.md:43` (`./changelog.md` link)

Why it matters:
- New contributors can follow stale instructions.

Cleanup sketch:
```text
Archive or refresh outdated top-level docs.
Fix template link targets or generate sibling files in template structure.
Add markdown link check to CI.
```

### P2-6: Legacy direct test executable in Go test folder

Problem:
- `test/test_library_loading.go` is a standalone `package main` executable and is not run by `go test`.

Where to look:
- `vm-system/test/test_library_loading.go:1`

Why it matters:
- Looks like test coverage but is not part of automated suite.

Cleanup sketch:
```text
Convert to `_test.go` integration test or move to `examples/` and label as manual diagnostic.
```

### P2-7: Environment placeholder warnings in UI HTML

Problem:
- Analytics placeholders are unresolved in default build.

Where to look:
- `vm-system-ui/client/index.html:20`

Why it matters:
- Build warnings reduce signal quality and can mask real regressions.

Cleanup sketch:
```text
Gate analytics script injection by env presence in Vite transform.
```

## Proposed Cleanup Plan (Sequenced)

### Phase 0: Unblock CI/Local Quality Gates (1-2 days)

1. Make `vm-system-ui` green:
   - Remove stale `useAppState` consumers or reintroduce a typed compatibility hook temporarily.
   - Unify execution/session typing model across `vmService.ts` and `types.ts`.
2. Enforce baseline checks in both repos:
   - `go test ./...`
   - `pnpm check`
   - `pnpm build`

Exit criteria:
- All three commands pass on clean checkout.

### Phase 1: Repo Hygiene + Determinism (0.5-1 day)

1. Resolve `test-goja-workspace` gitlink/submodule mismatch.
2. Add `vm-system/.gitignore` and remove accidental artifact tracking flow.
3. Decide policy for `.vm-cache/libraries` (runtime cache vs tracked fixture set).

Exit criteria:
- `git submodule status` no longer errors.
- Fresh clone does not include runtime artifacts in status after normal dev flow.

### Phase 2: Runtime Safety Corrections (1-2 days)

1. Make limit enforcement explicit (no silent bypass).
2. Replace startup `import` placeholder behavior with strict validation or true implementation.
3. Normalize runtime logging strategy and panic boundaries.

Exit criteria:
- New tests cover failure modes for limits/import.

### Phase 3: Structural Refactors (2-4 days)

1. Split large backend/frontend units by domain modules.
2. Remove dead pages or route them intentionally.
3. Consolidate UI data normalization boundaries.

Exit criteria:
- Largest files reduced below agreed thresholds (example target: <350 LOC for non-generated files).

### Phase 4: Docs + Auditability (0.5-1 day)

1. Refresh/archive stale root docs (`IMPLEMENTATION_SUMMARY.md`, historical diaries as needed).
2. Add markdown link check to CI.
3. Keep ticket docs linked to changed files via `docmgr doc relate`.

Exit criteria:
- No known broken markdown links except explicitly whitelisted template placeholders.

## Suggested Work Breakdown (Ticket Follow-ups)

- VM-012A: Build green + type-model unification in UI.
- VM-012B: Submodule/git hygiene + `.gitignore` policy.
- VM-012C: Runtime limit/import behavior corrections.
- VM-012D: Large-file decomposition pass.
- VM-012E: Test coverage expansion + markdown CI checks.

## Validation Commands Used During Audit

- `cd vm-system && go test ./...`
- `cd vm-system-ui && pnpm check`
- `cd vm-system-ui && pnpm build`
- `rg --files` and pattern scans across both repos
- markdown link scan across all `*.md`

## Final Assessment

Current state is recoverable but not cleanup-ready until P0 items are addressed. The fastest high-leverage path is:
1) restore UI build/type health,
2) fix repository determinism/hygiene, then
3) refactor safely behind stronger test coverage.
