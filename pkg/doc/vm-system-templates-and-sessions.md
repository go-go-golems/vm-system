---
Title: "Templates, Sessions, and Executions"
Slug: templates-and-sessions
Short: "How templates define runtime policy, sessions own live runtimes, and executions capture results."
Topics:
- vm-system
- templates
- sessions
- executions
- concepts
Commands:
- template
- session
- exec
IsTopLevel: true
IsTemplate: false
ShowPerDefault: true
SectionType: GeneralTopic
---

vm-system organizes JavaScript execution around three core concepts: templates,
sessions, and executions. Understanding the lifecycle of each is essential for
both using the CLI and contributing to the codebase.

## Templates — runtime policy profiles

A template is a persisted configuration that defines how a JavaScript runtime
should behave. It captures:

- **Engine** — currently `goja` (QuickJS, Node, custom planned)
- **Resource limits** — CPU time, wall time, memory, max events, max output
- **Resolver settings** — module roots, file extensions, import policy
- **Runtime settings** — ESM mode, strict mode, console availability
- **Startup files** — scripts executed in order when a session starts
- **Native modules** — host-provided capabilities (database, exec, fs)
- **Libraries** — third-party JS libraries loaded into the runtime (lodash, moment, etc.)
- **Capabilities** — metadata describing what features are enabled

### Creating and configuring a template

```bash
# Create template with defaults
vm-system template create --name my-runtime --engine goja

# Add startup files (executed in order_index order)
vm-system template add-startup <template-id> --path runtime/init.js --order 10 --mode eval
vm-system template add-startup <template-id> --path runtime/globals.js --order 20 --mode eval

# Add a native module
vm-system template add-module <template-id> --name console

# Add a third-party library
vm-system template add-library <template-id> --name lodash-4.17.21

# Add capability metadata
vm-system template add-capability <template-id> --kind module --name console --enabled

# Inspect the full template
vm-system template get <template-id>
```

### Default settings

Every new template gets default settings immediately. The defaults are:

| Category | Setting | Default |
|----------|---------|---------|
| Limits | CPU timeout | 5000ms |
| Limits | Wall timeout | 30000ms |
| Limits | Memory | 128MB |
| Limits | Max events | 1000 |
| Limits | Max output | 512KB |
| Resolver | Extensions | `.js`, `.mjs` |
| Resolver | Absolute imports | disabled |
| Runtime | ESM | disabled |
| Runtime | Strict | disabled |
| Runtime | Console | enabled |

Settings are stored as JSON blobs and can evolve without schema migrations.

### Available modules and libraries

vm-system ships with a catalog of configurable native modules and downloadable
libraries:

```bash
# See what native modules can be configured
vm-system template list-available-modules

# See what libraries can be downloaded and loaded
vm-system template list-available-libraries
```

**Native modules** (template-configurable):
- `database` — SQLite access (configure, query, exec, close)
- `exec` — Run external commands (run)
- `fs` — File read/write (readFileSync, writeFileSync)

**Built-in JavaScript globals** (JSON, Math, Date, etc.) are always available
and are not template-configurable. Attempting to add them as modules returns
`MODULE_NOT_ALLOWED`.

### Template deletion

Deleting a template cascades to settings, capabilities, and startup files via
foreign keys. Active sessions using that template are not affected (their
runtime is already created in memory).

## Sessions — live runtime instances

A session is a long-lived goja runtime instance bound to a template and a
workspace. The daemon process holds the runtime in memory.

### Session lifecycle

```
POST /api/v1/sessions  →  starting  →  ready
                                    →  crashed (startup failure)
                        ready       →  closed  (explicit close)
```

**Starting:** The daemon allocates a goja runtime, applies template settings,
loads libraries, and executes startup files in order.

**Ready:** The session is available for execution requests.

**Crashed:** Startup failed (bad script, missing library, etc.). The `last_error`
field captures the failure message.

**Closed:** The in-memory runtime is discarded. The session row persists in the
database with `closed_at` set.

### Creating a session

```bash
vm-system session create \
  --template-id <template-id> \
  --workspace-id ws-prod \
  --base-commit abc123 \
  --worktree-path /absolute/path/to/code
```

