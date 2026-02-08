# Tasks

## TODO

- [ ] Build and commit endpoint-to-CLI verb matrix from `pkg/vmtransport/http/server.go` and current `cmd/vm-system` command registrations
- [ ] Finalize session lifecycle verb policy (`close` vs `delete`) and record explicit decision in design doc + changelog
- [ ] Add `http ops` command group and wire it from root `http` command
- [ ] Add `http ops health` command mapped to `GET /api/v1/health`
- [ ] Add `http ops runtime-summary` command mapped to `GET /api/v1/runtime/summary`
- [ ] Add `http session close` command mapped to `POST /api/v1/sessions/{session_id}/close`
- [ ] Align `http session delete` semantics to destructive delete endpoint only, or remove it if close-only policy is chosen (no aliasing)
- [ ] Extend/adjust vmclient methods for any new command coverage requirements
- [ ] Add command registration tests for `http` + `http ops` + session lifecycle verbs
- [ ] Add/extend behavior tests validating session close/delete command semantics and output contracts
- [ ] Update scripts/runbooks to canonical HTTP command forms and remove non-canonical verb forms
- [ ] Update getting-started docs and quick references with final command taxonomy and examples
- [ ] Run final search guard for non-canonical command patterns and document intentional exceptions
- [ ] Validate full matrix: `go test ./... -count=1`, `go test ./pkg/vmtransport/http -count=1`, `./smoke-test.sh`, `./test-e2e.sh`
- [ ] Produce final handoff note summarizing decisions, changed surfaces, and reviewer-focused verification checklist
