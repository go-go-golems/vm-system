---
Title: Expand HTTP CLI verbs and command taxonomy
Ticket: VM-009-EXPAND-HTTP-CLI-VERBS
Status: complete
Topics:
    - backend
    - docs
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ttmp/2026/02/08/VM-009-EXPAND-HTTP-CLI-VERBS--expand-http-cli-verbs-and-command-taxonomy/changelog.md
      Note: Decision history
    - Path: ttmp/2026/02/08/VM-009-EXPAND-HTTP-CLI-VERBS--expand-http-cli-verbs-and-command-taxonomy/design-doc/01-http-cli-verb-coverage-and-taxonomy-expansion-plan.md
      Note: Primary design blueprint
    - Path: ttmp/2026/02/08/VM-009-EXPAND-HTTP-CLI-VERBS--expand-http-cli-verbs-and-command-taxonomy/tasks.md
      Note: Implementation checklist
ExternalSources: []
Summary: ""
LastUpdated: 2026-02-09T00:59:36.353003792-05:00
WhatFor: Track and execute expansion/normalization of HTTP-backed CLI verb coverage and taxonomy in vm-system.
WhenToUse: Use when implementing VM-009 tasks or reviewing CLI/API coverage consistency.
---



# Expand HTTP CLI verbs and command taxonomy

## Overview

This ticket defines follow-up CLI work after introducing the root `http` command namespace.  
Goal: ensure endpoint-to-command coverage is complete, verb semantics are unambiguous (`close` vs `delete`), and docs/tests/scripts enforce canonical command forms.

## Key Links

- Design doc: `design-doc/01-http-cli-verb-coverage-and-taxonomy-expansion-plan.md`
- Tasks: `tasks.md`
- Changelog: `changelog.md`

## Status

Current status: **active**

## Topics

- backend
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
