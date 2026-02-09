---
Title: "Getting Started with vm-system"
Slug: getting-started
Short: "Build, run the daemon, and master the template → session → execution workflow."
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
- ops
- libs
IsTopLevel: true
IsTemplate: false
ShowPerDefault: true
SectionType: Tutorial
---

vm-system runs JavaScript in managed goja runtimes. You define a **template**
(what the runtime can do), spin up a **session** (a live runtime), and
**execute** code in it — all through a daemon that persists everything to
SQLite.

The idea is that you separate the *configuration* of a JavaScript environment
from the *use* of it. A template captures decisions like "this runtime gets
database access and lodash, runs these setup scripts, and has a 5-second CPU
limit." Then you create as many sessions as you need from that template, each
bound to a specific working directory.

This tutorial takes you from first build to confident daily use.

## What you need

- **Go** compatible with `go 1.25.5` (check `go.mod`)
- **bash** or **zsh**
- **curl** if you want to poke the API directly

## Your first run

### 1. Build

```bash
GOWORK=off go build -o vm-system ./cmd/vm-system
./vm-system --help
```

You should see six command groups: `serve`, `template`, `session`, `exec`,
`ops`, and `libs`. If you see them, you're good.

### 2. Set up a throwaway workspace

Sessions execute code from a directory on disk called a "worktree." This is
the directory that `exec run-file` resolves paths against, and it's also
where startup scripts look for their files. Create a simple one with a startup
script and a file to run later:

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

`init.js` will run automatically when you create a session — it sets a global
variable that we'll use to verify everything works. `app.js` is a script you'll
execute manually in a moment to see file execution in action.

### 3. Start the daemon

```bash
./vm-system serve --db /tmp/vm-scratch.db --listen 127.0.0.1:3210
```

This stays in the foreground — open a second terminal for the rest. The `--db`
flag controls where templates, sessions, and execution history are persisted.
If the file doesn't exist yet, it's created automatically.

Quick health check to make sure the daemon is up:

```bash
curl -sS http://127.0.0.1:3210/api/v1/health
# {"status":"ok"}
```

### 4. Create a template

A template is a saved runtime profile. It captures the engine type, resource
limits, which startup scripts to run, which native modules to enable, and which
third-party libraries to load. You create it once and then stamp out sessions
from it:

```bash
./vm-system template create --name my-first --engine goja
# → Created template: my-first (ID: <template-id>)
```

Now attach the startup file. The `--order` flag controls execution order when
you have multiple startup files — lower numbers run first:

```bash
./vm-system template add-startup <template-id> \
  --path runtime/init.js --order 10 --mode eval
```

Peek at what you created to see the full picture — default limits, runtime
settings, and the startup file:

```bash
./vm-system template get <template-id>
```

You'll see the default settings (5s CPU timeout, 128MB memory, console enabled)
plus the startup file you just added.

### 5. Create a session

A session is a live goja runtime. When you create one, the daemon allocates a
fresh JavaScript runtime, applies the template settings, and immediately
executes any startup files. After that, the session sits in memory waiting for
execution requests:

```bash
./vm-system session create \
  --template-id <template-id> \
  --workspace-id ws-demo \
  --base-commit deadbeef \
  --worktree-path /tmp/vm-scratch
```

If you see `Status: ready`, everything worked — the daemon created the goja
runtime, ran `init.js` successfully, and the variable `seed` is now `40` in
that runtime's global scope.

### 6. Run some code

Try a REPL snippet. It runs in the same runtime where `init.js` already
executed, so `seed` is available:

```bash
./vm-system exec repl <session-id> 'seed + 2'
# → 42
```

Now run the file. The path is resolved relative to the worktree you specified
when creating the session:

```bash
./vm-system exec run-file <session-id> app.js
```

You'll see console output from `app.js` and a return value of `42`. Absolute
paths and `../` are rejected for safety — the runtime can only access files
inside its worktree.

### 7. Clean up

```bash
./vm-system session close <session-id>
# Ctrl+C the daemon
```

That's it — you just completed the full template → session → execution loop.

### What happened under the hood

Every CLI command you ran was actually a REST call to the daemon:

```
  You typed a CLI command
       │
       ▼
  vmclient sent a REST request to the daemon
       │
       ▼
  HTTP handler validated the request
       │
       ▼
  vmcontrol orchestrated the operation (transport-agnostic)
       │
       ├──► vmsession managed the goja runtime
       ├──► vmexec ran the code and captured events
       └──► vmstore persisted everything to SQLite
```

The daemon holds runtimes in memory. The CLI is just a thin REST client — it
never touches the database directly. SQLite is the durable store for templates,
sessions, executions, and events.

## Daily workflows

Now that you've done it once, here are the patterns you'll use every day.

### Configuring templates

Templates are meant to accumulate configuration over time. You start with a
basic template and then layer on startup files, modules, and libraries as your
use case evolves:

```bash
vm-system template create --name my-service --engine goja
```

**Startup files** run during session creation, in the order you specify. This
is where you set up global state, configure database connections, define helper
functions, or load polyfills. You can have as many as you want:

```bash
vm-system template add-startup <id> --path runtime/polyfills.js --order 10 --mode eval
vm-system template add-startup <id> --path runtime/globals.js --order 20 --mode eval
```

**Native modules** give the runtime access to host capabilities that aren't
part of standard JavaScript. These are implemented in Go and exposed to the
goja runtime. Three are available today:

- `database` — SQLite access with configure, query, exec, and close operations
- `exec` — run external shell commands from JavaScript
- `fs` — read and write files from the filesystem

