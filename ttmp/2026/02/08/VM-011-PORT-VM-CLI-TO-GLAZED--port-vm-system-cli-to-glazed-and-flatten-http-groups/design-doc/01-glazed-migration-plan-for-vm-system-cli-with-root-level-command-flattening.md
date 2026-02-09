---
Title: Glazed migration plan for vm-system CLI with root-level command flattening
Ticket: VM-011-PORT-VM-CLI-TO-GLAZED
Status: active
Topics:
    - backend
    - architecture
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: cmd/vm-system/cmd_exec.go
      Note: |-
        Execution command family and output behavior
        Execution command migration scope
    - Path: cmd/vm-system/cmd_http.go
      Note: |-
        HTTP command group to remove during flattening
        HTTP parent group slated for removal
    - Path: cmd/vm-system/cmd_http_test.go
      Note: Existing test tied to HTTP group that must be replaced
    - Path: cmd/vm-system/cmd_libs.go
      Note: Local library cache commands and init-time wiring
    - Path: cmd/vm-system/cmd_serve.go
      Note: Daemon serve command lifecycle
    - Path: cmd/vm-system/cmd_session.go
      Note: |-
        Session command family and close/delete ambiguity
        Session close/delete semantic changes
    - Path: cmd/vm-system/cmd_template.go
      Note: |-
        Template command family to port to glazed
        Template command family migration scope
    - Path: cmd/vm-system/main.go
      Note: |-
        Current Cobra root command wiring and persistent flags
        Root command taxonomy baseline to migrate
    - Path: docs/getting-started-from-first-vm-to-contributor-guide.md
      Note: |-
        Large command-reference surface impacted by command path changes
        Major docs surface requiring command path updates
    - Path: pkg/vmclient/executions_client.go
      Note: Execution API client calls used by CLI
    - Path: pkg/vmclient/rest_client.go
      Note: Shared HTTP client transport wrapper for CLI commands
    - Path: pkg/vmclient/sessions_client.go
      Note: Session API client calls and current close path
    - Path: pkg/vmclient/templates_client.go
      Note: Template API client calls used by CLI
    - Path: pkg/vmtransport/http/server.go
      Note: |-
        Authoritative route inventory and current delete-close aliasing
        Endpoint coverage and delete-close alias behavior
    - Path: ttmp/2026/02/08/VM-009-EXPAND-HTTP-CLI-VERBS--expand-http-cli-verbs-and-command-taxonomy/design-doc/01-http-cli-verb-coverage-and-taxonomy-expansion-plan.md
      Note: Prior plan that assumed an http parent group
ExternalSources: []
Summary: Implementation-ready migration blueprint for converting vm-system CLI commands from ad-hoc Cobra wiring to the new Glazed command API (schema/fields/values/sources), removing the root http group, and wiring built-in help docs.
LastUpdated: 2026-02-09T00:30:00Z
WhatFor: Provide a complete, command-by-command migration and rollout plan to port vm-system CLI to the new Glazed command API and flatten daemon-backed command groups to root.
WhenToUse: Use when implementing VM-011 and reviewing command-tree, new API usage, help-system setup, and test/doc updates.
---


# Glazed migration plan for vm-system CLI with root-level command flattening

## Executive Summary

This ticket ports `vm-system` CLI command implementation from custom Cobra handlers to glazed command definitions, and removes the `http` parent group so command families are exposed at root again (`template`, `session`, `exec`, plus `ops`).

The migration has three hard requirements:

1. Keep command behavior parity for all existing verbs while replacing command internals with Glazed command structs, schema sections, and value decoding.
2. Flatten CLI taxonomy by deleting `http` from the command path, while preserving daemon-backed semantics.
3. Wire glazed help system (`help.NewHelpSystem` + `help_cmd.SetupCobraRootCommand`) with embedded docs so command/help behavior remains coherent and self-documented.
4. Implement all ported commands directly against the new Glazed API (`schema`, `fields`, `values`, `sources`).

This document is exhaustive: it inventories every current command, maps each one to its target glazed form, defines package layout, identifies semantic issues (`session delete`), and provides phased execution, testing, and docs rollout.

