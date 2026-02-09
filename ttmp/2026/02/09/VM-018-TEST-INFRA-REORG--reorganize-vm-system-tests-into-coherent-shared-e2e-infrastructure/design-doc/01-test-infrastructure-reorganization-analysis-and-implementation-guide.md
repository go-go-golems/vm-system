---
Title: Test Infrastructure Reorganization Analysis and Implementation Guide
Ticket: VM-018-TEST-INFRA-REORG
Status: active
Topics:
    - backend
    - architecture
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: README.md
      Note: Public test command entrypoints that need alignment
    - Path: docs/getting-started-from-first-vm-to-contributor-guide.md
      Note: Contributor guidance that currently references old script architecture
    - Path: smoke-test.sh
    - Path: test-e2e.sh
    - Path: test-goja-library-execution.sh
    - Path: test-library-loading.sh
    - Path: test-library-requirements.sh
ExternalSources: []
Summary: |
    Detailed architecture analysis and implementation plan for reducing overlap in vm-system integration scripts and introducing a coherent scenario matrix.
LastUpdated: 2026-02-09T08:16:30-05:00
WhatFor: Drive a task-by-task migration from duplicated scripts to a shared infrastructure.
WhenToUse: Use while implementing VM-018 and when reviewing the resulting test architecture.
---


# VM-018 Test Infrastructure Reorganization Analysis and Implementation Guide

## Executive Summary

Current shell-driven integration coverage is valuable but fragmented. `smoke-test.sh`, `test-e2e.sh`, and three library-focused scripts all duplicate daemon/bootstrap/worktree/session scaffolding and overlap heavily in scenario coverage.

VM-018 introduces a single shared harness for lifecycle setup, a consolidated library capability matrix test, and one top-level test runner. The goal is useful tests with explicit ownership per scenario instead of five partially redundant scripts.

## Problem Statement

The current script set has three systemic issues:

1. Setup duplication and drift risk.
- Every script re-implements dynamic port allocation, daemon startup, health checks, temp DB/worktree creation, and cleanup.
- Behavior changes (for example module catalog policy) require edits in multiple places and stale assertions remain undetected.

2. Scenario overlap without explicit ownership.
- `smoke-test.sh` and `test-e2e.sh` both execute similar happy paths.
- `test-library-loading.sh`, `test-goja-library-execution.sh`, and `test-library-requirements.sh` all cover closely related library-load behavior with different levels of rigor.

3. Operational ambiguity for contributors.
- It is not obvious which script should be run for fast signal versus deep capability validation.
- Documentation references become stale when scripts diverge.

## Current-State Analysis

### Script overlap map

- `smoke-test.sh`
  - Should be fast sanity, but currently includes stale configurable-module assertion.
- `test-e2e.sh`
  - Full happy-path loop; overlaps significantly with smoke.
- `test-library-loading.sh`
  - Basic lodash path.
- `test-goja-library-execution.sh`
  - Rich lodash execution behavior.
- `test-library-requirements.sh`
  - With/without library semantics and post-hoc/multi-library checks.

### Architecture risk coverage gaps

- Strong today:
  - daemon-first lifecycle,
  - session creation and execution,
  - basic runtime/library behavior.
- Weak today:
  - maintainability and confidence in assertions due to duplication,
  - clear entry points for contributors/CI.

## Proposed Solution

### Design principles

1. Single source of truth for setup primitives.
- Introduce `test/lib/e2e-common.sh` with reusable helpers for temp resources, daemon lifecycle, command wrappers, and assertions.

2. Explicit scenario ownership.
- Keep only three script entry points:
  - `smoke-test.sh`: fast core health path.
  - `test-e2e.sh`: full command lifecycle.
  - `test-library-matrix.sh`: library/module capability matrix, including required negative-path checks.

3. Deterministic operator workflow.
- Add `test-all.sh` as top-level orchestrator with strict ordering and clear failure boundaries.

4. Documentation follows behavior.
- Update README and contributor guide to describe script intent and recommended execution order.

## Design Decisions

1. Keep shell scripts (do not migrate to Go test harness in this ticket).
- Rationale: quickest path to coherence with minimal architectural churn.

2. Consolidate library scripts into one matrix script.
- Rationale: current three scripts contain repeated scaffolding and can be merged without coverage loss.

3. Do not assert legacy module names in smoke.
- Rationale: smoke should validate stable runtime outcomes, not policy details that belong in focused matrix tests.

4. Use task-sized commits with diary updates.
- Rationale: explicit traceability for each migration phase.

## Alternatives Considered

1. Keep all existing scripts and only patch stale assertions.
- Rejected: does not solve duplication or maintenance burden.

2. Move all e2e scripts to Go integration tests.
- Rejected for VM-018 scope: useful longer-term, but too large for immediate stabilization.

3. Keep three library scripts but share only helper functions.
- Rejected: still leaves too many entry points with overlapping intent.

## Implementation Plan

## Task 1: Ticket scaffold and execution plan

Deliverables:
- Ticket created.
- Analysis + diary initialized.
- Task list finalized with commit boundaries.

Acceptance:
- VM-018 has actionable docs and reviewable task list.

## Task 2: Shared harness + smoke/e2e migration

Deliverables:
- `test/lib/e2e-common.sh` helpers.
- `smoke-test.sh` migrated to harness and stale module assertion removed.
- `test-e2e.sh` migrated to harness.

Acceptance:
- smoke/e2e pass using shared harness.

## Task 3: Library script consolidation

Deliverables:
- New `test-library-matrix.sh` replacing overlapping coverage from prior three scripts.
- Legacy scripts removed or converted to thin wrappers (decision documented in commit).

Acceptance:
- Matrix script validates both positive and negative library behavior and passes locally.

## Task 4: Runner + docs alignment

Deliverables:
- `test-all.sh` runs smoke, e2e, and matrix scripts in order.
- README + contributor docs updated with coherent test architecture and run commands.

Acceptance:
- A contributor can run one command for full shell-based integration validation.

## Task 5: Validation and closure

Deliverables:
- Validation evidence captured (`go test`, shell scripts).
- Tasks checked, diary/changelog updated, ticket closed.

Acceptance:
- VM-018 task list fully checked and status set complete.

## Open Questions

1. Should legacy script names be preserved as wrappers or removed immediately?
- Proposed answer for VM-018: preserve compatibility via wrappers only if negligible maintenance overhead; otherwise remove and update docs.

2. Should `libs download` run in smoke?
- Proposed answer for VM-018: no; keep smoke focused on core lifecycle and move library concerns to matrix script.

## References

- Existing scripts in repo root:
  - `smoke-test.sh`
  - `test-e2e.sh`
  - `test-goja-library-execution.sh`
  - `test-library-loading.sh`
  - `test-library-requirements.sh`
