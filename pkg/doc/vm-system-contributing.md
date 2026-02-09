---
Title: "Contributing to vm-system"
Slug: contributing
Short: "Contribution workflow, testing strategy, code review checklist, and safe first contribution ideas."
Topics:
- vm-system
- contributing
- testing
- development
IsTopLevel: true
IsTemplate: false
ShowPerDefault: true
SectionType: GeneralTopic
---

This guide explains how to contribute changes to vm-system safely: which layer
to change, how to test, what reviewers look for, and a set of bounded first
contributions.

## Choose the right layer

Before writing code, identify which layer owns the behavior you want to change:

| Change type | Layer | Files |
|------------|-------|-------|
| HTTP shape, status codes, validation | `pkg/vmtransport/http` | `server*.go` |
| Domain orchestration, policy | `pkg/vmcontrol` | `*_service.go`, `core.go` |
| Runtime semantics, goja execution | `pkg/vmsession`, `pkg/vmexec` | `session.go`, `executor.go` |
| Persistence, schema | `pkg/vmstore` | `vmstore*.go` |
| CLI UX, output formatting | `cmd/vm-system` | `cmd_*.go` |
| REST client | `pkg/vmclient` | `*_client.go` |

Do not mix concerns across layers in a single change unless the work is
intentionally cross-cutting.

## Testing strategy

### Integration tests (highest value)

Located in `pkg/vmtransport/http`, these tests spin up a real in-memory stack
(SQLite + vmcontrol + httptest server) with no mocks:

```bash
# Run all integration tests
GOWORK=off go test ./pkg/vmtransport/http -count=1

# Run a specific test
GOWORK=off go test ./pkg/vmtransport/http -run TestSessionLifecycleEndpoints -count=1
```

Test suites:

| File | Coverage area |
|------|--------------|
| `server_integration_test.go` | Cross-call continuity in one session |
| `server_templates_integration_test.go` | Template CRUD + nested resources |
| `server_sessions_integration_test.go` | Session lifecycle + runtime summary |
| `server_executions_integration_test.go` | REPL, run-file, events, after_seq |
| `server_error_contracts_integration_test.go` | Validation, not-found, conflict |
| `server_safety_integration_test.go` | Path traversal, output limits |
| `server_libraries_integration_test.go` | Library loading behavior |
| `server_native_modules_integration_test.go` | Module catalog behavior |
| `server_execution_contracts_integration_test.go` | Execution edge cases |

### Unit tests

```bash
GOWORK=off go test ./... -count=1
```

Covers `vmpath`, `vmmodels`, `vmdaemon/config`, `libloader`, `vmclient`, and
`vmcontrol/template_service`.

### Shell integration scripts

```bash
bash ./smoke-test.sh          # Fast confidence check (~10s)
bash ./test-e2e.sh            # Full CLI loop
bash ./test-library-matrix.sh # Library/module capability semantics
bash ./test-all.sh            # Everything above
```

All scripts use temporary databases, worktrees, and dynamic ports for
isolation.

### Recommended validation workflow

For normal changes:

```bash
GOWORK=off go test ./pkg/vmtransport/http -count=1
GOWORK=off go test ./... -count=1
bash ./smoke-test.sh
```

Before opening a PR:

```bash
bash ./test-all.sh
```

### What is well-covered today

- Template/session/execution endpoint families
- Major error contracts (400/404/409/422)
- Path traversal and output limit enforcement
- Runtime summary transitions on close/delete
- Library and module configuration semantics

### What needs more coverage

- Daemon restart recovery
- Load/performance/concurrency testing
- Deep library path/version consistency
- Process crash durability

## Standard change patterns

### Adding a new API endpoint

1. Add handler route in `server.go`
2. Add request DTO and validation in the handler
3. Add core service method in `pkg/vmcontrol` if needed
4. Add `vmclient` wrapper method
5. Add CLI command if user-facing
6. Add integration test asserting status code + error envelope

### Adding a new domain error

1. Add sentinel error in `pkg/vmmodels/models.go`
2. Add mapping in `writeCoreError` (`server_errors.go`)
3. Add integration test in the error contracts suite

### Changing session lifecycle behavior

1. Update `SessionManager` transitions in `pkg/vmsession`
2. Ensure store row updates stay consistent
3. Add runtime summary assertions in integration tests
4. Verify close/delete semantics remain deterministic

## Code review checklist

Before requesting review:

- [ ] Feature has deterministic tests
- [ ] Status code + error code + message are intentional
- [ ] No hidden coupling between transport and domain layers
- [ ] CLI output remains readable and stable
- [ ] Docs/changelog updated for user-visible changes

## Git workflow

- Inspect changes: `git diff` and `git diff --staged`
- Stage only intentional files
- Use clear scope prefixes in commit messages:

```
feat(api): add execution-not-found 404 contract
test(integration): add execution get not-found test
docs(help): add architecture glazed help page
```

## High-value first contributions

If you are looking for a bounded, meaningful first PR:

1. **Add metadata endpoint family** — New API surface for querying/updating
   template settings directly, with full integration coverage.

2. **Add structured logging** — Instrument key lifecycle events (session
   create/close, execution start/end) with zerolog fields for observability.

3. **Harden library cache consistency** — Unify filename expectations between
   the library downloader and runtime loader to prevent load failures.

4. **Add benchmark tests** — Performance characterization for execution
   throughput and event storage under load.

5. **Daemon restart policy** — Implement deterministic close-on-restart:
   mark active sessions as closed during daemon boot with integration tests.

Each is practical, bounded, and directly improves reliability or operability.

## Debugging playbook

### API-level

```bash
# Reproduce with curl and inspect full response
curl -i -sS http://127.0.0.1:3210/api/v1/templates/does-not-exist
```

Compare status code and error envelope with `writeCoreError` mapping.

### Runtime/session

```bash
vm-system session list --status ready
vm-system ops runtime-summary
```

Verify transitions match expected state machine.

### Execution/events

```bash
vm-system exec repl <session-id> 'minimal_test()'
vm-system exec events <execution-id> --after-seq 0
```

Inspect event types and ordering.

### Persistence

Use throwaway database paths per debugging run. Do not reuse a stale DB when
investigating schema or flow changes.

## Troubleshooting

| Problem | Cause | Solution |
|---------|-------|----------|
| Integration tests fail after handler change | Missing error mapping or changed response shape | Compare test assertions with `writeCoreError` and handler output |
| Smoke test passes but integration fails | Layer mismatch | Isolate whether CLI/transport/core layers diverge |
| `go test` warns about build tags | Missing `GOWORK=off` | Always prefix test commands with `GOWORK=off` |

## See Also

- `vm-system help architecture`
- `vm-system help api-reference`
- `vm-system help testing-guide`
