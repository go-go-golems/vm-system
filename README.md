# vm-system

A daemon-first JavaScript runtime service built on [goja](https://github.com/nicholasgasior/goja).
Define runtime profiles as **templates**, spin up long-lived **sessions**, and
execute JavaScript through **REPL snippets** or **file runs** — all managed
through a REST API and CLI. In merged-repo mode, the same daemon can also
serve the web UI when frontend assets are generated.

```bash
./vm-system serve &
./vm-system template create --name demo --engine goja
./vm-system session create --template-id <id> --workspace-id ws --base-commit HEAD --worktree-path /my/code
./vm-system exec repl <session-id> '1 + 1'   # → 2
```

## Why vm-system?

- **Templates** define what a runtime can do: engine, resource limits, startup
  scripts, native modules (database, fs, exec), and third-party libraries
  (lodash, moment, etc.).
- **Sessions** are live goja runtimes that hold state across executions —
  variables set in one REPL call persist in the next.
- **Executions** capture everything: console output, return values, exceptions,
  all as a queryable event stream with cursor-based pagination.
- **One binary, one process** — the daemon hosts runtimes in memory, persists
  everything to SQLite, and serves a REST API that the CLI consumes.

## Quick start

```bash
# Build
GOWORK=off go build -o vm-system ./cmd/vm-system

# Optional: build with embedded web UI assets
go generate ./internal/web
GOWORK=off go build -tags embed -o vm-system ./cmd/vm-system

# Start daemon (foreground — use a second terminal for the rest)
./vm-system serve --db /tmp/demo.db --listen 127.0.0.1:3210

# Create a template, add a startup script, start a session
./vm-system template create --name demo --engine goja
./vm-system template add-startup <template-id> --path runtime/init.js --order 10 --mode eval
./vm-system session create \
  --template-id <template-id> \
  --workspace-id ws-1 \
  --base-commit deadbeef \
  --worktree-path /abs/path/to/worktree

# Execute code
./vm-system exec repl <session-id> 'seed + 2'
./vm-system exec run-file <session-id> app.js

# Inspect and clean up
./vm-system ops runtime-summary
./vm-system session close <session-id>
```

## Web UI (merged repo)

```bash
# terminal 1: daemon API
make dev-backend

# terminal 2: Vite dev server (proxies /api/v1 to daemon)
make dev-frontend
```

Open `http://127.0.0.1:3000`.

## Architecture

```
CLI (template/session/exec/ops)
  │  REST via pkg/vmclient
  ▼
pkg/vmtransport/http ── REST routes, validation, error envelopes
  │
  ▼
pkg/vmcontrol ────────── Templates, Sessions, Executions (transport-agnostic)
  │
  ├─► pkg/vmsession ──── In-memory goja runtimes + execution lock
  ├─► pkg/vmexec ──────── Execution pipeline + event capture
  └─► pkg/vmstore ─────── SQLite persistence
```

`pkg/vmdaemon` wraps the above into a single long-lived process with graceful
shutdown.

## API

All endpoints live under `/api/v1`. Full contracts are in `vm-system help api-reference`.

| Resource | Endpoints |
|----------|-----------|
| Health | `GET /api/v1/health` |
| Runtime | `GET /api/v1/runtime/summary` |
| Templates | `GET/POST /api/v1/templates`, `GET/DELETE /api/v1/templates/{id}` |
| ↳ Capabilities | `GET/POST /api/v1/templates/{id}/capabilities` |
| ↳ Startup files | `GET/POST /api/v1/templates/{id}/startup-files` |
| ↳ Modules | `GET/POST /api/v1/templates/{id}/modules`, `DELETE .../modules/{name}` |
| ↳ Libraries | `GET/POST /api/v1/templates/{id}/libraries`, `DELETE .../libraries/{name}` |
| Sessions | `GET/POST /api/v1/sessions`, `GET /api/v1/sessions/{id}`, `POST .../close`, `DELETE` |
| Executions | `GET /api/v1/executions`, `POST .../repl`, `POST .../run-file` |
| ↳ Detail | `GET /api/v1/executions/{id}`, `GET .../events` |

## Tests

```bash
GOWORK=off go test ./...    # unit + integration (real stack, no mocks)
bash ./test/scripts/smoke-test.sh        # fast daemon sanity check (~10s)
bash ./test/scripts/test-e2e.sh          # full CLI loop
bash ./test/scripts/test-library-matrix.sh  # module/library capability semantics
bash ./test/scripts/test-all.sh          # everything
pnpm -C ui check                         # frontend type-check
pnpm -C ui run build                     # frontend production build
```

## Documentation

All docs are built into the binary as Glazed help pages:

```bash
vm-system help                         # list all topics
vm-system help getting-started         # first run + daily workflows
vm-system help architecture            # package layout and design
vm-system help templates-and-sessions  # core domain concepts
vm-system help api-reference           # REST endpoint contracts
vm-system help cli-command-reference   # every command with flags
vm-system help examples                # runnable recipes
vm-system help contributing            # how to contribute + testing
```
