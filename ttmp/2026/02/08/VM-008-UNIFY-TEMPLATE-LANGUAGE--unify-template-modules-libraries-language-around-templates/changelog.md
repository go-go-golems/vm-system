# Changelog

## 2026-02-08

- Initial workspace created

## 2026-02-08

Created detailed VM-008 review/design document and concrete implementation task list for unifying template/modules/libraries language around template-first API/CLI/docs with no backwards compatibility.

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ttmp/2026/02/08/VM-008-UNIFY-TEMPLATE-LANGUAGE--unify-template-modules-libraries-language-around-templates/design-doc/01-template-language-unification-review-and-implementation-plan.md - Deep review and implementation blueprint
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ttmp/2026/02/08/VM-008-UNIFY-TEMPLATE-LANGUAGE--unify-template-modules-libraries-language-around-templates/tasks.md - Concrete task breakdown for delegated implementation

## 2026-02-08

Task 1: Finalized VM-008 terminology contract to enforce template-centric user-facing language only (no vm-id/modules compatibility aliases) and documented explicit scope boundary for internal runtime model naming.

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ttmp/2026/02/08/VM-008-UNIFY-TEMPLATE-LANGUAGE--unify-template-modules-libraries-language-around-templates/design-doc/01-template-language-unification-review-and-implementation-plan.md — Added finalized terminology contract section
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ttmp/2026/02/08/VM-008-UNIFY-TEMPLATE-LANGUAGE--unify-template-modules-libraries-language-around-templates/tasks.md — Marked Task 1 complete


## 2026-02-08

Task 2: Added template module/library API endpoints for add/remove/list operations under /api/v1/templates/{template_id}/(modules|libraries), backed by template service logic so CLI no longer requires direct DB mutation paths for these operations.

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/vmcontrol/ports.go — Extended template store port for VM updates required by module/library mutations
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/vmcontrol/template_service.go — Added template module/library add/remove/list domain methods
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/vmcontrol/template_service_test.go — Updated test stub for new port method requirement
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/vmtransport/http/server.go — Added template module/library routes and handlers
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ttmp/2026/02/08/VM-008-UNIFY-TEMPLATE-LANGUAGE--unify-template-modules-libraries-language-around-templates/tasks.md — Marked Task 2 complete


## 2026-02-08

Task 4: Extended vmclient template client with module/library add/remove/list operations on new template routes to support CLI migration off direct DB writes.

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/vmclient/templates_client.go — Added template module/library request/response types and CRUD client methods
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ttmp/2026/02/08/VM-008-UNIFY-TEMPLATE-LANGUAGE--unify-template-modules-libraries-language-around-templates/tasks.md — Marked Task 4 complete


## 2026-02-08

Task 3: Removed ad-hoc command-side mutation logic from the legacy modules command by switching add-module/add-library to vmclient template API calls, keeping mutation flow in daemon template service/store paths only.

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/cmd/vm-system/cmd_modules.go — Replaced direct store mutation with template API client calls
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ttmp/2026/02/08/VM-008-UNIFY-TEMPLATE-LANGUAGE--unify-template-modules-libraries-language-around-templates/tasks.md — Marked Task 3 complete


## 2026-02-08

Task 5: Added template-native CLI subcommands for module/library add/remove/list plus available catalog listing under the template command surface.

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/cmd/vm-system/cmd_template.go — Added template module/library management and list-available commands
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ttmp/2026/02/08/VM-008-UNIFY-TEMPLATE-LANGUAGE--unify-template-modules-libraries-language-around-templates/tasks.md — Marked Task 5 complete


## 2026-02-08

Task 6: Deleted the legacy modules command file and removed modules command registration from the root CLI, leaving template as the only module/library command surface.

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/cmd/vm-system/cmd_modules.go — Deleted legacy modules command surface
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/cmd/vm-system/main.go — Removed modulesCmd root registration
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ttmp/2026/02/08/VM-008-UNIFY-TEMPLATE-LANGUAGE--unify-template-modules-libraries-language-around-templates/tasks.md — Marked Task 6 complete


## 2026-02-08

Task 7: Completed remaining CLI wording cleanup for template-targeted identifiers by replacing user-facing VM ID labels with Template ID in session command outputs; no template command flags use vm-id.

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/cmd/vm-system/cmd_session.go — Renamed user-facing VM ID labels to Template ID
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ttmp/2026/02/08/VM-008-UNIFY-TEMPLATE-LANGUAGE--unify-template-modules-libraries-language-around-templates/tasks.md — Marked Task 7 complete


## 2026-02-08

Task 8: Expanded integration coverage for template module/library routes and added CLI command-surface test coverage for template module/library command paths, replacing legacy assumptions about modules command usage.

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/cmd/vm-system/cmd_template_test.go — Added template command-path registration test
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/vmtransport/http/server_templates_integration_test.go — Added module/library nested resource integration assertions
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ttmp/2026/02/08/VM-008-UNIFY-TEMPLATE-LANGUAGE--unify-template-modules-libraries-language-around-templates/tasks.md — Marked Task 8 complete

