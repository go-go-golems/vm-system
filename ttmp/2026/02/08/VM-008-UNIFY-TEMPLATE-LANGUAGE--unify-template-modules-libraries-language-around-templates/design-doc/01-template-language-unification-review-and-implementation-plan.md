---
Title: Template language unification review and implementation plan
Ticket: VM-008-UNIFY-TEMPLATE-LANGUAGE
Status: active
Topics:
    - backend
    - docs
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: cmd/vm-system/cmd_modules.go
      Note: |-
        Legacy VM-oriented direct-DB mutations and `--vm-id` language
        Legacy mixed-language command surface targeted for removal
    - Path: cmd/vm-system/cmd_template.go
      Note: |-
        Template command surface to become authoritative
        Authoritative template command surface for new module/library operations
    - Path: cmd/vm-system/main.go
      Note: |-
        Current mixed command registration includes template and legacy modules surfaces
        Root command wiring to clean up
    - Path: docs/getting-started-from-first-vm-to-contributor-guide.md
      Note: |-
        Existing guide still documents mixed/legacy language caveats
        Getting-started guide language and command cleanup target
    - Path: pkg/vmclient/templates_client.go
      Note: Client methods to add for template module/library operations
    - Path: pkg/vmcontrol/template_service.go
      Note: |-
        Service methods to extend for unified template operations
        Template service operations to expand
    - Path: pkg/vmtransport/http/server.go
      Note: |-
        Template API routes currently missing module/library mutation endpoints
        Template API endpoints to extend
ExternalSources: []
Summary: Detailed review of template-vs-vm language drift and implementation plan to make template terminology and template-owned module/library operations the only supported path.
LastUpdated: 2026-02-08T12:12:26-05:00
WhatFor: Enable delegated cleanup of mixed command language and ownership boundaries without compatibility constraints.
WhenToUse: Use as the execution guide for VM-008 language and command-surface unification.
---


# Template language unification review and implementation plan

## Executive Summary

The repository has converged operationally on daemon-first `template/session/exec` flows, but still exposes legacy `modules` commands that use `vm` wording (`--vm-id`) and mutate storage directly. This creates two mental models and two control planes.

This ticket defines a no-backward-compatibility cleanup: `template` becomes the single authoritative language and command/API surface for modules/libraries, and legacy mixed-language surfaces are removed.

## Problem Statement

Current drift and confusion points:

1. Mixed root command surface:
   - `cmd/vm-system/main.go:26` registers both `template` and `modules`.
2. Legacy wording and direct DB mutation in command layer:
   - `cmd/vm-system/cmd_modules.go:49` uses `--vm-id`.
   - `cmd/vm-system/cmd_modules.go:56` opens store directly and bypasses daemon API.
3. Missing template API parity:
   - `pkg/vmtransport/http/server.go` has template CRUD/capability/startup endpoints but no template module/library mutation endpoints.
4. Documentation inconsistency:
   - `docs/getting-started-from-first-vm-to-contributor-guide.md:530` explicitly calls out mixed legacy modules path.
5. Script surface still reflects legacy command presence:
   - library scripts currently rely on `modules add-*` even though broader flow is template-first.

Consequences:

- Contributors are forced to understand both “template API” and “legacy VM DB mutation” concepts.
- Validation and policy behavior can differ because some operations bypass API/domain checks.
- Documentation includes caveats that should not exist in the intended clean system model.

## Proposed Solution

Adopt a strict template-first model across CLI/API/docs:

1. Add template-native module/library endpoints and service methods.
2. Add template-native CLI subcommands for module/library operations.
3. Remove legacy `modules` command entirely.
4. Remove `vm`-oriented user-facing terminology for template resources.
5. Update getting-started and scripts to only show template language.

Target command examples (post-cleanup):

```bash
vm-system template add-module <template-id> --name console
vm-system template add-library <template-id> --name lodash-4.17.21
vm-system template list-modules <template-id>
vm-system template list-libraries <template-id>
vm-system template list-available-modules
vm-system template list-available-libraries
```

Target route examples (post-cleanup):

