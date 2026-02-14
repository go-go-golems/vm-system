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

The vm-system daemon exposes a REST API under `/api/v1`. All request bodies
must be JSON with `Content-Type: application/json`, and all responses are
JSON. Every response includes an `X-Request-Id` header that you can use for
log correlation when debugging.

When frontend assets are present, the daemon may also serve the web UI from
`/` and `/assets/*`. Those routes are outside the REST API contract; API
behavior is scoped to `/api/v1/*`.

One important behavior: the server enforces strict JSON decoding with
`DisallowUnknownFields`. If you send a field the server doesn't expect, you
get `400 INVALID_REQUEST`. This catches typos early but can be surprising if
you're used to lenient APIs.

## Errors

Every error response uses the same envelope, regardless of the endpoint:

```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable explanation",
    "details": { "optional": "context" }
  }
}
```

The `code` field is the stable identifier you should match on in client code.
The `message` is for humans and may change between versions. Here's the
complete mapping:

| Situation | Status | Code |
|-----------|--------|------|
| Missing or invalid field in request | 400 | `VALIDATION_ERROR` |
| Malformed JSON or unknown fields | 400 | `INVALID_REQUEST` |
| Template not found | 404 | `TEMPLATE_NOT_FOUND` |
| Session not found | 404 | `SESSION_NOT_FOUND` |
| Execution not found | 404 | `EXECUTION_NOT_FOUND` |
| File not found in worktree | 404 | `FILE_NOT_FOUND` |
| Session not in `ready` state | 409 | `SESSION_NOT_READY` |
| Another execution already running | 409 | `SESSION_BUSY` |
| Path traversal or absolute path | 422 | `INVALID_PATH` |
| Output/event limit exceeded | 422 | `OUTPUT_LIMIT_EXCEEDED` |
| Unsupported startup mode | 422 | `STARTUP_MODE_UNSUPPORTED` |
| Adding a built-in as a module | 422 | `MODULE_NOT_ALLOWED` |
| Unhandled internal error | 500 | `INTERNAL` |

If you see `500 INTERNAL` for something that should have a specific error
code, that's a bug worth reporting.

## Health and operations

These endpoints let you check if the daemon is up and what it's doing.

**GET /api/v1/health** returns a simple readiness check. If you get a
response at all, the daemon is running:

```json
{"status":"ok"}
```

**GET /api/v1/runtime/summary** tells you what's actually alive in daemon
memory. This is different from querying sessions in the database — after a
daemon restart, the database still has session rows, but this endpoint
correctly shows zero active sessions:

```json
{
  "active_sessions": 2,
  "active_session_ids": ["session-a", "session-b"]
}
```

## Templates

Templates are persistent runtime profiles. They define what a JavaScript
runtime looks like: engine, limits, startup files, modules, and libraries.
Creating a template also initializes default settings, so you can start
creating sessions from it immediately.

**POST /api/v1/templates** creates a template:

```json
{ "name": "my-template", "engine": "goja" }
```

`name` is required. `engine` defaults to `goja` if omitted. Returns **201**
with the template object including its generated UUID.

**GET /api/v1/templates** lists all templates.

**GET /api/v1/templates/{template_id}** returns the template along with its
settings, capabilities, and startup files — everything you need to understand
what sessions created from it will look like.

**DELETE /api/v1/templates/{template_id}** deletes a template and all its
associated data (settings, capabilities, startup files) through cascading
foreign keys.

### Template sub-resources

Templates have four kinds of sub-resources. Each can be added, listed, and
in some cases removed independently.

**Capabilities** are metadata that describe what features the template
enables. They're descriptive — the runtime reads them during session setup:

- **POST /api/v1/templates/{id}/capabilities** — requires `kind` and `name`.
  Kinds are `module`, `global`, `fs`, `net`, `env`. Optional `enabled` flag
  and `config` JSON object.
- **GET /api/v1/templates/{id}/capabilities** — lists all capabilities.

**Startup files** are scripts that run during session creation. The
`order_index` determines the sequence — lower numbers run first:

- **POST /api/v1/templates/{id}/startup-files** — requires `path` (relative
  to the future session's worktree). Optional `order_index` (default 0) and
  `mode` (default `eval`, currently the only supported mode).
- **GET /api/v1/templates/{id}/startup-files** — lists startup files sorted
  by `order_index`.

**Modules** are host-provided native capabilities (like database access or
filesystem operations). Only modules from the configurable catalog can be
added — JavaScript built-ins like JSON and Math are always available and
attempting to add them returns `422 MODULE_NOT_ALLOWED`:

- **POST /api/v1/templates/{id}/modules** — body: `{"name":"database"}`.
- **GET /api/v1/templates/{id}/modules** — lists configured modules.
- **DELETE /api/v1/templates/{id}/modules/{name}** — removes a module.

**Libraries** are third-party JavaScript files that get loaded into the
runtime's global scope at session startup:

- **POST /api/v1/templates/{id}/libraries** — body:
  `{"name":"lodash-4.17.21"}`.
- **GET /api/v1/templates/{id}/libraries** — lists configured libraries.
- **DELETE /api/v1/templates/{id}/libraries/{name}** — removes a library.

### Default settings

When you create a template, it gets these defaults. You can see them in the
template detail response:

```json
{
  "limits": {
    "cpu_ms": 5000, "wall_ms": 30000, "mem_mb": 128,
    "max_events": 1000, "max_output_kb": 512
  },
  "resolver": {
    "roots": [], "extensions": [".js", ".mjs"],
    "allow_absolute_repo_imports": false
  },
  "runtime": {
    "esm": false, "strict": false, "console": true
  }
}
```

## Sessions

Sessions are live goja runtime instances. They exist in daemon memory and are
backed by persistent database rows. Creating a session triggers the full
startup sequence: runtime allocation, library loading, and startup file
execution.

**POST /api/v1/sessions** creates a session:

```json
{
  "template_id": "...",
  "workspace_id": "ws-1",
  "base_commit_oid": "deadbeef",
  "worktree_path": "/absolute/path/to/worktree"
}
```

All four fields are required. The worktree directory must exist on disk and
the path must be absolute. Returns **201** with the session object — the
`status` field will be `ready` if startup succeeded, or the creation will
fail if something went wrong.

**GET /api/v1/sessions** lists sessions. You can filter by status with
`?status=ready` (also accepts `starting`, `crashed`, `closed`). Without the
filter, all sessions are returned.

**GET /api/v1/sessions/{session_id}** returns full session detail including
`closed_at` and `last_error` when relevant. This is where you look when a
session crashed during creation.

**POST /api/v1/sessions/{session_id}/close** closes a session. The in-memory
runtime is discarded and the database row is updated with `closed_at`.

**DELETE /api/v1/sessions/{session_id}** is an alias for close.

## Executions

Executions are individual code runs inside a session. Each one produces a
persisted record and a stream of typed events.

**POST /api/v1/executions/repl** runs an inline JavaScript snippet:

```json
{ "session_id": "...", "input": "1 + 2" }
```

Returns **201** with the execution object and its events array. The events
capture everything: the input echo, any console output, the return value or
exception.

**POST /api/v1/executions/run-file** runs a file from the session worktree:

```json
{
  "session_id": "...",
  "path": "scripts/app.js",
  "args": {},
  "env": {}
}
```

The `path` must be relative to the worktree. Absolute paths and `../`
traversal are rejected with `422 INVALID_PATH` before any JavaScript runs.

**GET /api/v1/executions** lists executions. Requires `session_id` as a query
parameter. Optional `limit` (default 50, must be a positive integer).

**GET /api/v1/executions/{execution_id}** returns a single execution.

**GET /api/v1/executions/{execution_id}/events** returns the event stream for
an execution. The optional `after_seq` query parameter enables cursor-based
pagination — pass the `seq` of the last event you've seen, and you get only
newer events. This is how you poll for output in automation.

### Event types

Events are the atomic output of an execution. Each has a sequential `seq`
number starting from 1:

- **input_echo** — what was submitted (the code string or file path)
- **console** — captured from `console.log`, `console.warn`, `console.error`,
  `console.info`, and `console.debug`. Payload:
  `{"level":"log","text":"..."}`
- **value** — the return value of the expression. Payload includes a type
  name, a human-readable preview, and optional JSON:
  `{"type":"number","preview":"42","json":42}`
- **exception** — a thrown JavaScript error. Payload:
  `{"message":"ReferenceError: x is not defined","stack":"..."}`
- **system** — internal lifecycle messages from the runtime. Payload:
  `{"message":"...","level":"info"}`
- **stdout / stderr** — raw output capture (less common than console events)

## See Also

- `vm-system help getting-started` — hands-on walkthrough
- `vm-system help architecture` — how the server is structured internally
- `vm-system help cli-command-reference` — CLI equivalents for every endpoint
