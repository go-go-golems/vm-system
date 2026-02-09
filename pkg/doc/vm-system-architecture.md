---
Title: "vm-system Architecture"
Slug: architecture
Short: "Package layout, layered design, request lifecycle, and key design tradeoffs."
Topics:
- vm-system
- architecture
- design
IsTopLevel: true
IsTemplate: false
ShowPerDefault: true
SectionType: GeneralTopic
---

vm-system is organized as a layered daemon with clean separation between
transport, domain logic, and infrastructure. This page explains why the code
is structured the way it is, how a request travels through the system, and
what tradeoffs were made along the way.

If you're trying to understand the codebase for the first time, this is the
right page. If you want to *run* things first, start with
`vm-system help getting-started` and come back here.

## The big picture

The system follows a ports-and-adapters pattern. There are four layers, and
the arrows only point downward — the core never reaches up into HTTP or CLI
code:

```
  ┌─────────────────────────────────────────────────────┐
  │  cmd/vm-system          CLI (Cobra + Glazed)        │
  │  pkg/vmclient           REST client used by CLI     │
  ├─────────────────────────────────────────────────────┤
  │  pkg/vmtransport/http   REST routes + validation    │
  │  pkg/vmdaemon           Process host + shutdown     │
  ├─────────────────────────────────────────────────────┤
  │  pkg/vmcontrol          Core orchestration          │
  │    TemplateService · SessionService                 │
  │    ExecutionService · RuntimeRegistry               │
  ├─────────────────────────────────────────────────────┤
  │  pkg/vmsession          Live goja runtimes + locks  │
  │  pkg/vmexec             Execution + event capture   │
  │  pkg/vmstore            SQLite persistence          │
  │  pkg/vmmodels           Types, errors, IDs          │
  │  pkg/vmmodules          Module catalog              │
  │  pkg/vmpath             Path normalization          │
  │  pkg/libloader          Library cache loading       │
  └─────────────────────────────────────────────────────┘
```

The most important rule in the codebase: **vmcontrol is transport-agnostic.**
No HTTP request objects, no Cobra commands, no JSON tags on internal types —
just domain types in, domain types out. The HTTP layer translates between REST
DTOs and core types. The CLI layer does the same through vmclient. This
separation is what makes the core testable without spinning up an HTTP server,
and it's what would let you add a gRPC or WebSocket transport without touching
any business logic.

## What each package does

### Command layer (cmd/vm-system)

The CLI is a thin Cobra command tree. Each file maps to one command group, and
all they do is parse flags, call `vmclient` methods, and format the response
for humans:

- **main.go** — the root command. Wires the Glazed help system, sets up
  global `--db` and `--server-url` flags. This is where you go to see what's
  registered.
- **cmd_serve.go** — the only command that runs the daemon itself (all others
  are REST clients). It sets up the daemon via `vmdaemon` and blocks until
  shutdown.
- **cmd_template*.go** — template CRUD, modules, libraries, capabilities,
  startup files. This is split across three files because the template surface
  area is large — 16 subcommands in total.
- **cmd_session.go** — session create, list, get, and close.
- **cmd_ops.go** — health and runtime-summary — the two operational queries.

The important thing to understand about the CLI is that it never touches the
database directly. Every command creates a `vmclient.New(serverURL, nil)` and
makes REST calls. This means the daemon is always the source of truth.

### Core orchestration (pkg/vmcontrol)

This is the brain of the system. It's deliberately small — around 500 lines
of orchestration logic across five files:

- **core.go** is the composition root. It creates the four services and wires
  them together with their adapter dependencies. If you want to understand how
  the system is assembled, start here.
- **template_service.go** handles template creation and default settings
  initialization. When you create a template, this is where the default limits
  (5s CPU, 128MB memory, 1000 max events) come from.
- **session_service.go** orchestrates the complex session creation flow:
  look up the template, allocate a runtime, load libraries, execute startup
  files, handle crashes, update the database.
- **execution_service.go** wraps the raw executor with domain-level concerns:
  normalizing file paths, rejecting traversal attempts, enforcing output limits.
- **ports.go** defines the interfaces that separate core from adapters. This
  is the key to testability — integration tests can wire the same core with
  a real SQLite store, while unit tests could use stubs.

### HTTP transport (pkg/vmtransport/http)

The HTTP layer is a thin adapter. Every handler follows the exact same
pattern: decode the JSON request, validate required fields, call a core
service method, map any domain error to an HTTP status code and error
envelope, and encode the JSON response.

Three behaviors are worth knowing about:

- **Strict decoding:** `json.Decoder` is configured with
  `DisallowUnknownFields`. If you send a JSON field the server doesn't expect,
  you get `400 INVALID_REQUEST`. This catches typos early.
- **Error mapping:** `writeCoreError` in `server_errors.go` is the single
  place where domain errors (like `ErrSessionNotFound`) become HTTP responses
  (like `404 SESSION_NOT_FOUND`). If you add a new error, this is where you
  register it.
- **Request IDs:** Every response gets an `X-Request-Id` header from
  middleware. This is useful for correlating logs when debugging.

### Runtime layer (pkg/vmsession, pkg/vmexec)

**vmsession** owns the in-memory map of active goja runtimes. Each session
has a `TryLock` mutex — if you try to execute code while another execution
is already running, you get `SESSION_BUSY` immediately. There's no wait queue
and no timeout — the design deliberately keeps concurrency simple.

**vmexec** is the execution pipeline. It's where JavaScript actually runs.
The executor takes a session lock, creates an execution record, overrides
`console.log` to capture output as events, runs the code via `goja.RunString`,
records the return value or exception, and persists everything to the store.
All console calls, return values, and exceptions become typed events with
sequential `seq` numbers.