```http
POST   /api/v1/templates/{template_id}/modules
DELETE /api/v1/templates/{template_id}/modules/{module_name}
GET    /api/v1/templates/{template_id}/modules

POST   /api/v1/templates/{template_id}/libraries
DELETE /api/v1/templates/{template_id}/libraries/{library_name}
GET    /api/v1/templates/{template_id}/libraries
```

## Design Decisions

1. No backward compatibility for legacy module commands.
Rationale: carrying aliases keeps the ambiguity alive and prolongs cleanup.

2. Template API is the only mutating path for template-associated module/library configuration.
Rationale: one control plane means consistent validation, auth, and error contracts.

3. Preserve “VM” wording only where it is true runtime concept (for example session runtime internals), but not for template resource operations.
Rationale: avoid replacing useful runtime-domain wording indiscriminately.

4. Update guide/scripts in the same ticket to avoid another documentation lag cycle.
Rationale: user-facing language drift is part of the core problem.

## Terminology Contract (Finalized)

This ticket adopts a strict template-centric terminology contract for user-facing surfaces:

1. Use `template` / `template-id` for all template resource operations.
2. Do not use `vm` / `vm-id` in user-facing CLI help, examples, flags, or API route naming when the target is a template resource.
3. Remove the legacy `modules` command surface entirely rather than aliasing it.
4. Keep `VM` wording only where it represents true runtime internals/models (for example internal model names in `vmmodels` or session runtime concepts), not CLI/API language for template operations.
5. Use template-owned module/library operations through template service/API/CLI routes only; no direct command-side DB mutation path remains.

Out of scope for this ticket:

- internal core type/package renaming (`vmmodels.VM*`) where it does not affect user-facing command/API language.

## Alternatives Considered

1. Keep `modules` command as alias to new template subcommands.
Rejected: still presents two names for the same operation and undermines cleanup objective.

2. Keep direct DB mutation for `modules` while updating docs only.
Rejected: preserves dual control-plane behavior and bypasses API validation.

3. Partial rename in docs without command/API changes.
Rejected: creates misleading docs that do not match real tooling.

## Implementation Plan

Phase 1: API/domain foundation

1. Extend template service/ports for module/library operations.
2. Add store methods if needed to support add/remove/list semantics cleanly.
3. Add HTTP routes/handlers under `/api/v1/templates/{template_id}/...`.
4. Add vmclient methods for new routes.

Phase 2: CLI surface cleanup

1. Add template module/library subcommands in `cmd_template.go`.
2. Remove `cmd_modules.go` and root command registration.
3. Align flags and output language (`template-id`, template wording only).

Phase 3: Tests and scripts

1. Add/update integration tests for new template module/library endpoints.
2. Update scripts to use template module/library commands.
3. Remove tests that depend on legacy modules surface.

Phase 4: Documentation cleanup

1. Update getting-started guide sections and examples to template-only language.
2. Remove legacy caveat sections about `modules`/`vm-id`.
3. Run string-search cleanup for legacy phrases in user-facing docs/CLI help.

Definition of done:

- No `modules` command in CLI.
- No template mutation flow relies on direct DB writes from command handlers.
- Getting-started guide and scripts are template-only in language and commands.
- Integration tests cover new template module/library endpoints.

## Open Questions

1. Should module/library operations be represented as template capabilities internally, or remain explicit top-level template fields?
2. Should we rename core model types (`vmmodels.VM*`) in this same ticket, or keep this ticket focused on user/API language and follow with internal type rename?
3. Should available module/library catalogs remain under `template` command or move to a separate read-only `catalog` namespace?

## References

- `cmd/vm-system/main.go`
- `cmd/vm-system/cmd_modules.go`
- `cmd/vm-system/cmd_template.go`
- `pkg/vmtransport/http/server.go`
- `pkg/vmcontrol/template_service.go`
- `pkg/vmclient/templates_client.go`
- `docs/getting-started-from-first-vm-to-contributor-guide.md`
- `ttmp/2026/02/08/VM-006-REVIEW-VM-SYSTEM--review-vm-system-implementation-quality-testing-and-setup/design-doc/01-comprehensive-vm-system-implementation-quality-review.md`
