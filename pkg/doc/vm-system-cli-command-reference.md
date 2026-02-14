---
Title: "CLI Command Reference"
Slug: cli-command-reference
Short: "Complete reference for every vm-system CLI command, flag, and argument."
Topics:
- vm-system
- cli
- reference
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
SectionType: GeneralTopic
---

The vm-system CLI is split into six command groups. The `serve` command runs
the daemon itself; everything else is a REST client that talks to a running
daemon. If you're getting connection errors on any command except `serve`, the
daemon probably isn't running.

## Command tree

```
vm-system
├── serve                          start the daemon
├── template
│   ├── create / list / get / delete
│   ├── add-startup / list-startup
│   ├── add-capability / list-capabilities
│   ├── add-module / remove-module / list-modules
│   ├── add-library / remove-library / list-libraries
│   └── list-available-modules / list-available-libraries
├── session
│   ├── create / list / get / close
├── exec
│   ├── repl / run-file
│   └── list / get / events
├── ops
│   ├── health / runtime-summary
└── libs
    └── download
```

## Global flags

These flags apply to every command:

- **`--db PATH`** — path to the SQLite database file (default `vm-system.db`).
  This only matters for `serve` — it's where templates, sessions, and
  execution history are stored. If the file doesn't exist, it's created.
- **`--server-url URL`** — the daemon's HTTP address (default
  `http://127.0.0.1:3210`). Every command except `serve` uses this to connect
  to the daemon.
- **`--log-level LEVEL`** — controls logging verbosity. Accepts `debug`,
  `info`, `warn`, `error`. Default is `info`.

## serve

The `serve` command starts the daemon process. It initializes the database,
sets up the HTTP server, and blocks until you stop it with Ctrl+C (which
triggers graceful shutdown):

```bash
vm-system serve [--listen 127.0.0.1:3210]
```

The `--listen` flag sets the HTTP address and port. The daemon uses the `--db`
global flag for its SQLite database. You'll typically run this in one terminal
and use the other commands in another.

In merged-repo setups, `serve` can also host the web UI from `/` when frontend
assets are available under `internal/web/embed/public` (typically produced by
`go generate ./internal/web`). If assets are not available, the daemon runs in
API-only mode and still serves `/api/v1/*`.

## template

The `template` group manages runtime profiles. Templates are the blueprint
for sessions — they define what engine to use, what limits to enforce, what
scripts to run at startup, and what modules and libraries to make available.

### Creating and inspecting templates

```bash
vm-system template create --name NAME [--engine goja]
vm-system template list
vm-system template get TEMPLATE_ID
vm-system template delete TEMPLATE_ID
```

`--name` is the only required flag for `create`. The `--engine` flag defaults
to `goja`. When you create a template, default settings are initialized
automatically (5s CPU limit, 128MB memory, console enabled, etc.).

`delete` cascades — it removes the template's settings, capabilities, startup
files, modules, and libraries. Sessions already created from the template
are not affected.

### Startup files

Startup files execute during session creation, in the order specified by
`--order` (ascending). This is how you set up global state, define helpers,
or configure connections before any user code runs:

```bash
vm-system template add-startup TEMPLATE_ID --path PATH --order N [--mode eval]
vm-system template list-startup TEMPLATE_ID
```

The `--path` is relative to the session worktree (not to where you're running
the command). Only `eval` mode is supported today.

### Capabilities

Capabilities are metadata entries that describe what features a template
enables. They're used during session setup to configure the runtime:

```bash
vm-system template add-capability TEMPLATE_ID \
  --name NAME [--kind module] [--config '{}'] [--enabled]
vm-system template list-capabilities TEMPLATE_ID
```

The `--kind` flag accepts `module`, `global`, `fs`, `net`, and `env`.

### Modules

Modules are host-provided native capabilities — Go functions exposed to
JavaScript. You can only add modules from the configurable catalog; JavaScript
built-ins (JSON, Math, Date) are always available and will return
`MODULE_NOT_ALLOWED` if you try to add them:

```bash
vm-system template add-module TEMPLATE_ID --name NAME
vm-system template remove-module TEMPLATE_ID --name NAME
vm-system template list-modules TEMPLATE_ID
vm-system template list-available-modules     # shows: database, exec, fs
```

### Libraries

Libraries are third-party JavaScript files. They need to be in the local
cache before sessions can load them, so the usual workflow is: check what's
available, download, then attach to templates:

```bash
vm-system template list-available-libraries   # lodash, moment, axios, ...
vm-system template add-library TEMPLATE_ID --name NAME
vm-system template remove-library TEMPLATE_ID --name NAME
vm-system template list-libraries TEMPLATE_ID
```

## session

The `session` group manages live runtime instances. Creating a session
allocates a goja runtime, loads libraries and startup files, and leaves the
session in `ready` state (or `crashed` if something failed).

```bash
vm-system session create \
  --template-id ID \
  --workspace-id ID \
  --base-commit OID \
  --worktree-path /absolute/path
```

All four flags are required. The worktree directory must exist on disk and
the path must be absolute.

```bash
vm-system session list [--status ready]   # also: starting, crashed, closed
vm-system session get SESSION_ID
vm-system session close SESSION_ID
```

`session list` without `--status` shows everything. `session get` shows
timestamps, error messages, and other detail. `session close` discards the
in-memory runtime but keeps the database row as a historical record.

## exec

The `exec` group runs code inside sessions. Only one execution can run at a
time per session — if you try to run two concurrently, the second one gets
`SESSION_BUSY`.

### REPL

Run an inline JavaScript expression or statement. State persists across
calls — variables set in one REPL call are available in the next:

```bash
vm-system exec repl SESSION_ID 'your JavaScript here'
```

### File execution

Run a file from the session's worktree. The path must be relative and must
stay inside the worktree (no `../` traversal, no absolute paths):

```bash
vm-system exec run-file SESSION_ID path/to/file.js
```

### History and events

Every execution is recorded with its events (console output, return values,
exceptions). You can list executions for a session, inspect a specific one,
or page through its event stream:

```bash
vm-system exec list SESSION_ID [--limit 50]
vm-system exec get EXECUTION_ID
vm-system exec events EXECUTION_ID [--after-seq 0]
```

The `--after-seq` flag on `events` enables cursor-based pagination. Pass the
`seq` of the last event you've seen and you'll get only newer events.

## ops

Operational commands for checking on the running daemon:

```bash
vm-system ops health              # {"status":"ok"}
vm-system ops runtime-summary     # active session count and IDs
```

`runtime-summary` is particularly useful because it shows what's actually
alive in daemon memory. After a restart, the database still has session rows,
but `runtime-summary` correctly shows zero active sessions.

## libs

Library management for the local download cache:

```bash
vm-system libs download [--cache-dir .vm-cache/libraries]
```

This downloads every library in the built-in catalog (lodash, moment, axios,
ramda, dayjs, zustand) to the local cache. You need to do this before creating
sessions that use libraries — if a library isn't in the cache, session creation
will fail.

## See Also

- `vm-system help getting-started` — hands-on tutorial
- `vm-system help api-reference` — the REST endpoints these commands call
- `vm-system help templates-and-sessions` — deeper on the core concepts
