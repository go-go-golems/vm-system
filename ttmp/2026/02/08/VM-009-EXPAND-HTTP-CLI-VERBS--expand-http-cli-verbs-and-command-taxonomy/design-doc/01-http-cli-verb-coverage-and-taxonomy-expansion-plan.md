---
Title: HTTP CLI verb coverage and taxonomy expansion plan
Ticket: VM-009-EXPAND-HTTP-CLI-VERBS
Status: active
Topics:
    - backend
    - docs
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: cmd/vm-system/cmd_http.go
      Note: Current HTTP root grouping baseline
    - Path: cmd/vm-system/cmd_session.go
      Note: Current session lifecycle verb surface needing close/delete taxonomy decision
    - Path: docs/getting-started-from-first-vm-to-contributor-guide.md
      Note: Canonical user-facing command reference requiring follow-up updates
    - Path: pkg/vmtransport/http/server.go
      Note: Authoritative HTTP endpoint inventory for CLI coverage mapping
    - Path: ttmp/2026/02/08/VM-009-EXPAND-HTTP-CLI-VERBS--expand-http-cli-verbs-and-command-taxonomy/tasks.md
      Note: Execution checklist derived from design
ExternalSources: []
Summary: ""
LastUpdated: 2026-02-08T18:15:08.837613684-05:00
WhatFor: Define a complete, implementation-ready plan to expand and normalize HTTP-backed CLI verbs after introducing the root `http` command group.
WhenToUse: Use when implementing or reviewing VM-009 HTTP CLI coverage and command taxonomy changes.
---


# HTTP CLI verb coverage and taxonomy expansion plan

## Executive Summary

`vm-system` now separates daemon-backed client operations under `vm-system http ...`, which cleanly distinguishes HTTP client surfaces from local-only commands (`serve`, `libs`).  

However, command verb coverage and naming are still partially inconsistent:

1. Some HTTP endpoints do not have first-class CLI commands (`health`, `runtime summary`).
2. Session lifecycle naming is ambiguous (`delete` currently means close).
3. Verb taxonomy is not yet explicitly codified for future contributors.
4. CLI integration coverage is still selective, leaving room for command-surface drift.

This ticket defines a complete plan to:

- close endpoint-to-CLI coverage gaps,
- normalize command verb semantics,
- add explicit command taxonomy rules,
- and harden test/docs so future command additions remain consistent.

## Problem Statement

The project has daemon-backed APIs with stable routes, but CLI command discoverability and semantics are not fully aligned with those routes.

### Symptoms

1. Missing operations commands:
   - `GET /api/v1/health` has no direct CLI verb.
   - `GET /api/v1/runtime/summary` has no direct CLI verb.
2. Session lifecycle ambiguity:
   - `http session delete` currently performs close semantics via `POST /sessions/{id}/close`.
   - The verb suggests hard deletion while behavior is “close session”.
3. Taxonomy drift risk:
   - no written rule for how new HTTP routes map to verbs/nouns.
   - contributors can add ad-hoc naming patterns.
4. Coverage gaps:
   - command registration tests exist for selected surfaces, but not comprehensive command-to-route contract coverage.

### Why this matters

- Operator UX and script reliability depend on deterministic command semantics.
- Mismatched verbs increase onboarding and runbook errors.
- Without coverage/guardrails, CLI and API evolve out of sync.

## Proposed Solution

Adopt a strict HTTP CLI taxonomy and complete endpoint coverage for currently exposed route families.

### Command structure policy

- Keep root commands purpose-separated:
  - `serve`: daemon host runtime
  - `http`: daemon-backed API client surface
  - `libs`: local cache utility surface
- Under `http`, group by API noun:
  - `http template ...`
  - `http session ...`
  - `http exec ...`
  - `http ops ...` (new) or equivalent top-level ops verbs under `http`

### Endpoint-to-CLI coverage target

| HTTP Route | Current CLI | Target CLI |
|---|---|---|
| `GET /api/v1/health` | missing | `vm-system http ops health` |
| `GET /api/v1/runtime/summary` | missing | `vm-system http ops runtime-summary` |
| `POST /api/v1/sessions/{id}/close` | `http session delete` (ambiguous) | `http session close` |
| `DELETE /api/v1/sessions/{id}` | not mapped | `http session delete` (optional; only if semantics intentionally retained) |

Notes:

