---
Title: Cleanup vm-system and vm-system-ui code quality mess
Ticket: VM-012-CLEANUP-MESS
Status: complete
Topics:
    - backend
    - frontend
    - architecture
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ttmp/2026/02/08/VM-012-CLEANUP-MESS--cleanup-vm-system-and-vm-system-ui-code-quality-mess/design/01-cleanup-audit-report.md
      Note: Primary findings and phased cleanup plan
    - Path: ttmp/2026/02/08/VM-012-CLEANUP-MESS--cleanup-vm-system-and-vm-system-ui-code-quality-mess/reference/01-diary.md
      Note: Detailed execution log and reproducibility context
ExternalSources: []
Summary: |
    Exhaustive audit ticket for vm-system + vm-system-ui cleanup; includes prioritized findings, implementation diary, and phased remediation plan.
LastUpdated: 2026-02-08T23:45:57.790004115-05:00
WhatFor: Track and execute the cross-repo cleanup effort with clear severity and sequencing.
WhenToUse: Use as the entry point for VM-012 status, links, and execution tasks.
---



# Cleanup vm-system and vm-system-ui code quality mess

## Overview

This ticket captures a full-surface audit across `vm-system/` and `vm-system-ui/` (code, markdown, configuration, scripts, artifacts). The primary deliverable is a prioritized cleanup report with concrete file references and remediation sketches.

## Key Links

- Audit report: [design/01-cleanup-audit-report.md](./design/01-cleanup-audit-report.md)
- Detailed diary: [reference/01-diary.md](./reference/01-diary.md)
- Task list: [tasks.md](./tasks.md)
- Change log: [changelog.md](./changelog.md)

## Status

Current status: **active**

Current assessment:
- P0 blockers identified (UI build/type failures, gitlink/submodule mismatch).
- Cleanup plan drafted in phased order with exit criteria.
- Report and diary are complete; reMarkable upload completed and verified.

## Topics

- backend
- frontend
- architecture

## Structure

- `design/` - architecture and cleanup plans
- `reference/` - diary and reproducibility notes
- `playbooks/` - operational runbooks (future)
- `scripts/` - temporary automation for this ticket (future)
- `various/` - scratch notes
- `archive/` - deprecated/reference-only artifacts
