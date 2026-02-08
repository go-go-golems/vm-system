---
Title: Comprehensive vm-system analysis report
Ticket: VM-001-ANALYZE-VM
Status: active
Topics:
    - backend
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: vm-system/vm-system/cmd/vm-system/cmd_exec.go
      Note: Execution CLI behavior and argument contracts
    - Path: vm-system/vm-system/cmd/vm-system/cmd_session.go
      Note: Session lifecycle CLI and flags
    - Path: vm-system/vm-system/cmd/vm-system/cmd_vm.go
      Note: VM profile command surface and settings contract
    - Path: vm-system/vm-system/pkg/libloader/loader.go
      Note: Library download/cache naming strategy
    - Path: vm-system/vm-system/pkg/vmexec/executor.go
      Note: REPL/run-file execution and event capture
    - Path: vm-system/vm-system/pkg/vmsession/session.go
      Note: Session runtime lifecycle and in-memory map behavior
    - Path: vm-system/vm-system/pkg/vmstore/vmstore.go
      Note: SQLite schema and persistence semantics
    - Path: vm-system/vm-system/smoke-test.sh
      Note: Operational script drift and stale command usage
    - Path: vm-system/vm-system/test-goja-library-execution.sh
      Note: Stale flags and e2e mismatch
ExternalSources: []
Summary: Deep architecture and runtime analysis of vm-system with experiment-backed findings, strengths, risks, and an implementation roadmap.
LastUpdated: 2026-02-07T19:45:00-05:00
WhatFor: Understand what vm-system is for, how it works, what is strong, what is risky, and how to improve it pragmatically.
WhenToUse: Use when planning vm-system stabilization, runtime architecture changes, and test/tooling cleanup.
---


# Comprehensive vm-system analysis report

## Executive Summary

vm-system is a Go-based JavaScript virtual machine control layer built around three major ideas: persistent VM/session/execution metadata in SQLite, runtime execution via goja, and a CLI surface for managing profiles/sessions/executions. The core architecture is compact and readable. The package split is sensible (`cmd/` for CLI composition, `pkg/vmstore` for persistence, `pkg/vmsession` for session lifecycle, `pkg/vmexec` for code execution). For a reference implementation, this is a strong foundation.

The current implementation, however, behaves like a single-process prototype rather than a production-grade execution subsystem. The most consequential gap is process-scoped runtime state: sessions are stored in SQLite, but active goja runtimes are kept only in an in-memory map inside `SessionManager`. As soon as a new CLI invocation starts, it constructs a fresh `SessionManager` with an empty session map, and `exec repl` / `exec run-file` cannot find previously created sessions. In short: session metadata survives, runtime handles do not.

A second high-impact defect is library cache path mismatch. The downloader writes files as `<id>-<version>.js` (for example `lodash-4.17.21.js`), while session runtime loading expects `<id>.js` (for example `lodash.js`). This guarantees session creation failure when libraries are enabled, even after successful downloads.

A third class of issues is policy and contract drift. The data model advertises limits, capability allowlists, and startup modes, but several paths either do not enforce these constraints or treat them as metadata only. In parallel, shell scripts in repository root use stale command flags/subcommands and no longer align with current Cobra command contracts.

Despite these risks, the system has a clear path forward. The design already contains the right nouns (VM profile, session, execution, event stream, startup files, capabilities). A focused phase of runtime model hardening, policy enforcement, and workflow cleanup can turn this into a robust backend core.

## Problem Statement

The immediate problem is not lack of features; it is a mismatch between intended behavior and actual runtime guarantees.

- Intended behavior (as implied by docs and command structure):
- A user creates a VM profile and session, then executes multiple snippets/files over time.
- Capabilities/modules/libraries/settings meaningfully constrain runtime behavior.
- Execution lifecycle is observable and auditable through persistent events.
- Operational scripts provide quick confidence checks.

- Actual behavior observed:
- Session records persist, but runtime objects do not survive process boundaries.
- Library-enabled sessions can fail because loader and downloader disagree on filename contracts.
- Limits/capabilities are partially metadata-only in the execution path.
- Scripted smoke/e2e flows contain stale flags and fail against current CLI.

