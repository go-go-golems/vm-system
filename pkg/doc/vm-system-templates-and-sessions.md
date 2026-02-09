---
Title: "Templates, Sessions, and Executions"
Slug: templates-and-sessions
Short: "How templates define runtime policy, sessions own live runtimes, and executions capture results."
Topics:
- vm-system
- templates
- sessions
- executions
- concepts
Commands:
- template
- session
- exec
IsTopLevel: true
IsTemplate: false
ShowPerDefault: true
SectionType: GeneralTopic
---

Everything in vm-system revolves around three concepts. **Templates** describe
what a runtime is allowed to do. **Sessions** are live runtimes built from
templates. **Executions** are individual code runs inside a session, with every
event captured and queryable.

These three concepts form a natural hierarchy: you define templates once, create
sessions from them as needed, and run as many executions as you want inside
each session. Understanding how they relate is the key to using vm-system
effectively.

## Templates — what should the runtime look like?

A template is a saved configuration that controls every aspect of a JavaScript
runtime. Think of it as "a VM image you can stamp out sessions from." You
create a template once, configure it to your needs, and then create sessions
from it whenever you need a fresh runtime.

Templates capture these aspects of a runtime:

- **Engine** — which JavaScript engine to use. Currently `goja` is the only
  supported engine, with QuickJS, Node, and custom engines planned for the
  future.
- **Resource limits** — how much CPU time, wall time, and memory an execution
  gets, and how many events and output bytes it can produce. These prevent
  runaway scripts from consuming unbounded resources.
- **Startup files** — scripts that run automatically when a session is created,
  in a defined order. This is where you set up global state, configure
  database connections, define helper functions, or load polyfills.
- **Native modules** — host-provided capabilities that you opt into. These
  are Go functions exposed to JavaScript, giving the runtime access to things
  like SQLite, the filesystem, or shell commands.
- **Libraries** — third-party JavaScript files loaded into the runtime at
  startup. These are downloaded from CDN to a local cache and injected into
  the global scope.
- **Resolver settings** — how `import` statements are resolved (module roots,
  file extensions, whether absolute paths are allowed).
- **Runtime settings** — whether to enable ESM mode, strict mode, and the
  console shim.

### Setting up a template

The typical workflow is: create a template, then layer on configuration one
piece at a time:

```bash
# Start with a basic template — you get goja, console enabled, and sensible limits
vm-system template create --name my-runtime --engine goja

# Add startup files — they'll run in order_index order during session creation
vm-system template add-startup <id> --path runtime/init.js --order 10 --mode eval
vm-system template add-startup <id> --path runtime/globals.js --order 20 --mode eval

# Add a native module for database access
vm-system template add-module <id> --name database

# Add a library — download to the local cache first
vm-system libs download
vm-system template add-library <id> --name lodash-4.17.21

# See the complete template with all its configuration
vm-system template get <id>
```

### What you get by default

Every new template starts with sensible defaults so you can start using it
immediately:

- **Limits:** 5 seconds CPU time, 30 seconds wall time, 128MB memory, up to
  1000 events per execution, up to 512KB of output
- **Resolver:** recognizes `.js` and `.mjs` extensions, absolute imports
  disabled
- **Runtime:** console shim enabled (so `console.log` captures events), ESM
  and strict mode off

These settings are stored as JSON blobs in the database, which means adding
new settings fields is a one-line code change — no schema migrations needed.

### Modules and libraries — two kinds of extensions

vm-system distinguishes between native modules and JavaScript libraries, and
it's worth understanding why.

**Native modules** are Go code that's exposed to the JavaScript runtime. They
provide capabilities that pure JavaScript can't offer — like talking to SQLite,
running shell commands, or reading files from disk. Because these modules give
the runtime access to host resources, they must be explicitly enabled per
template. Three are available today:

- `database` — SQLite access with configure, query, exec, and close
- `exec` — run external shell commands and get the output
- `fs` — read and write files on the host filesystem

