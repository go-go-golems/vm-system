---
Title: Stop template-configuring JS built-ins; use go-go-goja registrable native modules
Ticket: VM-017-NATIVE-MODULE-REGISTRY
Status: complete
Topics:
    - backend
    - frontend
    - infrastructure
DocType: index
Intent: long-term
Owners: []
RelatedFiles:
    - Path: vm-system/vm-system/ttmp/2026/02/09/VM-017-NATIVE-MODULE-REGISTRY--stop-template-configuring-js-built-ins-use-go-go-goja-registrable-native-modules/design/01-analysis-and-implementation-guide.md
      Note: |-
        Primary implementation analysis and sequencing
        Primary VM-017 implementation design
    - Path: vm-system/vm-system/ttmp/2026/02/09/VM-017-NATIVE-MODULE-REGISTRY--stop-template-configuring-js-built-ins-use-go-go-goja-registrable-native-modules/reference/01-diary.md
      Note: |-
        Task-by-task implementation diary
        Execution diary for task-by-task implementation
ExternalSources: []
Summary: |
    Enforce module policy consistency by removing template configurability for JavaScript built-ins and adopting go-go-goja registrable native modules for template module configuration.
LastUpdated: 2026-02-09T00:42:24.294118397-05:00
WhatFor: Align runtime behavior, API contract, and UI semantics for template module configuration.
WhenToUse: Use as entry point for VM-017 implementation and review.
---



# Stop template-configuring JS built-ins; use go-go-goja registrable native modules

## Overview

VM-017 fixes module-policy drift by making built-ins always-on runtime features and restricting template-configurable modules to go-go-goja native module registrations.

## Key Links

- Analysis and implementation guide: [design/01-analysis-and-implementation-guide.md](./design/01-analysis-and-implementation-guide.md)
- Diary: [reference/01-diary.md](./reference/01-diary.md)
- Tasks: [tasks.md](./tasks.md)
- Changelog: [changelog.md](./changelog.md)

## Status

Current status: **active**

## Topics

- backend
- frontend
- infrastructure
