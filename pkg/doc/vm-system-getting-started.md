---
Title: "Getting Started with vm-system"
Slug: getting-started
Short: "Build, run the daemon, and master the template → session → execution workflow."
Topics:
- vm-system
- getting-started
- cli
- daemon
Commands:
- serve
- template
- session
- exec
- ops
- libs
IsTopLevel: true
IsTemplate: false
ShowPerDefault: true
SectionType: Tutorial
---

This tutorial walks you from first build through daily workflows. By the end
you will have run JavaScript inside a managed goja runtime, configured modules
and libraries, and know the patterns for day-to-day use.

## Prerequisites

- Go toolchain (compatible with `go 1.25.5`, see `go.mod`)
- A Unix shell (bash or zsh)
- `curl` (for optional direct API verification)

## First run in five minutes

### Build the binary

```bash
GOWORK=off go build -o vm-system ./cmd/vm-system
./vm-system --help
```

You should see command groups: `serve`, `template`, `session`, `exec`, `ops`,
`libs`.

### Prepare a scratch workspace

Sessions need a worktree directory on disk:

```bash
mkdir -p /tmp/vm-scratch/runtime

cat > /tmp/vm-scratch/runtime/init.js <<'JS'
console.log("startup loaded")
globalThis.seed = 40
JS

cat > /tmp/vm-scratch/app.js <<'JS'
console.log("running app.js")
seed + 2
JS
```

### Start the daemon

```bash
./vm-system serve --db /tmp/vm-scratch.db --listen 127.0.0.1:3210
```

The daemon stays in the foreground. Open a second terminal for the rest.

```bash
curl -sS http://127.0.0.1:3210/api/v1/health
# {"status":"ok"}
```

### Create a template and add a startup file

```bash
./vm-system template create --name my-first --engine goja
# Output: Created template: my-first (ID: <template-id>)

./vm-system template add-startup <template-id> \
  --path runtime/init.js --order 10 --mode eval

./vm-system template get <template-id>
```

### Create a session

```bash
./vm-system session create \
  --template-id <template-id> \
  --workspace-id ws-demo \
  --base-commit deadbeef \
  --worktree-path /tmp/vm-scratch
```

### Execute code

```bash
./vm-system exec repl <session-id> 'seed + 2'       # returns 42
./vm-system exec run-file <session-id> app.js        # runs the file
```

### Inspect and clean up

```bash
./vm-system session list
./vm-system ops runtime-summary
./vm-system session close <session-id>
# Ctrl+C the daemon
```

### What just happened

```
CLI command → vmclient REST call → HTTP handler → vmcontrol service → runtime/store
```

The daemon owns goja runtimes in memory. CLI commands are thin REST clients.
SQLite stores everything durably.

## Daily workflows

### Template configuration

Templates are persistent runtime profiles — engine, limits, startup files,
modules, and libraries:

```bash
# Create
vm-system template create --name my-service --engine goja

# Startup files (execute in order_index order during session creation)
vm-system template add-startup <id> --path runtime/polyfills.js --order 10 --mode eval
vm-system template add-startup <id> --path runtime/globals.js --order 20 --mode eval

# Native modules (database, exec, fs)
vm-system template list-available-modules
vm-system template add-module <id> --name database

# Third-party libraries
vm-system template list-available-libraries
vm-system libs download                              # populate cache first
vm-system template add-library <id> --name lodash-4.17.21

# Inspect, list, delete
vm-system template get <id>
vm-system template list
vm-system template delete <id>
```

Note: JavaScript built-ins (JSON, Math, Date) are always available and cannot
be added as modules — you will get `MODULE_NOT_ALLOWED`.

### Session management

A session is a live goja runtime bound to a template and worktree:

```bash
# Create (worktree dir must exist, use absolute path)
vm-system session create \
  --template-id <id> \
  --workspace-id my-ws \
  --base-commit abc123 \
  --worktree-path /absolute/path

# Monitor
vm-system session list
vm-system session list --status ready
vm-system session get <session-id>
vm-system ops runtime-summary

# Close (discards runtime, keeps DB row)
vm-system session close <session-id>
```

### Execution

