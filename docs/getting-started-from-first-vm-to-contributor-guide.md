
# Getting started from first VM to contributor workflow

## Overview

This guide is for a developer who has never touched `vm-system` and wants to move from:

1. Running the first daemon and first VM session.
2. Understanding how the runtime actually works under the hood.
3. Understanding what is and is not currently covered by automated tests.
4. Contributing safely with a repeatable implementation and review workflow.

The guide is intentionally practical. Every major concept is tied back to concrete commands, source files, and behavior that exists in the current codebase.

At a high level, `vm-system` is a daemon-first JavaScript runtime service built on goja with:

- template management APIs (`template` command group)
- long-lived runtime sessions (`session` command group)
- code execution APIs (`exec` command group)
- SQLite-backed persistence (`pkg/vmstore`)
- a transport-agnostic core (`pkg/vmcontrol`) shared by daemon and API transports

If you only need the first successful run, jump to the “Step-by-Step Guide” section. If you are here to contribute code, read the full guide in order once.

## Prerequisites

### Toolchain and environment

- Go toolchain compatible with `go 1.25.5` (see `go.mod`)
- A Unix-like shell (`bash`/`zsh`) for scripts and examples
- `curl` for direct API verification
- `python3` only if you run scripts that dynamically allocate ports (for example `smoke-test.sh`, `test-e2e.sh`)

### Repository location used in this guide

All commands assume you are at:

```bash
/home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system
```

If your local path differs, commands still apply, but absolute path examples will differ.

### Core concepts you should know before coding

- `Template` in the API maps to persisted VM profile/configuration.
- `Session` is a long-lived in-memory goja runtime plus a persisted session record.
- `Execution` is a single REPL or run-file invocation in a session.
- CLI runtime commands call daemon REST endpoints via `pkg/vmclient`.

## Step-by-Step Guide

## Step 0: Build a mental model before typing commands

The shortest path to success is to understand ownership boundaries first.

- `vm-system serve` starts one long-lived process.
- That process owns active session runtimes in memory.
- CLI commands like `template create`, `session create`, `exec repl` are client calls to that daemon.
- SQLite stores durable records for templates, sessions, executions, and events.

Minimal request path:

```text
CLI command -> vmclient REST call -> HTTP handler -> vmcontrol service -> runtime/store
```

This matters because if the daemon stops, in-memory sessions are gone even though session rows still exist in the DB.

## Step 1: Build the binary

From repo root:

```bash
GOWORK=off go build -o vm-system ./cmd/vm-system
```

Verify help output:

```bash
./vm-system --help
```

You should see command groups:

- `serve`
- `template`
- `session`
- `exec`
- `modules`
- `libs`

For daily backend work, `serve`, `template`, `session`, and `exec` are the main path.

## Step 2: Create a clean scratch workspace

Create a throwaway workspace to avoid polluting project files:

```bash
mkdir -p /tmp/vm-system-guide-worktree/runtime
cat > /tmp/vm-system-guide-worktree/runtime/init.js <<'JS'
console.log("startup init loaded")
globalThis.answerSeed = 40
JS

cat > /tmp/vm-system-guide-worktree/app.js <<'JS'
console.log("app.js executing")
answerSeed + 2
JS
```

Why this matters:

- startup files execute during session creation
- run-file execution must stay within worktree boundaries
- the path normalization logic rejects traversal

## Step 3: Start daemon process

Use a dedicated database for the run:

```bash
./vm-system serve --db /tmp/vm-system-guide.db --listen 127.0.0.1:3210
```

Expected behavior:

- daemon prints `vm-system daemon listening on 127.0.0.1:3210`
- process stays in foreground

In a second terminal, verify health:

```bash
curl -sS http://127.0.0.1:3210/api/v1/health
```

Expected response:

```json
{"status":"ok"}
```

## Step 4: Create first template

Use CLI in client mode (default `--server-url` already targets `127.0.0.1:3210`):

```bash
./vm-system template create --name guide-template --engine goja
```

Typical output:

```text
Created template: guide-template (ID: <template-id>)
```

Capture template ID for next commands.

### What happened internally

- CLI called `vmclient.CreateTemplate` (`pkg/vmclient/templates_client.go`)
- HTTP handler validated payload and called `core.Templates.Create`
- `TemplateService.Create` created VM row plus default `VMSettings`:
  - CPU, wall, memory, output/event limits
  - resolver defaults
  - runtime defaults

The default limits are created in `pkg/vmcontrol/template_service.go`.

## Step 5: Add startup and capability policy

Add startup file:

```bash
./vm-system template add-startup <template-id> --path runtime/init.js --order 10 --mode eval
```

Add capability metadata:

```bash
./vm-system template add-capability <template-id> --kind module --name console --enabled
```

Inspect template:

```bash
./vm-system template get <template-id>
```

You should see:

- settings JSON blobs
- one startup file (`runtime/init.js`)
- capability entry (`module:console`)

Important reality check:

- capability metadata is persisted and exposed through API
- runtime enforcement of capability metadata is not yet comprehensive
- do not assume capability rows imply strict deny-by-default execution policy

## Step 6: Create first session

```bash
./vm-system session create \
  --template-id <template-id> \
  --workspace-id ws-guide \
  --base-commit deadbeef \
  --worktree-path /tmp/vm-system-guide-worktree
```

Expected output includes:

- session ID
- status `ready`
- VM/template ID
- workspace/base-commit/worktree path

### What happened internally

- `SessionService.Create` calls runtime manager `CreateSession`
- `SessionManager` verifies template + settings + worktree path
- goja runtime is allocated
- console shim is set when runtime config enables console
- libraries configured on template are loaded if present in `.vm-cache/libraries`
- startup files are executed in order (`order_index`)
- session status transitions from `starting` to `ready`
- persisted session row is updated

If startup fails, status is set to `crashed` and `last_error` is persisted.

## Step 7: Execute a REPL snippet

```bash
./vm-system exec repl <session-id> 'answerSeed + 2'
```

Typical output:

- execution ID
- status
- result payload
- events list (console/input/value/exception depending on run)

Internally (`pkg/vmexec/executor.go`):

- session lock is acquired with `TryLock` to enforce one execution at a time
- execution row is created with status `running`
- input echo and console/value/exception events are recorded
- execution row is finalized with `ok` or `error`
- `ExecutionService` optionally enforces limits via event scan

## Step 8: Execute a file

```bash
./vm-system exec run-file <session-id> app.js
```

Notes:

