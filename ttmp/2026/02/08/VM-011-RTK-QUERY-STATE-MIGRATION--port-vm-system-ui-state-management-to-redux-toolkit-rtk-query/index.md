---
Title: Port vm-system-ui state management to Redux Toolkit + RTK Query
Ticket: VM-011-RTK-QUERY-STATE-MIGRATION
Status: complete
Topics:
    - frontend
    - state-management
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: vm-system/vm-system-ui/client/src/components/AppShell.tsx
      Note: Hand-rolled AppState context — to be replaced by Redux store
    - Path: vm-system/vm-system-ui/client/src/components/CreateSessionDialog.tsx
      Note: Session creation + template prop drilling
    - Path: vm-system/vm-system-ui/client/src/lib/api.ts
      Note: RTK Query createApi with 13 endpoints
    - Path: vm-system/vm-system-ui/client/src/lib/normalize.ts
      Note: Pure normalization functions
    - Path: vm-system/vm-system-ui/client/src/lib/store.ts
      Note: Redux store configuration
    - Path: vm-system/vm-system-ui/client/src/lib/types.ts
      Note: Domain types
    - Path: vm-system/vm-system-ui/client/src/lib/uiSlice.ts
      Note: Client-only Redux slice
    - Path: vm-system/vm-system-ui/client/src/lib/vmService.ts
      Note: Monolithic VMService class — the entire data layer to be replaced
    - Path: vm-system/vm-system-ui/client/src/pages/SessionDetail.tsx
      Note: REPL execution
    - Path: vm-system/vm-system-ui/client/src/pages/TemplateDetail.tsx
      Note: Module/library mutation + manual refresh
    - Path: vm-system/vm-system-ui/client/src/pages/Templates.tsx
      Note: Raw fetch bypass for template creation
ExternalSources: []
Summary: ""
LastUpdated: 2026-02-09T00:59:36.350315857-05:00
WhatFor: ""
WhenToUse: ""
---




# Port vm-system-ui state management to Redux Toolkit + RTK Query

## Overview

<!-- Provide a brief overview of the ticket, its goals, and current status -->

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

## Status

Current status: **active**

## Topics

- frontend
- state-management

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
