# Changelog

## 2026-02-09

- Created initial decomposition plan for P2-1 targets

## 2026-02-08

Added detailed implementation guide (design/02-implementation-guide.md) for staged monolith decomposition

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ttmp/2026/02/08/VM-014-DECOMPOSE-MONOLITHS--p2-1-decompose-monolithic-files-across-vm-system-and-vm-system-ui/design/02-implementation-guide.md — Detailed decomposition execution guide


## 2026-02-09

Completed Slice A: split HTTP server by resource domain (commit 204b74c)

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/vmtransport/http/server.go — Reduced to router and shared entrypoints
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/vmtransport/http/server_errors.go — Extracted parse/decode/error helpers
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/vmtransport/http/server_executions.go — Extracted execution handlers/types
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/vmtransport/http/server_sessions.go — Extracted session handlers/types
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/vmtransport/http/server_templates.go — Extracted template handlers/types


## 2026-02-09

Completed Slice B: split vmstore by persistence aggregate (commit 8d54ec3)

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/vmstore/vmstore.go — Reduced to store type/constructor/connection lifecycle
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/vmstore/vmstore_executions.go — Extracted execution/event persistence
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/vmstore/vmstore_migrations.go — Extracted schema initialization
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/vmstore/vmstore_sessions.go — Extracted session persistence
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/vmstore/vmstore_templates.go — Extracted template/settings/capability/startup persistence


## 2026-02-09

Completed Slice C: split template CLI command builders (commit 48c0311)

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/cmd/vm-system/cmd_template.go — Reduced to root command + shared settings types
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/cmd/vm-system/cmd_template_core.go — Extracted core command builders
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/cmd/vm-system/cmd_template_libraries.go — Extracted library command builders
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/cmd/vm-system/cmd_template_modules.go — Extracted module command builders
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/cmd/vm-system/cmd_template_startup.go — Extracted startup command builders


## 2026-02-09

Completed Slice D: split frontend API layer into transport + domain endpoint modules while preserving @/lib/api hook surface (vm-system-ui commit 625e94b)

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system-ui/client/src/lib/api.ts — Reduced to compatibility facade with stable hook exports
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system-ui/client/src/lib/vm/endpoints/executions.ts — Extracted execution/event endpoint operations
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system-ui/client/src/lib/vm/endpoints/sessions.ts — Extracted session endpoint operations and UI state mapping
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system-ui/client/src/lib/vm/endpoints/templates.ts — Extracted template endpoint operations
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system-ui/client/src/lib/vm/transport.ts — Extracted URL and fetch base query concerns


## 2026-02-09

Aligned Slice D guide paths with implemented api.ts decomposition to prevent doc drift

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ttmp/2026/02/08/VM-014-DECOMPOSE-MONOLITHS--p2-1-decompose-monolithic-files-across-vm-system-and-vm-system-ui/design/02-implementation-guide.md — Updated Slice D file targets to match implemented frontend modules


## 2026-02-09

All decomposition slices (A-D) completed; task list is fully done

