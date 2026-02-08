# Tasks

## TODO

- [x] Add tasks here

- [x] Map vm-system architecture and runtime flow
- [x] Run build/tests and CLI experiments
- [x] Write detailed diary and 10+ page report
- [x] Relate files and update changelog
- [x] Design daemonized runtime host architecture with REST API and CLI client model
- [x] Publish daemon architecture document and upload to reMarkable
- [x] Define v2 implementation backlog for daemon/core/template cutover
- [ ] Create pkg/vmcontrol reusable core with explicit ports and service wiring
- [ ] Implement pkg/vmdaemon process host and add vm-system serve command
- [ ] Implement pkg/vmtransport/http REST adapter for health, templates, sessions, executions, events, and runtime summary
- [ ] Implement pkg/vmclient REST client and switch CLI runtime commands to client mode by default
- [ ] Cut over CLI naming from vm to template and remove vm command registration
- [ ] Add execution/runtime safety hooks (path normalization and core limits scaffolding)
- [ ] Add integration tests proving cross-request session continuity through daemon HTTP API
- [ ] Update smoke/e2e scripts and README for daemon-first usage
- [ ] Record diary/changelog updates with commit hashes for each completed task