- path is normalized in `ExecutionService.normalizeRunFilePath`
- absolute paths and traversal (`../`) are rejected with `INVALID_PATH`
- executor reads file from session worktree and runs it in goja runtime

Try an intentionally invalid path to confirm protection:

```bash
curl -sS -X POST http://127.0.0.1:3210/api/v1/executions/run-file \
  -H 'Content-Type: application/json' \
  -d '{"session_id":"<session-id>","path":"../etc/passwd"}'
```

Expected error envelope:

```json
{"error":{"code":"INVALID_PATH","message":"Path escapes allowed worktree",...}}
```

## Step 9: Inspect session and execution history

List sessions:

```bash
./vm-system session list
./vm-system session list --status ready
```

List executions:

```bash
./vm-system exec list <session-id> --limit 50
```

Inspect one execution:

```bash
./vm-system exec get <execution-id>
./vm-system exec events <execution-id> --after-seq 0
```

Runtime summary for operations view:

```bash
curl -sS http://127.0.0.1:3210/api/v1/runtime/summary
```

Response shape:

```json
{"active_sessions":1,"active_session_ids":["..."]}
```

## Step 10: Close session and stop daemon

```bash
./vm-system session delete <session-id>
```

Or raw close endpoint:

```bash
curl -sS -X POST http://127.0.0.1:3210/api/v1/sessions/<session-id>/close -d '{}'
```

Then stop daemon with `Ctrl+C`.

At this point you have successfully completed the first full runtime loop.

## Architecture deep dive

This section explains how the implementation is structured so you can change code confidently.

## 1) Source map and package responsibilities

### Command layer (`cmd/vm-system`)

- `main.go`
  - root command
  - global flags `--db` and `--server-url`
  - registers command groups
- `cmd_serve.go`
  - daemon bootstrap command
  - wires `vmdaemon` + HTTP handler
- `cmd_template.go`
  - template CRUD + startup/capability operations
- `cmd_session.go`
  - create/list/get/close semantics
- `cmd_exec.go`
  - repl/run-file/get/list/events
- `cmd_modules.go`, `cmd_libs.go`
  - utility/legacy support for modules/libraries

### Core orchestration (`pkg/vmcontrol`)

- `core.go`
  - composition root for services
- `template_service.go`
  - template lifecycle and settings policy defaults
- `session_service.go`
  - session lifecycle orchestration
- `execution_service.go`
  - run-file path safety and limit checks
- `runtime_registry.go`
  - active runtime summary exposure
- `ports.go`
  - interfaces separating domain from adapters

This layer is transport-agnostic. No HTTP request objects should leak here.

### Daemon host (`pkg/vmdaemon`)

- wraps lifecycle around core + store + HTTP server
- owns process timeouts and graceful shutdown behavior

### HTTP transport (`pkg/vmtransport/http`)

- REST route registration
- JSON DTO decode/validation
- status/error envelope mapping
- request-id middleware

### Client adapter (`pkg/vmclient`)

- typed wrapper used by CLI commands
- one central `do()` helper for JSON request/response
- standardized API error parsing

### Runtime + persistence adapters

- `pkg/vmsession`
  - in-memory active sessions and goja runtime ownership
- `pkg/vmexec`
  - execution pipeline and event persistence
- `pkg/vmstore`
  - SQLite schema + query methods

## 2) Request lifecycle walkthroughs

### Template create (`template create`)

1. CLI command parses flags and calls `vmclient.CreateTemplate`.
2. Client POSTs `/api/v1/templates`.
3. HTTP handler validates `name` (and optional `engine`).
4. Core `TemplateService.Create` persists VM row.
5. Core writes default settings row (`vm_settings`).
6. API returns created template JSON.
7. CLI prints `Created template: ...`.

### Session create (`session create`)

1. CLI POSTs `/api/v1/sessions` with template/workspace/commit/worktree.
2. Handler validates all required fields.
3. Core `SessionService.Create` calls runtime manager.
4. Runtime manager allocates goja runtime, loads startup/libraries.
5. Status transitions `starting -> ready` or `crashed`.
6. Session row is returned to client.

### Execution (`exec repl` / `exec run-file`)

1. CLI POSTs execution request.
2. Handler validates request shape.
3. Core execution service may sanitize/normalize path (run-file).
4. Runtime executor acquires session execution lock.
5. Execution row + events are persisted.
6. Core may enforce output/event limits.
7. Response returns execution summary.

## 3) Data model and schema tour

SQLite schema lives in `pkg/vmstore/vmstore.go` inside `initSchema()`.

Key tables:

- `vm`
  - template identity and high-level profile
- `vm_settings`
  - limits/resolver/runtime JSON config
- `vm_capability`
  - capability metadata rows
- `vm_startup_file`
  - startup policy ordered by `order_index`
- `vm_session`
  - durable session records and status history
- `execution`
  - execution summaries and result/error blobs
- `execution_event`
  - event stream by (`execution_id`, `seq`)

Design implication:

- in-memory runtime state is not reconstructed automatically from DB rows today
- daemon restart recovery semantics are still a known gap

## 4) Error contract and API behavior

`writeCoreError` in `pkg/vmtransport/http/server.go` maps domain errors to status codes.

Current mapping highlights:

- `ErrVMNotFound` -> `404 TEMPLATE_NOT_FOUND`
- `ErrSessionNotFound` -> `404 SESSION_NOT_FOUND`
- `ErrSessionNotReady` -> `409 SESSION_NOT_READY`
- `ErrSessionBusy` -> `409 SESSION_BUSY`
- `ErrPathTraversal` -> `422 INVALID_PATH`
- `ErrOutputLimitExceeded` -> `422 OUTPUT_LIMIT_EXCEEDED`
- `ErrFileNotFound` -> `404 FILE_NOT_FOUND`
- fallback -> `500 INTERNAL`

Important nuance:

- `GET /api/v1/executions/{id}` for missing execution currently falls through to `500 INTERNAL` in integration tests, not a dedicated not-found contract.

