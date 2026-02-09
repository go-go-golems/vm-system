# vm-system

`vm-system` is a daemon-first JavaScript runtime service built on goja.

## What It Provides

- `template` management API/CLI for runtime policy and startup configuration.
- Long-lived `session` runtime ownership inside a daemon process.
- `exec` APIs for REPL and run-file execution with persisted event logs.
- A reusable orchestration core in `pkg/vmcontrol` shared by daemon + transports.

## Architecture (v2)

- `pkg/vmcontrol`: transport-agnostic orchestration core (templates, sessions, executions, runtime registry).
- `pkg/vmdaemon`: process host/lifecycle wrapper.
- `pkg/vmtransport/http`: REST adapter under `/api/v1`.
- `pkg/vmclient`: shared REST client used by CLI commands.

## Build

```bash
GOWORK=off go build -o vm-system ./cmd/vm-system
```

## Quick Start (Daemon-First)

```bash
# 1) Start daemon
./vm-system serve --db vm-system.db --listen 127.0.0.1:3210

# 2) In another terminal, create a template
./vm-system --server-url http://127.0.0.1:3210 template create --name demo --engine goja

# 3) Create a session from that template
./vm-system --server-url http://127.0.0.1:3210 session create \
  --template-id <template-id> \
  --workspace-id ws-1 \
  --base-commit deadbeef \
  --worktree-path /abs/path/to/worktree

# 4) Execute code
./vm-system --server-url http://127.0.0.1:3210 exec repl <session-id> '1+2'

# 5) Inspect daemon state and close the session
./vm-system --server-url http://127.0.0.1:3210 ops runtime-summary
./vm-system --server-url http://127.0.0.1:3210 session close <session-id>
```

## Developer Guide

For a full onboarding and contribution guide (first VM run, architecture deep dive,
API contracts, test coverage, and contribution workflow), see:

- `docs/getting-started-from-first-vm-to-contributor-guide.md`

## API Endpoints (Current)

- `GET /api/v1/health`
- `GET /api/v1/runtime/summary`
- `GET/POST /api/v1/templates`
- `GET/DELETE /api/v1/templates/{template_id}`
- `GET/POST /api/v1/templates/{template_id}/capabilities`
- `GET/POST /api/v1/templates/{template_id}/startup-files`
- `GET/POST /api/v1/sessions`
- `GET /api/v1/sessions/{session_id}`
- `POST /api/v1/sessions/{session_id}/close`
- `DELETE /api/v1/sessions/{session_id}`
- `GET /api/v1/executions`
- `POST /api/v1/executions/repl`
- `POST /api/v1/executions/run-file`
- `GET /api/v1/executions/{execution_id}`
- `GET /api/v1/executions/{execution_id}/events`

## Tests

```bash
GOWORK=off go test ./...
bash ./smoke-test.sh
bash ./test-e2e.sh
```