This gap creates a reliability cliff: users can successfully configure and create resources, then hit hard failures during execution workflows that appear valid.

## Proposed Solution

The solution is to treat vm-system as two layers with explicit contracts:

1. Control-plane persistence (already reasonably strong): VM/session/execution metadata in SQLite.
2. Runtime-plane execution host (currently weak): process lifecycle, runtime ownership, policy enforcement, and reproducibility.

The near-term architecture should evolve toward a long-lived runtime host process (daemon or server mode) that owns all live session runtimes and exposes command/API operations. The CLI then becomes a client or thin in-process wrapper for one-shot operations.

### Architectural Flow (Current)

```text
+-------------------+        +------------------+        +----------------------+
| CLI invocation #1 | -----> | SessionManager A | -----> | sessions map (memory) |
+-------------------+        +------------------+        +----------------------+
         |                              |
         |                              +--> SQLite row in vm_session (durable)
         v
   session create (works)

+-------------------+        +------------------+        +----------------------+
| CLI invocation #2 | -----> | SessionManager B | -----> | sessions map (empty)  |
+-------------------+        +------------------+        +----------------------+
         |
         v
   exec repl -> session not found
```

### Architectural Flow (Target)

```text
+------------------+       RPC/HTTP/IPC       +-------------------------+
| vm-system CLI    | <-----------------------> | vm-system runtime host |
| (client mode)    |                           | (long-lived process)   |
+------------------+                           +-------------------------+
                                                          |
                                                          v
                                             +---------------------------+
                                             | SessionManager (single)   |
                                             | goja runtime ownership    |
                                             +---------------------------+
                                                          |
                                                          v
                                             +---------------------------+
                                             | SQLite persistence        |
                                             | vm / session / execution  |
                                             +---------------------------+
```

### Policy Enforcement Model (Target)

```pseudo
function execute(sessionID, input):
    session = runtimeHost.getLiveSession(sessionID)
    profile = store.getVM(session.vmID)
    settings = store.getVMSettings(session.vmID)

    enforceSessionReady(session)
    enforceLimitsPreflight(settings.limits, input)

    sandbox = buildSandbox(profile.capabilities, profile.modules, profile.libraries)
    output = sandbox.run(input, timeout=settings.limits.wall_ms)

    metrics = measure(output)
    enforceOutputLimits(metrics, settings.limits)

    persistExecutionAndEvents(output, metrics)
    return output
```

## Design Decisions

### Decision 1: Keep SQLite as canonical control-plane store

SQLite schema in `vm-system/pkg/vmstore/vmstore.go` is compact, understandable, and already models required entities (`vm`, `vm_settings`, `vm_session`, `execution`, `execution_event`). This is a strength worth preserving.

### Decision 2: Separate runtime ownership from command process lifecycle

The root cause of session reuse failure is not SQL design; it is runtime ownership in ephemeral CLI processes. A single runtime host process is required for stable session continuity.

### Decision 3: Normalize library cache contract

Downloader and runtime loader must share one filename manifest strategy. Either:

- Store and load by exact `<id>-<version>.js` names.
- Or persist resolved cache path in DB at configuration time.

The current split contract is unsound.

### Decision 4: Treat scripts as product surface, not disposable artifacts

Root scripts currently encode stale CLI contracts. They should be maintained as first-class compatibility checks and generated/verified against Cobra command help where practical.

## Deep Architecture Walkthrough

### 1. CLI Composition Layer

- `vm-system/cmd/vm-system/main.go` assembles root command and subcommands (`vm`, `session`, `exec`, `modules`, `libs`).
- `cmd_vm.go` handles profile creation/settings/capabilities/startup files.
- `cmd_session.go` handles session lifecycle operations.
- `cmd_exec.go` handles repl/run-file/list/get/events.

Strengths:

- Command tree is discoverable and straightforward.
- Argument parsing is explicit and readable.

