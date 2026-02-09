---
Title: Implementation Guide
Ticket: VM-013A-BACKEND-RUNTIME-SESSION-STORE
Status: active
Topics:
    - backend
    - architecture
    - testing
DocType: design
Intent: long-term
Owners: []
RelatedFiles:
    - Path: pkg/vmtransport/http/server_libraries_integration_test.go
      Note: Primary lodash and cache semantics integration coverage
    - Path: pkg/vmtransport/http/server_native_modules_integration_test.go
      Note: Primary native module configured-vs-unconfigured integration coverage
ExternalSources: []
Summary: ""
LastUpdated: 2026-02-09T00:49:58-05:00
WhatFor: Execute VM-013A backend runtime/session/store high-risk coverage with deterministic integration tests.
WhenToUse: Use when implementing or reviewing VM-013A runtime semantics coverage.
---


# VM-013A Implementation Guide

## Scope

This ticket implements backend high-risk runtime coverage from VM-013 with emphasis on behavior users depend on today:
- JavaScript built-in availability (`JSON`),
- template-controlled library behavior (`lodash`),
- template-controlled native module behavior (`database`, `exec`),
- startup/session behavior when configured libraries are missing from cache.

## Implemented test files

- `pkg/vmtransport/http/server_libraries_integration_test.go`
- `pkg/vmtransport/http/server_native_modules_integration_test.go`

## Test plan delivered

1. Lodash configured and cache present:
- Create `.vm-cache/libraries/lodash-4.17.21.js` fixture.
- Configure template library `lodash-4.17.21`.
- Assert `_.chunk(...).length` succeeds.

2. Lodash not configured while cache exists:
- Keep same cache fixture.
- Use template without library.
- Assert `_` call fails with undefined/reference error.

3. Lodash configured but cache missing:
- Configure template library.
- Do not create cache file.
- Assert session creation fails with `INTERNAL` and message mentioning library load failure.

4. Native modules configured vs unconfigured:
- Without configured modules, `require("database")` and `require("exec")` fail.
- With configured modules, `exec.run(...)` and `database.configure/exec/query` succeed.

## Validation command

```bash
GOWORK=off go test ./pkg/vmtransport/http -count=1
```

## Notes

- Library cache loading is cwd-relative (`.vm-cache/libraries`) in current architecture, so tests explicitly control cwd.
- Session/runtime behavior is asserted via execution status and payload fields rather than snapshots.
