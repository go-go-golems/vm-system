---
Title: Remove executor internal duplication with no backwards compatibility
Ticket: VM-007-REFACTOR-EXECUTOR-PIPELINE
Status: active
Topics:
    - backend
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: pkg/vmexec/executor.go
      Note: Primary duplication hotspot and target of pipeline refactor
    - Path: pkg/vmcontrol/execution_service.go
      Note: Caller contract and error/limit semantics impacted by executor refactor
    - Path: pkg/vmtransport/http/server.go
      Note: API behavior contract that must remain intentional after internal refactor
    - Path: pkg/vmstore/vmstore.go
      Note: Persistence write/read paths used by execution lifecycle
    - Path: ttmp/2026/02/08/VM-007-REFACTOR-EXECUTOR-PIPELINE--remove-executor-internal-duplication-with-no-backwards-compatibility/design-doc/01-executor-internal-duplication-inspection-and-implementation-plan.md
      Note: Detailed inspection and implementation plan
ExternalSources: []
Summary: Ticket for removing high executor internal duplication using a single execution pipeline (no backwards compatibility constraints), with explicit runtime/persistence contract hardening.
LastUpdated: 2026-02-08T12:12:26-05:00
WhatFor: Provide a concrete implementation-ready plan to simplify executor internals, reduce divergence, and improve robustness of execution persistence/event behavior.
WhenToUse: Use when implementing vmexec refactor work and related tests for execution lifecycle correctness.
---

# Remove executor internal duplication with no backwards compatibility

## Overview

This ticket defines and tracks a no-backward-compatibility refactor of `pkg/vmexec/executor.go` to eliminate high internal duplication and make execution lifecycle behavior explicit, testable, and consistent across REPL and run-file paths.

The associated design document contains a concrete architecture proposal, acceptance criteria, and a phased implementation plan for delegated execution.

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **active**

## Topics

- backend

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
