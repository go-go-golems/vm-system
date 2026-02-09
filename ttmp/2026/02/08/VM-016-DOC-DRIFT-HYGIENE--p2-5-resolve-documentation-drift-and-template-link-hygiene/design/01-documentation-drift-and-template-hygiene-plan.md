---
Title: Documentation Drift and Template Hygiene Plan
Ticket: VM-016-DOC-DRIFT-HYGIENE
Status: active
Topics:
    - backend
    - frontend
    - architecture
DocType: design
Intent: long-term
Owners: []
RelatedFiles:
    - /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/IMPLEMENTATION_SUMMARY.md
    - /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/cmd/vm-system/main.go
    - /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ttmp/_templates/index.md
    - /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/docs/getting-started-from-first-vm-to-contributor-guide.md
ExternalSources: []
Summary: >
    P2-5 planning ticket to resolve stale docs, fix template-link hygiene, and add
    repeatable markdown validation checks.
LastUpdated: 2026-02-09T10:22:00-05:00
WhatFor: Improve documentation correctness and reduce stale guidance risk.
WhenToUse: Use when executing P2-5 cleanup work.
---

# Documentation Drift and Template Hygiene Plan

## Problem Areas

1. Stale narrative docs
- `IMPLEMENTATION_SUMMARY.md` mentions old CLI group structures that no longer match current command taxonomy.

2. Template markdown link integrity
- `ttmp/_templates/index.md` references sibling files (`tasks.md`, `changelog.md`) that may not exist in template context.

3. Large guide maintenance burden
- `docs/getting-started-from-first-vm-to-contributor-guide.md` is large and likely to drift without automated checks.

## Goals

- Align docs with current behavior and command taxonomy.
- Ensure template-generated docs don’t produce broken-link noise.
- Introduce repeatable link/integrity checks in CI or release checklist.

## Proposed Actions

1. Classify docs by source-of-truth level
- product contract docs
- historical diary/archive docs
- generated template scaffolding

2. Refresh or archive stale docs
- update current contract docs
- move historical one-off docs into clearly marked archive sections

3. Fix template strategy
- either generate required sibling files in scaffold
- or adjust template links to avoid non-existent local targets

4. Add markdown hygiene check
- automated broken-relative-link check across tracked docs
- allowlist intentional template placeholders if needed

## Acceptance Criteria

- Known stale top-level docs are refreshed or explicitly archived.
- Template markdown no longer triggers unresolved link noise.
- A repeatable markdown integrity command is documented and runnable in CI/local checks.
