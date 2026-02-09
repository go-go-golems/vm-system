---
Title: Reorganize vm-system tests into coherent shared e2e infrastructure
Ticket: VM-018-TEST-INFRA-REORG
Status: active
Topics:
    - backend
    - architecture
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: README.md
      Note: Updated test command guidance for consolidated scripts
    - Path: docs/getting-started-from-first-vm-to-contributor-guide.md
      Note: Updated script architecture and recommended validation workflow
    - Path: smoke-test.sh
      Note: |-
        Current smoke baseline and upcoming harness migration target
        Migrated to shared harness and stale module assertion removed
    - Path: test-all.sh
      Note: Top-level orchestrator for coherent shell integration suite
    - Path: test-e2e.sh
      Note: |-
        Current full-loop baseline and upcoming harness migration target
        Migrated to shared harness for consistent setup and teardown
    - Path: test-goja-library-execution.sh
      Note: |-
        Legacy overlapping library execution test to consolidate
        Legacy wrapper delegating to consolidated matrix test
    - Path: test-library-loading.sh
      Note: |-
        Legacy overlapping library test to consolidate
        Legacy wrapper delegating to consolidated matrix test
    - Path: test-library-matrix.sh
      Note: Canonical consolidated library capability matrix test
    - Path: test-library-requirements.sh
      Note: |-
        Legacy overlapping library requirements test to consolidate
        Legacy wrapper delegating to consolidated matrix test
    - Path: test/lib/e2e-common.sh
      Note: New shared harness for daemon-first shell integration scripts
    - Path: ttmp/2026/02/09/VM-018-TEST-INFRA-REORG--reorganize-vm-system-tests-into-coherent-shared-e2e-infrastructure/design-doc/01-test-infrastructure-reorganization-analysis-and-implementation-guide.md
      Note: Primary analysis and implementation sequencing for VM-018
    - Path: ttmp/2026/02/09/VM-018-TEST-INFRA-REORG--reorganize-vm-system-tests-into-coherent-shared-e2e-infrastructure/reference/01-diary.md
      Note: Task-by-task execution diary
ExternalSources: []
Summary: |
    Consolidate overlapping shell-based integration scripts into a shared,
    coherent test infrastructure with clear scenario ownership and predictable
    validation commands.
LastUpdated: 2026-02-09T08:16:30-05:00
WhatFor: Create a maintainable test surface that maps directly to architecture risks.
WhenToUse: Use while implementing or reviewing VM-018 test-suite reorganization work.
---





# Reorganize vm-system tests into coherent shared e2e infrastructure

## Overview

VM-018 resolves test-script sprawl and overlap by introducing a shared shell harness, a clear scenario matrix, and a top-level runner for predictable local/CI usage.

## Key Links

- Analysis and implementation guide: [design-doc/01-test-infrastructure-reorganization-analysis-and-implementation-guide.md](./design-doc/01-test-infrastructure-reorganization-analysis-and-implementation-guide.md)
- Diary: [reference/01-diary.md](./reference/01-diary.md)
- Tasks: [tasks.md](./tasks.md)
- Changelog: [changelog.md](./changelog.md)

## Status

Current status: **active**

## Topics

- backend
- architecture
