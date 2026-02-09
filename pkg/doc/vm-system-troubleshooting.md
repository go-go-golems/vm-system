---
Title: "Troubleshooting vm-system"
Slug: troubleshooting
Short: "Diagnosis and fixes for common daemon, session, execution, and library issues."
Topics:
- vm-system
- troubleshooting
- operations
- debugging
Commands:
- serve
- session
- exec
- ops
IsTopLevel: true
IsTemplate: false
ShowPerDefault: true
SectionType: GeneralTopic
---

This page collects common problems, their root causes, and concrete fixes.
Use it when something goes wrong during development or operations.

## Daemon issues

### Daemon does not start

**Symptoms:** `vm-system serve` exits immediately or prints an error.

**Checks:**

1. Is the port already in use? (`lsof -i :3210` or `ss -tlnp | grep 3210`)
2. Is the database path writable? Check directory permissions.
3. Is another daemon instance already running?

**Fix:** Kill the conflicting process or use a different `--listen` address and
`--db` path.

### Health endpoint unreachable

**Symptoms:** `curl http://127.0.0.1:3210/api/v1/health` connection refused.

**Checks:**

1. Confirm the daemon is running (check the terminal where you started it)
2. Confirm the listen address matches your `--server-url`
3. If using a remote host, check firewall rules

**Fix:** Start the daemon or correct the URL.

### CLI commands fail with connection errors

**Symptoms:** All client commands return transport-level errors.

**Root cause:** Daemon is not running, or `--server-url` does not match the
daemon's `--listen` address.

**Fix:**

```bash
# Verify daemon is up
curl -sS http://127.0.0.1:3210/api/v1/health

# If using non-default address
vm-system --server-url http://HOST:PORT ops health
```

## Template issues

### Template create fails with VALIDATION_ERROR

**Root cause:** Missing `--name` flag.

**Fix:** Always provide `--name`:

```bash
vm-system template create --name my-template
```

### MODULE_NOT_ALLOWED when adding a module

**Root cause:** You tried to add a JavaScript built-in (JSON, Math, Date, etc.)
as a template module. Built-ins are always available and not configurable.

**Fix:** Only add modules from the configurable catalog:

```bash
vm-system template list-available-modules
# Available: database, exec, fs
```

### Template not found after daemon restart

**Root cause:** Templates are persisted in SQLite. If you used a different
`--db` path, the templates are in a different database file.

**Fix:** Use the same `--db` path consistently, or create new templates.

## Session issues

### Session create fails with worktree error

**Root cause:** The `--worktree-path` directory does not exist on disk.

**Fix:** Create the directory before creating the session:

```bash
mkdir -p /path/to/worktree
vm-system session create --worktree-path /path/to/worktree ...
```

Always use absolute paths.

### Session status is "crashed"

**Root cause:** One of the startup files threw an exception, or a configured
library could not be loaded.

**Diagnosis:**

```bash
vm-system session get <session-id>
# Look at the "Last Error" field
```

**Common causes:**

- Startup script has a syntax error
- Startup script references a file that does not exist in the worktree
- A library was configured but not downloaded to the cache

**Fix:** Fix the startup script, ensure the file exists in the worktree, or
download the library:

```bash
vm-system libs download
```

### SESSION_NOT_READY (409)

**Root cause:** Session exists but is not in `ready` state. It may be
`starting`, `crashed`, or `closed`.

**Diagnosis:**

```bash
vm-system session get <session-id>
```

**Fix:** If crashed, create a new session with fixed configuration. If closed,
create a new session. If starting, wait (though this is usually very fast).

### SESSION_BUSY (409)

**Root cause:** Another execution is currently running in this session.
vm-system allows only one execution at a time per session.

**Fix:** Wait for the current execution to complete, then retry. There is no
execution queue — the client must retry.

### Sessions gone after daemon restart

**Root cause:** In-memory goja runtimes are not persisted. Daemon restart
discards all active runtimes. Session rows remain in the database but have no
backing runtime.

**Fix:** Create new sessions after restarting the daemon. This is a known
architectural limitation — full recovery semantics are planned but not yet
implemented.

## Execution issues

### INVALID_PATH (422) on run-file

**Root cause:** The file path is either absolute (starts with `/`) or contains
`../` traversal.

**Fix:** Use a relative path within the session worktree:

```bash
# Wrong
vm-system exec run-file <session-id> /absolute/path/to/file.js
vm-system exec run-file <session-id> ../outside/file.js

# Right
vm-system exec run-file <session-id> scripts/transform.js
```

### FILE_NOT_FOUND (404) on run-file