Weaknesses:

- Global singletons (`sessionManager`, `executor`) are process-scoped conveniences, not durable runtime infrastructure.

### 2. Data Model Layer

`vm-system/pkg/vmmodels/models.go` defines a rich domain vocabulary:

- `VM`, `VMSettings`, `VMCapability`, `VMStartupFile`
- `VMSession`, `Execution`, `ExecutionEvent`
- Named errors and enumerated status constants

Strengths:

- Models are expressive enough for future policy enforcement.
- Event model supports auditability.

Weaknesses:

- Some model fields imply guarantees not yet enforced in execution path (limits, resolver constraints, capability deny-by-default semantics).

### 3. Persistence Layer

`vm-system/pkg/vmstore/vmstore.go` centralizes schema and CRUD behavior.

Strengths:

- Schema contains useful referential constraints and indexes.
- Event stream persistence model is clear.

Weaknesses:

- No migration/version mechanism; schema is implicit and mutable in-place.
- JSON marshalling/unmarshalling errors are occasionally ignored (`json.Unmarshal` return values dropped in VM reads).

### 4. Session Runtime Layer

`vm-system/pkg/vmsession/session.go` creates and tracks live sessions.

Strengths:

- Runtime initialization path is easy to follow.
- Startup file execution order is deterministic by `order_index`.

Weaknesses:

- Live sessions only in memory map -> process boundary breakage.
- Library loading path mismatch (`<id>.js` expectation vs versioned files).
- `import` startup mode currently aliases to eval with comment TODO.

### 5. Execution Layer

`vm-system/pkg/vmexec/executor.go` executes code and records events.

Strengths:

- Event capture for console/input/value/exception is well structured.
- Session-level mutex (`TryLock`) correctly prevents concurrent execution per session.

Weaknesses:

- Capability/module enforcement not integrated into execution path.
- Limits (CPU/wall/memory/output/event count) are modeled but not enforced.
- Errors from `e.store.AddEvent` and `e.store.UpdateExecution` are largely ignored.
- `consoleOutput` buffer in `ExecuteREPL` is accumulated but unused.

## How vm-system Is Built And Used

### Build

From module root (`vm-system/vm-system`):

```bash
GOWORK=off go build -o ./vm-system ./cmd/vm-system
```

`GOWORK=off` is currently necessary in this workspace because the top-level `go.work` does not include this module.

### Typical current usage pattern

```bash
./vm-system --db my.db vm create --name demo --engine goja
./vm-system --db my.db session create --vm-id <id> --workspace-id ws --base-commit abc --worktree-path /tmp/ws
./vm-system --db my.db exec repl <session-id> '1+2'
```

Reality today: the third command fails in a fresh process with `session not found` because runtime state is not reconstructed.

## Experimental Findings (Observed)

### Experiment A: Module/build health

Command:

```bash
GOWORK=off go test ./...
```

Result:

- Succeeds, but packages report `no test files`.
- This indicates low automated behavioral coverage despite significant runtime logic.

### Experiment B: Session continuity across commands

Sequence:

1. Create VM.
2. Create session with valid startup file and worktree.
3. In separate command invocation, execute repl/run-file.

Observed:

- `session create` prints ready status.
- `exec repl` and `exec run-file` return `Error: session not found`.

Interpretation:

- Persistent session record exists in SQLite, but live runtime object is absent in new process.

### Experiment C: Library loading with enabled lodash

Sequence:

1. `libs download` successfully fetches versioned files (`lodash-4.17.21.js`, etc.).
2. VM configured with library `lodash`.
3. `session create` attempted.

Observed failure:

- `library lodash not found in cache ... stat .vm-cache/libraries/lodash.js: no such file or directory`

Interpretation:

- Runtime loader filename contract diverges from downloader contract.

### Experiment D: Script compatibility

- `smoke-test.sh` fails at build due go.work assumptions.
- `test-goja-library-execution.sh` uses stale flags (`--worktree`, `--session-id`, `--file`).
- `smoke-test.sh` uses obsolete subcommands (`add-startup-file`, `list-startup-files`).