Response envelope format:

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "...",
    "details": {"session_id":"..."}
  }
}
```

## 5) Runtime lifecycle and concurrency

Session states in model:

- `starting`
- `ready`
- `crashed`
- `closed`

Execution concurrency:

- one execution at a time per session runtime via `ExecutionLock.TryLock()`
- concurrent request while running returns `SESSION_BUSY`

Why this matters for contributors:

- any change that touches execution flow must preserve lock semantics and deterministic status transitions
- adding async features requires explicit policy for lock scope and event ordering

## 6) Startup files and libraries behavior

Startup files:

- fetched per template from `vm_startup_file`
- resolved relative to session worktree
- executed in ascending `order_index`
- `mode=eval` and `mode=import` currently both execute with runtime `RunString`

Libraries:

- downloaded by `libs download` into `.vm-cache/libraries` by default
- runtime loader currently expects `<cache>/<lib>.js`
- downloader writes versioned filenames `<lib>-<version>.js`

Practical consequence:

- there is a known mismatch risk in library loading path conventions
- library integration tests and scripts should be treated as partially legacy until normalized

## 7) CLI/API cutover status

Current CLI runtime commands are daemon-oriented (`template`, `session`, `exec`), which is good.

Still present:

- `modules add-module` and `modules add-library` mutate DB directly and use `--vm-id` terminology
- this is partly legacy vocabulary and bypasses API-level validation paths

As a contributor, avoid expanding legacy direct-DB command paths unless intentionally maintaining backward compatibility.

## Implementation walkthrough: concrete files to read in order

If you want to understand the full implementation fast, read these in sequence.

1. `cmd/vm-system/main.go`
2. `cmd/vm-system/cmd_serve.go`
3. `pkg/vmdaemon/app.go`
4. `pkg/vmcontrol/core.go`
5. `pkg/vmtransport/http/server.go`
6. `pkg/vmcontrol/template_service.go`
7. `pkg/vmcontrol/session_service.go`
8. `pkg/vmcontrol/execution_service.go`
9. `pkg/vmsession/session.go`
10. `pkg/vmexec/executor.go`
11. `pkg/vmstore/vmstore.go`

Use this read order for a new teammate and they will understand 90% of runtime behavior with minimal backtracking.

## Testing and quality: what is covered and what is not

## Test inventory

### Integration tests (Go, HTTP level)

Located in `pkg/vmtransport/http`:

- `server_integration_test.go`
  - continuity across separate API calls in one session runtime
- `server_templates_integration_test.go`
  - template create/list/get/delete plus startup/capability nested resources
- `server_sessions_integration_test.go`
  - create/list/get/filter/close/delete and runtime summary transitions
- `server_executions_integration_test.go`
  - repl/run-file/get/list/events and `after_seq` filtering
- `server_error_contracts_integration_test.go`
  - validation/not-found/conflict/unprocessable contracts
- `server_safety_integration_test.go`
  - path traversal rejection and output/event limit exceeded behavior

These tests run a real in-memory stack:

- SQLite store
- vmcontrol core
- httptest server

No mocks are required for core endpoint behavior.

### Smoke script

`smoke-test.sh` validates:

- build success
- daemon health
- template create
- capability and startup add
- session create
- repl and run-file execution
- runtime summary check

It uses:

- isolated temp worktree
- unique DB path
- dynamic free TCP port

This script is good for fast confidence before committing.

### E2E script

`test-e2e.sh` validates a richer full loop:

- build
- daemon startup
- workspace + startup files
- template create + startup registration
- session create
- repl and run-file execution
- session and execution listing
- runtime summary

It is also parallel-safe via dynamic temp resources.

## Does current e2e/integration coverage exercise all capabilities?

Short answer: no, but it covers the highest-risk operational path well.

Strong coverage today:

- template/session/execution endpoint families
- major error contracts (400/404/409/422)
- path traversal handling
- output/event limit enforcement
- runtime summary transitions with close/delete
- script-level daemon-first happy path

Still weak or missing:

- daemon restart recovery semantics
- metadata endpoint family (not implemented)
- load/performance/concurrency beyond single-session lock tests
- fuzz/property testing for JSON payload edge cases
- deep library loading compatibility and path/version consistency
- durability semantics under abrupt process crash

Practical interpretation for contributors:

- if you change endpoint behavior, existing tests will likely catch regressions
- if you change daemon lifecycle/recovery, you must add new tests because current suite is thin there

## Recommended local validation workflow

For normal backend changes:

```bash
GOWORK=off go test ./pkg/vmtransport/http -count=1
GOWORK=off go test ./... -count=1
bash ./smoke-test.sh
```

For end-to-end confidence before opening PR:

```bash
bash ./test-e2e.sh
```

For focused test loops:

```bash
GOWORK=off go test ./pkg/vmtransport/http -run TestSessionLifecycleEndpoints -count=1
GOWORK=off go test ./pkg/vmtransport/http -run TestExecutionEndpointsLifecycle -count=1
```

## Contribution playbook

This section gives a practical end-to-end approach for making changes safely.

## 1) Start from a ticketed plan

For non-trivial changes, use ticketed docs in `ttmp/...`:

- define tasks first
- implement one task at a time
- keep changelog + diary updated
- commit per task boundary where practical

Why:

- this project already uses docmgr and ticket workspaces heavily
- consistent ticket records reduce review ambiguity

## 2) Choose the right layer first

Before coding, ask which layer owns the behavior.

If you are changing:

- HTTP shape/status -> `pkg/vmtransport/http`
- domain orchestration/policy -> `pkg/vmcontrol`
- runtime semantics/goja execution -> `pkg/vmsession` or `pkg/vmexec`
- persistence schema/query behavior -> `pkg/vmstore`
- CLI UX/output -> `cmd/vm-system` and maybe `pkg/vmclient`

Do not mix multiple ownership concerns in one file unless the change is deliberately cross-cutting.

## 3) Standard change patterns

### Adding a new API endpoint

1. Add handler route in `server.go`.
2. Add request DTO and validation.
3. Add core service method if needed.
4. Add vmclient wrapper method.
5. Add CLI command or extend command output if required.
6. Add integration test with status and envelope assertions.

### Adding a new execution policy check

1. Put policy logic in `ExecutionService` if it is domain-level.
2. Keep executor focused on runtime operation and event emission.
3. Map domain errors through `writeCoreError`.
4. Add focused integration tests in error/safety suites.

### Changing session lifecycle behavior

1. Update `SessionManager` transitions.
2. Ensure store row updates stay consistent.
3. Add runtime summary assertions.
4. Verify close/delete semantics remain deterministic.

## 4) Code review checklist

Before asking for review, verify:

- feature behavior has deterministic tests
- status code + error code + message contract remains intentional
- no hidden coupling was introduced between transport and domain layers
- CLI output is still readable and stable
- docs/diary/changelog are updated if this is a ticketed task

If your change impacts user-visible API behavior, include before/after examples in the PR description.

## 5) Git workflow checklist

Use clean, auditable commits.

- inspect with `git status --short`
- inspect diff with `git diff` and `git diff --staged`
- stage only intentional files
- use commit messages with clear scope prefix where possible

Example commit sequence:

1. `docs(getting-started): add full onboarding and architecture guide`
2. `docs(ticket): update VM-005 tasks diary and changelog`

## 6) Debugging playbook

When behavior is unclear, use this sequence.

### API-level debugging

- reproduce with `curl`
- capture status code and full JSON envelope
- compare with `writeCoreError` mapping

### Runtime/session debugging

- list sessions from API
- inspect `runtime/summary`
- verify session status transitions in DB/API output

### Execution/event debugging

- run `exec repl` with a minimal snippet
- fetch events with `after_seq=0` and then with filters
- inspect event types and payload ordering

### Persistence debugging

- use temporary DB paths per run
- avoid reusing stale DB while debugging schema/flow changes

### Script debugging

- run `smoke-test.sh` first for quick signal
- run `test-e2e.sh` for full command loop
- if script passes but integration test fails (or opposite), isolate layer mismatch

## 7) High-value first contributions for new developers

If you want a meaningful first PR after reading this guide, pick one:

1. Normalize execution-not-found behavior to dedicated 404 contract instead of generic `INTERNAL`.
2. Add daemon restart recovery integration tests and explicit policy behavior.
3. Unify library cache filename expectations between downloader and runtime loader.
4. Add metadata endpoint family with full integration coverage.
5. Add structured metrics counters around execution and session lifecycle.

All five are practical, bounded, and directly improve reliability.

## Troubleshooting

## Daemon does not become healthy

Symptoms:

- `curl /api/v1/health` fails
- CLI commands return transport errors

Checks:

1. confirm daemon is running in terminal and listening on expected address
2. verify `--server-url` matches listen host/port
3. ensure no port conflict
4. check daemon stderr output for startup errors

## Session create fails with worktree error

Symptoms:

- API/CLI error indicates worktree path does not exist

Fix:

- create the directory before session creation
- pass absolute path for clarity

## `SESSION_BUSY` on execution

Cause:

- one execution is already holding lock for that session

Fix:

- retry after current execution completes
- avoid concurrent `exec` calls on same session unless expected

## `INVALID_PATH` on run-file

Cause:

- provided path is absolute or escapes worktree

Fix:

- pass path relative to session worktree
- avoid `../` segments

## `OUTPUT_LIMIT_EXCEEDED`

Cause:

- configured template limits exceeded by event count/payload bytes

Fix:

- reduce output volume
- adjust template limits if policy allows

## Library load failures

Symptoms:

- session creation fails when template has libraries configured
- loader cannot locate cached library file

Checks:

1. run `./vm-system libs download`
2. inspect `.vm-cache/libraries`
3. verify filename expectations match loader behavior

## Verification

Use this final checklist to confirm you completed the tutorial and understand the system.

1. You can build and run daemon.
2. You can create template, add startup file, and inspect template detail.
3. You can create session and execute both REPL and run-file.
4. You can inspect execution events and runtime summary.
5. You can explain request flow from CLI to core to store/runtime.
6. You can describe what integration/e2e tests cover and major remaining gaps.
7. You can propose at least one safe first contribution and where in code it belongs.

If all seven are true, you are ready to contribute.

## Appendix A: Command quick reference

### Build and run

```bash
GOWORK=off go build -o vm-system ./cmd/vm-system
./vm-system serve --db /tmp/vm-system-guide.db --listen 127.0.0.1:3210
```

### Template workflow

```bash
./vm-system template create --name demo --engine goja
./vm-system template list
./vm-system template add-capability <template-id> --kind module --name console --enabled
./vm-system template add-startup <template-id> --path runtime/init.js --order 10 --mode eval
./vm-system template get <template-id>
./vm-system template list-capabilities <template-id>
./vm-system template list-startup <template-id>
./vm-system template delete <template-id>
```

### Session workflow

```bash
./vm-system session create --template-id <template-id> --workspace-id ws-1 --base-commit deadbeef --worktree-path /abs/path
./vm-system session list
./vm-system session list --status ready
./vm-system session get <session-id>
./vm-system session delete <session-id>
```

### Execution workflow

```bash
./vm-system exec repl <session-id> '1+2'
./vm-system exec run-file <session-id> app.js
./vm-system exec list <session-id>
./vm-system exec get <execution-id>
./vm-system exec events <execution-id> --after-seq 0
```

### Direct API checks

```bash
curl -sS http://127.0.0.1:3210/api/v1/health
curl -sS http://127.0.0.1:3210/api/v1/runtime/summary
curl -sS "http://127.0.0.1:3210/api/v1/sessions?status=ready"
curl -sS "http://127.0.0.1:3210/api/v1/executions?session_id=<session-id>&limit=10"
```

## Appendix B: Example architecture diagram

```text
                          +----------------------------+
                          |        vm-system CLI       |
                          | template/session/exec cmds |
                          +-------------+--------------+
                                        |
                                        | REST (vmclient)
                                        v
