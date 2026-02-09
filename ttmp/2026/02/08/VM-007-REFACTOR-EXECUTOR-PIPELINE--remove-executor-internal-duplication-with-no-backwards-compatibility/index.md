---
Title: Remove executor and core model/helper duplication with no backwards compatibility
Ticket: VM-007-REFACTOR-EXECUTOR-PIPELINE
Status: complete
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
      Note: Persistence write/read paths plus duplicated JSON helper semantics
    - Path: pkg/vmcontrol/types.go
      Note: Shared config aliases and JSON helper behavior in control layer
    - Path: pkg/vmmodels/models.go
      Note: Core config model types used as single source-of-truth target
    - Path: ttmp/2026/02/08/VM-007-REFACTOR-EXECUTOR-PIPELINE--remove-executor-internal-duplication-with-no-backwards-compatibility/design-doc/01-executor-internal-duplication-inspection-and-implementation-plan.md
      Note: Detailed inspection and implementation plan
ExternalSources: []
Summary: Ticket for removing high executor internal duplication and remaining core model/helper duplication (finding 8 + 9) with no backwards compatibility constraints.
LastUpdated: 2026-02-09T00:59:36.345223748-05:00
WhatFor: Provide a concrete implementation-ready plan to simplify executor internals and eliminate duplicated core helper/model behavior.
WhenToUse: Use when implementing vmexec refactor work plus shared model/helper deduplication tasks.
---


# Remove executor and core model/helper duplication with no backwards compatibility

## Overview

This ticket defines and tracks a no-backward-compatibility refactor that combines two VM-006 findings:

- finding 9: high executor internal duplication in `pkg/vmexec/executor.go`
- finding 8: remaining core model/helper duplication (notably duplicated `mustMarshalJSON` semantics)

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