## Strengths (What Is Good)

- Clear package boundaries and readable command architecture.
- Strong starting schema design for session/execution/event persistence.
- Useful event model that can support debugging and audit trails.
- Good foundation for module/library metadata exposure.
- Straightforward developer ergonomics for profile/session management commands.

## Problems And Improvement Opportunities

## Runtime Continuity Defect

### Issue 1: Sessions are not reusable across CLI process boundaries

Problem: Session metadata persists, but runtime handles are process-local and lost between invocations.

Where to look:

- `vm-system/cmd/vm-system/cmd_exec.go:47`
- `vm-system/cmd/vm-system/cmd_exec.go:332`
- `vm-system/pkg/vmsession/session.go:149`

Example:

```go
if sessionManager == nil {
    sessionManager = getSessionManager(store)
}
...
func getSessionManager(store *vmstore.VMStore) *vmsession.SessionManager {
    return vmsession.NewSessionManager(store)
}
```

Why it matters:

- Core REPL/run-file workflow appears functional but fails in normal CLI usage.

Cleanup sketch:

```pseudo
if daemonMode:
    resolve session in long-lived runtime host
else:
    reject execution with explicit message:
    "session runtime unavailable in one-shot mode; start runtime host"
```

## Library Loading Contract Defect

### Issue 2: Cache filename mismatch between download and runtime load

Problem: Downloader writes `<id>-<version>.js`, runtime expects `<id>.js`.

Where to look:

- `vm-system/pkg/libloader/loader.go:80`
- `vm-system/pkg/vmsession/session.go:253`

Example:

```go
filename := fmt.Sprintf("%s-%s.js", lib.ID, lib.Version) // downloader
...
libPath := filepath.Join(cacheDir, libName+".js") // runtime loader
```

Why it matters:

- Library-enabled sessions fail deterministically.

Cleanup sketch:

```pseudo
buildLibraryManifest(): map[id]versionedPath
persist manifest in cache metadata
loadLibraries(vm): resolve by manifest[id]
```

## Policy Enforcement Gap

### Issue 3: Capability and limits model is present but partially non-operative

Problem: Execution path does not fully enforce capabilities/limits represented in settings and capability tables.

Where to look:

- `vm-system/pkg/vmexec/executor.go:1034`
- `vm-system/pkg/vmstore/vmstore.go:255`
- `vm-system/pkg/vmmodels/models.go:43`

Example:

```go
context := map[string]any{console, Math, Date, Array, Object, String, Number, Boolean, JSON}
```

Why it matters:

- Security and resource guarantees are weaker than API surface suggests.

Cleanup sketch:

```pseudo
caps = store.listCapabilities(vmID)
context = buildContextFromCapabilities(caps)
enforce max_events / max_output / wall timeout per execution
```

## Script Drift And Operational Reliability

### Issue 4: Existing scripts diverge from current CLI contract

Problem: Several scripts use removed/renamed flags and subcommands.

Where to look:

- `vm-system/smoke-test.sh:247`
- `vm-system/smoke-test.sh:257`
- `vm-system/test-goja-library-execution.sh:97`
- `vm-system/test-goja-library-execution.sh:121`

Example:

```bash
vm add-startup-file ...
vm list-startup-files ...
exec run-file --session-id ... --file ...
session create --worktree ...
```

Why it matters:

- Smoke tests can no longer be treated as trustworthy verification.

Cleanup sketch:

```pseudo
generate-cli-contract-json from cobra help
lint scripts against contract
fail CI if stale flags/subcommands detected
```

## Error Handling And Data Integrity Hygiene

### Issue 5: Ignored errors in critical persistence/event paths

Problem: errors from store writes are frequently dropped; failed event writes may silently degrade audit trail.

Where to look:

- `vm-system/pkg/vmexec/executor.go:96`
- `vm-system/pkg/vmexec/executor.go:138`
- `vm-system/pkg/vmexec/executor.go:171`

Why it matters:

- Observability and consistency can degrade silently under I/O issues.

Cleanup sketch:

```pseudo
if err := store.AddEvent(...); err != nil:
    append fallback in-memory error event
    mark execution as degraded_observability
```

## Alternatives Considered

### Alternative A: Keep pure CLI process model and reconstruct runtime lazily from DB

Rejected for now. Reconstructing goja runtime deterministically from session metadata/startup files per command is possible, but brittle for stateful REPL semantics and expensive for repeated execution.

### Alternative B: Persist runtime snapshots in DB

Rejected for now. goja snapshot persistence is non-trivial and introduces serialization complexity before core runtime host model is stable.

### Alternative C: Remove session abstraction and make everything stateless

Rejected. Session abstraction is valuable and already reflected in schema and command model.

## Implementation Plan

### Phase 1: Correctness Hotfixes (High Priority)

1. Fix library cache contract.
2. Add explicit one-shot mode guardrail message for exec commands when live session runtime unavailable.
3. Update stale scripts to current CLI contract.
4. Add minimal integration tests for session create + exec behavior expectations.

### Phase 2: Runtime Host Introduction (High Priority)

1. Introduce daemon mode owning a singleton `SessionManager`.
2. Move `exec` commands to host-backed RPC/HTTP calls.
3. Add lifecycle management (startup/shutdown, health endpoint, graceful close).

### Phase 3: Policy Enforcement (Medium Priority)

1. Enforce capability/module allowlist in runtime context construction.
2. Enforce wall timeout and output/event limits.
3. Emit structured metrics in `Execution.Metrics`.

### Phase 4: Quality and Operational Maturity (Medium Priority)

1. Add migration/version table for schema evolution.
2. Expand tests for error paths and script compatibility.
3. Improve documentation to separate “implemented” vs “planned”.

## Open Questions

- Should daemon mode be embedded in this binary (`vm-system serve`) or separated into a dedicated service binary?
- What compatibility contract is required for current script users?
- Should capabilities default to explicit deny-all for builtins, or keep current permissive defaults until enforcement is complete?

## References

- `vm-system/cmd/vm-system/main.go`
- `vm-system/cmd/vm-system/cmd_vm.go`
- `vm-system/cmd/vm-system/cmd_session.go`
- `vm-system/cmd/vm-system/cmd_exec.go`
- `vm-system/pkg/vmstore/vmstore.go`
- `vm-system/pkg/vmsession/session.go`
- `vm-system/pkg/vmexec/executor.go`
- `vm-system/pkg/libloader/loader.go`
- `vm-system/smoke-test.sh`
- `vm-system/test-goja-library-execution.sh`

## Chapter: Formal Domain Model And Invariants

A useful way to reason about vm-system is to treat it as a small operating system for JavaScript execution contexts. The core value is not raw script execution; it is controlled execution with identity, lifecycle, and observability.

### Entity graph

```text
VM (profile)
  ├─ VMSettings (limits, resolver, runtime)
  ├─ VMCapability[*] (module/global/fs/net/env)
  ├─ VMStartupFile[*] (ordered bootstrap files)
  └─ VMSession[*]
       └─ Execution[*]
            └─ ExecutionEvent[*]
```

### Invariants that should always hold

1. Every `VMSession` references an existing `VM`.
2. Every `Execution` references an existing `VMSession`.
3. Event sequence numbers for an execution are strictly increasing and unique.
4. A `ready` session has exactly one live runtime handle in the runtime host.
5. A session marked `closed` cannot accept new executions.
6. Session runtime context must reflect VM capabilities/settings at time of execution.

The current implementation enforces (1)-(3) mostly through schema + code path discipline, but does not yet guarantee (4)-(6) under multi-process CLI usage.

### State machine model

```text
Session:
  starting -> ready -> closed
      |         |
      v         v
    crashed <---+

Execution:
  running -> ok
  running -> error
  running -> timeout
  running -> cancelled
```

The state model is present in `vmmodels`, which is excellent. The missing piece is comprehensive transition enforcement and reconciliation between DB state and live runtime state.

