---
Title: Implementation Guide
Ticket: VM-013B-BACKEND-CLIENT-LOADER-DAEMON
Status: active
Topics:
    - backend
    - architecture
    - testing
DocType: design
Intent: long-term
Owners: []
RelatedFiles:
    - Path: pkg/libloader/loader_test.go
      Note: Defines cache/download/checksum contract coverage
    - Path: pkg/vmclient/rest_client_test.go
      Note: Defines vmclient support-layer contract coverage
    - Path: pkg/vmdaemon/config_test.go
      Note: Defines daemon config/new-app contract coverage
ExternalSources: []
Summary: ""
LastUpdated: 2026-02-09T00:52:31-05:00
WhatFor: Execute VM-013B support-package coverage for vmclient, libloader, and vmdaemon.
WhenToUse: Use when implementing or reviewing VM-013B backend support coverage.
---


# VM-013B Implementation Guide

## Scope

This ticket adds support-layer backend coverage where failures are easy to miss in endpoint-level integration tests:
- `pkg/vmclient` transport/decoding/error-envelope behavior,
- `pkg/libloader` cache/discovery/download/checksum behavior,
- `pkg/vmdaemon` default configuration and app construction contract.

## Implemented test files

- `pkg/vmclient/rest_client_test.go`
- `pkg/libloader/loader_test.go`
- `pkg/vmdaemon/config_test.go`

## Test plan delivered

1. `vmclient`:
- API error envelope mapping into `*APIError`.
- malformed/non-JSON error responses.
- malformed success payload decode failures.
- `withQuery` empty-value skipping + query encoding.

2. `libloader`:
- existing cache discovery for known builtin library filenames.
- download path determinism (`<id>-<version>.js`).
- cache-hit short circuit (second call succeeds without network).
- loaded code content and checksum assertions.
- non-200 HTTP failure behavior.

3. `vmdaemon`:
- default config duration/address contracts.
- `New(...)` server wiring of timeout fields/handler/core.
- explicit error path when DB parent path is missing.

## Validation command

```bash
GOWORK=off go test ./pkg/vmclient ./pkg/libloader ./pkg/vmdaemon -count=1
```

## Notes

- `vmclient` tests use local `httptest.Server` responses so failures are deterministic.
- `libloader` tests avoid external network and pin assertions to local fixture servers/files.