## Problem Statement

`vm-system` CLI has grown through direct Cobra `RunE` handlers in `cmd/vm-system/*.go`. This created four major issues.

### 1) Command architecture is not standardized

Current commands manually parse flags and print strings, with repeated patterns:

- instantiate `vmclient.New(serverURL, nil)` in each `RunE`
- parse JSON flags ad hoc (`--args`, `--env`, `--config`)
- print hand-formatted text tables manually

This blocks consistent structured output and makes reuse/testability harder.

### 2) Root taxonomy mismatch and drift

Current root wiring in `cmd/vm-system/main.go` registers:

- `serve`
- `http`
- `libs`

But `README.md` quick start already uses root-level `template/session/exec` commands (without `http`). This means docs and CLI are currently inconsistent.

### 3) Semantic ambiguity in session lifecycle

`DELETE /api/v1/sessions/{session_id}` currently aliases `close` (`handleSessionDelete -> handleSessionClose` in `pkg/vmtransport/http/server.go`). CLI only exposes `session delete`, which is semantically destructive but behavior is transition-to-closed.

### 4) No built-in glazed help system wiring

Root command currently relies on default Cobra help only. For larger command surfaces and migration to glazed, we need structured help docs and command metadata-driven help behavior.

### 5) New API consistency at command boundaries

The port must use one consistent API vocabulary across all command boundaries:

- `schema` for sections
- `fields` for field definitions/types
- `values` for decoded values
- `sources` for resolver/source chains

## Current State Inventory

### Current root command tree

- `vm-system serve`
- `vm-system http template ...`
- `vm-system http session ...`
- `vm-system http exec ...`
- `vm-system libs ...`

Global flags:

- `--db` (used by `serve`)
- `--server-url` (used by daemon-backed commands)

### Current command inventory and target mapping (exhaustive)

| Current command | Target command after flattening | Transport/API | Current implementation locus | Migration notes |
|---|---|---|---|---|
| `serve` | `serve` | local daemon host | `cmd_serve.go` | Port to glazed Writer/Bare command; retain signal lifecycle |
| `libs download` | `libs download` | local filesystem/cache | `cmd_libs.go` | Port to glazed writer command; keep cache path default |
| `libs list` | `libs list` | local metadata | `cmd_libs.go` | Prefer glazed tabular output rows |
| `libs cache-info` | `libs cache-info` | local filesystem/cache | `cmd_libs.go` | Preserve no-cache message behavior |
| `http template create` | `template create` | `POST /api/v1/templates` | `cmd_template.go` | Required flags: `name`; default engine `goja` |
| `http template list` | `template list` | `GET /api/v1/templates` | `cmd_template.go` | Convert formatted table to rows |
| `http template get` | `template get` | `GET /api/v1/templates/{id}` | `cmd_template.go` | Multi-section output; row + optional detail rows |
| `http template delete` | `template delete` | `DELETE /api/v1/templates/{id}` | `cmd_template.go` | Keep destructive semantics |
| `http template add-module` | `template add-module` | `POST /templates/{id}/modules` | `cmd_template.go` | Required `--name` |
| `http template remove-module` | `template remove-module` | `DELETE /templates/{id}/modules/{name}` | `cmd_template.go` | Required `--name` |
| `http template list-modules` | `template list-modules` | `GET /templates/{id}/modules` | `cmd_template.go` | Emit one row per module |
| `http template add-library` | `template add-library` | `POST /templates/{id}/libraries` | `cmd_template.go` | Required `--name` |
| `http template remove-library` | `template remove-library` | `DELETE /templates/{id}/libraries/{name}` | `cmd_template.go` | Required `--name` |
| `http template list-libraries` | `template list-libraries` | `GET /templates/{id}/libraries` | `cmd_template.go` | Emit one row per library |
| `http template list-available-modules` | `template list-available-modules` | local catalog | `cmd_template.go` | No daemon call; still in template family |
| `http template list-available-libraries` | `template list-available-libraries` | local catalog | `cmd_template.go` | No daemon call; still in template family |
| `http template add-capability` | `template add-capability` | `POST /templates/{id}/capabilities` | `cmd_template.go` | Parse JSON `--config`; validate early |
| `http template list-capabilities` | `template list-capabilities` | `GET /templates/{id}/capabilities` | `cmd_template.go` | Emit rows; avoid truncation when output=json |
| `http template add-startup` | `template add-startup` | `POST /templates/{id}/startup-files` | `cmd_template.go` | Required `--path`; defaults mode/order retained |
| `http template list-startup` | `template list-startup` | `GET /templates/{id}/startup-files` | `cmd_template.go` | Emit rows with order/mode/path |
| `http session create` | `session create` | `POST /api/v1/sessions` | `cmd_session.go` | Required 4 flags retained |
| `http session list` | `session list` | `GET /api/v1/sessions` | `cmd_session.go` | Preserve optional `--status` filter |
| `http session get` | `session get` | `GET /api/v1/sessions/{id}` | `cmd_session.go` | Preserve closed/last-error conditional fields |
| `http session delete` (close semantics) | `session close` | `POST /api/v1/sessions/{id}/close` | `cmd_session.go` + server alias | Rename for semantic correctness |
| `http exec repl` | `exec repl` | `POST /api/v1/executions/repl` | `cmd_exec.go` | Keep positional args `[session-id] [code]` |
| `http exec run-file` | `exec run-file` | `POST /api/v1/executions/run-file` | `cmd_exec.go` | Parse JSON args/env via typed settings |
| `http exec list` | `exec list` | `GET /api/v1/executions` | `cmd_exec.go` | Preserve `--limit`, output rows |
| `http exec get` | `exec get` | `GET /api/v1/executions/{id}` | `cmd_exec.go` | Preserve conditional fields |
| `http exec events` | `exec events` | `GET /api/v1/executions/{id}/events` | `cmd_exec.go` | Preserve `--after-seq` |
| (missing) | `ops health` | `GET /api/v1/health` | none | Add new glazed command |
| (missing) | `ops runtime-summary` | `GET /api/v1/runtime/summary` | none | Add new glazed command |

