---
Title: Scope Plugin Actions and State for WebVM
Ticket: WEBVM-001-SCOPE-PLUGIN-ACTIONS
Status: active
Topics:
    - architecture
    - plugin
    - state-management
DocType: index
Intent: long-term
Owners: []
RelatedFiles: []
ExternalSources: []
Summary: "Landing page for WEBVM-001 with links to the simplified v1 scoping model and QuickJS worker replacement design documents."
LastUpdated: 2026-02-08T13:49:00-05:00
WhatFor: "Track the plugin identity/action/state scoping architecture investigation and implementation strategy."
WhenToUse: "Use as the landing page for WEBVM-001 design docs, decisions, and deliverables."
---

# Scope Plugin Actions and State for WebVM

## Overview

This ticket investigates plugin identity, action scoping, and state scoping in the `plugin-playground` system. The current decision is a simplified v1 API (`selectPluginState`, `selectGlobalState`, plugin/global dispatch actions with `dispatchId`) plus dedicated QuickJS worker replacement designs for removing mock runtime paths.

## Key Links

- Design doc 01 (state/action scoping): `design-doc/01-plugin-action-and-state-scoping-architecture-review.md`
- Design doc 02 (QuickJS isolation): `design-doc/02-quickjs-isolation-architecture-and-mock-runtime-removal-plan.md`
- Design doc 03 (QuickJS worker replacement deep design): `design-doc/03-quickjs-worker-replacement-detailed-analysis-and-design.md`
- reMarkable upload (doc 01): `/ai/2026/02/08/WEBVM-001-SCOPE-PLUGIN-ACTIONS/01-plugin-action-and-state-scoping-architecture-review`
- reMarkable upload (bundle docs 01+02): `/ai/2026/02/08/WEBVM-001-SCOPE-PLUGIN-ACTIONS/WEBVM-001-scoping-and-quickjs-review`
- Changelog: `changelog.md`
- Tasks: `tasks.md`

## Status

Current status: **active**

Latest outcome:

- Completed a detailed architecture assessment and migration plan.
- Updated the assessment with the simplified v1 selector/action model.
- Added a dedicated QuickJS isolation and mock-runtime removal plan.
- Added a new detailed QuickJS worker replacement architecture and migration design.
- Uploaded the combined docs bundle to reMarkable.

## Topics

- architecture
- plugin
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
