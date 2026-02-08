---
Title: Daemon/API test coverage matrix and expansion plan
Ticket: VM-004-EXPAND-E2E-COVERAGE
Status: active
Topics:
    - backend
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: vm-system/pkg/vmtransport/http/server.go
      Note: Route surface and error contract mapping to test
    - Path: vm-system/pkg/vmtransport/http/server_integration_test.go
      Note: Existing single continuity test baseline
    - Path: vm-system/pkg/vmtransport/http/server_templates_integration_test.go
      Note: Template endpoint CRUD and nested-resource coverage
    - Path: vm-system/pkg/vmtransport/http/server_sessions_integration_test.go
      Note: Session lifecycle plus runtime summary transition coverage
    - Path: vm-system/pkg/vmtransport/http/server_executions_integration_test.go
      Note: Execution endpoint lifecycle coverage
    - Path: vm-system/pkg/vmtransport/http/server_error_contracts_integration_test.go
      Note: Validation/not-found/conflict/unprocessable error contract coverage
    - Path: vm-system/pkg/vmtransport/http/server_safety_integration_test.go
      Note: Path traversal and limit-enforcement safety coverage
    - Path: vm-system/smoke-test.sh
      Note: Current daemon-first smoke validation script
    - Path: vm-system/test-e2e.sh
      Note: Current daemon-first e2e script behavior and gaps
ExternalSources: []
Summary: Coverage matrix, implemented test expansion, and residual risk report for daemon/API integration and e2e behavior.
LastUpdated: 2026-02-08T10:43:00-05:00
WhatFor: Establish comprehensive automated coverage for vm-system daemon/API behavior and close high-risk regression gaps.
WhenToUse: Use when implementing or reviewing test coverage for template/session/execution APIs, runtime safety checks, and smoke/e2e workflows.
---

# Daemon/API test coverage matrix and expansion plan

## Executive Summary

Current daemon-first tests validate the happy-path runtime flow and one key continuity property, but they do not provide broad endpoint and error-contract coverage. This ticket expands automated integration coverage across template/session/execution APIs, safety paths, and script robustness.

The plan prioritizes high-risk regressions first: endpoint correctness, status/error codes, path traversal protection, and runtime summary/session-close behavior.

## Problem Statement

The current test suite has these structural gaps:

1. Only one Go integration test exists (`TestSessionContinuityAcrossAPIRequests`), so most endpoints and error paths are not regression-protected.
2. Smoke/e2e scripts focus on happy paths and share fixed paths/ports, making concurrent runs fragile.
3. Error contract behavior (`400`, `404`, `409`, `422`) is largely untested despite being part of the API design.
4. Runtime safety behavior (path traversal, limits handling) has minimal automated assertion depth.

These gaps increase risk of unnoticed regressions in routing, status codes, and runtime policy behavior.

## Proposed Solution

Expand test coverage in three layers:

1. **Go integration tests (primary):**
   - Add table-driven tests for template/session/execution route families.
   - Add negative-path tests asserting status code + error code contracts.
   - Add safety-focused tests for traversal and limits errors.
   - Extend runtime summary assertions before/after session close.

2. **Shell workflow validation (secondary):**
   - Keep smoke/e2e scripts as operator-facing checks.
   - Make scripts parallel-safe via isolated temp directories and dynamic ports.

3. **Coverage documentation (governance):**
   - Maintain a route-level matrix showing what is tested and what remains intentionally out of scope.

### Baseline Coverage Matrix (Before This Ticket)

| Surface | Current coverage | Status |
|---|---|---|
| `GET /api/v1/health` | smoke + e2e scripts | Partial |
| `GET /api/v1/runtime/summary` | smoke + e2e + continuity test | Partial |
| Template endpoints (`/templates*`) | smoke/e2e exercise subset (create/add startup/add capability) | Partial |
| Session endpoints (`/sessions*`) | smoke/e2e create/list/get; continuity test create | Partial |
| Execution endpoints (`/executions*`) | smoke/e2e repl/run-file/list/events; continuity repl+events | Partial |
| Error contract mapping | minimal direct assertion | Weak |
| Safety path traversal/limits | minimal assertion | Weak |
| Script parallel safety | fixed path/port, race-prone when parallelized | Weak |