**Libraries** are plain JavaScript files. They're downloaded from CDN to a
local cache directory (`.vm-cache/libraries/`) and loaded into the runtime's
global scope at session startup. The built-in catalog includes popular libraries
like lodash, moment, axios, ramda, dayjs, and zustand.

One important gotcha: **JavaScript built-ins like JSON, Math, and Date are
always available.** They're part of the language, not something you configure
per template. If you try to add them as modules, you'll get
`MODULE_NOT_ALLOWED` — the system is telling you they're already there.

### Template deletion

When you delete a template, the deletion cascades to its settings, capabilities,
and startup files through database foreign keys. But sessions that were already
created from that template keep running — their goja runtime is already
allocated in memory and doesn't depend on the template anymore.

## Sessions — live runtimes

A session is a running goja instance. The daemon holds it in memory, and it
persists across execution requests until you explicitly close it. This is what
makes vm-system different from a one-shot script runner — you can set up state
in one execution and use it in the next.

### Lifecycle

When you create a session, a lot happens behind the scenes:

```
  Creating session
       │
       ▼
  starting ─── allocate goja runtime, apply settings,
       │       load libraries, execute startup files in order
       │
       ├──► ready    (success — waiting for executions)
       │
       └──► crashed  (something failed — check last_error)

  ready ───────► closed   (you called session close)
```

The daemon follows these steps:

1. **Looks up the template** and its settings in the database.
2. **Allocates a fresh goja runtime** — a clean JavaScript environment.
3. **Configures the console shim** so that `console.log`, `console.warn`, etc.
   are captured as events instead of being lost.
4. **Loads configured libraries** from the local cache into the global scope.
5. **Executes startup files** in ascending `order_index` order. Each file runs
   in the same runtime, so earlier files can set up state that later files use.
6. **If everything succeeds** → the session moves to `ready` and is available
   for execution requests.
7. **If anything throws** → the session moves to `crashed` and the error
   message is saved in `last_error`.

### Creating a session

```bash
vm-system session create \
  --template-id <id> \
  --workspace-id ws-prod \
  --base-commit abc123 \
  --worktree-path /absolute/path/to/code
```

All four parameters are required, and each serves a specific purpose:

- **template-id** — determines which settings, startup files, modules, and
  libraries the session gets.
- **workspace-id** — a logical identifier you choose for your own tracking.
  The system stores it but doesn't interpret it.
- **base-commit** — a git commit OID for reproducibility context. Like
  workspace-id, this is metadata for your use.
- **worktree-path** — the directory on disk where `exec run-file` resolves
  paths. Must exist and must be absolute. This is the sandbox boundary for
  file execution.

### Things to know about sessions

**State carries across executions.** This is the key feature that makes
sessions useful. If you run `var x = 42` in one REPL call, `x` is still `42`
in the next call. Startup file mutations to `globalThis` are visible to every
later execution. This makes it possible to set up a rich environment once and
then run many operations against it.

**One execution at a time.** Each session has a mutex. If you fire two
execution requests concurrently, one succeeds and the other gets
`SESSION_BUSY` (409) immediately. There's no wait queue — the client is
responsible for retrying. This is a deliberate design choice: goja runtimes
are single-threaded, and serializing concurrent executions would add
complexity without real concurrency benefits.

**Daemon restart loses runtimes.** Session rows survive in the database, but
the in-memory goja runtime is gone. You'll need to create new sessions after
a restart. The database rows remain as historical records, but there's nothing
to reconnect them to. This is a known limitation — daemon restart recovery is
planned but not yet implemented.

## Executions — running code

An execution is a single REPL snippet or file run inside a session. It
produces a persisted record and a stream of typed events that capture
everything that happened — console output, return values, exceptions, and
timing.

### Three kinds of executions

- **repl** — an inline JavaScript string you pass on the command line or via
  the API. Great for interactive exploration and quick queries.
