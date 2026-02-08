---
Title: Unify template/modules/libraries language around templates
Ticket: VM-008-UNIFY-TEMPLATE-LANGUAGE
Status: active
Topics:
    - backend
    - docs
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: cmd/vm-system/cmd_template.go
      Note: Primary CLI surface to become authoritative for template/module/library operations
    - Path: cmd/vm-system/cmd_modules.go
      Note: Legacy mixed-language command surface targeted for removal
    - Path: cmd/vm-system/main.go
      Note: Root command wiring currently exposes legacy modules command
    - Path: pkg/vmtransport/http/server.go
      Note: Template API routes requiring extension for module/library operations
    - Path: pkg/vmcontrol/template_service.go
      Note: Domain service layer for template operation unification
    - Path: pkg/vmclient/templates_client.go
      Note: Client APIs that need template-language-aligned module/library methods
    - Path: docs/getting-started-from-first-vm-to-contributor-guide.md
      Note: Contributor-facing language and command examples to normalize
    - Path: ttmp/2026/02/08/VM-008-UNIFY-TEMPLATE-LANGUAGE--unify-template-modules-libraries-language-around-templates/design-doc/01-template-language-unification-review-and-implementation-plan.md
      Note: Detailed review and delegated implementation plan
ExternalSources: []
Summary: Ticket for unifying all template/module/library user-facing and internal language around templates, removing legacy VM-oriented command vocabulary without backwards compatibility.
LastUpdated: 2026-02-08T12:12:26-05:00
WhatFor: Provide an implementation-ready plan to remove mixed template/vm language and direct-DB legacy command paths across CLI/API/docs.
WhenToUse: Use when implementing language and command-surface cleanup centered on templates.
---

# Unify template/modules/libraries language around templates

## Overview

This ticket scopes a no-backward-compatibility cleanup to make `template` the single authoritative vocabulary and control surface for modules/libraries.

It includes CLI/API/service/store naming and behavior cleanup plus documentation updates (especially the getting started guide) so new contributors see one coherent mental model.

## Key Links

- **Related Files**: See frontmatter RelatedFiles field
- **External Sources**: See frontmatter ExternalSources field

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
