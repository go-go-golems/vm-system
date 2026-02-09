---
Title: "Getting Started with vm-system"
Slug: getting-started
Short: "Build, run the daemon, create your first template, session, and execution in under five minutes."
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
IsTopLevel: true
IsTemplate: false
ShowPerDefault: true
SectionType: Tutorial
---

This tutorial walks you through building vm-system, starting the daemon, and
completing a full runtime loop: template → session → execution. By the end you
will have run JavaScript inside a managed goja runtime and inspected the result
through the REST API.

## Prerequisites

You need:

- Go toolchain (compatible with `go 1.25.5`, see `go.mod`)
- A Unix shell (bash or zsh)
- `curl` (for optional direct API verification)

## Step 1 — Build the binary

From the repository root:

```bash
GOWORK=off go build -o vm-system ./cmd/vm-system
./vm-system --help
```

You should see the command groups `serve`, `template`, `session`, `exec`, `ops`,
and `libs`. If the build fails, check your Go version and module path.

## Step 2 — Prepare a scratch workspace

vm-system sessions need a worktree directory on disk. Create a throwaway one:

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

The `init.js` file will run during session startup. `app.js` is a script you
will execute later with `exec run-file`.

## Step 3 — Start the daemon

Use a dedicated database file so you do not pollute anything:

```bash
./vm-system serve --db /tmp/vm-scratch.db --listen 127.0.0.1:3210
```

The daemon stays in the foreground and prints a listening message. Open a second
terminal for the remaining commands.

Verify health:

```bash
curl -sS http://127.0.0.1:3210/api/v1/health
# {"status":"ok"}
```

## Step 4 — Create a template

A template is a persistent runtime profile — it defines engine type, resource
limits, startup files, modules, and libraries.

```bash
./vm-system template create --name my-first --engine goja
```

Capture the template ID from the output (e.g. `Created template: my-first (ID: <template-id>)`).

Add a startup file to the template:

```bash
./vm-system template add-startup <template-id> \
  --path runtime/init.js --order 10 --mode eval
```

Inspect the template to confirm:

```bash
./vm-system template get <template-id>
```

You should see the settings JSON, the startup file entry, and any modules or
libraries (none yet).

## Step 5 — Create a session

A session is a long-lived goja runtime bound to a template and a worktree:

```bash
./vm-system session create \
  --template-id <template-id> \
  --workspace-id ws-demo \
  --base-commit deadbeef \
  --worktree-path /tmp/vm-scratch
```

The output shows the session ID and status `ready`. Internally the daemon
allocated a goja runtime, injected the console shim, and executed `init.js`.

## Step 6 — Execute code

### REPL snippet

```bash
./vm-system exec repl <session-id> 'seed + 2'
```

The result should be `42` — the startup script set `globalThis.seed = 40`.

### Run a file

```bash
./vm-system exec run-file <session-id> app.js
```

The path is resolved relative to the session worktree. Absolute paths and
`../` traversal are rejected.

## Step 7 — Inspect state and clean up

List sessions and check the runtime summary:

```bash
./vm-system session list
./vm-system ops runtime-summary
```

Close the session and stop the daemon:

```bash
./vm-system session close <session-id>
```

Then press Ctrl+C in the daemon terminal.

## What just happened

The request flow you exercised is:

```
CLI command → vmclient REST call → HTTP handler → vmcontrol service → runtime/store
```

The daemon owns active goja runtimes in memory. CLI commands are thin REST
clients. SQLite stores templates, sessions, executions, and events durably.

## Next steps

- Read `vm-system help architecture` for the package layout and design reasoning.
- Read `vm-system help api-reference` for complete endpoint contracts.
- Read `vm-system help templates-and-sessions` for deeper template/session semantics.
- Read `vm-system help troubleshooting` for common failure modes and fixes.

## Troubleshooting

| Problem | Cause | Solution |
|---------|-------|----------|
| `curl /api/v1/health` fails | Daemon not running or port mismatch | Verify daemon output and `--server-url` flag |
| Session create fails with worktree error | Directory does not exist | Create the directory first; use absolute path |
| `INVALID_PATH` on run-file | Path is absolute or escapes worktree | Pass a relative path without `../` |
| `SESSION_BUSY` | Another execution is in progress | Wait for it to finish; one execution at a time per session |

## See Also

- `vm-system help architecture`
- `vm-system help api-reference`
- `vm-system help templates-and-sessions`
- `vm-system help cli-command-reference`
