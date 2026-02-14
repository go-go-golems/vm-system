# Changelog

## 2026-02-14

- Initial workspace created


## 2026-02-14

Completed deep dual-repo analysis and authored merge design recommending history-preserving subdirectory import (subtree primary) plus phased Go-served SPA integration plan; updated detailed diary and task tracking.

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ttmp/2026/02/14/VM-017-MERGE-UI-VM--merge-vm-system-ui-into-vm-system/design-doc/01-vm-system-vm-system-ui-merge-integration-design-and-analysis.md — Primary analysis deliverable
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ttmp/2026/02/14/VM-017-MERGE-UI-VM--merge-vm-system-ui-into-vm-system/reference/01-diary.md — Detailed implementation diary
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ttmp/2026/02/14/VM-017-MERGE-UI-VM--merge-vm-system-ui-into-vm-system/tasks.md — Execution checklist


## 2026-02-14

Uploaded design+diary bundle to reMarkable as 'VM-017-MERGE-UI-VM Analysis and Integration Plan.pdf'. Cloud listing verification attempted twice but blocked by DNS resolution errors to remarkable cloud endpoints in current environment.

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ttmp/2026/02/14/VM-017-MERGE-UI-VM--merge-vm-system-ui-into-vm-system/reference/01-diary.md — Diary captures upload and verification errors
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ttmp/2026/02/14/VM-017-MERGE-UI-VM--merge-vm-system-ui-into-vm-system/tasks.md — Marked upload done and verification blocked


## 2026-02-14

Cleared reMarkable cloud verification blocker, created implementation branch vm-017-merge-ui-implementation, and expanded ticket task list into commit-sized implementation checkpoints.

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ttmp/2026/02/14/VM-017-MERGE-UI-VM--merge-vm-system-ui-into-vm-system/reference/01-diary.md — Step 7 records checkpoint kickoff
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ttmp/2026/02/14/VM-017-MERGE-UI-VM--merge-vm-system-ui-into-vm-system/tasks.md — Implementation checklist and verification status


## 2026-02-14

Imported vm-system-ui into vm-system as ui/ using git subtree with history preserved; handled dirty-tree precondition by temporarily stashing unrelated README.md and restoring it after import.

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ttmp/2026/02/14/VM-017-MERGE-UI-VM--merge-vm-system-ui-into-vm-system/reference/01-diary.md — Step 8 records subtree import workflow
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ttmp/2026/02/14/VM-017-MERGE-UI-VM--merge-vm-system-ui-into-vm-system/tasks.md — Task 9 checked
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ui/package.json — Imported frontend project root
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ui/vite.config.ts — Imported frontend build/proxy config


## 2026-02-14

Implemented internal/web static serving layer (disk + embed) and wired cmd serve to compose API and SPA handlers without shadowing /api; checked tasks 10 and 11 and validated compilation.

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/cmd/vm-system/cmd_serve.go — Serve command handler composition
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/internal/web/publicfs_disk.go — Disk-mode public FS resolver
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/internal/web/publicfs_embed.go — Embed-mode public FS loader
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/internal/web/spa.go — SPA/static handler with index fallback and /api bypass
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ttmp/2026/02/14/VM-017-MERGE-UI-VM--merge-vm-system-ui-into-vm-system/tasks.md — Tasks 10 and 11 checked


## 2026-02-14

Added go-generate frontend bridge under internal/web/tools to build ui and copy dist/public into internal/web/embed/public; checked task 12 and validated package compilation.

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/internal/web/generate.go — go:generate entrypoint
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/internal/web/tools/main.go — Frontend build-and-copy implementation
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ttmp/2026/02/14/VM-017-MERGE-UI-VM--merge-vm-system-ui-into-vm-system/tasks.md — Task 12 checked


## 2026-02-14

Added root Makefile targets for merged backend/frontend workflow (dev, install/check/build, web-generate, embed build) and checked task 13.

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/Makefile — Developer and CI command entrypoints
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ttmp/2026/02/14/VM-017-MERGE-UI-VM--merge-vm-system-ui-into-vm-system/tasks.md — Task 13 checked