+----------------------+    +-----------------------------+    +----------------------+
|    vmdaemon.App      |    |    vmtransport/http Server  |    |  Request ID + JSON   |
| lifecycle + timeouts |<-->| routes + validation + errors|--->| error envelope       |
+-----------+----------+    +---------------+-------------+    +----------------------+
            |                                   |
            |                                   v
            |                    +-------------------------------+
            |                    |         vmcontrol.Core        |
            |                    | Templates / Sessions / Execs  |
            |                    +---------------+---------------+
            |                                    |
            v                                    v
+----------------------+           +-------------------------------+
|   vmsession.Manager  |           |          vmstore             |
| goja runtimes + lock |           | SQLite templates/sessions/...|
+----------------------+           +-------------------------------+
            |
            v
+----------------------+
|      vmexec          |
| execute + events     |
+----------------------+
```

## Related Resources

- Ticket design context: `ttmp/2026/02/07/VM-001-ANALYZE-VM--analyze-vm-system-architecture-behavior-and-quality/design-doc/02-daemonized-vm-system-architecture-backend-runtime-host-rest-api-and-cli.md`
- Coverage matrix and residual risks: `ttmp/2026/02/08/VM-004-EXPAND-E2E-COVERAGE--expand-vm-system-daemon-api-e2e-and-integration-coverage/design-doc/01-daemon-api-test-coverage-matrix-and-expansion-plan.md`
- Primary runtime API implementation: `pkg/vmtransport/http/server.go`
- Runtime/session/execution implementation: `pkg/vmsession/session.go`, `pkg/vmexec/executor.go`

## Part II: Endpoint contracts in detail

This section is intentionally explicit and repetitive. New contributors make fewer mistakes when endpoint contracts are concrete, including failure paths and field names.

## Contract conventions

### Content type and body decoding

- Request body must be JSON for all POST routes.
- Handler decode uses `DisallowUnknownFields`, so unknown keys are rejected.
- Missing required fields produce `VALIDATION_ERROR` with `400`.

### Error envelope shape

All error responses are serialized as:

```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Human readable message",
    "details": {
      "optional": "context"
    }
  }
}
```

### Request ID

Every request receives an `X-Request-Id` response header generated by middleware.

Use this for log correlation when adding structured logging in future.

## Health and runtime ops endpoints

### `GET /api/v1/health`

Purpose:

- lightweight process readiness check

Response:

```json
{"status":"ok"}
```

Notes:

- this is process-level health, not deep dependency health
- if future dependencies are added, keep this route fast and stable

### `GET /api/v1/runtime/summary`

Purpose:

- observe active in-memory runtime state

Response shape:

```json
{
  "active_sessions": 2,
  "active_session_ids": ["session-a", "session-b"]
}
```

Notes:

- list is sorted by session ID in `RuntimeRegistry`
- values represent in-memory runtime ownership, not merely persisted DB rows

## Template endpoints

### `POST /api/v1/templates`

Request:

```json
{
  "name": "my-template",
  "engine": "goja"
}
```

Behavior:

- `name` required
- `engine` optional (defaults to `goja`)
- initializes default settings row in `vm_settings`

Response (`201`): template object

Failure examples:

- invalid JSON -> `400 INVALID_REQUEST`
- missing name -> `400 VALIDATION_ERROR`
- duplicate name (from DB unique constraint) -> currently falls to `500 INTERNAL` unless specialized handling is added

### `GET /api/v1/templates`

Response: list of template objects ordered by created timestamp in store query.

### `GET /api/v1/templates/{template_id}`

Response envelope includes:

- `template`
- `settings`
- `capabilities`
- `startup_files`

Failure:

- unknown template -> `404 TEMPLATE_NOT_FOUND`

### `DELETE /api/v1/templates/{template_id}`

Response:

```json
{"status":"ok","template_id":"..."}
```

Notes:

- row deletion cascades to settings/capabilities/startup files through foreign keys

### `POST /api/v1/templates/{template_id}/capabilities`

Request:

```json
{
  "kind": "module",
  "name": "console",
  "enabled": true,
  "config": {}
}
```

Behavior:

- `kind` and `name` required
- missing config defaults to `{}`

Failure:

- invalid payload -> `400 INVALID_REQUEST` or `400 VALIDATION_ERROR`
- unknown template -> `404 TEMPLATE_NOT_FOUND`

### `GET /api/v1/templates/{template_id}/capabilities`

Returns a list of capability rows.

### `POST /api/v1/templates/{template_id}/startup-files`

Request:

```json
{
  "path": "runtime/init.js",
  "order_index": 10,
  "mode": "eval"
}
```

Behavior:

- `path` required
- `mode` defaults to `eval` when omitted

### `GET /api/v1/templates/{template_id}/startup-files`

Returns startup rows sorted by `order_index`.

## Session endpoints

### `POST /api/v1/sessions`

Request:

```json
{
  "template_id": "...",
  "workspace_id": "ws-1",
  "base_commit_oid": "deadbeef",
  "worktree_path": "/abs/path"
}
```

All four fields are required.

Behavior:

- validates template exists
- validates worktree path exists
- creates in-memory runtime and persisted session row
- executes startup files

Response (`201`): session object

Possible statuses at return time:

- usually `ready`
- may fail and return error if startup/libraries fail

### `GET /api/v1/sessions`

Optional query:

- `status=ready|starting|crashed|closed`

If status omitted, all persisted sessions are returned.

### `GET /api/v1/sessions/{session_id}`

Returns persisted session row including `closed_at` and `last_error` when set.

### `POST /api/v1/sessions/{session_id}/close`

Behavior:

- removes in-memory runtime ownership
- updates DB status to `closed` and sets `closed_at`

### `DELETE /api/v1/sessions/{session_id}`

Alias of close in current implementation.

## Execution endpoints

### `POST /api/v1/executions/repl`

Request:

```json
{"session_id":"...","input":"1+2"}
```

Behavior:

- validates session and readiness
- enforces per-session lock
- records execution and events

Response (`201`): execution object

### `POST /api/v1/executions/run-file`

Request:

```json
{
  "session_id": "...",
  "path": "app.js",
  "args": {"key":"value"},
  "env": {"NAME":"value"}
}
```

Behavior:

- path normalized against worktree
- traversal and absolute paths rejected

### `GET /api/v1/executions`

Required query:

- `session_id`

Optional query:

- `limit` (positive integer; default 50)

Failure:

- missing `session_id` -> `400 VALIDATION_ERROR`
- invalid `limit` -> `400 VALIDATION_ERROR`

### `GET /api/v1/executions/{execution_id}`

Returns execution summary row.

Known contract rough edge:

- unknown ID currently maps to `500 INTERNAL` in integration coverage

### `GET /api/v1/executions/{execution_id}/events`

Optional query:

- `after_seq` (non-negative integer)

Failure:

- invalid `after_seq` -> `400 VALIDATION_ERROR`

## Part III: Deep internal code walk with pseudocode

This section is designed as a narrative code walk for contributors who want implementation-level understanding before touching behavior.

## 1) Core wiring (`pkg/vmcontrol/core.go`)

Pseudocode:

```text
func NewCore(store):
  sessionRuntime = NewSessionManager(store)
  executionRuntime = NewExecutor(store, sessionRuntime)
  return Core{
    Templates: NewTemplateService(store),
    Sessions: NewSessionService(store, sessionRuntime),
    Executions: NewExecutionService(executionRuntime, store, store),
    Registry: NewRuntimeRegistry(sessionRuntime),
  }
