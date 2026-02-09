---
Title: "vm-system Examples"
Slug: examples
Short: "Runnable recipes: REPL sessions, startup files, libraries, file execution, events, and API automation."
Topics:
- vm-system
- examples
- recipes
Commands:
- template
- session
- exec
IsTopLevel: true
IsTemplate: false
ShowPerDefault: true
SectionType: Example
---

Each example below is self-contained and assumes a running daemon. If you
don't have one yet, start with `vm-system serve --db /tmp/examples.db`.

The examples are ordered from simple to complex. If you've completed the
getting-started tutorial, you can jump to whichever pattern interests you.

## Minimal REPL session

The absolute simplest workflow — create a template, create a session, execute
code, clean up. No startup files, no modules, no libraries — just the raw
goja runtime with default settings:

```bash
vm-system template create --name minimal --engine goja
mkdir -p /tmp/vm-minimal

vm-system session create \
  --template-id <TEMPLATE_ID> --workspace-id ws \
  --base-commit HEAD --worktree-path /tmp/vm-minimal

vm-system exec repl <SESSION_ID> '2 + 2'                          # → 4
vm-system exec repl <SESSION_ID> 'JSON.stringify({hello:"world"})' # → '{"hello":"world"}'

vm-system session close <SESSION_ID>
```

Even with no configuration, you get the full JavaScript language including
JSON, Math, Date, and all built-in globals.

## Startup files that set up global state

Startup files are the primary way to initialize a runtime before user code
runs. Anything a startup file puts on `globalThis` is available to every
later execution. This is useful for configuration, helper functions, database
setup, or anything that should be "just there" when you start working:

```bash
mkdir -p /tmp/vm-startup/runtime
cat > /tmp/vm-startup/runtime/init.js <<'JS'
globalThis.config = { appName: "my-app", version: "1.0.0", debug: true }
console.log("Config initialized:", JSON.stringify(config))
JS

vm-system template create --name with-startup --engine goja
vm-system template add-startup <TEMPLATE_ID> --path runtime/init.js --order 10 --mode eval

vm-system session create \
  --template-id <TEMPLATE_ID> --workspace-id ws \
  --base-commit HEAD --worktree-path /tmp/vm-startup

vm-system exec repl <SESSION_ID> 'config.appName'   # → "my-app"
vm-system exec repl <SESSION_ID> 'config.version'    # → "1.0.0"
```

## Multiple startup files with ordering

When your initialization is complex, split it across multiple files and use
`--order` to control the sequence. Lower numbers run first, so you can build
up layers — a base layer that defines data structures, then an extensions
layer that adds functions:

```bash
mkdir -p /tmp/vm-ordered/runtime

cat > /tmp/vm-ordered/runtime/01-base.js <<'JS'
globalThis.log = []
log.push("base loaded")
JS

cat > /tmp/vm-ordered/runtime/02-extensions.js <<'JS'
log.push("extensions loaded")
globalThis.greet = function(name) { return "Hello, " + name }
JS

vm-system template create --name ordered --engine goja
vm-system template add-startup <ID> --path runtime/01-base.js --order 10 --mode eval
vm-system template add-startup <ID> --path runtime/02-extensions.js --order 20 --mode eval

vm-system session create --template-id <ID> --workspace-id ws \
  --base-commit HEAD --worktree-path /tmp/vm-ordered

vm-system exec repl <SESSION_ID> 'log'               # → ["base loaded", "extensions loaded"]
vm-system exec repl <SESSION_ID> 'greet("World")'    # → "Hello, World"
```

## Running files from a worktree

For anything more than a one-liner, put the code in a file and use
`exec run-file`. This is how you'd run pipeline steps, data transformations,
or test scripts. The file executes in the same runtime as previous REPL calls
and startup files, so it has access to all existing state:

```bash
mkdir -p /tmp/vm-project/scripts

cat > /tmp/vm-project/scripts/fibonacci.js <<'JS'
function fib(n) { return n <= 1 ? n : fib(n-1) + fib(n-2) }
var result = []
for (var i = 0; i < 10; i++) result.push(fib(i))
console.log("Fibonacci:", JSON.stringify(result))
result
JS

vm-system template create --name project --engine goja
vm-system session create \
  --template-id <TEMPLATE_ID> --workspace-id ws \
  --base-commit HEAD --worktree-path /tmp/vm-project

vm-system exec run-file <SESSION_ID> scripts/fibonacci.js
# console: "Fibonacci: [0,1,1,2,3,5,8,13,21,34]"
# value:   [0,1,1,2,3,5,8,13,21,34]
```

## Stateful REPL — building up data across calls

One of the most useful features of vm-system is that REPL calls are stateful.
Variables, functions, and objects you create in one call persist into the next.
This makes the REPL great for exploratory data work — build up a dataset
step by step and query it interactively:

