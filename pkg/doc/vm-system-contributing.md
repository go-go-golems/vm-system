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

This page covers everything you need to make changes to vm-system safely:
where code belongs, how to test it, what reviewers look for, and how to
debug when things break.

If you haven't run vm-system yet, start with `vm-system help getting-started`
and come back here when you're ready to change code.

## Where does my change go?

vm-system has clean layer boundaries, and respecting them is the most
important thing you can do for code quality. Before writing code, figure out
which layer owns the behavior you're changing:

- **HTTP shape, status codes, request validation** →
  `pkg/vmtransport/http` (`server*.go`). This is where JSON is decoded,
  fields are validated, and domain errors are mapped to HTTP status codes.
- **Domain orchestration and policy** →
  `pkg/vmcontrol` (`*_service.go`, `core.go`). Business rules like "what
  defaults does a new template get?" and "is this file path safe?" live here.
- **Runtime semantics, goja execution** →
  `pkg/vmsession` and `pkg/vmexec`. This is where JavaScript actually runs,
  where the console shim captures output, and where execution locks are
  managed.
- **Persistence and schema** →
  `pkg/vmstore` (`vmstore*.go`). All database queries and schema definitions.
- **CLI output and UX** →
  `cmd/vm-system` (`cmd_*.go`). Flag parsing, output formatting, and help
  text.
- **REST client** →
  `pkg/vmclient` (`*_client.go`). The typed HTTP client used by CLI commands.

The most common mistake is leaking HTTP-specific logic into vmcontrol, or
putting domain policy in a handler. When in doubt, ask yourself: "Does this
code need to know about HTTP requests?" If no, it belongs in vmcontrol or
below.

## Testing

### Quick start

For most changes, this is all you need:

```bash
GOWORK=off go test ./pkg/vmtransport/http -count=1    # integration tests
GOWORK=off go test ./... -count=1                      # all Go tests
bash ./smoke-test.sh                                   # daemon sanity (~10s)
```

Before opening a PR, run the full suite:

```bash
bash ./test-all.sh
```

Always prefix Go test commands with `GOWORK=off` — without it, workspace mode
can cause confusing import errors.

### Integration tests — where most of the value is

The integration tests in `pkg/vmtransport/http` are the most important tests
in the codebase. Each test spins up a complete in-memory stack — a real SQLite
store, the full vmcontrol core, and an `httptest.Server` with the actual HTTP
handler. No mocks anywhere. This means they exercise the entire request path
from HTTP decoding through domain logic to database persistence.

Here's what each test file covers:

- **server_integration_test.go** — verifies that state carries across REPL
  calls within a single session (set a variable in call 1, read it in call 2)
- **server_templates_integration_test.go** — full template CRUD including
  startup files, capabilities, modules, and libraries
- **server_sessions_integration_test.go** — session lifecycle (create, list,
  filter by status, close, delete) and runtime summary transitions
- **server_executions_integration_test.go** — REPL and run-file execution,
  execution listing, events, and `after_seq` cursor-based pagination
- **server_error_contracts_integration_test.go** — every error path: missing
  fields (400), unknown resources (404), concurrent execution (409), path
  traversal (422)
- **server_safety_integration_test.go** — path traversal rejection with
  various attack patterns, plus output and event limit enforcement
- **server_libraries_integration_test.go** — library loading and configuration
- **server_native_modules_integration_test.go** — module allowlist enforcement,
  including the `MODULE_NOT_ALLOWED` response for JavaScript built-ins
- **server_execution_contracts_integration_test.go** — execution edge cases
  and contract refinements

### Unit tests

Several packages have focused unit tests that don't need the full stack:
`vmpath` (path normalization), `vmmodels` (ID parsing, JSON helpers),
`vmdaemon/config` (configuration defaults), `libloader` (loading files from
cache), `vmclient` (URL construction, error parsing), `vmcontrol/template_service`
(template creation defaults), and `vmexec` (executor behavior and persistence
failure handling).

### Shell scripts

The shell scripts test the system from the outside, the way a real user would:

- **smoke-test.sh** (~10 seconds) — builds the binary, starts a daemon,
  creates a template and session, runs REPL and file execution, and checks
  the runtime summary. This is your fastest feedback loop.
- **test-e2e.sh** — a thorough CLI loop from build through close, covering
  every major command.
- **test-library-matrix.sh** — specifically tests module and library semantics:
  that JSON works without configuration, that JSON can't be added as a module,
  that lodash fails without configuration, and that lodash succeeds with it.
- **test-all.sh** — runs everything in sequence.

All scripts use temporary databases, temporary worktrees, and dynamically
allocated ports. They're safe to run in parallel with other work.

### What's well-covered, and what isn't

The test suite has strong coverage of the core operational path:
template/session/execution endpoints, error contracts for all status codes,
path traversal rejection, output limit enforcement, module allowlist behavior,
and the full CLI happy path.

Known gaps that could use attention: daemon restart recovery (what happens to
sessions that are "ready" in the DB but have no runtime), load and concurrency
testing beyond the basic `SESSION_BUSY` contract, library cache filename
consistency between the downloader and the loader, and process crash durability
testing.

