# Changelog

## 2026-02-07

- Initial workspace created


## 2026-02-07

Completed deep vm-system audit: architecture map, runtime experiments, script compatibility checks, and textbook-style report with remediation roadmap.

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/libloader/loader.go — Library cache contract mismatch evidence
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/vmsession/session.go — Session continuity defect evidence
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ttmp/2026/02/07/VM-001-ANALYZE-VM--analyze-vm-system-architecture-behavior-and-quality/design-doc/01-comprehensive-vm-system-analysis-report.md — Final analysis report
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ttmp/2026/02/07/VM-001-ANALYZE-VM--analyze-vm-system-architecture-behavior-and-quality/reference/01-diary.md — Detailed implementation diary


## 2026-02-08

Added a dedicated daemon architecture design for vm-system covering runtime host process model, template/session lifecycle, REST API, CLI client design, internals, migration phases, and DoD for real UI integration.

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ttmp/2026/02/07/VM-001-ANALYZE-VM--analyze-vm-system-architecture-behavior-and-quality/design-doc/02-daemonized-vm-system-architecture-backend-runtime-host-rest-api-and-cli.md — New detailed daemon architecture design doc


## 2026-02-08

Reworked daemon architecture design to separate reusable session management core (`pkg/vmcontrol`) from daemon/HTTP adapters; updated phased implementation plan, alternatives, and Definition of Done; added detailed diary step documenting the separation rationale and execution details.

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ttmp/2026/02/07/VM-001-ANALYZE-VM--analyze-vm-system-architecture-behavior-and-quality/design-doc/02-daemonized-vm-system-architecture-backend-runtime-host-rest-api-and-cli.md — Core/daemon separation plan update
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ttmp/2026/02/07/VM-001-ANALYZE-VM--analyze-vm-system-architecture-behavior-and-quality/reference/01-diary.md — Added frequent detailed diary step for reusable core separation


## 2026-02-08

Follow-up clarification: reusable orchestration core is pkg/vmcontrol; daemon and HTTP remain adapters on top of this core.

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ttmp/2026/02/07/VM-001-ANALYZE-VM--analyze-vm-system-architecture-behavior-and-quality/design-doc/02-daemonized-vm-system-architecture-backend-runtime-host-rest-api-and-cli.md — Explicit vmcontrol naming and boundary clarification

## 2026-02-08

Updated daemon architecture doc to explicit hard cutover model: removed migration/backward-compatibility strategy, removed vm alias/deprecation path, removed migration matrix, and defined single post-cutover CLI/API surface.

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ttmp/2026/02/07/VM-001-ANALYZE-VM--analyze-vm-system-architecture-behavior-and-quality/design-doc/02-daemonized-vm-system-architecture-backend-runtime-host-rest-api-and-cli.md — Hard cutover update with no migration/backward-compatibility paths


## 2026-02-08

Added detailed v2 implementation task backlog and completed Task 8 to begin sequential implementation workflow.

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ttmp/2026/02/07/VM-001-ANALYZE-VM--analyze-vm-system-architecture-behavior-and-quality/reference/01-diary.md — Recorded Step 7 with prompt context and planning rationale


## 2026-02-08

Step 8: added reusable vmcontrol core package with explicit ports, service wiring, and runtime registry (commit a257a5a6b3e9eba9ac9b4aaf90a5b5eff46d03b4).

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/vmcontrol/core.go — Core constructor and dependency wiring
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/vmcontrol/execution_service.go — Execution orchestration service
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/vmcontrol/ports.go — Transport-agnostic store/runtime interfaces
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/vmcontrol/session_service.go — Session lifecycle orchestration service
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/vmcontrol/template_service.go — Template orchestration service
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ttmp/2026/02/07/VM-001-ANALYZE-VM--analyze-vm-system-architecture-behavior-and-quality/reference/01-diary.md — Added detailed Step 8 implementation diary


## 2026-02-08

Step 9: added daemon host package and serve command with graceful lifecycle handling (commit a89ddcedaa8a6035bfcf939550c657dcc2934483).

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/cmd/vm-system/cmd_serve.go — Serve command entrypoint with signal-based shutdown
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/vmdaemon/app.go — Daemon process host and HTTP server lifecycle
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/vmdaemon/config.go — Daemon listen and timeout config defaults
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ttmp/2026/02/07/VM-001-ANALYZE-VM--analyze-vm-system-architecture-behavior-and-quality/reference/01-diary.md — Added detailed Step 9 diary entry


## 2026-02-08

Step 10: added vmcontrol-backed REST transport with template/session/execution/event/runtime-summary endpoints and daemon wiring (commit 5046108bd0d1c3930aa4bf45fdda8714f2ac1301).

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/cmd/vm-system/cmd_serve.go — Daemon now installs vmhttp router
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/vmtransport/http/server.go — HTTP adapter and endpoint handlers backed by vmcontrol
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ttmp/2026/02/07/VM-001-ANALYZE-VM--analyze-vm-system-architecture-behavior-and-quality/reference/01-diary.md — Added detailed Step 10 diary entry


## 2026-02-08

Step 11: added vmclient package and switched session/exec CLI commands to daemon API client mode by default (commit 7cbdea6a4f99672a327c691625fcaa8eea15e47f).

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/cmd/vm-system/cmd_exec.go — Execution CLI now backed by daemon API
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/cmd/vm-system/cmd_session.go — Session CLI now backed by daemon API
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/vmclient/executions_client.go — Execution API client methods
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/vmclient/rest_client.go — Shared REST client core and API error decoding
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/vmclient/sessions_client.go — Session API client methods
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ttmp/2026/02/07/VM-001-ANALYZE-VM--analyze-vm-system-architecture-behavior-and-quality/reference/01-diary.md — Added detailed Step 11 diary entry

