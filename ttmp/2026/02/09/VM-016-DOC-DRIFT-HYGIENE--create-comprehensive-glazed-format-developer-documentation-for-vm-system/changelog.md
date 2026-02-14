# Changelog

## 2026-02-09

### Created comprehensive glazed-format developer documentation

Created 8 new glazed help pages and rewrote 2 existing stubs:

**New pages:**
- `vm-system-getting-started.md` (Tutorial) — Build, run daemon, first template→session→execution loop
- `vm-system-architecture.md` (GeneralTopic) — Package layout, layered design, request lifecycle, tradeoffs
- `vm-system-api-reference.md` (GeneralTopic) — Complete endpoint contracts with request/response shapes
- `vm-system-templates-and-sessions.md` (GeneralTopic) — Template policy, session lifecycle, execution semantics
- `vm-system-cli-command-reference.md` (GeneralTopic) — Every CLI command with flags and arguments
- `vm-system-contributing.md` (GeneralTopic) — Contribution workflow, testing strategy, review checklist
- `vm-system-testing-guide.md` (GeneralTopic) — Test inventory, coverage analysis, how to add tests
- `vm-system-troubleshooting.md` (GeneralTopic) — Diagnosis and fixes for common issues
- `vm-system-examples.md` (Example) — 10 runnable recipes covering REPL, files, libraries, API automation

**Updated pages:**
- `vm-system-how-to-use.md` (Tutorial) — Expanded from 20-line stub to full practical guide
- `vm-system-command-map.md` (GeneralTopic) — Expanded from 6-line stub to full tree with role groupings

All pages follow glazed help format (YAML frontmatter with Title/Slug/Short/Topics/Commands/SectionType),
are embedded via `pkg/doc/doc.go`, and accessible via `vm-system help <slug>`.

**Files:**
- `pkg/doc/vm-system-getting-started.md`
- `pkg/doc/vm-system-architecture.md`
- `pkg/doc/vm-system-api-reference.md`
- `pkg/doc/vm-system-templates-and-sessions.md`
- `pkg/doc/vm-system-cli-command-reference.md`
- `pkg/doc/vm-system-contributing.md`
- `pkg/doc/vm-system-testing-guide.md`
- `pkg/doc/vm-system-troubleshooting.md`
- `pkg/doc/vm-system-examples.md`
- `pkg/doc/vm-system-how-to-use.md`
- `pkg/doc/vm-system-command-map.md`

## 2026-02-14

Ticket closed per request.