## Proposed Solution

### A) New root taxonomy (http group removed)

Target root tree:

- `serve`
- `template <verb>`
- `session <verb>`
- `exec <verb>`
- `ops <verb>`
- `libs <verb>`

`http` is removed entirely from user-facing paths.

### B) Glazed command architecture

Use explicit command-group packages and one command file per verb.

Proposed layout:

- `cmd/vm-system/main.go` (root bootstrap + help/logging)
- `cmd/vm-system/cmds/root.go`
- `cmd/vm-system/cmds/common/common.go` (shared builders, parser config)
- `cmd/vm-system/cmds/template/*.go`
- `cmd/vm-system/cmds/session/*.go`
- `cmd/vm-system/cmds/exec/*.go`
- `cmd/vm-system/cmds/ops/*.go`
- `cmd/vm-system/cmds/libs/*.go`
- `cmd/vm-system/cmds/serve/*.go`

### C) Standard glazed implementation pattern per verb

For each verb command:

1. Define command struct embedding `*cmds.CommandDescription`.
2. Define settings struct with `glazed.parameter` tags.
3. Use `fields.New` + `cmds.WithFlags` for command fields.
4. Attach standard schema sections (`schema.NewGlazedSchema`, `cli.NewCommandSettingsLayer`).
5. Decode settings via `values.DecodeSectionInto(vals, schema.DefaultSlug, &settings)`.
6. Emit row-based output via processor (for commands with structured results).

### D) Required new API usage (mandatory for VM-011)

Every ported command must follow these rules:

- Define flags with `fields.New(...)`.
- Build schemas/sections via `schema` APIs.
- Decode command settings with `values.DecodeSectionInto(...)`.
- Use `sources` wrappers for explicit precedence chains.
- Avoid introducing pre-migration command APIs in new command code.

### E) Shared parser and middleware configuration

Adopt one helper for `cli.BuildCobraCommand` with:

- `ShortHelpLayers: []string{schema.DefaultSlug}`
- `MiddlewaresFunc: cli.CobraCommandDefaultMiddlewares`