```

Interpretation:

- core is compositional, not magical
- session runtime and execution runtime are concrete adapters
- core service objects stay small and mostly orchestration-focused

## 2) Template creation path (`template_service.go`)

Pseudocode:

```text
CreateTemplate(input):
  engine = input.engine or "goja"
  vm = new VM row
  store.CreateVM(vm)

  defaults = {
    limits: {cpu, wall, mem, max_events, max_output_kb},
    resolver: {roots, extensions, allow_absolute_repo_imports},
    runtime: {esm, strict, console},
  }
  store.SetVMSettings(defaults)

  return vm
```

Important operational detail:

- settings are initialized immediately
- code that expects settings can assume row exists for freshly created templates

## 3) Session creation path (`vmsession/session.go`)

Pseudocode:

```text
CreateSession(templateID, workspaceID, baseCommitOID, worktreePath):
  vm = store.GetVM(templateID)
  settings = store.GetVMSettings(templateID)
  ensure worktree exists

  session = {
    status: starting,
    runtime: goja.New(),
  }
  store.CreateSession(sessionRow)

  configure runtime (console, libraries)
  run startup files

  if startup failed:
    status = crashed
    last_error = ...
    store.UpdateSession(...)
    return error

  status = ready
  store.UpdateSession(...)
  register in active session map
  return session
