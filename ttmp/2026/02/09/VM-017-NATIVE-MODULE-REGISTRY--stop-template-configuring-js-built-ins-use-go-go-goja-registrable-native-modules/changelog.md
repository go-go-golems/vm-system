# Changelog

## 2026-02-09

- Initial workspace created


## 2026-02-09

Initialized VM-017 ticket with analysis guide, diary scaffold, and four-step implementation task plan.

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ttmp/2026/02/09/VM-017-NATIVE-MODULE-REGISTRY--stop-template-configuring-js-built-ins-use-go-go-goja-registrable-native-modules/design/01-analysis-and-implementation-guide.md — Defines module policy and implementation sequencing
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ttmp/2026/02/09/VM-017-NATIVE-MODULE-REGISTRY--stop-template-configuring-js-built-ins-use-go-go-goja-registrable-native-modules/reference/01-diary.md — Captures implementation loop and progress evidence


## 2026-02-09

Implemented backend native-module policy: built-in JS modules now rejected for template configuration, go-go-goja registry-backed modules enabled at session startup, and runtime/API integration tests added.

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/vmcontrol/template_service.go — Template add-module validation against native module policy
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/vmmodules/registry.go — Registry-backed module validation and runtime enablement
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/vmsession/session.go — Session runtime now installs configured native modules via require registry
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/vmtransport/http/server.go — ErrModuleNotAllowed mapped to MODULE_NOT_ALLOWED API contract
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/pkg/vmtransport/http/server_native_modules_integration_test.go — Covers require(fs) runtime behavior and JSON built-in semantics