### Persistence (pkg/vmstore)

The store is a straightforward `database/sql` adapter with hand-written SQL.
There's no ORM, no query builder, no migration framework — just explicit
CREATE TABLE statements in `initSchema()` and CRUD methods.

The schema covers seven tables:

- **vm** + **vm_settings** — template identity and configuration. Settings
  (limits, resolver config, runtime config) are stored as JSON blobs rather
  than normalized columns, which makes it easy to add fields without schema
  migrations.
- **vm_capability** + **vm_startup_file** — template policy metadata.
  Capabilities describe what features a template enables. Startup files have
  an `order_index` that controls execution order.
- **vm_session** — durable session records with status, timestamps, and
  error messages. These survive daemon restarts even though the runtimes don't.
- **execution** + **execution_event** — execution summaries (kind, status,
  timing) and event streams (each event has an `execution_id` and a `seq`
  number for ordered retrieval).

## How a request flows

Here's what happens when you run `vm-system exec repl <session> '1+1'`. This
is the most interesting flow because it touches every layer:

```
  CLI parses flags, creates vmclient
       │
       ▼
  vmclient.ExecuteREPL()
       │  POST /api/v1/executions/repl
       │  {"session_id":"...","input":"1+1"}
       ▼
  HTTP handler decodes JSON, validates fields
       │
       ▼
  core.Executions.ExecuteREPL()
       │  looks up the session
       ▼
  executor.TryLock(session)
       │  ← fails fast with SESSION_BUSY if already locked
       ▼
  Creates execution row in SQLite (status: running)
       │
       ▼
  Overrides console.log to capture events
       │
       ▼
  goja.RunString("1+1")
       │  captures return value → "value" event
       │  would capture exception if code threw
       ▼
  Finalizes execution (status: ok)
       │  persists all events to SQLite
       ▼
  201 JSON response with execution + events
```

Template creation and session creation follow the same layered pattern. Session
creation is the most complex — it allocates a runtime, configures the console
shim, loads libraries, and executes startup files, all within a single request.
If any startup file throws, the session transitions to `crashed` instead of
`ready`.

## Data model

**Sessions** move through four states. The happy path is simple, and the crash
path captures what went wrong:

```
  starting ──► ready ──► closed
      │
      └──► crashed (startup failed — check last_error)
```

**Executions** have their own state machine. Most executions end in `ok` or
`error`:

```
  running ──► ok        (expression returned a value)
         ──► error     (JavaScript exception was thrown)
         ──► timeout   (resource limit exceeded)
```

**Events** are the atomic unit of execution output. Every execution produces a
stream of events, each with a sequential `seq` number that enables cursor-based
retrieval. The event types are:

- **input_echo** — what the user submitted (the code or file path)
- **console** — from `console.log`, `console.warn`, etc. (captured by the
  console shim)
- **value** — the return value of the expression (type, human-readable preview,
  and optional JSON representation)
- **exception** — a thrown error with message and stack trace
- **system** — internal lifecycle messages from the runtime

## Design tradeoffs

Every architecture involves tradeoffs. Here are the ones that matter most in
vm-system, and why they were made:

**In-memory runtimes with no state persistence.** The goja runtime for each
session lives entirely in daemon memory. If the daemon restarts, active
runtimes are gone — the database rows survive, but there's nothing to
reconnect them to. The alternative would be serializing V8/goja heap state,
which is complex and fragile. The current design keeps things simple and fast.
Recovery semantics (like closing stale sessions on startup) are a known future
work item.

**JSON blobs for settings.** Template settings (limits, resolver config,
runtime config) are stored as JSON text columns rather than normalized tables.
This was chosen because settings evolve frequently — adding a new limit field
is a one-line code change, not a schema migration. The cost is that you can't
write SQL queries like "find all templates with CPU limit > 10s" without
parsing JSON. In practice this hasn't been needed.

**One lock per session, no execution queue.** Each session has a single mutex.
If an execution is already running, the next request gets `SESSION_BUSY` (409)
immediately — there's no queue, no retry logic, no priority system. This is
deliberately simple. goja runtimes are single-threaded, so true concurrent
execution would require serialization anyway. The lock makes the concurrency
model obvious and easy to reason about.

**Plain SQL, no ORM.** The store uses `database/sql` directly with
hand-written queries. This keeps behavior transparent and dependencies minimal.
The cost is repetition — each CRUD method has its own SQL string. But the
queries are simple enough that an ORM would add complexity without much benefit.

## Reading order for new contributors

If you want to understand the codebase quickly, read these files in order.
Each one builds on the previous, and together they cover about 90% of the
system's behavior:

1. **cmd/vm-system/main.go** — entry point. See what commands are registered.
2. **cmd/vm-system/cmd_serve.go** — how the daemon starts up.
3. **pkg/vmdaemon/app.go** — process lifecycle and graceful shutdown.
4. **pkg/vmcontrol/core.go** — how services are composed. This is the map.
5. **pkg/vmtransport/http/server.go** — all routes in one place.
6. **pkg/vmcontrol/template_service.go** — template creation and defaults.
7. **pkg/vmcontrol/session_service.go** — session orchestration.
8. **pkg/vmcontrol/execution_service.go** — path safety and limit checks.
9. **pkg/vmsession/session.go** — runtime allocation and locking.
10. **pkg/vmexec/executor.go** — where JavaScript actually runs.
11. **pkg/vmstore/vmstore.go** — schema and persistence.

## See Also

- `vm-system help getting-started` — first run walkthrough
- `vm-system help api-reference` — endpoint contracts and error codes
- `vm-system help contributing` — how to make changes safely