```

Potential pitfalls when changing this path:

- avoid reordering persisted state writes in a way that loses crash visibility
- preserve deterministic status updates
- ensure runtime map and DB rows do not drift during failure handling

## 4) Execution flow (`vmexec/executor.go`)

Pseudocode:

```text
ExecuteREPL(sessionID, input):
  session = sessionManager.GetSession(sessionID)
  ensure status ready
  try lock session.ExecutionLock else session busy

  exec = CreateExecution(status=running)
  add input_echo event
  override console.log to record console events

  value, err = runtime.RunString(input)
  if err:
    add exception event
    mark execution error
    return exec

  add value event (preview + optional JSON export)
  mark execution ok
  return exec
```

For run-file:

- read file content from worktree path
- execute file text in same runtime
- collect console/exception events similarly

Concurrency implication:

- execution lock is the critical guard for single-threaded session execution semantics

## 5) Execution policy layer (`vmcontrol/execution_service.go`)

This layer wraps executor and adds domain-level constraints.

Key pieces:

- `normalizeRunFilePath` rejects absolute paths and traversal attempts
- `enforceLimits` loads template limits and scans persisted events

Important caveat from source:

- limit enforcement path includes soft-fail behavior for some errors while scaffolding matures
- if limit metadata cannot be loaded, execution is not hard-failed by that loading error

When hardening behavior, contributors should decide intentionally whether to keep or remove this soft-fail posture.

## 6) Store semantics (`vmstore/vmstore.go`)

The store layer is a direct `database/sql` adapter.

Characteristics:

- simple, explicit SQL statements
- no ORM abstractions
- timestamps persisted as Unix seconds
- JSON blobs stored in text columns and marshaled/unmarshaled in Go

When adding schema:

1. keep migration additive in `initSchema`
2. update CRUD methods
3. add integration tests that exercise both HTTP contract and store persistence

## 7) HTTP handler style (`server.go`)

Pattern repeated consistently:

1. decode JSON and validate required fields
2. call core service
3. map errors using `writeCoreError`
4. encode JSON response

This consistency is valuable. Preserve this style instead of introducing route-specific special snowflake behavior.

## Part IV: Design reasoning and tradeoffs

A new contributor should understand why the current architecture looks like this.

## Tradeoff A: in-memory runtime ownership vs full persistence

Current model chooses simplicity and runtime speed:

- active runtime is memory-only
- durable DB rows capture metadata and execution history

Benefits:

- low complexity
- straightforward session map + lock behavior

Costs:

- daemon restart loses active runtime state
- reconstruction/recovery semantics not fully implemented

## Tradeoff B: JSON blobs for settings vs normalized tables

Current settings schema uses JSON blobs for limits/resolver/runtime.

Benefits:

- easy to evolve fields without frequent table redesign
- easier to pass through to runtime config objects

Costs:

- weaker static queryability
- more runtime unmarshal failure surface

## Tradeoff C: coarse execution lock

One lock per session guarantees deterministic single execution at a time.

Benefits:

- very clear safety model
- no interleaved state mutation in one runtime

Costs:

- no concurrent execution in same session
- clients must handle `SESSION_BUSY`

## Tradeoff D: explicit SQL and minimal abstractions

Benefits:

- easy to inspect actual database behavior
- lower dependency complexity

Costs:

- repetitive code patterns
- manual consistency responsibilities

## Part V: Guided contribution example

This section walks through a real feature design and change plan from proposal to tests.

### Example feature

Normalize `GET /api/v1/executions/{id}` not-found response from `500 INTERNAL` to `404 EXECUTION_NOT_FOUND`.

### Why this is a good onboarding feature

- bounded surface area
- user-visible contract improvement
- forces contributor to understand store errors, core propagation, and HTTP mapping
- requires integration test update

### Step-by-step implementation plan

1. Introduce a domain error for missing execution if one does not exist.
2. Ensure store `GetExecution` returns that error on missing row.
3. Ensure core execution service passes it through.
4. Add mapping in `writeCoreError` to `404 EXECUTION_NOT_FOUND`.
5. Update/add integration tests to assert status and code.
6. Update guide/changelog/diary for contract change.

### Potential side effects to watch

- CLI currently returns generic API error formatting; ensure new code/message is still understandable
- confirm no other paths depend on previous `INTERNAL` behavior

### Sample test case outline

```go
func TestExecutionGetNotFoundReturns404(t *testing.T) {
  // setup integration server
  // call GET /api/v1/executions/non-existent
  // assert status 404
  // assert envelope error.code == "EXECUTION_NOT_FOUND"
}
```

## Part VI: Operational runbooks for contributors

These runbooks reduce friction when debugging or validating behavior.

## Runbook 1: Fresh daemon-first demo in under 2 minutes

```bash
GOWORK=off go build -o vm-system ./cmd/vm-system
./vm-system serve --db /tmp/vm-demo.db --listen 127.0.0.1:3210

# terminal 2
mkdir -p /tmp/vm-demo/runtime
printf 'globalThis.seed=41\n' >/tmp/vm-demo/runtime/init.js
printf 'seed+1\n' >/tmp/vm-demo/app.js

TEMPLATE_ID=$(./vm-system template create --name quick-demo | sed -n 's/.*(ID: \(.*\)).*/\1/p')
./vm-system template add-startup "$TEMPLATE_ID" --path runtime/init.js --order 10 --mode eval
SESSION_ID=$(./vm-system session create --template-id "$TEMPLATE_ID" --workspace-id ws --base-commit deadbeef --worktree-path /tmp/vm-demo | awk '/Created session:/ {print $3}')
./vm-system exec run-file "$SESSION_ID" app.js
```

If this fails, debug in this order:

1. daemon health endpoint
2. template creation
3. session creation
4. execution call

## Runbook 2: Validate API contracts after handler changes

```bash
GOWORK=off go test ./pkg/vmtransport/http -count=1
```

Then manually spot-check:

```bash
curl -i -sS http://127.0.0.1:3210/api/v1/templates/does-not-exist
curl -i -sS "http://127.0.0.1:3210/api/v1/executions?limit=3"
```

Confirm status and error code match intended contract.

## Runbook 3: Reproduce `SESSION_BUSY`

Use two near-simultaneous execution requests against the same session.

Expected:

- one succeeds
- one gets `409 SESSION_BUSY`

For deterministic reproduction in tests, hold `ExecutionLock` directly in harness (as done in existing error-contract integration test).

## Runbook 4: Validate path traversal protection

```bash
curl -sS -X POST http://127.0.0.1:3210/api/v1/executions/run-file \
  -H 'Content-Type: application/json' \
  -d '{"session_id":"<session-id>","path":"../outside.js"}'
