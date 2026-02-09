---
Title: "Contributing to vm-system"
Slug: contributing
Short: "How to contribute: layer ownership, testing strategy, test inventory, review checklist, and debugging."
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

## Choose the right layer

| Change type | Layer | Key files |
|------------|-------|-----------|
| HTTP shape, status codes, validation | `pkg/vmtransport/http` | `server*.go` |
| Domain orchestration, policy | `pkg/vmcontrol` | `*_service.go`, `core.go` |
| Runtime semantics, goja execution | `pkg/vmsession`, `pkg/vmexec` | `session.go`, `executor.go` |
| Persistence, schema | `pkg/vmstore` | `vmstore*.go` |
| CLI UX, output formatting | `cmd/vm-system` | `cmd_*.go` |
| REST client | `pkg/vmclient` | `*_client.go` |

Do not mix concerns across layers unless the change is intentionally
cross-cutting.

## Running tests

```bash
# All Go tests
GOWORK=off go test ./... -count=1

# Integration tests only (highest value)
GOWORK=off go test ./pkg/vmtransport/http -count=1

# Specific test
GOWORK=off go test ./pkg/vmtransport/http -run TestSessionLifecycleEndpoints -count=1 -v

# Shell integration
bash ./smoke-test.sh          # ~10 seconds
bash ./test-e2e.sh            # full CLI loop
bash ./test-library-matrix.sh # library/module semantics
bash ./test-all.sh            # everything
```

Always use `GOWORK=off`. All shell scripts use temporary databases, worktrees,
and dynamic ports for isolation.

### Recommended workflow

Normal changes: integration tests → all Go tests → smoke test.
Before PR: `bash ./test-all.sh`.

## Test inventory

### Integration tests (pkg/vmtransport/http)

These exercise the full stack — store, core, HTTP handler — through a real
`httptest.Server` with in-memory SQLite. No mocks.

| File | Coverage |
|------|----------|
| `server_integration_test.go` | Cross-execution state continuity |
| `server_templates_integration_test.go` | Template CRUD + nested resources |
| `server_sessions_integration_test.go` | Session lifecycle + runtime summary |
| `server_executions_integration_test.go` | REPL, run-file, events, `after_seq` |
| `server_error_contracts_integration_test.go` | 400/404/409/422 contracts |
| `server_safety_integration_test.go` | Path traversal, output limits |
| `server_libraries_integration_test.go` | Library loading |
| `server_native_modules_integration_test.go` | Module allowlist (`MODULE_NOT_ALLOWED`) |
| `server_execution_contracts_integration_test.go` | Execution edge cases |

### Unit tests

`vmpath`, `vmmodels` (IDs, JSON helpers), `vmdaemon/config`, `libloader`,
`vmclient`, `vmcontrol/template_service`, `vmexec` (executor, persistence
failures).

### Shell scripts

- **smoke-test.sh** — Build, daemon health, template/session/exec happy path (~10s)
- **test-e2e.sh** — Full CLI loop: build → daemon → workspace → template → session → exec → close
- **test-library-matrix.sh** — Built-in vs configurable module semantics, lodash loading
- **test-all.sh** — Runs everything in sequence

### Coverage gaps

| Gap | Impact |
|-----|--------|
| Daemon restart recovery | Sessions in DB but no runtime — undefined |
| Concurrent execution load | Only `SESSION_BUSY` tested; no throughput |
| Library cache path consistency | Downloader vs loader naming mismatch |
| Process crash durability | No abrupt-kill + DB integrity test |

### Adding a new integration test

```go
func TestMyNewBehavior(t *testing.T) {
    store := vmstore.NewMemoryStore(t)
    core := vmcontrol.NewCore(store)
    server := httptest.NewServer(vmhttp.NewHandler(core))
    defer server.Close()

    template := createTestTemplate(t, server.URL)
    session := createTestSession(t, server.URL, template.ID)

    resp, err := http.Post(server.URL+"/api/v1/executions/repl", ...)
    assert.Equal(t, 201, resp.StatusCode)
}
```

Assert on both HTTP status and `error.code` in the JSON envelope.

## Standard change patterns

### New API endpoint

1. Add route in `server.go`
2. Add request DTO + validation in handler
3. Add core service method if needed
4. Add `vmclient` wrapper
5. Add CLI command if user-facing
6. Add integration test (status code + error envelope)

### New domain error

1. Add sentinel in `pkg/vmmodels/models.go`
2. Add mapping in `writeCoreError` (`server_errors.go`)
3. Add integration test in error contracts suite

### Session lifecycle change

1. Update `SessionManager` transitions in `pkg/vmsession`
2. Keep store row updates deterministic
3. Add runtime summary assertions
4. Verify close/delete remain deterministic

## Code review checklist

- [ ] Feature has deterministic tests
- [ ] Status code + error code + message are intentional
- [ ] No hidden coupling between transport and domain layers
- [ ] CLI output remains readable and stable
- [ ] Docs/changelog updated for user-visible changes

## Git workflow

```
feat(api): add execution-not-found 404 contract
test(integration): add execution get not-found test
docs(help): update api-reference for new endpoint
```

Stage intentionally; inspect with `git diff --staged`.

## High-value first contributions

1. **Add metadata endpoint family** — query/update template settings directly
2. **Structured logging** — zerolog instrumentation on session/execution lifecycle
3. **Library cache consistency** — unify downloader/loader filename expectations
4. **Benchmark tests** — execution throughput and event storage under load
5. **Daemon restart policy** — close-on-restart with integration tests

## Debugging playbook

### API-level

```bash
curl -i -sS http://127.0.0.1:3210/api/v1/templates/does-not-exist
```

Compare status and envelope with `writeCoreError` mapping.

### Runtime/session

```bash
vm-system session list --status ready
vm-system ops runtime-summary
```

### Execution/events

```bash
vm-system exec repl <session-id> 'minimal_test()'
vm-system exec events <execution-id> --after-seq 0
```

### Persistence

Use throwaway DB paths. Do not reuse stale DBs when debugging schema changes.

### Test failures

- Integration: read assertion message → check `writeCoreError` → run with `-v`
- Shell: find last successful output → check daemon started → run failing command manually
- Flaky: check for port conflicts, leftover temp files, stale daemon processes

## Troubleshooting

| Problem | Cause | Solution |
|---------|-------|----------|
| `go test` import errors | Missing `GOWORK=off` | Always prefix with `GOWORK=off` |
| Shell script hangs | Daemon did not start | Check stderr; verify port is free |
| Integration passes locally, fails CI | Go version mismatch | Check CI matches `go.mod` |
| Smoke passes but integration fails | Layer mismatch | Isolate CLI vs transport vs core |

## See Also

- `vm-system help architecture`
- `vm-system help api-reference`
- `vm-system help getting-started`
