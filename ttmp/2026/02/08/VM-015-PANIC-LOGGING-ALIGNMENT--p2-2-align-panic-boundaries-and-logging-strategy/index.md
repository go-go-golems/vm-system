---
Title: 'P2-2: Align panic boundaries and logging strategy'
Ticket: VM-015-PANIC-LOGGING-ALIGNMENT
Status: active
Topics:
    - backend
    - architecture
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: cmd/vm-system/cmd_serve.go
      Note: Structured daemon startup logging
    - Path: cmd/vm-system/glazed_support.go
      Note: Removed panic-based command construction in favor of error-returning helpers and fallback error handling
    - Path: pkg/libloader/loader.go
      Note: Structured logging for library cache download lifecycle
    - Path: pkg/vmmodels/ids.go
      Note: Removed Must* panic helpers to enforce error-returning ID parsing
    - Path: pkg/vmmodels/ids_test.go
      Note: Updated tests to validate parse errors instead of panic behavior
    - Path: pkg/vmsession/session.go
      Note: Structured logging for startup console and library runtime load
    - Path: ttmp/2026/02/08/VM-015-PANIC-LOGGING-ALIGNMENT--p2-2-align-panic-boundaries-and-logging-strategy/design/01-panic-boundary-and-logging-alignment-plan.md
      Note: Policy and migration plan
    - Path: ttmp/2026/02/08/VM-015-PANIC-LOGGING-ALIGNMENT--p2-2-align-panic-boundaries-and-logging-strategy/design/02-implementation-guide.md
      Note: Updated inventory to current post-completion state
    - Path: ttmp/2026/02/08/VM-015-PANIC-LOGGING-ALIGNMENT--p2-2-align-panic-boundaries-and-logging-strategy/reference/01-diary.md
      Note: Task-by-task VM-015 implementation diary
ExternalSources: []
Summary: P2-2 planning ticket for panic/logging cleanup.
LastUpdated: 2026-02-09T10:20:00-05:00
WhatFor: Track panic boundary and logging consistency work.
WhenToUse: Entry point for P2-2 execution.
---



# P2-2: Align panic boundaries and logging strategy

## Overview

Define and execute panic/logging policy changes for safer runtime behavior and consistent observability.

## Key Links

- Plan: [design/01-panic-boundary-and-logging-alignment-plan.md](./design/01-panic-boundary-and-logging-alignment-plan.md)
- Implementation guide: [design/02-implementation-guide.md](./design/02-implementation-guide.md)
- Tasks: [tasks.md](./tasks.md)
- Changelog: [changelog.md](./changelog.md)
