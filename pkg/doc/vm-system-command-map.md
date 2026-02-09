---
Title: "Command Map"
Slug: command-map
Short: "Visual topology of all vm-system commands, subcommands, and their relationships."
Topics:
- vm-system
- commands
- reference
IsTopLevel: true
IsTemplate: false
ShowPerDefault: false
SectionType: GeneralTopic
---

This page shows the full command tree of vm-system with brief descriptions.
Use it as a quick lookup when you cannot remember a subcommand name.

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

All commands accept:

- `--db PATH` — SQLite database path (used by `serve`, default `vm-system.db`)
- `--server-url URL` — Daemon address (used by client commands, default `http://127.0.0.1:3210`)
- `--log-level LEVEL` — Logging verbosity (debug, info, warn, error)

## Command groups by role

### For users / operators

```bash
vm-system serve                  # Start daemon
vm-system ops health             # Check daemon
vm-system ops runtime-summary    # Active session count
```

### For runtime configuration

```bash
vm-system template create ...    # Define runtime profile
vm-system template add-*  ...   # Configure modules/libraries/startup
vm-system libs download          # Populate library cache
```

### For execution

```bash
vm-system session create ...     # Start a live runtime
vm-system exec repl ...          # Run snippets
vm-system exec run-file ...      # Run files
vm-system session close ...      # Stop runtime
```

### For inspection

```bash
vm-system template get ...       # Template detail
vm-system session get ...        # Session detail
vm-system exec get ...           # Execution detail
vm-system exec events ...        # Execution event stream
vm-system exec list ...          # Execution history
```

## See Also

- `vm-system help cli-command-reference`
- `vm-system help how-to-use`
- `vm-system help getting-started`
