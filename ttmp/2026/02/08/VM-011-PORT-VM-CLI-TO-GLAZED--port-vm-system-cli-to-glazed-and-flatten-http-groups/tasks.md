# Tasks

## Execution Checklist (new API only)

- [x] T01: Add glazed dependency and ensure imports compile with `schema/fields/values/sources` APIs.
- [x] T02: Add common glazed command helpers in `cmd/vm-system` for building command descriptions and Cobra bindings.
- [x] T03: Rebuild CLI root in `cmd/vm-system/main.go` with help-system wiring and logging setup (`help.NewHelpSystem`, `help_cmd.SetupCobraRootCommand`, `logging.AddLoggingSectionToRootCommand`, `logging.InitLoggerFromCobra`).
- [x] T04: Add embedded help docs package (`pkg/doc`) and wire docs loading into root help system.
- [x] T05: Remove `http` parent command from root registration.
- [x] T06: Port `serve` command to Glazed command implementation.
- [x] T07: Port `libs` command group (`download`, `list`, `cache-info`) to Glazed command implementations.
- [x] T08: Port `template` command group to Glazed command implementations (all existing verbs).
- [x] T09: Port `session` command group to Glazed command implementations; expose canonical `close` command.
- [x] T10: Remove `session delete` CLI verb from registration.
- [x] T11: Port `exec` command group (`repl`, `run-file`, `list`, `get`, `events`) to Glazed command implementations.
- [x] T12: Add new `ops` command group with `health` and `runtime-summary`.
- [x] T13: Add/extend vmclient support for ops endpoints used by CLI.
- [x] T14: Remove obsolete legacy command files and references tied to `http` parent wiring.
- [x] T15: Replace `cmd_http_test.go` with root topology tests asserting `http` is absent and `template/session/exec/ops/libs/serve` are present.
- [x] T16: Add tests for `session close` registration and keep template subgroup coverage tests green.
- [x] T17: Update `README.md` command examples to root taxonomy (`template/session/exec/ops`).
- [x] T18: Update `docs/getting-started-from-first-vm-to-contributor-guide.md` from `vm-system http ...` to root forms.
- [ ] T19: Run focused test pass (`go test ./cmd/vm-system -count=1`), then broader test pass (`go test ./... -count=1`) and capture any breakage.
- [ ] T20: Update ticket diary/changelog with implementation sequence and commit evidence.