**Root cause:** The file does not exist at the resolved path within the session
worktree.

**Fix:** Check that the file exists relative to the worktree path used when
creating the session:

```bash
ls /path/to/worktree/scripts/transform.js
```

### OUTPUT_LIMIT_EXCEEDED (422)

**Root cause:** The execution produced more events or output bytes than the
template limits allow.

**Diagnosis:** Check the template settings:

```bash
vm-system template get <template-id>
# Look at limits: max_events, max_output_kb
```

**Fix:** Reduce output volume in your script (fewer `console.log` calls), or
create a template with higher limits.

### Execution returns undefined

**Root cause:** The REPL input is a statement (`var x = 1`) rather than an
expression (`x`). Statements do not produce a return value.

**Fix:** Use an expression as the last statement:

```bash
# Returns undefined
vm-system exec repl <session-id> 'var x = 1'

# Returns 1
vm-system exec repl <session-id> 'var x = 1; x'
```

### Exception in REPL execution

**Root cause:** The JavaScript code threw an error.

**Diagnosis:**

```bash
vm-system exec events <execution-id> --after-seq 0
# Look for "exception" type events with message and stack
```

## Library issues

### Library load failure during session creation

**Root cause:** The library was configured on the template but not downloaded to
the cache.

**Fix:**

```bash
# Download all libraries
vm-system libs download

# Verify cache contents
ls .vm-cache/libraries/
```

### Library cache naming mismatch

**Root cause:** The library downloader writes versioned filenames
(`lodash-4.17.21.js`) but the runtime loader may expect a different naming
convention.

**Status:** This is a known issue. Check `.vm-cache/libraries/` to see actual
filenames and compare with what the loader expects.

## API debugging

### Reproducing issues with curl

Always start debugging by reproducing with curl to isolate CLI vs API behavior:

```bash
# Health check
curl -i -sS http://127.0.0.1:3210/api/v1/health

# Template operations
curl -i -sS http://127.0.0.1:3210/api/v1/templates
curl -i -sS http://127.0.0.1:3210/api/v1/templates/<id>

# Session operations
curl -i -sS "http://127.0.0.1:3210/api/v1/sessions?status=ready"

# Execution with full headers
curl -i -sS -X POST http://127.0.0.1:3210/api/v1/executions/repl \
  -H 'Content-Type: application/json' \
  -d '{"session_id":"<id>","input":"1+1"}'
```

The `-i` flag shows response headers including `X-Request-Id` for log
correlation.

### 400 INVALID_REQUEST on valid-looking JSON

**Root cause:** The JSON body contains a field the server does not expect.
`DisallowUnknownFields` is enforced on all decode operations.

**Fix:** Remove extra fields from the request body. Check the API reference for
the exact field names.

### 500 INTERNAL for unexpected errors

**Root cause:** A domain error is not mapped in `writeCoreError`. This is a
bug in error handling.

**Fix:** File a bug. The error should be mapped to a specific status code and
error code. Check `pkg/vmtransport/http/server_errors.go` for the current
mapping.

## Troubleshooting summary table

| Symptom | Error code | Likely cause | Quick fix |
|---------|-----------|-------------|-----------|
| Connection refused | — | Daemon not running | Start `vm-system serve` |
| 400 VALIDATION_ERROR | `VALIDATION_ERROR` | Missing required field | Check API reference for required fields |
| 400 INVALID_REQUEST | `INVALID_REQUEST` | Malformed or extra JSON fields | Fix request body |
| 404 TEMPLATE_NOT_FOUND | `TEMPLATE_NOT_FOUND` | Wrong ID or wrong database | Verify with `template list` |
| 404 SESSION_NOT_FOUND | `SESSION_NOT_FOUND` | Wrong ID or session deleted | Verify with `session list` |
| 409 SESSION_BUSY | `SESSION_BUSY` | Concurrent execution | Wait and retry |
| 409 SESSION_NOT_READY | `SESSION_NOT_READY` | Session crashed or closed | Check status with `session get` |
| 422 INVALID_PATH | `INVALID_PATH` | Absolute or traversal path | Use relative path |
| 422 OUTPUT_LIMIT_EXCEEDED | `OUTPUT_LIMIT_EXCEEDED` | Too much output | Reduce output or increase limits |
| 422 MODULE_NOT_ALLOWED | `MODULE_NOT_ALLOWED` | Adding built-in as module | Use `list-available-modules` |
| 500 INTERNAL | `INTERNAL` | Unhandled error | File a bug |

## See Also

- `vm-system help getting-started`
- `vm-system help api-reference`
- `vm-system help templates-and-sessions`