Use this helper in all command-group register functions to avoid drift.

For explicit precedence chains, use `sources` wrappers:

- defaults < config files < env < flags
- `sources.Execute(schema_, vals, ...sources...)`

### F) Output schema policy

- If a command implements `cmds.GlazeCommand`, `cli.BuildCobraCommand(...)` will ensure glazed output layer wiring.
- Only add output schema explicitly when needed (for example custom schema composition), using `schema.NewGlazedSchema()`.

### G) Session verb semantics policy

- Rename current close behavior to `session close`.
- Remove `session delete` until/if true destructive delete is implemented end-to-end.
- If true delete is later needed, implement a dedicated `vmclient.DeleteSession` and separate server behavior first, then expose `session delete`.

### H) Help system setup (required)

Root setup:

1. `helpSystem := help.NewHelpSystem()`
2. `doc.AddDocToHelpSystem(helpSystem)` from new `pkg/doc` package with embedded markdown.
3. `help_cmd.SetupCobraRootCommand(helpSystem, rootCmd)`

Doc embedding package:

- `pkg/doc/doc.go` with `//go:embed *` and `LoadSectionsFromFS`.
- Markdown files for top-level CLI map and command families.
- Validate frontmatter carefully (quote strings containing colons).

Recommended initial help docs:

- `pkg/doc/vm-system-how-to-use.md`
- `pkg/doc/vm-system-command-map.md`
- `pkg/doc/vm-system-glazed-output.md`

### I) Logging layer setup (recommended)

At root:

- `logging.AddLoggingLayerToRootCommand(rootCmd, "vm-system")`
- `PersistentPreRunE` initializes logger from cobra parsing.

This keeps migration aligned with glazed conventions in sibling repos.

## Design Decisions

1. Remove `http` command group rather than aliasing it.
Rationale: explicit product requirement in this ticket and existing doc drift back toward root-level commands.

2. Keep noun-first groups (`template`, `session`, `exec`, `ops`, `libs`) instead of flattening every verb to root.
Rationale: preserves discoverability and avoids collisions (`list`, `get`, `delete` repeated across families).

3. Add `ops` group for health/runtime routes.
Rationale: these are daemon operational endpoints and do not belong under template/session/exec.

4. Eliminate `session delete` until true delete semantics exist.
Rationale: current behavior is close-only, and semantic mismatch is operationally risky.

5. Migrate all command families in one coordinated taxonomy release.
Rationale: mixed Cobra/glazed trees with path changes produce user confusion and brittle docs/tests.

## Alternatives Considered

1. Keep direct Cobra handlers and only remove `http`.
Rejected: does not address maintainability/output/help consistency goals.

2. Keep `http` as hidden alias for one transition release.
Rejected for this ticket scope: requirement is to remove group; no compatibility shim requested.

3. Flatten verbs fully to root (for example `vm-system create-template`).
Rejected: degrades command discoverability and scales poorly as API surface grows.

4. Keep `session delete` name while documenting close semantics.
Rejected: semantic ambiguity is exactly the problem.

## Implementation Plan

### Phase 0: Baseline and branch hygiene

1. Snapshot current command help outputs (`--help`) for reference.
2. Capture current command matrix and test baseline.
3. Mark VM-009 assumptions as superseded for taxonomy decisions.

### Phase 1: Root/bootstrap migration

1. Add glazed dependencies to `go.mod`.
2. Create `cmd/vm-system/cmds/root.go` to build root command.
3. Move persistent/global flag definitions into root setup with shared defaults.
4. Add logging layer and `PersistentPreRunE` logger init.
5. Wire help system and embedded docs package.
6. Add migration guard checks for new API adoption:
   - ensure ported command files use `schema/fields/values/sources` imports
   - ensure settings decoding uses `values.DecodeSectionInto(...)`

### Phase 2: Command family migration

