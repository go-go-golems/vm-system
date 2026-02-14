---
Title: 'Import strategy: plugin-playground history into vm-system/frontend'
Ticket: VM-022-IMPORT-FRONTEND-VM
Status: active
Topics:
    - frontend
    - architecture
    - integration
DocType: analysis
Intent: long-term
Owners: []
RelatedFiles:
    - Path: go-go-labs/cmd/experiments/2026-02-08--simulated-communication/plugin-playground/client/src/pages/WorkbenchPage.tsx
      Note: Representative frontend VM workbench implementation
    - Path: go-go-labs/cmd/experiments/2026-02-08--simulated-communication/plugin-playground/package.json
      Note: Source project root to import into vm-system/frontend
    - Path: vm-system/frontend/client/src/pages/WorkbenchPage.tsx
      Note: Imported isolated VM workbench implementation now in destination repo
    - Path: vm-system/frontend/package.json
      Note: Imported frontend project root in vm-system
    - Path: vm-system/ui/client/src/App.tsx
      Note: Representative existing UI entrypoint compared against imported frontend
    - Path: vm-system/ui/package.json
      Note: Existing vm-system frontend package used for overlap/risk analysis
ExternalSources: []
Summary: Recommends a tested filter-repo workflow to import plugin-playground into frontend/ with queryable history.
LastUpdated: 2026-02-14T15:39:52.266805449-05:00
WhatFor: Decision and runbook for history-preserving frontend import from go-go-labs.
WhenToUse: Use before running the VM-022 import on vm-system and during code review of the migration PR.
---



# VM-022 Analysis: Import plugin-playground into `vm-system/frontend` with history

## Scope and Objective

Import the frontend VM prototype from:

- `go-go-labs/cmd/experiments/2026-02-08--simulated-communication/plugin-playground`

into:

- `vm-system/frontend`

Requirements:

- Preserve commit history (not a squash copy).
- Import directly into `frontend/` (not `frontend/plugin-playground/`).

## Execution Status

- Executed on branch `task/vm-022-import-frontend-vm`.
- Import commit in `vm-system`: `79ef15f`.
- Post-import validation passed in `frontend`:
  - `pnpm check`
  - `pnpm test:unit`
  - `pnpm test:integration`
  - `pnpm build`

## Observed Repository Facts

- Destination repo (`vm-system`) currently has no `frontend/` directory.
- Source path has `40` commits and `104` tracked files on current branch (`task/plugin-playground` in this workspace).
- Existing `vm-system/ui` has `103` tracked files; `27` relative paths overlap with source (e.g. `package.json`, `pnpm-lock.yaml`, `client/src/App.tsx`).  
  This does not block import because destination is a different directory, but it increases future convergence cost if `ui/` and `frontend/` are both maintained.

## Evaluated Import Methods

## 1) `git subtree split` + `git subtree add`

Tested. Content imports correctly into `frontend/`, but path-scoped history under `frontend/` is not preserved as a full chain in practical queries (`git log -- frontend` shows only the import commit in the destination repo).

Conclusion: acceptable only if historical provenance through merged parents is enough and path-local history is not required.

## 2) `git filter-repo` rewrite + merge unrelated histories

Tested and validated in `/tmp`:

- `frontend/` ends with `104` tracked files (expected).
- `git rev-list --count HEAD -- frontend` returns `40` (expected).
- `git log -- frontend` shows the full imported series.

Conclusion: best fit for "import with history into exactly `frontend/`".

## Recommended Procedure (Path-Preserving History)

Run from this workspace root:

```bash
cd /home/manuel/workspaces/2026-02-08/plugin-playground

# 0) Preconditions
cd go-go-labs
git status --short
cd ../vm-system
git status --short

# 1) Create a throwaway filtered clone from go-go-labs
SRC_PREFIX='cmd/experiments/2026-02-08--simulated-communication/plugin-playground'
TMPDIR="$(mktemp -d /tmp/vm022-import-XXXX)"

git clone ../go-go-labs "$TMPDIR/go-go-labs-filtered"
cd "$TMPDIR/go-go-labs-filtered"

# Keep only the experiment subtree and rewrite it under frontend/
git filter-repo --force \
  --subdirectory-filter "$SRC_PREFIX" \
  --to-subdirectory-filter frontend

FILTERED_COMMIT="$(git rev-parse HEAD)"

# 2) Merge rewritten history into vm-system
cd /home/manuel/workspaces/2026-02-08/plugin-playground/vm-system
git checkout -b task/vm-022-import-frontend-vm

git remote add vm022-filtered "$TMPDIR/go-go-labs-filtered"
git fetch vm022-filtered "$FILTERED_COMMIT"

git merge --allow-unrelated-histories --no-ff FETCH_HEAD \
  -m "import(frontend): bring plugin-playground VM prototype with history"

# 3) Verification
git ls-tree -r --name-only HEAD frontend | wc -l
git rev-list --count HEAD -- frontend
git log --oneline -- frontend | head -n 20
```

## Why This Is the Recommendation

- Satisfies "into `frontend/` directly" without extra nesting.
- Preserves commit-level history for path-based inspection tools.
- Avoids subtree edge cases where provenance exists but is not ergonomically visible for `frontend/` path logs.

## Risks and Mitigations

- Risk: two similar frontends (`ui/` and `frontend/`) coexist.  
  Mitigation: define a follow-up plan (converge, archive one, or explicitly scope both).

- Risk: large lockfile and dependencies imported unchanged.  
  Mitigation: import first, then perform a separate normalization pass (dependency dedupe, scripts alignment, build integration).

- Risk: `git filter-repo` not installed in another environment.  
  Mitigation: install via package manager or use a prebuilt filtered mirror branch generated by one maintainer.

## Rollback Plan

If merge result is not accepted:

```bash
git reset --hard HEAD~1
git remote remove vm022-filtered
```

(Use only before sharing/pushing. After push, revert with a new commit instead.)