## Chapter: Sequence Walkthroughs

### Sequence 1: Profile creation

```text
User
  -> CLI vm create
     -> vmstore.CreateVM
     -> vmstore.SetVMSettings(defaults)
     -> print VM ID
```

This is one of the strongest paths: deterministic, explicit defaults, durable persistence.

### Sequence 2: Session creation

```text
User
  -> CLI session create
     -> vmstore.GetVM / GetVMSettings
     -> validate worktree path exists
     -> vmstore.CreateSession(status=starting)
     -> goja.New()
     -> runtime config parse
     -> library load (optional)
     -> startup file loop
     -> vmstore.UpdateSession(status=ready)
```

Observations:

- The flow is comprehensible and well-ordered.
- Any failure in library/startup path sets crashed status and writes `LastError`.
- Runtime handle then resides only in `SessionManager.sessions` map.

### Sequence 3: REPL execution

```text
User
  -> CLI exec repl
     -> resolve session in sessionManager
     -> lock session
     -> create execution row
     -> override console for event capture
     -> run goja code
     -> append events
     -> update execution status
```

Critical dependence: “resolve session in sessionManager” fails when this invocation has a fresh manager instance without reconstructed runtime map.

### Sequence 4: Run-file execution

```text
User
  -> CLI exec run-file
     -> resolve session
     -> lock session
     -> resolve file path in worktree
     -> read file
     -> execute in runtime
     -> persist events/status
```

This path has the same continuity dependency as REPL.

## Chapter: Concurrency, Throughput, And Resource Behavior

### Concurrency model today

- Global session map lock for session lookup/registration.
- Per-session execution mutex with `TryLock` (single execution at a time per session).

This is a sound minimal concurrency design for single-process mode. The lock granularity is appropriate for current scale: session map lock is short-lived; execution mutex prevents runtime corruption from concurrent script runs.

### Throughput implications

- Each command invocation opens a new SQLite connection and initializes schema if needed.
- Single-session throughput is intentionally serialized.
- Multi-session throughput in current CLI is limited by process lifecycle and absent runtime host.

### Resource controls

Although limits are modeled (`cpu_ms`, `wall_ms`, `mem_mb`, `max_events`, `max_output_kb`), they are not yet enforced uniformly in execution loops. This means runtime behavior depends more on host defaults than VM profile declarations.

### Suggested enforcement hooks

```pseudo
before execute:
  start wall timer
  install interrupt callback for timeout

after each emitted event:
  if event_count > max_events: abort(timeout/error)
  if output_bytes > max_output_kb * 1024: abort(output_limit)

after execute:
  persist metrics {duration_ms, events, output_bytes}
```

## Chapter: Failure Mode Catalogue

### FM-1: Session not found after successful creation

- Trigger: create session in one command, execute in another.
- Cause: runtime map is process-local.
- Impact: core workflow broken for realistic CLI usage.

### FM-2: Library-enabled session creation fails despite successful download

- Trigger: add library and create session.
- Cause: filename contract mismatch.
- Impact: library feature appears broken.

### FM-3: Scripted smoke tests fail with stale flags/subcommands

- Trigger: running legacy scripts.
- Cause: CLI contract drift.
- Impact: false negatives, wasted debugging time.

### FM-4: Capability declarations give false assurance

- Trigger: assume capability rows actively gate runtime context.
- Cause: partial/no enforcement in execution context construction.
- Impact: policy mismatch and security ambiguity.

### FM-5: Event persistence degradation under DB write failure

- Trigger: intermittent DB I/O failure while adding events.
- Cause: ignored errors in event writes.
- Impact: incomplete observability with no explicit escalation.

### FM-6: Schema evolution risk

- Trigger: changing tables in-place with implicit `CREATE IF NOT EXISTS` only.
- Cause: no migration metadata/versioning.
- Impact: hard-to-debug drift across environments.

## Chapter: Practical Usage Guidance (Today)

Until runtime host mode is implemented, treat vm-system as:

