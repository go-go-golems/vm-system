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

This reference covers every command in the vm-system CLI. All client-mode
commands communicate with a running daemon at the address specified by
`--server-url` (default `http://127.0.0.1:3210`).

## Global flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--db` | string | `vm-system.db` | SQLite database path (used by `serve`) |
| `--server-url` | string | `http://127.0.0.1:3210` | Daemon base URL (used by client commands) |
| `--log-level` | string | `warn` | Logging level (debug, info, warn, error) |

## serve — Start the daemon

```bash
vm-system serve [--listen ADDRESS]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--listen` | string | `127.0.0.1:3210` | HTTP listen address |

Starts the daemon process in the foreground. Uses the `--db` global flag for
SQLite storage. Stop with Ctrl+C for graceful shutdown.

## template — Manage templates

### template create

```bash
vm-system template create --name NAME [--engine ENGINE]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--name` | string | (required) | Template name |
| `--engine` | string | `goja` | Engine type |

Creates a template with default settings (limits, resolver, runtime config).

### template list

```bash
vm-system template list
```

Lists all templates with ID, name, engine, and active status.

### template get

```bash
vm-system template get TEMPLATE_ID
```

Shows full template detail: settings JSON, capabilities, startup files, modules,
and libraries.

### template delete

```bash
vm-system template delete TEMPLATE_ID
```

Deletes a template and all associated settings, capabilities, and startup files.

### template add-startup

```bash
vm-system template add-startup TEMPLATE_ID --path PATH --order N [--mode MODE]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--path` | string | (required) | Script path relative to session worktree |
| `--order` | int | 0 | Execution order (ascending) |
| `--mode` | string | `eval` | Execution mode (`eval` only currently) |

### template list-startup

```bash
vm-system template list-startup TEMPLATE_ID
```

### template add-capability

```bash
vm-system template add-capability TEMPLATE_ID --name NAME [--kind KIND] [--config JSON] [--enabled]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--name` | string | (required) | Capability name |
| `--kind` | string | `module` | Kind: module, global, fs, net, env |
| `--config` | string | `{}` | Config JSON |
| `--enabled` | bool | `true` | Enable the capability |

### template list-capabilities

```bash
vm-system template list-capabilities TEMPLATE_ID
```

### template add-module

```bash
vm-system template add-module TEMPLATE_ID --name NAME
```

Adds a native module from the configurable catalog. Built-in JS globals cannot
be added (returns `MODULE_NOT_ALLOWED`).

### template remove-module

```bash
vm-system template remove-module TEMPLATE_ID --name NAME
```

### template list-modules

```bash
vm-system template list-modules TEMPLATE_ID
```

### template add-library

```bash
vm-system template add-library TEMPLATE_ID --name NAME
```

Adds a third-party library to the template.

### template remove-library

```bash
vm-system template remove-library TEMPLATE_ID --name NAME
```

### template list-libraries

```bash
vm-system template list-libraries TEMPLATE_ID
```

### template list-available-modules

```bash
vm-system template list-available-modules
```

Shows the catalog of native modules that can be configured per template:
`database`, `exec`, `fs`.

### template list-available-libraries

```bash
vm-system template list-available-libraries
```

Shows the built-in library catalog: lodash, moment, axios, ramda, dayjs, zustand.

## session — Manage sessions

### session create

```bash
vm-system session create \
  --template-id ID \
  --workspace-id ID \
  --base-commit OID \
  --worktree-path /absolute/path
```

| Flag | Type | Description |
|------|------|-------------|
| `--template-id` | string | Template ID (required) |
| `--workspace-id` | string | Logical workspace identifier (required) |
| `--base-commit` | string | Git commit OID (required) |
| `--worktree-path` | string | Absolute worktree directory (required, must exist) |

Creates a session, allocates goja runtime, loads libraries and startup files.

### session list

```bash
vm-system session list [--status STATUS]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--status` | string | (all) | Filter: starting, ready, crashed, closed |

### session get

```bash
vm-system session get SESSION_ID
```

Shows full session detail including timestamps, status, and last error.

### session close

```bash
vm-system session close SESSION_ID
```

Discards the in-memory runtime and marks the session closed.

## exec — Execute code

### exec repl

```bash
vm-system exec repl SESSION_ID 'JAVASCRIPT_CODE'
```

Executes a JavaScript snippet in the session runtime. Returns the execution
record with captured events.

### exec run-file

```bash
vm-system exec run-file SESSION_ID FILE_PATH
```

Executes a file from the session worktree. Path must be relative and within the
worktree boundary.

### exec list

```bash
vm-system exec list SESSION_ID [--limit N]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--limit` | int | 50 | Maximum executions to return |

### exec get

```bash
vm-system exec get EXECUTION_ID
```

Shows execution detail: kind, status, input/path, result, timing.

### exec events

```bash
vm-system exec events EXECUTION_ID [--after-seq N]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--after-seq` | int | 0 | Return events after this sequence number |

## ops — Operational commands

### ops health

```bash
vm-system ops health
```

Returns daemon health status as JSON.

### ops runtime-summary

```bash
vm-system ops runtime-summary
```

Returns active session count and IDs as JSON.

## libs — Library management

### libs download

```bash
vm-system libs download [--cache-dir DIR]
```

Downloads libraries from the built-in catalog into the local cache
(default `.vm-cache/libraries`).

## Common patterns

### Full loop from scratch

```bash
./vm-system serve --db /tmp/demo.db --listen 127.0.0.1:3210 &
TEMPLATE=$(./vm-system template create --name demo | grep -oP 'ID: \K[^ )]+')
./vm-system template add-startup "$TEMPLATE" --path init.js --order 10 --mode eval
SESSION=$(./vm-system session create --template-id "$TEMPLATE" \
  --workspace-id ws --base-commit dead --worktree-path /tmp/demo-ws \
  | awk '/Created session:/ {print $3}')
./vm-system exec repl "$SESSION" '1+1'
./vm-system session close "$SESSION"
```

### Inspect execution history

```bash
vm-system exec list <session-id> --limit 20
vm-system exec get <execution-id>
vm-system exec events <execution-id> --after-seq 0
```

### Direct API verification

```bash
curl -sS http://127.0.0.1:3210/api/v1/health
curl -sS http://127.0.0.1:3210/api/v1/runtime/summary
curl -sS http://127.0.0.1:3210/api/v1/templates
curl -sS "http://127.0.0.1:3210/api/v1/sessions?status=ready"
```

## Troubleshooting

| Problem | Cause | Solution |
|---------|-------|----------|
| All client commands fail with connection error | Daemon not running | Start `vm-system serve` first |
| Wrong daemon contacted | `--server-url` mismatch | Pass explicit `--server-url http://host:port` |
| `MODULE_NOT_ALLOWED` on add-module | Tried to add a built-in (JSON, Math) | Only catalog modules are configurable; check `list-available-modules` |
| Template ID not recognized | Copy-paste error or template was deleted | Verify with `template list` |

## See Also

- `vm-system help getting-started`
- `vm-system help api-reference`
- `vm-system help templates-and-sessions`
