# Changelog

## 2026-02-09

- Initial workspace created


## 2026-02-09

Scaffolded VM-018 with detailed analysis guide, diary, and task-by-task execution plan for coherent test infrastructure reorganization.

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ttmp/2026/02/09/VM-018-TEST-INFRA-REORG--reorganize-vm-system-tests-into-coherent-shared-e2e-infrastructure/design-doc/01-test-infrastructure-reorganization-analysis-and-implementation-guide.md — Defines overlap analysis and implementation sequencing
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ttmp/2026/02/09/VM-018-TEST-INFRA-REORG--reorganize-vm-system-tests-into-coherent-shared-e2e-infrastructure/reference/01-diary.md — Captures step-by-step execution trail
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ttmp/2026/02/09/VM-018-TEST-INFRA-REORG--reorganize-vm-system-tests-into-coherent-shared-e2e-infrastructure/tasks.md — Establishes task boundaries and acceptance criteria


## 2026-02-09

Implemented shared daemon-first harness (test/lib/e2e-common.sh), migrated smoke/e2e to use it, and removed stale smoke module assertion behavior tied to legacy module catalogs.

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/smoke-test.sh — Now validates stable core behavior and uses template add-module fs instead of legacy console capability path
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/test-e2e.sh — Uses shared harness for deterministic setup and reduced duplication
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/test/lib/e2e-common.sh — Centralizes temporary resource setup


## 2026-02-09

Consolidated three overlapping library scripts into test-library-matrix.sh with explicit JSON built-in and lodash configured/unconfigured assertions; legacy script names now delegate as thin wrappers.

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/test-goja-library-execution.sh — Wrapper migration to consolidated matrix
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/test-library-loading.sh — Wrapper migration to consolidated matrix
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/test-library-matrix.sh — Defines single authoritative library/module capability matrix
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/test-library-requirements.sh — Wrapper migration to consolidated matrix
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ttmp/2026/02/09/VM-018-TEST-INFRA-REORG--reorganize-vm-system-tests-into-coherent-shared-e2e-infrastructure/design-doc/01-test-infrastructure-reorganization-analysis-and-implementation-guide.md — Records wrapper decision in open questions


## 2026-02-09

Added test-all.sh orchestrator and aligned README/getting-started docs with the coherent script architecture (smoke, e2e, library-matrix, full-suite runner).

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/README.md — Now documents full suite runner and individual script commands
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/docs/getting-started-from-first-vm-to-contributor-guide.md — Explains consolidated script responsibilities and debugging order
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/test-all.sh — Runs smoke/e2e/library-matrix scripts and summarizes pass/fail


## 2026-02-09

Final validation complete: GOWORK=off go test ./... -count=1 and ./test-all.sh both passed; VM-018 is ready to close.

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/test-all.sh — Provides full shell integration validation evidence path
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/test-library-matrix.sh — Confirms JSON builtin and lodash configuration semantics
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ttmp/2026/02/09/VM-018-TEST-INFRA-REORG--reorganize-vm-system-tests-into-coherent-shared-e2e-infrastructure/reference/01-diary.md — Captures final validation and closure notes


## 2026-02-09

VM-018 complete: test scripts reorganized into shared harness, matrix coverage, and unified runner with aligned docs.

