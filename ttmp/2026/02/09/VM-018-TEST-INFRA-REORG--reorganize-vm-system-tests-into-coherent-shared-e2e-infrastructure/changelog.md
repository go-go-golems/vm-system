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