```

Expected:

- `422 INVALID_PATH`

## Runbook 5: Validate output/event limit behavior

Set tight limits in template settings and run execution producing multiple events.

Expected:

- API returns `422 OUTPUT_LIMIT_EXCEEDED`

Use integration tests as reference for deterministic setup (`setTightLimitsForTemplate` helper in safety test suite).

## Part VII: Contributor FAQ

### Q1: Why does the CLI still expose `--db` if runtime commands use daemon API?

`--db` is still used by `serve` and legacy direct-DB commands. For daemon API client mode commands, `--server-url` is the primary runtime selector.

### Q2: Is this project multi-tenant or auth-enabled?

Not in current code path. There is no full authn/authz layer in the main daemon transport.

### Q3: Should I add business logic to HTTP handlers?

No. Keep handlers as thin transport adapters. Put orchestration and policy in `pkg/vmcontrol`.

### Q4: Where should I add cross-cutting telemetry?

Prefer middleware and dedicated instrumentation seams. Avoid scattering ad-hoc prints across core service methods.

### Q5: Can session runtimes be recovered after daemon restart?

Not fully today. This is explicitly a known gap and a strong candidate for future ticket work.

### Q6: Why do some policies look declared but not fully enforced?

The architecture has policy metadata and enforcement scaffolding. Some enforcement paths are mature (path traversal, event/output limits), others remain partial and need explicit hardening tickets.

### Q7: Where should I add new domain errors?

Start in `pkg/vmmodels/models.go` and ensure mapping in `writeCoreError` plus integration tests.

### Q8: What test type should I add first for API changes?

Integration tests in `pkg/vmtransport/http` should be first, because they assert full contract behavior over real store/core/runtime wiring.

### Q9: Is there a style expectation for docs in this repo?

Yes: ticketed docs via docmgr are expected for substantial work, with tasks, changelog, and diary updates.

### Q10: What is the fastest way to understand current behavior drift?

Read integration tests plus smoke/e2e scripts. They capture most practical behavior better than older prose docs alone.

## Part VIII: Advanced architecture notes for maintainers

This section targets developers who may lead larger refactors.

## 1) Toward cleaner adapter boundaries

Current core layer is already transport-agnostic, which is strong. Next maturity step is making runtime adapters more interface-driven, especially around library loading and startup module semantics.

Potential improvement:

- extract explicit `LibraryLoaderPort` and inject into session runtime construction
- avoid direct filesystem policy hardcoding in runtime loader path

## 2) Recovery strategy options

If implementing daemon restart recovery, pick one policy and encode it clearly.

Option A: strict close-on-restart

- mark persisted active sessions `closed` during daemon boot
- requires no runtime reconstruction
- safest operationally but loses continuity

Option B: best-effort rehydrate

- recreate runtime from template + startup scripts
- likely loses transient in-memory state anyway
- needs careful semantics and explicit status markers

Option C: snapshot/restore (future)

- complex and likely outside immediate scope

Recommendation for near-term contributors:

- implement deterministic close-on-restart semantics plus tests first

## 3) API evolution strategy

When changing contracts:

1. prefer additive changes when possible
2. update integration tests before rollout
3. keep CLI formatting robust to new optional fields
4. document breaking changes in ticket changelog and guide updates

## 4) Potential performance work

Areas likely to need attention under load:

- event payload storage volume
- execution list query efficiency with high cardinality
- lock contention in high-frequency execution use cases

Before optimizing, add benchmark or load tests to avoid premature complexity.

## 5) Security hardening opportunities

Current baseline includes traversal protection and some limit checks.

Future hardening areas:

- stronger input validation around JSON size and payload fields
- stricter capability enforcement path in executor/runtime setup
- audit logging and request-level attribution
- optional authn/authz middleware for daemon endpoints

## Part IX: Reading map for a first week on the project

Use this plan if onboarding someone new.

### Day 1: first run and API familiarity

- run quickstart loop from this guide
- call health, template, session, execution endpoints manually
- read `README.md`

### Day 2: source walk

- read command layer files
- read `server.go` end-to-end
- read `core.go`, template/session/execution services

### Day 3: runtime internals

- read `vmsession/session.go` and `vmexec/executor.go`
- trace startup and execution event generation

### Day 4: persistence and contracts

- read `vmstore.go` schema and CRUD methods
- study error mapping in HTTP layer

### Day 5: testing and first contribution

- run integration tests + smoke + e2e
- pick one bounded improvement ticket
- implement with tests and doc updates

This five-day plan gives a new developer high confidence without overwhelming breadth.

## Part X: Contribution templates

Copy these templates when starting work.

### Template: feature plan

```markdown
Goal:
Scope:
Out of scope:
Files to change:
API contract impact:
Test plan:
Rollout risk:
```

### Template: review checklist for API changes

```markdown
- [ ] Route added/updated in server.go
- [ ] Request validation covers required fields and type constraints
- [ ] Domain error mapping updated in writeCoreError
- [ ] Integration tests assert status + error code
- [ ] CLI output still readable
- [ ] Docs/ticket updated
```

### Template: post-implementation note

```markdown
What changed:
Why changed:
How validated:
Known follow-ups:
```

These templates keep quality consistent across contributors.

## Part XI: Known risks snapshot (as of 2026-02-08)

Use this as quick decision support during planning.

1. Missing dedicated execution-not-found contract.
2. Daemon restart recovery semantics not fully implemented.
3. Library cache naming mismatch risk (`id.js` vs `id-version.js`).
4. Some policy declarations are broader than current enforcement.
5. Load/perf behavior under large concurrency not deeply characterized.

When planning large work, reference these risks and decide whether to absorb, defer, or mitigate explicitly.

## Part XII: Final guidance for new contributors

If you remember only six rules from this entire guide, use these.

1. Keep transport logic in handlers and domain logic in core services.
2. Preserve deterministic session/execution lifecycle semantics.
3. Protect API contracts with integration tests first.
4. Use throwaway DB/worktree resources while developing.
5. Prefer small, auditable commits with clear intent.
6. Keep ticket docs current so reviewers can reason from context quickly.

Following these six rules will keep your changes aligned with the architecture and make reviews materially easier.

## Part XIII: End-to-end feature implementation drill

This drill is intentionally concrete. It trains the full contribution loop from concept to validated change.

### Drill objective

Add a new API validation behavior:

- reject excessively large `limit` values for `GET /api/v1/executions`
- return `400 VALIDATION_ERROR` with clear message when above threshold

This is an example only. Even if you do not implement this exact change now, follow the sequence to learn the workflow.

### Step A: define behavior contract first

Write the intended contract before coding.

Example contract:

- accepted range: `1 <= limit <= 1000`
- missing limit uses default `50`
- non-integer or out-of-range values -> `400 VALIDATION_ERROR`

Reasoning:

- upper bound prevents accidental huge query scans
- explicit contract improves client predictability

### Step B: identify exact ownership location

This behavior is transport input validation. Primary ownership is in:

- `pkg/vmtransport/http/server.go`

Potential secondary updates:

- CLI help text for `exec list --limit`
- tests in `server_error_contracts_integration_test.go`

### Step C: update tests first (or at least in same commit)

Add integration assertions:

1. `limit=0` -> existing invalid
2. `limit=-1` -> existing invalid
3. `limit=1001` -> new invalid contract
4. `limit=1000` -> accepted

Why integration tests here:

- contract is HTTP-level
- request parsing and response envelope are the behavior under test

### Step D: implement handler validation

In `handleExecutionList`:

- parse `limit`
- enforce range
- keep clear message text

Example pseudocode:

```text
if rawLimit provided:
  parsed = Atoi(rawLimit)
  if parsed <= 0 || parsed > 1000:
    VALIDATION_ERROR