1. Port `template` group (all 16 verbs).
2. Port `session` group (`create/list/get/close`).
3. Port `exec` group (`repl/run-file/list/get/events`).
4. Port `libs` group (`download/list/cache-info`).
5. Port `serve` command.
6. Add new `ops` group (`health/runtime-summary`).
7. Apply the new API recipe while porting each verb:
   - use `schema/fields/values/sources` directly in command implementations
   - replace parameter definitions with `fields.New(...)`
   - decode via `values.DecodeSectionInto(...)`
   - use `sources.Execute(...)` where explicit precedence chains are needed

### Phase 3: Remove legacy Cobra files and http group

1. Delete `cmd_http.go` and direct `newHTTPCommand()` wiring.
2. Replace `cmd_http_test.go` with root command topology tests.
3. Remove obsolete constructors from legacy files after new code is active.

### Phase 4: Test hardening

1. Unit tests:
   - root registers `template/session/exec/ops/libs/serve`
   - `http` group absent
   - required flags and arg validation parity
   - new API checks pass for ported files (import/use assertions)
2. Integration/E2E tests:
   - CLI command to route coverage (including new ops commands)
   - session close semantics
3. Help tests:
   - `vm-system help` works with embedded docs loaded
   - help docs frontmatter validation

### Phase 5: Documentation migration

1. Update command examples from `vm-system http ...` to root group forms.
2. `docs/getting-started-from-first-vm-to-contributor-guide.md` currently has 51 `http`-prefixed references and must be fully normalized.
3. Update README and ticket docs to reflect final command tree.

### Phase 6: Validation and handoff

Run:

- `GOWORK=off go test ./... -count=1`
- `./smoke-test.sh`
- `./test-e2e.sh`

Then verify manual smoke paths:

- `vm-system template create/list/get`
- `vm-system session create/list/get/close`
- `vm-system exec repl/run-file/list/get/events`
- `vm-system ops health/runtime-summary`

## Acceptance Criteria

1. `http` root group no longer exists.
2. Command families are root-level (`template/session/exec/ops/libs/serve`).
3. Every command in the inventory table has a glazed implementation.
4. `session close` is canonical and `session delete` is removed unless true delete semantics are introduced.
5. Built-in help system is wired and serves embedded docs.
6. Ported command files use the new APIs (`schema/fields/values/sources`) at API edges.
7. Tests and docs are updated and green.

## Risks and Mitigations

1. Risk: regression in positional arg handling for `exec repl` and `exec run-file`.
Mitigation: dedicated argument-validation tests and parity checks.

2. Risk: output format drift for scripts.
Mitigation: add machine-readable output pathways (`--output json`) and update scripts to explicit output mode.

3. Risk: docs drift during taxonomy switch.
Mitigation: grep-based audit for `vm-system http` in repository docs and tests before merge.

4. Risk: duplicate or conflicting flags from schema composition.
Mitigation: centralize schema setup and avoid attaching identical sections at multiple levels.

## Open Questions

1. Should `ops health` / `ops runtime-summary` remain grouped, or be promoted to root verbs (`health`, `runtime-summary`) for shorter command paths?
2. Should we include a hidden, temporary `http` alias for one release despite current requirement, or enforce hard cutover immediately?
3. For commands that currently print blended text sections, do we emit normalized rows only, or provide dual-mode with explicit `--output` defaults?

## References

- `cmd/vm-system/main.go`
- `cmd/vm-system/cmd_http.go`
- `cmd/vm-system/cmd_template.go`
- `cmd/vm-system/cmd_session.go`
- `cmd/vm-system/cmd_exec.go`
- `cmd/vm-system/cmd_libs.go`
- `cmd/vm-system/cmd_serve.go`
- `pkg/vmclient/rest_client.go`
- `pkg/vmclient/templates_client.go`
- `pkg/vmclient/sessions_client.go`
- `pkg/vmclient/executions_client.go`
- `pkg/vmtransport/http/server.go`
- `docs/getting-started-from-first-vm-to-contributor-guide.md`
- `ttmp/2026/02/08/VM-009-EXPAND-HTTP-CLI-VERBS--expand-http-cli-verbs-and-command-taxonomy/design-doc/01-http-cli-verb-coverage-and-taxonomy-expansion-plan.md`
