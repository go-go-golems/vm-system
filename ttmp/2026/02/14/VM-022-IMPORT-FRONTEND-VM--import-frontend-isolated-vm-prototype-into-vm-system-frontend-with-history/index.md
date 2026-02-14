---
Title: Import frontend isolated VM prototype into vm-system/frontend with history
Ticket: VM-022-IMPORT-FRONTEND-VM
Status: complete
Topics:
    - frontend
    - architecture
    - integration
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: go-go-labs/cmd/experiments/2026-02-08--simulated-communication/plugin-playground/package.json
      Note: Source subtree root for import
    - Path: vm-system/frontend/client/src/pages/WorkbenchPage.tsx
      Note: Core imported workbench page now under frontend/
    - Path: vm-system/frontend/package.json
      Note: Imported frontend root created by VM-022 merge
    - Path: vm-system/frontend/packages/plugin-runtime/src/runtimeService.ts
      Note: Imported runtime service validated by unit/integration tests
    - Path: vm-system/ttmp/2026/02/14/VM-022-IMPORT-FRONTEND-VM--import-frontend-isolated-vm-prototype-into-vm-system-frontend-with-history/analysis/01-import-strategy-plugin-playground-history-into-vm-system-frontend.md
      Note: Primary analysis and recommended import procedure
    - Path: vm-system/ttmp/2026/02/14/VM-022-IMPORT-FRONTEND-VM--import-frontend-isolated-vm-prototype-into-vm-system-frontend-with-history/reference/01-diary.md
      Note: Frequent investigation diary with command outcomes
    - Path: vm-system/ui/package.json
      Note: Current vm-system frontend package used in overlap assessment
ExternalSources: []
Summary: Analyze and document a history-preserving import of plugin-playground into vm-system/frontend.
LastUpdated: 2026-02-14T18:56:37.219224491-05:00
WhatFor: Ticket hub for VM-022 import analysis, diary, tasks, and changelog.
WhenToUse: Use when planning or executing the frontend VM prototype import into vm-system/frontend.
---




# Import frontend isolated VM prototype into vm-system/frontend with history

## Overview

Analyze and document how to import the isolated frontend VM prototype from `go-go-labs` into `vm-system/frontend` while preserving git history and avoiding `frontend/plugin-playground` nesting.

Current recommendation: use a `git filter-repo` rewrite (`subdirectory-filter` + `to-subdirectory-filter`) in a temporary clone, then merge unrelated histories into `vm-system`.

Execution result: completed on branch `task/vm-022-import-frontend-vm` with import commit `79ef15f` and successful frontend checks/tests/build.

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field
- **Analysis**: `analysis/01-import-strategy-plugin-playground-history-into-vm-system-frontend.md`
- **Diary**: `reference/01-diary.md`

## Status

Current status: **active**

## Topics

- frontend
- architecture
- integration

## Tasks

See [tasks.md](./tasks.md) for the current task list.

Completed so far:

- Ticket created and structured.
- Source/destination inventory collected.
- `subtree` and `filter-repo` workflows dry-run tested in `/tmp`.
- Analysis + diary documented in ticket.
- Import executed into `vm-system/frontend` with preserved path history.
- `pnpm check`, `pnpm test:unit`, `pnpm test:integration`, and `pnpm build` passed in `frontend/`.

## Changelog

See [changelog.md](./changelog.md) for recent changes and decisions.

## Structure

- design/ - Architecture and design documents
- reference/ - Prompt packs, API contracts, context summaries
- playbooks/ - Command sequences and test procedures
- scripts/ - Temporary code and tooling
- various/ - Working notes and research
- archive/ - Deprecated or reference-only artifacts
