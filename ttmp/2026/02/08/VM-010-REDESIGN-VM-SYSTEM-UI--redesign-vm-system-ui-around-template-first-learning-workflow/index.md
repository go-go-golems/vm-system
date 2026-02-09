---
Title: Redesign vm-system-ui around template-first learning workflow
Ticket: VM-010-REDESIGN-VM-SYSTEM-UI
Status: active
Topics:
    - frontend
    - docs
DocType: index
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: ""
LastUpdated: 2026-02-08T22:15:23.013236221-05:00
WhatFor: ""
WhenToUse: ""
---

# Redesign vm-system-ui around template-first learning workflow

## Overview

Redesign vm-system-ui to match the current backend object model and learner mental model.  
Primary product goal: make template management explicit, then session creation/interaction explicit, so first-time users understand runtime object boundaries.

Core design thesis:

- Template is configuration source-of-truth.
- Session is runtime instance created from template.
- Execution and execution events are session-scoped runtime outputs.

This ticket contains a full product/UX plan with implementation phases and ASCII wireframes.

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field
- **Design Plan**: [design-doc/01-template-first-learner-ui-redesign-plan-with-object-model-and-wireframes.md](./design-doc/01-template-first-learner-ui-redesign-plan-with-object-model-and-wireframes.md)

## Status

Current status: **active**

## Topics

- frontend
- docs

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
