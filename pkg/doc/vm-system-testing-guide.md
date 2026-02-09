---
Title: "Testing Guide"
Slug: testing-guide
Short: "How to run tests, what each suite covers, and how to add new test cases."
Topics:
- vm-system
- testing
- integration
- development
Commands:
- serve
IsTopLevel: true
IsTemplate: false
ShowPerDefault: true
SectionType: GeneralTopic
---

vm-system uses a layered testing strategy: Go integration tests for API
contract behavior, Go unit tests for isolated logic, and shell scripts for
end-to-end CLI validation. This guide explains what each layer covers, how to
run it, and how to add new tests.

## Quick reference

```bash
# All Go tests
GOWORK=off go test ./... -count=1

# Integration tests only (highest value)
GOWORK=off go test ./pkg/vmtransport/http -count=1

# Specific test
GOWORK=off go test ./pkg/vmtransport/http -run TestExecutionEndpointsLifecycle -count=1

# Shell integration
bash ./smoke-test.sh          # ~10 seconds
bash ./test-e2e.sh            # Full CLI loop
bash ./test-library-matrix.sh # Library/module semantics
bash ./test-all.sh            # Everything
```

Always use `GOWORK=off` to avoid workspace-mode interference.

## Integration tests (pkg/vmtransport/http)

These are the most valuable tests. They exercise the full stack — store, core,
HTTP handler — through a real `httptest.Server` with an in-memory SQLite
database. No mocks are used.

### How they work

Each test function:

1. Creates a fresh SQLite store (`:memory:` or temp file)
2. Wires `vmcontrol.NewCore` with real adapters
3. Starts an `httptest.Server` with `vmhttp.NewHandler`
4. Makes HTTP requests and asserts on status codes, response bodies, and headers

### Test suite inventory

| File | What it covers |
|------|---------------|
| `server_integration_test.go` | Cross-execution state continuity (var set in REPL 1, read in REPL 2) |
| `server_templates_integration_test.go` | Template create, list, get, delete; startup file and capability CRUD |
| `server_sessions_integration_test.go` | Session create, list, get by status, close, delete; runtime summary transitions |
| `server_executions_integration_test.go` | REPL, run-file, execution get/list, events with `after_seq` filtering |
| `server_error_contracts_integration_test.go` | Validation errors (400), not-found (404), conflict (409), unprocessable (422) |
| `server_safety_integration_test.go` | Path traversal rejection, absolute path rejection, output/event limit enforcement |
| `server_libraries_integration_test.go` | Library loading and configuration behavior |
| `server_native_modules_integration_test.go` | Module catalog allowlist enforcement (`MODULE_NOT_ALLOWED` for built-ins) |
| `server_execution_contracts_integration_test.go` | Execution edge cases and contract refinements |

### Adding a new integration test

Follow this pattern:

```go
func TestMyNewBehavior(t *testing.T) {
    // 1. Setup test server
    store := vmstore.NewMemoryStore(t)
    core := vmcontrol.NewCore(store)
    server := httptest.NewServer(vmhttp.NewHandler(core))
    defer server.Close()

    // 2. Create prerequisite resources
    template := createTestTemplate(t, server.URL)
    session := createTestSession(t, server.URL, template.ID)

    // 3. Exercise behavior
    resp, err := http.Post(server.URL+"/api/v1/executions/repl", ...)

    // 4. Assert status code and response body
    assert.Equal(t, 201, resp.StatusCode)
    // assert on JSON fields...
}
```

When testing error contracts, assert on both the HTTP status code and the
`error.code` field in the JSON envelope.

## Unit tests

### vmpath/path_test.go

Tests path normalization and traversal detection. These are pure functions
with no I/O.

### vmmodels/ids_test.go, json_helpers_test.go

Tests typed ID parsing and JSON marshal/unmarshal helpers.

### vmdaemon/config_test.go

Tests daemon configuration defaults and override behavior.

### libloader/loader_test.go

Tests library file loading from cache directory.

### vmclient/rest_client_test.go

Tests REST client URL construction and error parsing.

### vmcontrol/template_service_test.go

Tests template creation defaults and validation.

### vmexec/executor_test.go, executor_persistence_failures_test.go

Tests executor event capture and behavior under persistence failures.

## Shell scripts

### smoke-test.sh

Fast sanity check (~10 seconds). Validates:

- Binary builds successfully
- Daemon starts and becomes healthy
- Template create and module catalog commands work
- Session create with startup files succeeds
- REPL and run-file execution produce expected output
- Runtime summary reports correct session count

Uses isolated temp resources and dynamic port allocation.

### test-e2e.sh

Full CLI loop covering the same path as a real user:

- Build → daemon start → workspace setup → template → session → execution → listing → close

### test-library-matrix.sh

Tests library and module capability semantics:

- JSON built-in works without template configuration
- JSON cannot be added as a template module (`MODULE_NOT_ALLOWED`)
- lodash fails when not configured
- lodash succeeds when configured
- Post-hoc library configuration works

### test-all.sh

Runs all shell scripts in sequence. Use before opening a PR.

## Coverage analysis

### Well-covered areas

| Area | Test layer |
|------|-----------|
| Template CRUD + nested resources | Integration |
| Session lifecycle + state transitions | Integration |
| Execution REPL + run-file + events | Integration |
| Error code contracts (400/404/409/422) | Integration |
| Path traversal rejection | Integration + shell |
| Output/event limit enforcement | Integration |
| Module allowlist enforcement | Integration |
| CLI command flow | Shell scripts |

### Known gaps

| Area | Why it matters |
|------|---------------|
| Daemon restart recovery | Sessions in DB but no runtime — undefined behavior |
| Concurrent execution load | Only `SESSION_BUSY` tested; no throughput characterization |
| Library cache path consistency | Downloader and loader may disagree on filenames |
| Process crash durability | No test for abrupt kill + DB integrity |
| Fuzz testing for JSON payloads | Edge cases in decode/validation |

## Debugging test failures

### Integration test fails

1. Read the assertion message — it usually shows expected vs actual status/body
2. Check if `writeCoreError` maps the domain error correctly
3. Run the specific test with `-v` for full request/response logging:

```bash
GOWORK=off go test ./pkg/vmtransport/http -run TestFailingTest -count=1 -v
```

### Shell script fails

1. Scripts print each step; find the last successful output
2. Check if the daemon started (look for "listening on" line)
3. Check if template/session IDs were captured correctly
4. Run the failing curl/CLI command manually

### Flaky tests

All tests use isolated resources (temp DB, temp dirs, dynamic ports). If tests
are flaky, check for:

- Port conflicts (another daemon still running)
- Leftover temp files from a crashed previous run
- Race conditions in event ordering (unlikely with current synchronous model)

## Troubleshooting

| Problem | Cause | Solution |
|---------|-------|----------|
| `go test` fails with import errors | Missing `GOWORK=off` | Always prefix with `GOWORK=off` |
| Shell script hangs | Daemon did not start | Check daemon stderr; verify port is free |
| Integration test passes locally, fails in CI | Different Go version or missing tooling | Check CI Go version matches `go.mod` |
| Test creates real files | Missing temp directory cleanup | Ensure `t.TempDir()` or manual cleanup in defer |

## See Also

- `vm-system help contributing`
- `vm-system help architecture`
- `vm-system help api-reference`
