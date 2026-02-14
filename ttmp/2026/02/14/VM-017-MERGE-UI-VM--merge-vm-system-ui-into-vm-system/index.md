---
Title: Merge vm-system-ui into vm-system
Ticket: VM-017-MERGE-UI-VM
Status: active
Topics:
    - architecture
    - frontend
    - backend
    - integration
    - monorepo
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: vm-system/vm-system-ui/vite.config.ts
      Note: Frontend proxy/build baseline
    - Path: vm-system/vm-system/pkg/vmtransport/http/server.go
      Note: Backend API route contract baseline
    - Path: vm-system/vm-system/ttmp/2026/02/14/VM-017-MERGE-UI-VM--merge-vm-system-ui-into-vm-system/design-doc/01-vm-system-vm-system-ui-merge-integration-design-and-analysis.md
      Note: Primary design and merge strategy analysis
    - Path: vm-system/vm-system/ttmp/2026/02/14/VM-017-MERGE-UI-VM--merge-vm-system-ui-into-vm-system/reference/01-diary.md
      Note: Step-by-step implementation diary
ExternalSources: []
Summary: Analyze and design the best path to merge vm-system-ui into vm-system as a subdirectory while preserving history and converging deployment/runtime architecture.
LastUpdated: 2026-02-14T16:44:00-05:00
WhatFor: Provide an executable blueprint for repository merge mechanics and post-merge runtime integration.
WhenToUse: Use when preparing the merge PR, sequencing implementation phases, and validating follow-up tasks.
---


# Merge vm-system-ui into vm-system

## Overview

This ticket covers the architecture analysis and design for merging `vm-system-ui` into `vm-system`. The recommended path is a history-preserving import into `ui/` (subtree-first), followed by phased production consolidation where the Go daemon serves the SPA while preserving existing `/api/v1` contracts.

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field
- **Design Deliverable**: `design-doc/01-vm-system-vm-system-ui-merge-integration-design-and-analysis.md`
- **Diary**: `reference/01-diary.md`

## Status

Current status: **active**
Analysis and documentation complete; reMarkable upload pending.

## Topics

- architecture
- frontend
- backend
- integration
- monorepo

## Tasks

See [tasks.md](./tasks.md) for the current task list.

## Changelog

See [changelog.md](./changelog.md) for recent changes and decisions.

## Structure

- design/ - Architecture and design documents
- reference/ - Prompt packs, API contracts, context summaries
- playbooks/ - Command sequences and test procedures
- scripts/ - Temporary code and tooling
- various/ - Working notes and research
- archive/ - Deprecated or reference-only artifacts
