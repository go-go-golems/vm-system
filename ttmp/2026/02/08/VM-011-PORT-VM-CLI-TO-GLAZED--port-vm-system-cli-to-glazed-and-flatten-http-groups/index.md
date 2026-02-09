---
Title: Port vm-system CLI to glazed and flatten HTTP groups
Ticket: VM-011-PORT-VM-CLI-TO-GLAZED
Status: complete
Topics:
    - backend
    - architecture
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ttmp/2026/02/08/VM-011-PORT-VM-CLI-TO-GLAZED--port-vm-system-cli-to-glazed-and-flatten-http-groups/design-doc/01-glazed-migration-plan-for-vm-system-cli-with-root-level-command-flattening.md
      Note: Primary migration analysis and phased implementation plan
    - Path: ttmp/2026/02/08/VM-011-PORT-VM-CLI-TO-GLAZED--port-vm-system-cli-to-glazed-and-flatten-http-groups/reference/01-diary.md
      Note: Detailed implementation diary with commands, findings, and follow-ups
    - Path: ttmp/2026/02/08/VM-011-PORT-VM-CLI-TO-GLAZED--port-vm-system-cli-to-glazed-and-flatten-http-groups/tasks.md
      Note: Execution checklist for migration and validation
    - Path: cmd/vm-system/main.go
      Note: Current root CLI entrypoint and command registration baseline
ExternalSources: []
Summary: Tracks the plan to migrate vm-system CLI commands to glazed, remove the http parent group, and wire built-in help docs.
LastUpdated: 2026-02-09T00:59:36.32968964-05:00
WhatFor: Coordinate VM-011 implementation work to port command wiring to glazed and flatten command groups to root.
WhenToUse: Use when implementing, reviewing, or validating VM-011 command taxonomy and help-system changes.
---



# Port vm-system CLI to glazed and flatten HTTP groups

## Overview

This ticket captures the complete migration plan to move `vm-system` CLI commands from direct Cobra handlers to glazed command definitions, while removing the `http` group and exposing daemon-backed command families directly at root.

Core outcomes:

- root taxonomy: `serve`, `template`, `session`, `exec`, `ops`, `libs`
- command implementations ported to glazed patterns
- built-in help docs wired via glazed help system
- test and docs updates to prevent taxonomy drift

## Key Links

- Design doc: `design-doc/01-glazed-migration-plan-for-vm-system-cli-with-root-level-command-flattening.md`
- Diary: `reference/01-diary.md`
- Tasks: `tasks.md`
- Changelog: `changelog.md`

## Status

Current status: **complete**

## Topics

- backend
- architecture

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