- If deletion semantics are intentionally unsupported operationally, remove CLI `delete` and use `close` only.
- If hard delete remains, it must map to `DELETE /sessions/{id}` explicitly and be documented as destructive.

### Verb taxonomy contract

Use these verb meanings consistently:

- `create`: create durable resource
- `list`: list/query resources
- `get`: fetch one resource
- `add` / `remove`: nested collection membership mutation
- `close`: transition active resource to closed state
- `delete`: destructive removal (must not alias close)
- `repl` / `run-file`: execution actions (imperative verbs)
- `ops *`: operational/diagnostic read APIs (`health`, `runtime-summary`)

### Output and scripting policy

- Keep text output for human-friendly default.
- Add/expand predictable JSON output option for automation-sensitive commands (if currently missing), especially new `ops` commands.
- Ensure scripts and docs use only canonical command forms.

## Design Decisions

1. **Keep `http` as the single daemon-client namespace.**  
Rationale: clear boundary between client calls and local process/cache commands.

2. **No semantic aliases for destructive/transition verbs.**  
Rationale: `close` and `delete` must never be conflated.

3. **Add explicit operations subgroup (`http ops`) for non-resource routes.**  
Rationale: health/runtime endpoints are first-class but not template/session/exec resources.

4. **Treat CLI coverage as route-contract coverage, not optional sugar.**  
Rationale: for this project, scripts and runbooks rely on CLI as stable API shim.

5. **No backwards-compatibility wrappers unless explicitly ticketed.**  
Rationale: consistent with current cleanup direction and avoids long-tail ambiguity.

## Alternatives Considered

1. Keep route coverage gaps and rely on `curl` for ops endpoints.  
Rejected: adds operational friction and inconsistent tooling expectations.

2. Keep `session delete` as alias for close.  
Rejected: semantically misleading and high risk in automation.

3. Flatten all commands at root (`template/session/exec` directly).  
Rejected: reintroduces blending of local and HTTP-backed command families.

4. Add wrappers for legacy command names.  
Rejected: contradicts clean-cut direction and increases maintenance surface.

## Implementation Plan

### Phase 1: Inventory and contract locking

1. Build endpoint-to-command matrix from `pkg/vmtransport/http/server.go` and `cmd/vm-system`.
2. Finalize verb semantics for `session close` vs `session delete`.
3. Define canonical usage examples for each route family.

### Phase 2: CLI implementation

1. Add `http ops` command with:
   - `health`
   - `runtime-summary`
2. Add `http session close`.
3. Decide and implement one of:
   - map `http session delete` to hard delete endpoint only, or
   - remove `http session delete` from CLI if deletion is not intended surface.
4. Update client methods if missing for any added command.

### Phase 3: Tests

1. Add/expand command registration tests for `http`, `http ops`, and session verbs.
2. Add HTTP integration assertions for operations commands where needed.
3. Add CLI-level behavior tests for session close/delete semantics.
4. Ensure script suites pass with canonical command forms only.

### Phase 4: Docs and runbooks

1. Update getting-started and quick references with final command taxonomy.
2. Update script playbooks and onboarding docs.
3. Add final guard search for non-canonical command forms.

### Phase 5: Validation and handoff

1. Run:
   - `GOWORK=off go test ./... -count=1`
   - `GOWORK=off go test ./pkg/vmtransport/http -count=1`
   - `./smoke-test.sh`
   - `./test-e2e.sh`
2. Record evidence in changelog and handoff notes.

## Acceptance Criteria

1. Every active HTTP endpoint family has intentional CLI coverage or a documented exclusion.
2. `close` vs `delete` semantics are unambiguous in both code and help text.
3. `http ops health` and `http ops runtime-summary` are available and tested.
4. Docs/scripts reference canonical verbs only.
5. Full validation matrix is green.

## Open Questions

1. Should `DELETE /sessions/{id}` remain exposed operationally, or should close-only be enforced at CLI level?
2. Should machine-readable output mode be standardized now (`--output json`) across all HTTP commands or deferred?
3. Should future route families always require same-ticket CLI coverage before merge?

## References

- `cmd/vm-system/main.go`
- `cmd/vm-system/cmd_http.go`
- `cmd/vm-system/cmd_template.go`
- `cmd/vm-system/cmd_session.go`
- `cmd/vm-system/cmd_exec.go`
- `pkg/vmtransport/http/server.go`
- `docs/getting-started-from-first-vm-to-contributor-guide.md`