```bash
vm-system template list-available-modules
vm-system template add-module <id> --name database
```

One thing that trips people up: JavaScript built-in globals like JSON, Math,
and Date are always present in every runtime. They're part of the language, not
something you configure. If you try to add them as modules, you'll get
`MODULE_NOT_ALLOWED`.

**Third-party libraries** like lodash or moment are JavaScript files that get
loaded into the runtime at session startup. They need to be downloaded to a
local cache first:

```bash
vm-system template list-available-libraries
vm-system libs download
vm-system template add-library <id> --name lodash-4.17.21
```

### Managing sessions

Sessions are the live runtimes. Most of the time you're creating them, checking
on them, and closing them when you're done:

```bash
vm-system session list                   # see everything
vm-system session list --status ready    # just the ones you can execute in
vm-system session get <session-id>       # full detail including timestamps and errors
vm-system ops runtime-summary            # what's actually alive in memory
vm-system session close <session-id>     # discard runtime, keep DB record
```

The `runtime-summary` is worth highlighting — it shows what's actually alive in
daemon memory, not just what's in the database. After a daemon restart, the
database still has session rows but the runtimes are gone. `runtime-summary`
tells you the truth.

The worktree directory must exist before you create a session, and the path
must be absolute. If you forget either, you'll get an error immediately.

### Running code

**REPL calls are stateful** — this is one of the most useful features.
Variables you set in one call are still there in the next. This makes it easy
to build up state interactively:

```bash
vm-system exec repl <session-id> 'var counter = 0'
vm-system exec repl <session-id> 'counter += 1; counter'   # → 1
vm-system exec repl <session-id> 'counter += 1; counter'   # → 2
```

**File execution** runs a script from the session's worktree. The file can
access everything that previous executions and startup scripts set up:

```bash
vm-system exec run-file <session-id> scripts/transform.js
```

**Events** capture everything that happened during an execution — console
output, return values, exceptions, and more. Each event has a sequential `seq`
number, which makes cursor-based polling easy:

```bash
vm-system exec list <session-id> --limit 20
vm-system exec events <execution-id> --after-seq 0
```

### Talking to the API directly

Every CLI command maps to a REST endpoint. For scripting, debugging, or
building integrations, you can use curl directly. This is especially useful
when you need to automate vm-system in CI pipelines or integrate it with
other tools:

```bash
curl -sS http://127.0.0.1:3210/api/v1/templates
curl -sS "http://127.0.0.1:3210/api/v1/sessions?status=ready"
curl -sS "http://127.0.0.1:3210/api/v1/executions?session_id=<id>&limit=5"
```

## Patterns worth knowing

### ETL pipeline with database access

This pattern shows how to use native modules and file execution together. You
set up a template with database access, create a session bound to your data
directory, and then run pipeline steps as separate files — each building on
the state left by the previous one:

```bash
vm-system template create --name etl --engine goja
vm-system template add-module <id> --name database
vm-system template add-startup <id> --path runtime/db-setup.js --order 10 --mode eval

vm-system session create --template-id <id> --workspace-id etl \
  --base-commit main --worktree-path /data/etl

vm-system exec run-file <session-id> steps/extract.js
vm-system exec run-file <session-id> steps/transform.js
vm-system exec run-file <session-id> steps/load.js
vm-system exec repl <session-id> 'db.query("SELECT count(*) FROM output_table")'
vm-system session close <session-id>
```

### Interactive development with lodash

This pattern is great for exploratory work. You create a session with lodash
loaded, then use REPL calls to interactively prototype data transformations.
Because state persists across calls, you can build up complex pipelines one
step at a time:

```bash
vm-system template create --name dev --engine goja
vm-system libs download
vm-system template add-library <id> --name lodash-4.17.21

vm-system session create --template-id <id> --workspace-id dev \
  --base-commit HEAD --worktree-path /project

vm-system exec repl <session-id> '_.chunk([1,2,3,4,5,6], 2)'
vm-system exec repl <session-id> '_.groupBy(["one","two","three"], "length")'
```

## When things go wrong

**Daemon won't connect?** Make sure it's running and your `--server-url`
matches the `--listen` address. The most common issue is having the daemon on
one port and the CLI pointing at another.

**Session crashed on creation?** A startup script threw an error or a library
isn't downloaded. Run `vm-system session get <id>` and look at the `Last Error`
field — it tells you exactly what went wrong. If it's a library issue, run
`vm-system libs download` and try again.

**`SESSION_BUSY`?** Only one execution runs at a time per session. This is
by design — goja runtimes are single-threaded, and interleaving executions
would corrupt state. Wait for the current execution to finish and retry.

**`INVALID_PATH` on run-file?** The path must be relative to the worktree
and must not contain `../`. This prevents the runtime from accessing files
outside its designated directory.

**REPL returns `undefined`?** This usually means you wrote a statement
(`var x = 1`) instead of an expression. Statements don't produce a return
value in JavaScript. Add the variable at the end: `var x = 1; x`.

**Sessions gone after daemon restart?** That's expected. Runtimes live in
daemon memory and don't survive restarts. The session rows stay in the
database, but there's nothing to reconnect them to. Create new sessions
after a restart.

For a full error code reference, see `vm-system help api-reference`.

## See Also

- `vm-system help architecture` — how the code is organized and why
- `vm-system help templates-and-sessions` — deep dive on the three core concepts
- `vm-system help api-reference` — every endpoint, request shape, and error code
- `vm-system help cli-command-reference` — every command and flag
- `vm-system help examples` — more runnable recipes
- `vm-system help contributing` — how to change the code