```bash
vm-system exec repl <SESSION_ID> 'var users = []'
vm-system exec repl <SESSION_ID> 'users.push({name: "Alice", age: 30}); users.length'  # → 1
vm-system exec repl <SESSION_ID> 'users.push({name: "Bob", age: 25}); users.length'    # → 2
vm-system exec repl <SESSION_ID> 'users.filter(u => u.age > 28)'  # → [{name:"Alice",age:30}]
```

## Inspecting the event stream

Every execution produces typed events with sequential `seq` numbers. Events
are how you see exactly what happened — not just the return value, but every
console call, in order. This is especially useful for debugging or for
building automation that reacts to specific output:

```bash
vm-system exec repl <SESSION_ID> 'console.log("hello"); console.warn("careful"); 42'
vm-system exec events <EXECUTION_ID> --after-seq 0
```

You'll see something like:

```
seq 1  input_echo   console.log("hello"); console.warn("careful"); 42
seq 2  console      {"level":"log","text":"hello"}
seq 3  console      {"level":"warn","text":"careful"}
seq 4  value        {"type":"number","preview":"42","json":42}
```

The `--after-seq` parameter is the key to polling. If you pass `--after-seq 3`,
you get only events after seq 3. This means you can poll periodically without
re-fetching the entire history — just remember the last `seq` you saw.

## Using lodash

Third-party libraries are downloaded to a local cache and then loaded into
the runtime at session startup. This example shows lodash, but the same
pattern works for moment, axios, ramda, dayjs, and zustand:

```bash
vm-system libs download
vm-system template create --name with-lodash --engine goja
vm-system template add-library <TEMPLATE_ID> --name lodash-4.17.21

mkdir -p /tmp/vm-lodash
vm-system session create --template-id <TEMPLATE_ID> \
  --workspace-id ws --base-commit HEAD --worktree-path /tmp/vm-lodash

vm-system exec repl <SESSION_ID> '_.chunk([1,2,3,4,5,6], 2)'
# → [[1,2],[3,4],[5,6]]

vm-system exec repl <SESSION_ID> '_.groupBy(["one","two","three"], "length")'
# → {"3":["one","two"],"5":["three"]}
```

## Error handling and debugging

It's worth knowing what different failure modes look like so you can debug
them quickly. Here are the most common ones:

```bash
# Syntax error — you get an exception event with the parse error
vm-system exec repl <SESSION_ID> 'function('
# → exception: "SyntaxError: Unexpected token )"

# Reference error — the variable doesn't exist in this session
vm-system exec repl <SESSION_ID> 'undefinedVar.method()'
# → exception: "ReferenceError: undefinedVar is not defined"

# Path traversal — blocked before any JavaScript runs
vm-system exec run-file <SESSION_ID> '../etc/passwd'
# → 422 INVALID_PATH: "Path escapes allowed worktree"
```

For all of these, the event stream has the details. Run
`vm-system exec events <execution-id> --after-seq 0` to see the exception
message and stack trace.

## API automation with curl

When you need to integrate vm-system into CI pipelines, monitoring scripts,
or other tools, the REST API is the way to go. This example shows the
complete template → session → execute → cleanup flow using only curl and jq:

```bash
TEMPLATE=$(curl -sS -X POST http://127.0.0.1:3210/api/v1/templates \
  -H 'Content-Type: application/json' \
  -d '{"name":"ci-runner","engine":"goja"}' | jq -r '.id')

mkdir -p /tmp/ci-ws
SESSION=$(curl -sS -X POST http://127.0.0.1:3210/api/v1/sessions \
  -H 'Content-Type: application/json' \
  -d "{\"template_id\":\"$TEMPLATE\",\"workspace_id\":\"ci\",
       \"base_commit_oid\":\"HEAD\",\"worktree_path\":\"/tmp/ci-ws\"}" \
  | jq -r '.id')

RESULT=$(curl -sS -X POST http://127.0.0.1:3210/api/v1/executions/repl \
  -H 'Content-Type: application/json' \
  -d "{\"session_id\":\"$SESSION\",\"input\":\"1+1\"}" \
  | jq -r '.events[] | select(.type=="value") | .payload.preview')

echo "Result: $RESULT"   # → Result: 2

curl -sS -X POST "http://127.0.0.1:3210/api/v1/sessions/$SESSION/close" -d '{}'
```

## Monitoring with runtime-summary

The runtime summary endpoint tells you exactly what's alive in the daemon's
memory. This is useful for monitoring dashboards, health checks, and for
verifying that sessions were properly created or cleaned up:

```bash
# One-shot check
curl -sS http://127.0.0.1:3210/api/v1/runtime/summary | jq .
# {"active_sessions":0,"active_session_ids":[]}

# Continuous monitoring (updates every 2 seconds)
watch -n 2 'curl -sS http://127.0.0.1:3210/api/v1/runtime/summary | jq .'
```

## See Also

- `vm-system help getting-started` — full walkthrough from build to close
- `vm-system help cli-command-reference` — every flag and argument
- `vm-system help templates-and-sessions` — deeper on the core concepts
