---
Title: "vm-system Examples"
Slug: examples
Short: "Practical recipes: REPL sessions, file execution, library usage, database access, and automation patterns."
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

This page contains self-contained, runnable examples for common vm-system
use cases. Each example assumes a running daemon and starts from scratch.

## Example 1 — Minimal REPL session

The simplest possible workflow: create a template, create a session, run code.

```bash
# Create template
vm-system template create --name minimal --engine goja
# Output: Created template: minimal (ID: <TEMPLATE_ID>)

# Create worktree
mkdir -p /tmp/vm-minimal

# Create session
vm-system session create \
  --template-id <TEMPLATE_ID> \
  --workspace-id ws \
  --base-commit HEAD \
  --worktree-path /tmp/vm-minimal

# Execute
vm-system exec repl <SESSION_ID> '2 + 2'
# Result: value event with preview "4"

vm-system exec repl <SESSION_ID> 'JSON.stringify({hello: "world"})'
# Result: value event with preview '{"hello":"world"}'

# Cleanup
vm-system session close <SESSION_ID>
```

## Example 2 — Startup files with persistent state

Startup files run during session creation and set up global state:

```bash
# Create worktree with startup script
mkdir -p /tmp/vm-startup/runtime
cat > /tmp/vm-startup/runtime/init.js <<'JS'
globalThis.config = {
  appName: "my-app",
  version: "1.0.0",
  debug: true
}
console.log("Config initialized:", JSON.stringify(config))
JS

# Create and configure template
vm-system template create --name with-startup --engine goja
vm-system template add-startup <TEMPLATE_ID> \
  --path runtime/init.js --order 10 --mode eval

# Create session (init.js runs automatically)
vm-system session create \
  --template-id <TEMPLATE_ID> \
  --workspace-id ws \
  --base-commit HEAD \
  --worktree-path /tmp/vm-startup

# Startup state is available in subsequent executions
vm-system exec repl <SESSION_ID> 'config.appName'
# Result: "my-app"

vm-system exec repl <SESSION_ID> 'config.version'
# Result: "1.0.0"
```

## Example 3 — Running files from a worktree

Execute scripts from a project directory:

```bash
mkdir -p /tmp/vm-project/scripts

cat > /tmp/vm-project/scripts/fibonacci.js <<'JS'
function fib(n) {
  if (n <= 1) return n
  return fib(n - 1) + fib(n - 2)
}
var result = []
for (var i = 0; i < 10; i++) {
  result.push(fib(i))
}
console.log("Fibonacci:", JSON.stringify(result))
result
JS

# Create template and session
vm-system template create --name project --engine goja
vm-system session create \
  --template-id <TEMPLATE_ID> \
  --workspace-id ws \
  --base-commit HEAD \
  --worktree-path /tmp/vm-project

# Run the file
vm-system exec run-file <SESSION_ID> scripts/fibonacci.js
# Console event: "Fibonacci: [0,1,1,2,3,5,8,13,21,34]"
# Value event: [0,1,1,2,3,5,8,13,21,34]
```

## Example 4 — Stateful REPL workflow

Variables persist across REPL calls within the same session:

```bash
vm-system exec repl <SESSION_ID> 'var users = []'
vm-system exec repl <SESSION_ID> 'users.push({name: "Alice", age: 30}); users.length'
# Result: 1

vm-system exec repl <SESSION_ID> 'users.push({name: "Bob", age: 25}); users.length'
# Result: 2

vm-system exec repl <SESSION_ID> 'users.filter(u => u.age > 28)'
# Result: [{name: "Alice", age: 30}]
```

## Example 5 — Inspecting execution events

Every execution produces a sequence of typed events:

```bash
# Run code that produces multiple event types
vm-system exec repl <SESSION_ID> 'console.log("hello"); console.warn("careful"); 42'

# Fetch all events
vm-system exec events <EXECUTION_ID> --after-seq 0
```

Expected event sequence:

| seq | type | content |
|-----|------|---------|
| 1 | `input_echo` | The input code |
| 2 | `console` | `{"level":"log","text":"hello"}` |
| 3 | `console` | `{"level":"warn","text":"careful"}` |
| 4 | `value` | `{"type":"number","preview":"42","json":42}` |