```bash
# REPL — state persists across calls
vm-system exec repl <session-id> 'var counter = 0'
vm-system exec repl <session-id> 'counter += 1; counter'  # 1
vm-system exec repl <session-id> 'counter += 1; counter'  # 2

# Run files (path relative to worktree, no ../ allowed)
vm-system exec run-file <session-id> scripts/transform.js

# Inspect history and events
vm-system exec list <session-id> --limit 20
vm-system exec get <execution-id>
vm-system exec events <execution-id> --after-seq 0
```

### Operations and direct API access

```bash
vm-system ops health
vm-system ops runtime-summary

# curl for scripting
curl -sS http://127.0.0.1:3210/api/v1/templates
curl -sS "http://127.0.0.1:3210/api/v1/sessions?status=ready"
curl -sS "http://127.0.0.1:3210/api/v1/executions?session_id=<id>&limit=5"
```

## Common patterns

### ETL pipeline with database access

```bash
vm-system template create --name etl --engine goja
vm-system template add-module <id> --name database
vm-system template add-startup <id> --path runtime/db-setup.js --order 10 --mode eval

vm-system session create --template-id <id> --workspace-id etl \
  --base-commit main --worktree-path /data/etl

vm-system exec run-file <session-id> steps/extract.js
vm-system exec run-file <session-id> steps/transform.js
vm-system exec run-file <session-id> steps/load.js
vm-system exec repl <session-id> 'db.query("SELECT count(*) FROM output_table")'
vm-system session close <session-id>
```

### Interactive development with lodash

```bash
vm-system template create --name dev --engine goja
vm-system libs download
vm-system template add-library <id> --name lodash-4.17.21

vm-system session create --template-id <id> --workspace-id dev \
  --base-commit HEAD --worktree-path /project

vm-system exec repl <session-id> '_.chunk([1,2,3,4,5,6], 2)'
vm-system exec repl <session-id> '_.groupBy(["one","two","three"], "length")'
```

## Troubleshooting

| Symptom | Error code | Cause | Fix |
|---------|-----------|-------|-----|
| Connection refused | — | Daemon not running | Start `vm-system serve` |
| Session create fails | — | Worktree directory missing | `mkdir -p` the path; use absolute paths |
| Session status `crashed` | — | Startup script or library error | Check `session get` for `Last Error`; fix script or `libs download` |
| 400 `VALIDATION_ERROR` | `VALIDATION_ERROR` | Missing required field | Check required flags/fields |
| 400 `INVALID_REQUEST` | `INVALID_REQUEST` | Extra or malformed JSON field | Remove unknown fields |
| 404 `TEMPLATE_NOT_FOUND` | `TEMPLATE_NOT_FOUND` | Wrong ID or different DB | Verify with `template list` |
| 404 `SESSION_NOT_FOUND` | `SESSION_NOT_FOUND` | Wrong ID or session deleted | Verify with `session list` |
| 409 `SESSION_BUSY` | `SESSION_BUSY` | Concurrent execution | Wait and retry |
| 409 `SESSION_NOT_READY` | `SESSION_NOT_READY` | Session crashed or closed | Check status; create new session |
| 422 `INVALID_PATH` | `INVALID_PATH` | Absolute or `../` path | Use relative path within worktree |
| 422 `OUTPUT_LIMIT_EXCEEDED` | `OUTPUT_LIMIT_EXCEEDED` | Too much output | Reduce output or increase template limits |
| 422 `MODULE_NOT_ALLOWED` | `MODULE_NOT_ALLOWED` | Adding built-in as module | Use `list-available-modules` |
| 500 `INTERNAL` | `INTERNAL` | Unmapped domain error | File a bug |
| REPL returns undefined | — | Input is a statement, not expression | End with expression: `var x=1; x` |
| Library load fails | — | Not downloaded to cache | `vm-system libs download` then check `.vm-cache/libraries/` |
| Sessions gone after restart | — | In-memory runtimes lost | Create new sessions (known limitation) |

## See Also

- `vm-system help architecture`
- `vm-system help api-reference`
- `vm-system help templates-and-sessions`
- `vm-system help cli-command-reference`
- `vm-system help contributing`
- `vm-system help examples`