- A robust VM/session metadata manager.
- A single-process demonstration runtime for immediate session use only.

### Safe operating pattern for current CLI

1. Use one shell process and sequence operations quickly.
2. Prefer direct smoke checks that do not require cross-invocation runtime reuse.
3. Avoid relying on library-enabled sessions until path contract is fixed.
4. Treat scripts in root as historical references unless updated.

### Suggested temporary command checklist

```bash
# Build
GOWORK=off go build -o ./vm-system ./cmd/vm-system

# Create profile/session
./vm-system --db /tmp/demo.db vm create --name demo --engine goja
./vm-system --db /tmp/demo.db session create --vm-id <id> --workspace-id ws --base-commit abc --worktree-path /tmp/ws

# Inspect persisted state
./vm-system --db /tmp/demo.db session list
./vm-system --db /tmp/demo.db session get <session-id>
```

## Chapter: Improvement Blueprint In Pseudocode

### Blueprint A: Runtime host process

```pseudo
main():
  cfg = loadConfig()
  store = openStore(cfg.db)
  host = newRuntimeHost(store)
  host.restoreDurableSessionsMetadata()
  startAPIServer(host)
  waitForShutdown()
```

```pseudo
RuntimeHost.createSession(req):
  vm = store.getVM(req.vmID)
  settings = store.getVMSettings(req.vmID)
  runtime = newGojaRuntime(settings)
  loadLibraries(runtime, vm.libraries)
  runStartup(runtime, vm.startupFiles)
  session = registerLiveSession(runtime)
  store.updateSessionStatus(session.id, ready)
  return session
```

### Blueprint B: Capability-aware context construction

```pseudo
buildContext(caps):
  ctx = {}
  if caps.module.console.enabled: ctx.console = hostConsole()
  if caps.module.math.enabled:    ctx.Math = Math
  if caps.module.json.enabled:    ctx.JSON = JSON
  ...
  for each allowedLibrary in caps.libraries:
     ctx[allowedLibrary.global] = libraryRegistry.get(allowedLibrary.id)
  return ctx
```

### Blueprint C: Script contract verification

```pseudo
contracts = cobra.exportCommandContract(binary)
for each script in scripts/:
  parsed = parseCLIInvocations(script)
  assert parsed.flags ⊆ contracts.flags(parsed.command)
  assert parsed.subcommand exists in contracts
```

## Chapter: Measurement And Test Strategy

The project currently has little automated behavioral testing despite execution complexity. Introduce three levels:

### Level 1: Unit tests

- vmstore CRUD and schema expectations.
- library loader filename resolution and manifest behavior.
- session manager state transitions.

### Level 2: Integration tests

- create VM -> create session -> execute repl -> retrieve events.
- error cases: startup failure, missing files, disabled capability.
- enforcement cases: output/event/timeout limits.

### Level 3: Contract tests for CLI

- command/flag snapshots from `--help`.
- script validity checks against snapshot.

### Example integration test skeleton

```go
func TestSessionLifecycleAndExecution(t *testing.T) {
  store := newTempStore(t)
  vmID := createVM(t, store)
  sm := vmsession.NewSessionManager(store)
  ex := vmexec.NewExecutor(store, sm)

  s := createSession(t, sm, vmID)
  result := mustExecREPL(t, ex, s.ID, "1+2")
  require.Equal(t, "ok", result.Status)
  events := mustEvents(t, ex, result.ID)
  require.NotEmpty(t, events)
}
```

## Chapter: Long-Term Architectural Direction

In mature form, vm-system can become a reliable execution substrate for multiple clients (CLI, web UI, automation agents). The current model already contains most required concepts. The strategic path is incremental hardening, not replacement.

Milestones:

1. **Stabilize correctness**: fix continuity and cache contract defects.
2. **Establish runtime host**: durable session runtime ownership.
3. **Enforce policy**: limits and capabilities become real guarantees.
4. **Operational maturity**: scripts, migrations, tests, observability.

At that point, vm-system moves from “reference implementation” to “trustworthy runtime service”.
