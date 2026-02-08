# Tasks

## TODO

- [x] Review VM-001, VM-004, and VM-005 ticket context and scope VM-006 review criteria
- [x] Run compile/tests/scripts dynamically (unit, integration, smoke, e2e, legacy script surface)
- [x] Perform deep static audit across cmd, core, runtime, transport, store, and script setup layers
- [x] Reproduce high-risk edge cases dynamically (path safety, startup failure lifecycle, error contracts, limit semantics)
- [x] Write comprehensive VM-006 quality review report with severity ordering and cleanup sketches
- [x] Write detailed VM-006 diary with commands, failures, findings, and rationale
- [x] Relate key files and update ticket docs metadata
- [x] Upload VM-006 report to reMarkable and record verification output

## Type-System Follow-Up

- [x] Introduce typed worktree path model (`WorktreeRoot`, `RelWorktreePath`, `ResolvedWorktreePath`) in a shared package and unit test it
- [x] Replace run-file normalization with typed resolver and block symlink escapes; extend safety integration coverage
- [x] Validate and resolve startup file paths through typed path model (API + session runtime) and add traversal/symlink safety tests
- [x] Add typed `ErrExecutionNotFound` contract end-to-end (store + transport + integration tests)
- [ ] Remove duplicated VM settings config structs from `vmcontrol` by reusing `vmmodels` types
