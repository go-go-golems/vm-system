---
Title: "REST API Reference"
Slug: api-reference
Short: "Complete endpoint contracts: request shapes, response envelopes, status codes, and error codes."
Topics:
- vm-system
- api
- rest
- reference
Commands:
- serve
IsTopLevel: true
IsTemplate: false
ShowPerDefault: true
SectionType: GeneralTopic
---

All endpoints live under `/api/v1`. Request bodies must be JSON with
`Content-Type: application/json`. Every response includes an `X-Request-Id`
header for log correlation. Unknown JSON fields are rejected.

## Error envelope

All error responses use this shape:

```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable message",
    "details": { "optional": "context" }
  }
}
```

## Error code mapping

| Domain error | HTTP status | Code |
|-------------|-------------|------|
| Template not found | 404 | `TEMPLATE_NOT_FOUND` |
| Session not found | 404 | `SESSION_NOT_FOUND` |
| Execution not found | 404 | `EXECUTION_NOT_FOUND` |
| File not found | 404 | `FILE_NOT_FOUND` |
| Session not ready | 409 | `SESSION_NOT_READY` |
| Session busy | 409 | `SESSION_BUSY` |
| Path traversal | 422 | `INVALID_PATH` |
| Output limit exceeded | 422 | `OUTPUT_LIMIT_EXCEEDED` |
| Startup mode unsupported | 422 | `STARTUP_MODE_UNSUPPORTED` |
| Module not allowed | 422 | `MODULE_NOT_ALLOWED` |
| Validation failure | 400 | `VALIDATION_ERROR` |
| Decode failure | 400 | `INVALID_REQUEST` |
| Unhandled | 500 | `INTERNAL` |

## Health and operations

### GET /api/v1/health

Lightweight process readiness check.

**Response (200):**

```json
{ "status": "ok" }
```

### GET /api/v1/runtime/summary

Reports in-memory active sessions (not just persisted rows).

**Response (200):**

```json
{
  "active_sessions": 2,
  "active_session_ids": ["session-a", "session-b"]
}
```

## Templates

### POST /api/v1/templates

Create a new template (runtime profile).

**Request:**

```json
{
  "name": "my-template",
  "engine": "goja"
}
```

`name` is required. `engine` defaults to `goja`.

**Response (201):** Template object with generated ID, default settings, empty
capabilities and startup files.

**Errors:** `400 VALIDATION_ERROR` (missing name), `400 INVALID_REQUEST`
(malformed JSON).

### GET /api/v1/templates

List all templates.

**Response (200):** Array of template objects.

### GET /api/v1/templates/{template_id}

Get template with settings, capabilities, and startup files.

**Response (200):** Envelope with `template`, `settings`, `capabilities`,
`startup_files` fields.

**Errors:** `404 TEMPLATE_NOT_FOUND`.

### DELETE /api/v1/templates/{template_id}

Delete a template. Cascades to settings, capabilities, and startup files via
foreign keys.

**Response (200):**

```json
{ "status": "ok", "template_id": "..." }
```

**Errors:** `404 TEMPLATE_NOT_FOUND`.

### POST /api/v1/templates/{template_id}/capabilities

Add a capability to a template.

**Request:**

```json
{
  "kind": "module",
  "name": "console",
  "enabled": true,
  "config": {}
}
```

`kind` and `name` are required. Capability kinds: `module`, `global`, `fs`,
`net`, `env`.

### GET /api/v1/templates/{template_id}/capabilities

List capabilities for a template.

### POST /api/v1/templates/{template_id}/startup-files

Add a startup file to a template.

**Request:**

```json
{
  "path": "runtime/init.js",
  "order_index": 10,
  "mode": "eval"
}
```

`path` is required. `mode` defaults to `eval`. Files execute in ascending
`order_index` during session creation.

### GET /api/v1/templates/{template_id}/startup-files

List startup files sorted by `order_index`.

### POST /api/v1/templates/{template_id}/modules

Add a native module to a template. Only modules from the configurable catalog
are allowed; built-in JavaScript globals (JSON, Math, Date) are always available
and cannot be configured.

**Request:**

```json
{ "name": "console" }
```

**Errors:** `422 MODULE_NOT_ALLOWED` if the module is a non-configurable built-in.

### GET /api/v1/templates/{template_id}/modules

List configured modules.

### DELETE /api/v1/templates/{template_id}/modules/{module_name}

Remove a module.

### POST /api/v1/templates/{template_id}/libraries