- **run_file** — a file path relative to the session worktree. Used for
  running scripts, pipeline steps, or any code that's more than a one-liner.
- **startup** — a startup file executed automatically during session creation.
  You don't trigger these directly — they happen as part of session setup.

### Lifecycle

```
  running ──► ok        (expression returned a value)
         ──► error     (JavaScript exception was thrown)
         ──► timeout   (resource limit exceeded)
```

Most executions end in `ok` or `error`. The `timeout` state happens when the
code exceeds the CPU or wall time limits configured on the template.

### REPL execution

```bash
vm-system exec repl <session-id> 'Math.random()'
```

The snippet runs in the session runtime. The return value becomes a `value`
event with a type, a human-readable preview, and an optional JSON
representation. Any `console.log` calls become `console` events. If the code
throws, you get an `exception` event with the message and stack trace.

One subtlety: if your REPL input is a statement like `var x = 1`, the return
value is `undefined` because statements don't produce values in JavaScript.
To see a value, end with an expression: `var x = 1; x`.

### File execution

```bash
vm-system exec run-file <session-id> scripts/transform.js
```

The path is validated before the file is even read:

- It must be **relative** — no leading `/`
- It must stay **inside the worktree** — `../` is rejected with `INVALID_PATH`
- It must point to a **file that exists** — otherwise you get `FILE_NOT_FOUND`

These checks happen before any JavaScript runs, so path traversal attacks
can't reach the executor.

### Events — the execution trace

Every execution produces a stream of events, each with a sequential `seq`
number. Events are the primary way to see what happened during an execution:

```bash
vm-system exec events <execution-id> --after-seq 0
```

A typical REPL execution that calls `console.log` and returns a value looks
like this:

```
  seq 1  input_echo   "console.log('hi'); 42"
  seq 2  console      {"level":"log", "text":"hi"}
  seq 3  value        {"type":"number", "preview":"42", "json":42}
```

The `--after-seq` parameter enables cursor-based pagination. Pass the `seq` of
the last event you've seen, and you'll get only newer events. This is how you
poll for output in automation scripts without re-fetching everything.

### Resource limits

Templates define limits that protect the host from runaway scripts:

- **max_events** — caps how many events one execution can produce. A script
  that spams `console.log` in a tight loop will hit this limit.
- **max_output_kb** — caps the total output payload size across all events.
- **cpu_ms / wall_ms** — caps how long the code can run.

Exceeding any limit produces `OUTPUT_LIMIT_EXCEEDED` (422). If your scripts
are chatty, you have two options: reduce the output volume, or create a
template with higher limits for that use case.

## Putting it all together

A typical workflow chains all three concepts. Here's a concrete example that
sets up a database-backed runtime, runs some pipeline steps, inspects the
results, and cleans up:

```bash
# 1. Define the runtime — what capabilities does it need?
vm-system template create --name etl-runner --engine goja
vm-system template add-module <id> --name database
vm-system template add-startup <id> --path runtime/setup-db.js --order 10 --mode eval

# 2. Create a session bound to a specific workspace
vm-system session create --template-id <id> --workspace-id etl-ws \
  --base-commit abc123 --worktree-path /data/etl

# 3. Run code — each execution builds on the state left by the previous one
vm-system exec repl <session-id> 'db.query("SELECT count(*) FROM users")'
vm-system exec run-file <session-id> transforms/normalize.js
vm-system exec run-file <session-id> transforms/export.js

# 4. Look at what happened
vm-system exec list <session-id>
vm-system exec events <last-execution-id> --after-seq 0

# 5. Done — close the session to free the runtime
vm-system session close <session-id>
```

## See Also

- `vm-system help getting-started` — hands-on walkthrough from build to close
- `vm-system help api-reference` — endpoint contracts for all of the above
- `vm-system help cli-command-reference` — every flag and argument
