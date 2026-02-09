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


## 2026-02-08

Step 12: completed hard CLI naming cutover from vm to template and removed vm command registration (commit 10a6c7382a228b4dff9f2a9cd33dc248ed16359c).

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/cmd/vm-system/cmd_session.go — Renamed session create flag to --template-id
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/cmd/vm-system/cmd_template.go — New template command group backed by daemon API
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/cmd/vm-system/cmd_vm.go — Removed legacy vm command implementation
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/vmclient/templates_client.go — Template REST client wrappers
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ttmp/2026/02/07/VM-001-ANALYZE-VM--analyze-vm-system-architecture-behavior-and-quality/reference/01-diary.md — Added detailed Step 12 diary entry


## 2026-02-08

Step 13: added core run-file path traversal protection and execution limit scaffolding, with HTTP error mapping updates (commit a645a4a87190a3ead23d35b8c9de8395369409f0).

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/vmcontrol/execution_service.go — Path normalization guard and post-execution limits scaffolding
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/vmmodels/models.go — Added ErrPathTraversal error contract
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/vmtransport/http/server.go — Mapped new core safety errors to API response codes
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ttmp/2026/02/07/VM-001-ANALYZE-VM--analyze-vm-system-architecture-behavior-and-quality/reference/01-diary.md — Added detailed Step 13 diary entry


## 2026-02-08

Step 14: added integration test proving daemon session continuity across independent API requests (commit c4dae0d2002711d4a8ed65274515398c3e89d64d).

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/vmtransport/http/server_integration_test.go — Cross-request session continuity integration coverage
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ttmp/2026/02/07/VM-001-ANALYZE-VM--analyze-vm-system-architecture-behavior-and-quality/reference/01-diary.md — Added detailed Step 14 diary entry


## 2026-02-08

Step 15: rewrote smoke/e2e scripts and README for daemon-first template/session/exec workflows, with sequential script validation (commit 1ef14f69b4fd6612576eb592694bc9626e3c7771).

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/README.md — Updated v2 architecture and quickstart documentation
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/smoke-test.sh — Daemon-first smoke flow with template/session/exec checks
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/test-e2e.sh — Daemon-first end-to-end workflow script
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ttmp/2026/02/07/VM-001-ANALYZE-VM--analyze-vm-system-architecture-behavior-and-quality/reference/01-diary.md — Added detailed Step 15 diary entry


## 2026-02-08

Step 16: finalized task-to-commit bookkeeping and completed task 17 with full implementation traceability across tasks 8-16.

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ttmp/2026/02/07/VM-001-ANALYZE-VM--analyze-vm-system-architecture-behavior-and-quality/reference/01-diary.md — Added final task-to-commit closure step
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ttmp/2026/02/07/VM-001-ANALYZE-VM--analyze-vm-system-architecture-behavior-and-quality/tasks.md — Marked task 17 complete after commit mapping audit


## 2026-02-08

Ticket closed


## 2026-02-09

Ticket re-closed per request.