Use `--after-seq 3` to get only events after seq 3 (cursor-based pagination).

## Example 6 — Multiple startup files with ordering

Control initialization order with `order_index`:

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

vm-system exec repl <SESSION_ID> 'log'
# Result: ["base loaded", "extensions loaded"]

vm-system exec repl <SESSION_ID> 'greet("World")'
# Result: "Hello, World"
```

## Example 7 — Template with lodash library

Use third-party libraries for richer JavaScript capabilities:

```bash
# Download libraries to cache first
vm-system libs download

# Create template with lodash
vm-system template create --name with-lodash --engine goja
vm-system template add-library <TEMPLATE_ID> --name lodash-4.17.21

# Create session (lodash loaded during startup)
vm-system session create --template-id <TEMPLATE_ID> \
  --workspace-id ws --base-commit HEAD \
  --worktree-path /tmp/vm-lodash

# Use lodash
vm-system exec repl <SESSION_ID> '_.chunk([1,2,3,4,5,6], 2)'
# Result: [[1,2],[3,4],[5,6]]

vm-system exec repl <SESSION_ID> '_.groupBy(["one","two","three"], "length")'
```

## Example 8 — Error handling and debugging

See what happens when things go wrong:

```bash
# Syntax error
vm-system exec repl <SESSION_ID> 'function('
# Execution status: error, exception event with message

# Reference error
vm-system exec repl <SESSION_ID> 'undefinedVariable.method()'
# Exception event: "ReferenceError: undefinedVariable is not defined"

# Path traversal attempt
vm-system exec run-file <SESSION_ID> '../etc/passwd'
# Error: 422 INVALID_PATH "Path escapes allowed worktree"
```

## Example 9 — API-level automation with curl

For scripting and CI pipelines, use the REST API directly:

```bash
# Create template
TEMPLATE=$(curl -sS -X POST http://127.0.0.1:3210/api/v1/templates \
  -H 'Content-Type: application/json' \
  -d '{"name":"ci-runner","engine":"goja"}' | jq -r '.id')

# Create session
SESSION=$(curl -sS -X POST http://127.0.0.1:3210/api/v1/sessions \
  -H 'Content-Type: application/json' \
  -d "{\"template_id\":\"$TEMPLATE\",\"workspace_id\":\"ci\",\"base_commit_oid\":\"HEAD\",\"worktree_path\":\"/tmp/ci-ws\"}" \
  | jq -r '.id')

# Execute and extract result
RESULT=$(curl -sS -X POST http://127.0.0.1:3210/api/v1/executions/repl \
  -H 'Content-Type: application/json' \
  -d "{\"session_id\":\"$SESSION\",\"input\":\"1+1\"}" \
  | jq -r '.events[] | select(.type=="value") | .payload.preview')

echo "Result: $RESULT"

# Cleanup
curl -sS -X POST http://127.0.0.1:3210/api/v1/sessions/$SESSION/close -d '{}'
```

## Example 10 — Runtime summary monitoring

Monitor daemon state in a loop:

```bash
# Watch active sessions
watch -n 2 'curl -sS http://127.0.0.1:3210/api/v1/runtime/summary | jq .'

# Check session count before and after operations
curl -sS http://127.0.0.1:3210/api/v1/runtime/summary
# {"active_sessions":0,"active_session_ids":[]}

# ... create sessions ...

curl -sS http://127.0.0.1:3210/api/v1/runtime/summary
# {"active_sessions":2,"active_session_ids":["session-1","session-2"]}
```

## Troubleshooting

| Problem | Cause | Solution |
|---------|-------|----------|
| Examples fail to run | Daemon not started | Start `vm-system serve` first |
| `libs download` fails | Network issue or URL unreachable | Check connectivity; libraries are downloaded from CDN |
| Startup file not found | Wrong path relative to worktree | Verify the file exists at `<worktree-path>/<startup-path>` |
| Variable not defined in REPL | Typo or wrong session | Variables are per-session; check session ID |

## See Also

- `vm-system help getting-started`
- `vm-system help how-to-use`
- `vm-system help cli-command-reference`
- `vm-system help templates-and-sessions`