Add a third-party library to a template.

**Request:**

```json
{ "name": "lodash-4.17.21" }
```

### GET /api/v1/templates/{template_id}/libraries

List configured libraries.

### DELETE /api/v1/templates/{template_id}/libraries/{library_name}

Remove a library.

## Sessions

### POST /api/v1/sessions

Create a new session from a template.

**Request:**

```json
{
  "template_id": "...",
  "workspace_id": "ws-1",
  "base_commit_oid": "deadbeef",
  "worktree_path": "/absolute/path/to/worktree"
}
```

All four fields are required. The worktree directory must exist on disk.

**Response (201):** Session object. Status is usually `ready`; may fail if
startup or library loading fails.

**Errors:** `404 TEMPLATE_NOT_FOUND`, `400 VALIDATION_ERROR`.

### GET /api/v1/sessions

List sessions. Optional query parameter: `status` (`starting`, `ready`,
`crashed`, `closed`).

### GET /api/v1/sessions/{session_id}

Get a single session including `closed_at` and `last_error` when set.

**Errors:** `404 SESSION_NOT_FOUND`.

### POST /api/v1/sessions/{session_id}/close

Close a session. Removes the in-memory runtime and updates the DB row.

**Errors:** `404 SESSION_NOT_FOUND`.

### DELETE /api/v1/sessions/{session_id}

Alias for close.

## Executions

### POST /api/v1/executions/repl

Execute a REPL snippet in a session.

**Request:**

```json
{
  "session_id": "...",
  "input": "1 + 2"
}
```

**Response (201):** Execution object with events array.

**Errors:** `404 SESSION_NOT_FOUND`, `409 SESSION_NOT_READY`,
`409 SESSION_BUSY`, `422 OUTPUT_LIMIT_EXCEEDED`.

### POST /api/v1/executions/run-file

Execute a file from the session worktree.

**Request:**

```json
{
  "session_id": "...",
  "path": "app.js",
  "args": {},
  "env": {}
}
```

`path` must be relative to the worktree. Absolute paths and `../` traversal are
rejected.

**Errors:** `422 INVALID_PATH`, `404 FILE_NOT_FOUND`, `409 SESSION_BUSY`.

### GET /api/v1/executions

List executions for a session.

**Query parameters:**

- `session_id` (required)
- `limit` (optional, positive integer, default 50)

**Errors:** `400 VALIDATION_ERROR` (missing session_id or invalid limit).

### GET /api/v1/executions/{execution_id}

Get a single execution.

### GET /api/v1/executions/{execution_id}/events

Get events for an execution. Optional query: `after_seq` (non-negative integer)
for cursor-based pagination.

**Errors:** `400 VALIDATION_ERROR` (invalid after_seq).

## Event types

Events are emitted during execution and stored with sequential `seq` numbers:

| Type | When emitted | Payload shape |
|------|-------------|---------------|
| `input_echo` | Start of every execution | `{ "text": "..." }` |
| `console` | `console.log/warn/error/info/debug` | `{ "level": "log", "text": "..." }` |
| `value` | Successful expression result | `{ "type": "number", "preview": "42", "json": 42 }` |
| `exception` | Runtime error | `{ "message": "...", "stack": "..." }` |
| `system` | Internal lifecycle events | `{ "message": "...", "level": "info" }` |
| `stdout` | Standard output capture | Raw text |
| `stderr` | Standard error capture | Raw text |

## Template default settings

When a template is created, default settings are initialized:

```json
{
  "limits": { "cpu_ms": 5000, "wall_ms": 30000, "mem_mb": 128, "max_events": 1000, "max_output_kb": 512 },
  "resolver": { "roots": [], "extensions": [".js", ".mjs"], "allow_absolute_repo_imports": false },
  "runtime": { "esm": false, "strict": false, "console": true }
}
```

## Troubleshooting

| Problem | Cause | Solution |
|---------|-------|----------|
| 400 on POST with valid-looking JSON | Unknown field in body | Remove extra fields; `DisallowUnknownFields` is enforced |
| 409 SESSION_BUSY | Concurrent execution attempt | Wait for current execution to finish |
| 422 INVALID_PATH | Absolute path or traversal in run-file | Use relative path without `../` |
| 500 INTERNAL on unknown resource | Missing specific error mapping | File a bug; this should be a 404 |

## See Also

- `vm-system help getting-started`
- `vm-system help architecture`
- `vm-system help cli-command-reference`
