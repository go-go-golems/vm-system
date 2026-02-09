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

## Command tree

```
vm-system
├── serve                        Start the daemon host
├── template                     Manage templates (runtime profiles)
│   ├── create                   Create a new template
│   ├── list                     List all templates
│   ├── get                      Get template detail with settings
│   ├── delete                   Delete a template (cascades)
│   ├── add-startup              Add a startup file to a template
│   ├── list-startup             List startup files
│   ├── add-capability           Add capability metadata
│   ├── list-capabilities        List capabilities
│   ├── add-module               Add a native module
│   ├── remove-module            Remove a native module
│   ├── list-modules             List configured modules
│   ├── add-library              Add a third-party library
│   ├── remove-library           Remove a library
│   ├── list-libraries           List configured libraries
│   ├── list-available-modules   Show configurable module catalog
│   └── list-available-libraries Show downloadable library catalog
├── session                      Manage VM sessions (live runtimes)
│   ├── create                   Create session from template
│   ├── list                     List sessions (optionally by status)
│   ├── get                      Get session detail
│   └── close                    Close session (discard runtime)
├── exec                         Execute code in sessions
│   ├── repl                     Run a REPL snippet
│   ├── run-file                 Run a file from the worktree
│   ├── list                     List executions for a session
│   ├── get                      Get execution detail
│   └── events                   Get execution events
├── ops                          Operational commands
│   ├── health                   Daemon health check
│   └── runtime-summary          Active runtime state
└── libs                         Library cache management
    └── download                 Download libraries to cache
```

## Global flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--db` | string | `vm-system.db` | SQLite database path (used by `serve`) |
| `--server-url` | string | `http://127.0.0.1:3210` | Daemon base URL (used by client commands) |
| `--log-level` | string | `warn` | Logging level (debug, info, warn, error) |

## serve

```bash
vm-system serve [--listen ADDRESS]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--listen` | string | `127.0.0.1:3210` | HTTP listen address |

Starts the daemon in the foreground. Uses `--db` for SQLite storage.

## template

### template create

```bash
vm-system template create --name NAME [--engine ENGINE]
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--name` | string | (required) | Template name |
| `--engine` | string | `goja` | Engine type |

### template list / get / delete

```bash
vm-system template list
vm-system template get TEMPLATE_ID
vm-system template delete TEMPLATE_ID
```

### template add-startup / list-startup

```bash
vm-system template add-startup TEMPLATE_ID --path PATH --order N [--mode MODE]
vm-system template list-startup TEMPLATE_ID
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--path` | string | (required) | Script path relative to session worktree |
| `--order` | int | 0 | Execution order (ascending) |
| `--mode` | string | `eval` | Execution mode (`eval` only currently) |

### template add-capability / list-capabilities

```bash
vm-system template add-capability TEMPLATE_ID --name NAME [--kind KIND] [--config JSON] [--enabled]
vm-system template list-capabilities TEMPLATE_ID
```

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--name` | string | (required) | Capability name |
| `--kind` | string | `module` | Kind: module, global, fs, net, env |
| `--config` | string | `{}` | Config JSON |
| `--enabled` | bool | `true` | Enable the capability |

### template add-module / remove-module / list-modules

```bash
vm-system template add-module TEMPLATE_ID --name NAME
vm-system template remove-module TEMPLATE_ID --name NAME
vm-system template list-modules TEMPLATE_ID
```

Only modules from the configurable catalog are allowed (`MODULE_NOT_ALLOWED`
for built-ins like JSON, Math).

### template add-library / remove-library / list-libraries

```bash
vm-system template add-library TEMPLATE_ID --name NAME
vm-system template remove-library TEMPLATE_ID --name NAME
vm-system template list-libraries TEMPLATE_ID
```

### template list-available-modules / list-available-libraries

```bash
vm-system template list-available-modules     # database, exec, fs
vm-system template list-available-libraries   # lodash, moment, axios, ramda, dayjs, zustand
```

## session

### session create

```bash
vm-system session create \
  --template-id ID --workspace-id ID \
  --base-commit OID --worktree-path /absolute/path
```

All four flags required. Worktree directory must exist.

### session list / get / close

```bash
vm-system session list [--status STATUS]    # starting, ready, crashed, closed
vm-system session get SESSION_ID
vm-system session close SESSION_ID
```

## exec

### exec repl

```bash
vm-system exec repl SESSION_ID 'JAVASCRIPT_CODE'
```

### exec run-file

```bash
vm-system exec run-file SESSION_ID FILE_PATH
```

Path must be relative to worktree; no `../` traversal.

### exec list / get / events

```bash
vm-system exec list SESSION_ID [--limit N]          # default 50
vm-system exec get EXECUTION_ID
vm-system exec events EXECUTION_ID [--after-seq N]  # default 0
```

## ops

```bash
vm-system ops health              # daemon health (JSON)
vm-system ops runtime-summary     # active session count and IDs (JSON)
```

## libs

```bash
vm-system libs download [--cache-dir DIR]   # default .vm-cache/libraries
```

## See Also

- `vm-system help getting-started`
- `vm-system help api-reference`
- `vm-system help templates-and-sessions`