### Coverage Matrix (After Tasks 3-9)

| Surface | Added coverage artifacts | Status |
|---|---|---|
| Template endpoints (`/templates*`) | `TestTemplateEndpointsCRUDAndNestedResources` | Strong |
| Session endpoints (`/sessions*`) | `TestSessionLifecycleEndpoints` with status filters + close/delete | Strong |
| Execution endpoints (`/executions*`) | `TestExecutionEndpointsLifecycle` with list/get/events/after_seq | Strong |
| Error contracts (`400/404/409/422`) | `TestAPIErrorContractsValidationNotFoundConflictAndUnprocessable` | Strong |
| Safety (path traversal + limits) | `TestSafetyPathTraversalAndOutputLimitEnforcement` | Strong |
| Runtime summary transitions | session lifecycle test asserts active session count 2 -> 1 -> 0 | Strong |
| Script parallel safety | `smoke-test.sh` + `test-e2e.sh` use temp dirs, temp db files, dynamic ports | Strong |

### Residual Risk Gaps

1. Metadata endpoints (`/api/v1/metadata/*`) are not implemented yet, therefore not covered.
2. Session recovery/restart semantics across daemon restart are not covered.
3. Library/module command families (`modules`, `libs`) still have limited automated assertions in daemon-first context.
4. Performance/load characteristics (large concurrent session counts) are not covered.
5. Authentication/authorization behavior is out of scope for this phase.

## Design Decisions

1. Use route-family table-driven tests in Go as primary artifact.
Rationale: deterministic, fast, easy to assert status codes and payloads.
2. Retain shell scripts for integration/operator validation, not fine-grained assertions.
Rationale: scripts are valuable for workflow checks but brittle for matrix-depth API contracts.
3. Assert both HTTP status and API `error.code` for negative paths.
Rationale: status-only checks miss contract drift that breaks clients.
4. Keep coverage matrix in ticket design-doc.
Rationale: explicit tested-vs-untested mapping supports review and future planning.

## Alternatives Considered

1. Rely on smoke/e2e scripts only.
Rejected: too coarse; poor error-path and contract verification.
2. Add unit tests only at service layer (`vmcontrol`) and skip API tests.
Rejected: misses routing/DTO/error-envelope regressions.
3. Adopt full external black-box test framework first.
Rejected for now: higher setup overhead than needed for current API size.

## Implementation Plan

1. Define and publish route/coverage baseline matrix.
2. Add template endpoint integration tests.
3. Add session lifecycle and filter integration tests.
4. Add execution endpoint integration tests.
5. Add negative/error contract tests.
6. Add safety-specific traversal/limits tests.
7. Extend runtime summary close-state assertions.
8. Make smoke/e2e scripts parallel-safe.
9. Refresh matrix with completed coverage + residual gaps.

## Open Questions

1. Should limit-exceeded behavior mutate persisted execution status in addition to returning API error?
2. Should script validation become fully temp-dir/port auto-isolated by default in all repo scripts?
3. Do we want dedicated contract snapshot tests for error envelope schema?

## References

- `vm-system/pkg/vmtransport/http/server.go`
- `vm-system/pkg/vmtransport/http/server_integration_test.go`
- `vm-system/smoke-test.sh`
- `vm-system/test-e2e.sh`
- `vm-system/ttmp/2026/02/07/VM-001-ANALYZE-VM--analyze-vm-system-architecture-behavior-and-quality/design-doc/02-daemonized-vm-system-architecture-backend-runtime-host-rest-api-and-cli.md`
