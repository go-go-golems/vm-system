---
Title: "How to Use vm-system"
Slug: how-to-use
Short: "Practical guide covering daemon startup, template configuration, session management, and code execution workflows."
Topics:
- vm-system
- cli
- guide
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

This guide covers the practical workflows you need for daily vm-system usage.
It assumes you have already completed the getting-started tutorial and have a
built binary.

## Starting the daemon

vm-system is daemon-first. All runtime operations require a running daemon:

```bash
vm-system serve --db /path/to/state.db --listen 127.0.0.1:3210
```

The daemon stays in the foreground. Use `--db` to control where templates,
sessions, and execution history are persisted. Use `--listen` to set the
HTTP address.

Verify the daemon is healthy:

```bash
vm-system ops health
# or
curl -sS http://127.0.0.1:3210/api/v1/health
```

## Template workflows

Templates define runtime policy. Think of them as VM configuration profiles.

### Create a basic template

```bash
vm-system template create --name my-service --engine goja
```

This gives you a template with default resource limits, console enabled, and
no modules or libraries.

### Configure startup files

Startup files execute during session creation, in `order_index` order:

```bash
vm-system template add-startup <template-id> --path runtime/polyfills.js --order 10 --mode eval
vm-system template add-startup <template-id> --path runtime/globals.js --order 20 --mode eval
```

Paths are resolved relative to the session worktree at session creation time.

### Add native modules

Native modules provide host capabilities (database access, file I/O, command
execution):

```bash
# See what is available
vm-system template list-available-modules

# Add modules
vm-system template add-module <template-id> --name database
vm-system template add-module <template-id> --name fs
```

Note: JavaScript built-ins (JSON, Math, Date) are always available and cannot
be added as modules.

### Add libraries

Third-party libraries are loaded into the runtime at session startup:

```bash
# See what is available
vm-system template list-available-libraries

# Download to local cache first
vm-system libs download

# Add to template
vm-system template add-library <template-id> --name lodash-4.17.21
```

### Inspect a template

```bash
vm-system template get <template-id>
```

This shows settings, capabilities, startup files, modules, and libraries.

### List and delete templates

```bash
vm-system template list
vm-system template delete <template-id>
```

## Session workflows

A session is a live goja runtime instance. It holds state across executions.

### Create a session

```bash
vm-system session create \
  --template-id <template-id> \
  --workspace-id my-workspace \
  --base-commit abc123 \
  --worktree-path /absolute/path/to/code
```

The worktree directory must exist. Startup files from the template are executed
immediately. The session is `ready` when creation succeeds.

### Monitor sessions

```bash
# All sessions
vm-system session list

# Only active sessions
vm-system session list --status ready

# Detailed view
vm-system session get <session-id>

# Runtime view (in-memory state)
vm-system ops runtime-summary
```

### Close a session

```bash
vm-system session close <session-id>
```

This discards the in-memory runtime. The session row persists in the database.

## Execution workflows

### REPL — run snippets

```bash
vm-system exec repl <session-id> 'Math.random()'
vm-system exec repl <session-id> 'db.query("SELECT count(*) FROM users")'
```

State persists across REPL calls within the same session:

```bash
vm-system exec repl <session-id> 'var counter = 0'
vm-system exec repl <session-id> 'counter += 1; counter'  # returns 1
vm-system exec repl <session-id> 'counter += 1; counter'  # returns 2
```

### Run files

```bash
vm-system exec run-file <session-id> scripts/transform.js
```

The path must be relative to the session worktree. No absolute paths, no
`../` traversal.

### Inspect execution history

```bash
# List recent executions
vm-system exec list <session-id> --limit 20

# Get execution detail
vm-system exec get <execution-id>

# Get events (console output, values, exceptions)
vm-system exec events <execution-id> --after-seq 0
```

Use `--after-seq` for cursor-based polling of event streams.

## Operations workflows

### Health and runtime state

```bash
vm-system ops health
vm-system ops runtime-summary
```

The runtime summary shows in-memory active sessions — not just database rows.
This is what you want when checking if sessions are actually alive.

### Direct API access

For scripting and debugging, use curl directly:

```bash
curl -sS http://127.0.0.1:3210/api/v1/health
curl -sS http://127.0.0.1:3210/api/v1/runtime/summary
curl -sS http://127.0.0.1:3210/api/v1/templates
curl -sS "http://127.0.0.1:3210/api/v1/sessions?status=ready"
curl -sS "http://127.0.0.1:3210/api/v1/executions?session_id=<id>&limit=5"
```

## Common multi-step patterns

### ETL pipeline with database access

```bash
# Setup
vm-system template create --name etl --engine goja
vm-system template add-module <id> --name database
vm-system template add-startup <id> --path runtime/db-setup.js --order 10 --mode eval

# Create session
vm-system session create --template-id <id> --workspace-id etl \
  --base-commit main --worktree-path /data/etl

# Run pipeline steps
vm-system exec run-file <session-id> steps/extract.js
vm-system exec run-file <session-id> steps/transform.js
vm-system exec run-file <session-id> steps/load.js

# Check results
vm-system exec repl <session-id> 'db.query("SELECT count(*) FROM output_table")'

# Cleanup
vm-system session close <session-id>
```

### Interactive development with lodash

```bash
vm-system template create --name dev --engine goja
vm-system libs download
vm-system template add-library <id> --name lodash-4.17.21
vm-system session create --template-id <id> --workspace-id dev \
  --base-commit HEAD --worktree-path /project

# Use lodash interactively
vm-system exec repl <session-id> '_.map([1,2,3], n => n * 2)'
vm-system exec repl <session-id> '_.groupBy(["one","two","three"], "length")'
```

## Troubleshooting

| Problem | Cause | Solution |
|---------|-------|----------|
| Commands fail with connection error | Daemon not running | Start `vm-system serve` |
| Session crashes on creation | Startup script error | Check `session get` for `Last Error` |
| Library not found at runtime | Not downloaded | Run `vm-system libs download` first |
| `SESSION_BUSY` | Concurrent execution | Wait for current execution to complete |

## See Also

- `vm-system help getting-started`
- `vm-system help cli-command-reference`
- `vm-system help templates-and-sessions`
- `vm-system help troubleshooting`
