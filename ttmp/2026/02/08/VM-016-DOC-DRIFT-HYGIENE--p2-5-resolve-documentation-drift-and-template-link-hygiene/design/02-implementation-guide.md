---
Title: Implementation Guide
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
    Detailed implementation guide for resolving documentation drift and enforcing
    template/link hygiene checks.
LastUpdated: 2026-02-09T10:43:00-05:00
WhatFor: Execute P2-5 documentation cleanup with repeatable validation.
WhenToUse: Use while implementing VM-016.
---

# VM-016 Implementation Guide

## Objective

Make docs trustworthy and maintainable by aligning content with current behavior and enforcing automated link hygiene.

## Step 1: Source-of-Truth Audit

1. Identify contract-bearing docs (must stay current):
- CLI/HTTP behavior docs
- contributor workflow docs

2. Identify historical docs (can be archived):
- one-off implementation diaries
- obsolete summaries

Deliverable:
- classification table (`contract` vs `historical`) for top-level docs.

## Step 2: Drift Remediation

1. Compare docs against current commands/routes:
- use `cmd/vm-system/main.go` and live route definitions as truth.

2. For each stale section:
- update if still relevant
- archive with explicit “historical” marker if obsolete.

Targets to check first:
- `IMPLEMENTATION_SUMMARY.md`
- `docs/getting-started-from-first-vm-to-contributor-guide.md`

## Step 3: Template Hygiene Fix

1. Resolve template link issue in `ttmp/_templates/index.md`:
- either ensure scaffold always creates linked siblings (`tasks.md`, `changelog.md`), or
- remove/replace links with non-broken placeholders.

2. Keep template behavior consistent with `docmgr ticket create-ticket` outputs.

## Step 4: Add Repeatable Markdown Validation

1. Add local command to run markdown link checks for tracked docs.
2. Wire check into CI or release checklist.
3. Define allowlist strategy for intentional unresolved placeholders.

Suggested acceptance command set:
- markdown link check script/tool over `*.md`
- `docmgr doctor --ticket <id> --stale-after 30`

## Step 5: Closeout

- Update affected ticket docs/changelogs.
- Verify no broken links in contract docs.
- Ensure archived docs are clearly marked as non-authoritative.

## Risks and Mitigations

- Risk: aggressive edits accidentally remove useful context.
  - Mitigation: archive rather than delete when uncertain.
- Risk: template change breaks existing docmgr flows.
  - Mitigation: run ticket creation smoke checks after template edits.