## Common change patterns

### Adding a new API endpoint

This is the most common type of change. Follow the existing pattern:

1. Add the route in `server.go` (look at how existing routes are registered)
2. Write the handler: decode JSON → validate required fields → call core → map errors
3. Add a core service method in `pkg/vmcontrol` if the endpoint needs new
   business logic
4. Add a `vmclient` wrapper so the CLI can call it
5. Add a CLI command if the endpoint is user-facing
6. Add an integration test that asserts on both the HTTP status code and the
   `error.code` field in the JSON envelope

### Adding a new domain error

When the system needs to distinguish a new failure case:

1. Add a sentinel error variable in `pkg/vmmodels/models.go`
2. Add a case in `writeCoreError` in `server_errors.go` — this is the single
   place where domain errors become HTTP responses
3. Add an integration test in the error contracts suite to lock in the behavior

### Changing session lifecycle

Session state transitions are the most sensitive part of the system:

1. Update `SessionManager` in `pkg/vmsession`
2. Make sure database row updates happen in a deterministic order — partial
   updates can leave the database and memory out of sync
3. Add runtime summary assertions in integration tests to verify the in-memory
   state matches expectations
4. Verify that close and delete still behave consistently

### Writing a new integration test

Follow the pattern established by existing tests:

```go
func TestMyBehavior(t *testing.T) {
    store := vmstore.NewMemoryStore(t)
    core := vmcontrol.NewCore(store)
    server := httptest.NewServer(vmhttp.NewHandler(core))
    defer server.Close()

    template := createTestTemplate(t, server.URL)
    session := createTestSession(t, server.URL, template.ID)

    resp, err := http.Post(server.URL+"/api/v1/executions/repl", ...)
    assert.Equal(t, 201, resp.StatusCode)
    // parse response body and assert on specific fields...
}
```

The test helpers (`createTestTemplate`, `createTestSession`) handle the
boilerplate of creating prerequisite resources.

## Code review checklist

Before requesting review, check these:

- [ ] Your change has deterministic tests (no flaky assertions)
- [ ] Status codes and error codes are intentional, not accidental
- [ ] No transport-domain coupling was introduced (HTTP logic stays in
  handlers, business rules stay in vmcontrol)
- [ ] CLI output is still readable and stable
- [ ] Docs or changelog are updated for user-visible changes

## Git conventions

Use scope prefixes in commit messages so the history is scannable:

```
feat(api): add execution-not-found 404 contract
test(integration): add execution get not-found test
docs(help): update api-reference for new endpoint
```

Stage intentionally. Check `git diff --staged` before committing — it's easy
to accidentally include debug prints or scratch files.

## Good first contributions

If you're looking for something bounded and meaningful for a first PR, here
are five options that are practical, well-scoped, and directly improve the
system:

1. **Structured logging** — add zerolog fields to key lifecycle events (session
   create/close, execution start/end) for observability
2. **Library cache consistency** — the downloader writes versioned filenames
   and the loader looks for them, but the naming convention isn't always
   consistent. Unifying this would prevent subtle load failures.
3. **Benchmark tests** — characterize execution throughput and event storage
   under load. This would establish a baseline for performance work.
4. **Daemon restart policy** — implement close-on-restart: when the daemon
   starts, mark any "ready" sessions in the database as "closed" since their
   runtimes are gone. Add integration tests.
5. **Metadata endpoint** — add an API for querying and updating template
   settings directly, so you don't have to delete and recreate templates to
   change limits.

## Debugging

**API issues:** Always reproduce with curl first. This isolates CLI formatting
from actual server behavior:

```bash
curl -i -sS http://127.0.0.1:3210/api/v1/templates/does-not-exist
```

Compare the status code and error envelope against the `writeCoreError`
mapping in `server_errors.go`. If the error code doesn't match what you
expect, that's where to look.

**Session and runtime issues:** Check what the daemon thinks is alive:

```bash
vm-system session list --status ready
vm-system ops runtime-summary
```

If `session list` shows sessions as "ready" but `runtime-summary` shows zero
active sessions, the daemon was restarted and the runtimes are gone.

**Execution issues:** Run a minimal REPL snippet and inspect every event:

```bash
vm-system exec repl <session-id> '1+1'
vm-system exec events <execution-id> --after-seq 0
```

Look at the event types and ordering. If you're getting unexpected exceptions,
the stack trace in the exception event usually points to the problem.

**Persistence issues:** Always use throwaway database paths when debugging.
Stale databases from previous runs cause confusing behavior — you'll see old
templates, old sessions, and wonder why your changes aren't taking effect.

**Test failures:** For integration tests, read the assertion message first —
it usually shows expected vs actual status codes or response bodies. Then check
whether `writeCoreError` maps the error correctly. Re-run with `-v` for full
output. For shell scripts, find the last line of successful output and try
the failing command manually.

## See Also

- `vm-system help architecture` — how the layers fit together and why
- `vm-system help api-reference` — endpoint contracts and error codes
- `vm-system help getting-started` — first-run walkthrough