```

### Step E: validate locally

```bash
GOWORK=off go test ./pkg/vmtransport/http -run TestAPIErrorContractsValidationNotFoundConflictAndUnprocessable -count=1
GOWORK=off go test ./pkg/vmtransport/http -count=1
GOWORK=off go test ./... -count=1
```

### Step F: run quick behavior checks manually

```bash
curl -i -sS "http://127.0.0.1:3210/api/v1/executions?session_id=s1&limit=1001"
curl -i -sS "http://127.0.0.1:3210/api/v1/executions?session_id=s1&limit=1000"
```

Manual checks confirm your contract and test assumptions align.

### Step G: document and commit

Commit message example:

```text
test(api): enforce max execution list limit and cover validation
```

Ticket notes should include:

- what changed
- why changed
- validation commands
- any follow-up questions

### Why this drill is valuable

You practice all core project muscles:

- identifying correct ownership layer
- writing contract tests
- making minimal code changes
- validating at both automated and manual levels
- producing reviewable change records

## Part XIV: Glossary and field semantics

Use this glossary when reading API payloads and persistence rows.

### Template terms

- `template_id`:
  - identifier for template/VM profile
  - maps to `vm.id` in store
- `engine`:
  - runtime engine label (currently defaulted to `goja`)
- `capabilities`:
  - declared allowlist/permissions metadata
- `startup_files`:
  - ordered scripts run at session boot

### Session terms

- `session_id`:
  - identifier for runtime session
- `workspace_id`:
  - client workspace correlation string
- `base_commit_oid`:
  - commit correlation string for workspace snapshot context
- `worktree_path`:
  - filesystem root used for startup/run-file resolution
- `status`:
  - lifecycle state (`starting`, `ready`, `crashed`, `closed`)

### Execution terms

- `execution_id`:
  - identifier for one execution attempt
- `kind`:
  - `repl` or `run_file`
- `input`:
  - code snippet for REPL execution
- `path`:
  - relative path for run-file execution
- `result`:
  - serialized value payload on success
- `error`:
  - serialized exception payload on failure

### Event terms

- `seq`:
  - monotonic per-execution event sequence number
- `ts`:
  - event timestamp
- `type`:
  - `input_echo`, `console`, `value`, `exception`, etc.
- `payload`:
  - JSON payload shape depends on event type

## Part XV: Field-by-field payload examples

This section is useful when implementing clients or debugging serialization changes.

### Example: successful REPL execution response

```json
{
  "id": "e-123",
  "session_id": "s-123",
  "kind": "repl",
  "input": "21*2",
  "args": [],
  "env": {},
  "status": "ok",
  "started_at": "2026-02-08T12:00:00Z",
  "ended_at": "2026-02-08T12:00:00Z",
  "result": {
    "type": "int64",
    "preview": "42",
    "json": 42
  },
  "metrics": {}
}
```

### Example: exception event payload

```json
{
  "execution_id": "e-123",
  "seq": 3,
  "ts": "2026-02-08T12:00:00Z",
  "type": "exception",
  "payload": {
    "message": "ReferenceError: x is not defined",
    "stack": "..."
  }
}
```

### Example: template detail response

```json
{
  "template": {
    "id": "t-123",
    "name": "demo",
    "engine": "goja",
    "is_active": true,
    "exposed_modules": [],
    "libraries": [],
    "created_at": "...",
    "updated_at": "..."
  },
  "settings": {
    "vm_id": "t-123",
    "limits": {"cpu_ms":2000,"wall_ms":5000,"mem_mb":128,"max_events":50000,"max_output_kb":256},
    "resolver": {"roots":["."],"extensions":[".js",".mjs"],"allow_absolute_repo_imports":true},
    "runtime": {"esm":true,"strict":true,"console":true}
  },
  "capabilities": [
    {"id":"c-1","vm_id":"t-123","kind":"module","name":"console","enabled":true,"config":{}}
  ],
  "startup_files": [
    {"id":"sf-1","vm_id":"t-123","path":"runtime/init.js","order_index":10,"mode":"eval"}
  ]
}
```

When changing payload shape, update both integration tests and CLI formatting code.

## Part XVI: Final onboarding checklist for tech leads

Use this with every new developer joining the backend.

1. They can run daemon + full first VM flow without help.
2. They can explain why session continuity requires long-lived daemon ownership.
3. They can locate ownership layers for transport/core/runtime/store concerns.
4. They can run and interpret integration, smoke, and e2e checks.
5. They can describe current major coverage gaps honestly.
6. They can propose one well-scoped improvement and a test plan.
7. They can produce ticketed docs with tasks/changelog/diary updates.

If all seven are true, onboarding is complete.
