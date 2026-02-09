# Changelog

## 2026-02-07

- Initial workspace created


## 2026-02-07

Created a full migration blueprint to make vm-system-ui a real backend client: target architecture, API contract, phased backend/frontend plan, risks, testing, rollout, and WBS.

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ttmp/2026/02/07/VM-001-ANALYZE-VM--analyze-vm-system-architecture-behavior-and-quality/design-doc/01-comprehensive-vm-system-analysis-report.md — Referenced backend findings
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ttmp/2026/02/07/VM-002-ANALYZE-VM-SYSTEM-UI--analyze-vm-system-ui-architecture-behavior-and-quality/design-doc/01-comprehensive-vm-system-ui-analysis-report.md — Referenced frontend findings
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ttmp/2026/02/07/VM-003-MAKE-WEB-UI-REAL--make-vm-system-ui-real-by-integrating-with-vm-system-backend/reference/01-diary.md — Detailed process diary


## 2026-02-07

Uploaded bundled VM-003 analysis+diary PDF to reMarkable at /ai/2026/02/08/VM-003-MAKE-WEB-UI-REAL and verified cloud listing.

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ttmp/2026/02/07/VM-003-MAKE-WEB-UI-REAL--make-vm-system-ui-real-by-integrating-with-vm-system-backend/design-doc/01-make-web-ui-real-backend-integration-analysis-and-implementation-plan.md — Uploaded as primary analysis payload
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ttmp/2026/02/07/VM-003-MAKE-WEB-UI-REAL--make-vm-system-ui-real-by-integrating-with-vm-system-backend/reference/01-diary.md — Uploaded as companion process narrative

## 2026-02-08

Implemented the VM-003 clean-cut web UI migration from browser-local mock runtime to real daemon REST integration.

Behavioral changes:
- UI now calls real template/session/execution endpoints (`/api/v1/templates`, `/api/v1/sessions`, `/api/v1/executions`) instead of in-browser simulation.
- REPL execution is daemon-owned only; no browser `new Function` execution remains.
- Session actions now reflect backend contracts (close semantics, template-backed session creation requirements).
- Template module/library toggles persist through template API endpoints.
- Development server now proxies `/api/v1` to configurable daemon target.

Validation:
- `pnpm check` passed in `vm-system-ui`
- `pnpm build` passed in `vm-system-ui` (existing analytics/chunk-size warnings unchanged)

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system-ui/client/src/lib/vmService.ts — Replaced mock runtime service with real REST-backed service and mapping logic
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system-ui/client/src/pages/Home.tsx — Switched startup/session orchestration to async backend initialization
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system-ui/client/src/components/VMConfig.tsx — Wired module/library toggles to template endpoints
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system-ui/client/src/components/SessionManager.tsx — Removed mock-only reload/delete semantics and GC copy
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system-ui/client/src/components/ExecutionConsole.tsx — Hardened event rendering for backend event shapes
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system-ui/client/src/components/ExecutionLogViewer.tsx — Added robust kind/payload formatting for backend events
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system-ui/client/src/pages/SystemOverview.tsx — Updated implementation note to daemon-backed reality
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system-ui/client/src/pages/Docs.tsx — Corrected execution/session behavior docs
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system-ui/vite.config.ts — Added `/api/v1` proxy support for dev
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system-ui/DESIGN.md — Added runtime integration/env configuration notes

## 2026-02-09

Closed per consolidation pass before VM-014 implementation