All four parameters are required:

- **template-id** — Which template (and therefore which settings/startup/modules) to use
- **workspace-id** — Logical workspace identifier for your tracking
- **base-commit** — Git commit OID for reproducibility context
- **worktree-path** — Absolute filesystem path; must exist; used for run-file resolution

### Important runtime behavior

- **One execution at a time** — Each session has a `TryLock` mutex. Concurrent
  requests get `SESSION_BUSY` (409), not queued.
- **State carries across executions** — Variables set in one REPL call persist
  in subsequent calls within the same session.
- **Startup state persists** — `globalThis` mutations from startup files are
  visible to all later executions.
- **Daemon restart loses runtimes** — In-memory session state does not survive
  daemon process restart. DB rows remain, but the runtime is gone.

### Listing and filtering sessions

```bash
vm-system session list                  # all sessions
vm-system session list --status ready   # only active sessions
vm-system session get <session-id>      # full detail with timestamps
```

## Executions — discrete code runs

An execution is a single REPL snippet or run-file invocation within a session.
Each execution produces a persisted record and a stream of events.

### Execution types

| Kind | Trigger | Input |
|------|---------|-------|
| `repl` | `exec repl` | Inline JavaScript string |
| `run_file` | `exec run-file` | File path relative to worktree |
| `startup` | Session creation | Startup file from template |

### Execution lifecycle

```
created (running)  →  ok       (value returned)
                   →  error    (exception thrown)
                   →  timeout  (limit exceeded)
```

### REPL execution

```bash
vm-system exec repl <session-id> 'Math.random()'
```

The snippet runs in the session runtime. The return value becomes a `value`
event. Console calls become `console` events. Exceptions become `exception`
events.

### File execution

```bash
vm-system exec run-file <session-id> scripts/transform.js
```

The path is normalized and validated:

- Must be relative (no leading `/`)
- Must not escape the worktree (`../` rejected → `INVALID_PATH`)
- File must exist within the worktree (→ `FILE_NOT_FOUND`)

### Event stream

Every execution produces events with sequential `seq` numbers:

```bash
# Get all events
vm-system exec events <execution-id> --after-seq 0

# Get events after a cursor (for polling)
vm-system exec events <execution-id> --after-seq 5
```

Events capture the full execution trace: input echo, console output, return
values, and exceptions.

### Resource limits

Template settings define limits that are enforced during execution:

- **max_events** — Maximum event count per execution
- **max_output_kb** — Maximum total output payload size
- **cpu_ms / wall_ms** — Execution time limits

Exceeding limits produces `OUTPUT_LIMIT_EXCEEDED` (422).

## Putting it all together

A typical workflow chains all three concepts:

```bash
# 1. Define the runtime policy
vm-system template create --name etl-runner --engine goja
vm-system template add-module <id> --name database
vm-system template add-startup <id> --path runtime/setup-db.js --order 10 --mode eval

# 2. Create a session for a specific workspace
vm-system session create --template-id <id> --workspace-id etl-ws \
  --base-commit abc123 --worktree-path /data/etl-workspace

# 3. Run multiple executions in the same session
vm-system exec repl <session-id> 'db.query("SELECT count(*) FROM users")'
vm-system exec run-file <session-id> transforms/normalize.js
vm-system exec run-file <session-id> transforms/export.js

# 4. Inspect results
vm-system exec list <session-id>
vm-system exec events <last-execution-id> --after-seq 0

# 5. Clean up
vm-system session close <session-id>
```

## Troubleshooting

| Problem | Cause | Solution |
|---------|-------|----------|
| Session status is `crashed` | Startup script failed or library not found | Check `last_error` via `session get`; fix startup file or download library with `libs download` |
| `SESSION_BUSY` | Overlapping execution requests | Wait for current execution; sessions allow one at a time |
| REPL returns `undefined` | Expression has no return value | Use an expression, not a statement (e.g. `x` not `var x = 1`) |
| Library not loaded | Cache filename mismatch | Verify library is in `.vm-cache/libraries` with expected filename |

## See Also

- `vm-system help getting-started`
- `vm-system help api-reference`
- `vm-system help cli-command-reference`
