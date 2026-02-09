---
Title: Panic Boundary and Logging Alignment Plan
Ticket: VM-015-PANIC-LOGGING-ALIGNMENT
Status: active
Topics:
    - backend
    - architecture
DocType: design
Intent: long-term
Owners: []
RelatedFiles:
    - /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/vmmodels/ids.go
    - /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/vmsession/session.go
    - /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/libloader/loader.go
    - /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/cmd/vm-system/glazed_helpers.go
ExternalSources: []
Summary: >
    P2-2 planning ticket to define where panics are acceptable and align runtime logging
    on a consistent structured strategy.
LastUpdated: 2026-02-09T10:20:00-05:00
WhatFor: Reduce crash risk and improve observability consistency.
WhenToUse: Use when implementing panic/logging policy changes.
---

# Panic Boundary and Logging Alignment Plan

## Problem Statement

Current codebase mixes:
- panic-based helper methods (`Must*` style) in shared model code
- direct `fmt.Print*` logging inside runtime/session/loader paths
- structured logging initialization at CLI root

This creates inconsistent error handling and observability behavior.

## Target Policy

1. Panics allowed only in:
- tests
- clearly marked programmer-invariant helpers not reachable from request/runtime paths

2. Runtime/service/daemon paths should:
- return typed errors
- log through one structured logger strategy

## Candidate Changes

1. `pkg/vmmodels/ids.go`
- evaluate `Must*` usage sites
- keep if test-only; otherwise replace with parse+error return path

2. `pkg/vmsession/session.go` and `pkg/libloader/loader.go`
- replace `fmt.Print*` with injected logger interface or centralized logging helper

3. CLI helpers (`cmd/vm-system/glazed_helpers.go`)
- preserve panic use only for startup-time command wiring invariants
- ensure those code paths are not reachable during request handling

## Migration Steps

1. Inventory panic call sites and categorize by risk.
2. Add logger interface in runtime components.
3. Replace direct prints in runtime/session/loader.
4. Add tests around behavior where panic paths were converted.

## Acceptance Criteria

- No panics in request/runtime/session execution path.
- Runtime logs follow one style and include enough context for diagnosis.
- Existing CLI behavior remains functionally unchanged.
