# Tasks

## TODO

- [x] Finalize terminology contract for this ticket: template-centric naming only, no compatibility aliases (`vm`, `vm-id`, legacy modules command) in user-facing surfaces
- [x] Add/extend template API endpoints for module and library mutation/query so CLI no longer needs direct DB writes for these operations
- [ ] Extend `vmcontrol.TemplateService` and ports with explicit template module/library methods (add/remove/list) and remove ad-hoc command-side mutation logic
- [ ] Extend `vmclient/templates_client.go` with module/library operations on template routes
- [ ] Add new template CLI subcommands for module/library management and available-catalog listing
- [ ] Delete `cmd/vm-system/cmd_modules.go` and remove `modulesCmd` registration from `cmd/vm-system/main.go`
- [ ] Rename remaining CLI flags/wording from `vm-id` to `template-id` where template resources are targeted
- [ ] Update integration tests for new template module/library endpoints and command paths; remove/replace tests that assume `modules` command
- [ ] Update script surfaces (`test-library-*.sh`, other helpers) to template-first commands only
- [ ] Update `docs/getting-started-from-first-vm-to-contributor-guide.md` to remove legacy caveats and present only template vocabulary and flows
- [ ] Run final search guard and cleanup pass for user-facing legacy terms (`vm create`, `vm get`, `--vm-id`, `modules add-*`) and document any intentional exceptions
- [ ] Validate full matrix: `go test ./... -count=1`, smoke/e2e scripts, and doc review of getting-started walkthrough end-to-end
