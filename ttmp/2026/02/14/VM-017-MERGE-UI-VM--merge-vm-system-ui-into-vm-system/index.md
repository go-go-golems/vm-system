---
Title: Merge vm-system-ui into vm-system
Ticket: VM-017-MERGE-UI-VM
Status: complete
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
LastUpdated: 2026-02-14T18:56:42.029286671-05:00
WhatFor: Provide and track an executable blueprint plus implementation results for repository/runtime consolidation.
WhenToUse: Use when reviewing VM-017 implementation outcomes, remaining tasks, and follow-up planning.
---



# Merge vm-system-ui into vm-system

## Overview

This ticket started as architecture/design analysis and now includes an active implementation stream. `vm-system-ui` has been imported into `ui/`, daemon-side static serving and `go generate` asset bridging are in place, and merged dev/build commands are available via root `Makefile` targets.

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field
- **Design Deliverable**: `design-doc/01-vm-system-vm-system-ui-merge-integration-design-and-analysis.md`
- **Diary**: `reference/01-diary.md`

## Status

Current status: **active**
Implementation mostly complete; final documentation polish and implementation artifact upload remain.

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
