# Changelog

## 2026-02-08

- Initial workspace created.

## 2026-02-09

- Added exhaustive design doc for glazed migration with command-by-command inventory and root flattening plan.
- Rewrote design doc to be strictly new-API-first (`schema/fields/values/sources`) and removed legacy vocabulary/mapping content.
- Added detailed diary with command traces, failure logs, decisions, and review instructions.
- Replaced placeholder tasks with phased implementation checklist covering root wiring, porting, tests, docs, and upload.
- Replaced broad ticket tasks with implementation-ordered checklist (`T01`-`T20`) for commit-by-commit execution.
- Uploaded bundled plan + diary PDFs to reMarkable at `/ai/2026/02/09/VM-011-PORT-VM-CLI-TO-GLAZED` (`VM-011 Glazed Migration Plan` and `VM-011 Glazed Migration Plan Final`).
- Implemented root flattening slice: added Glazed dependency, wired help/logging in root, removed `http` from root registration, and added new `ops` command + vmclient ops methods.
- Added shared Glazed helper layer and ported `serve` + `libs` command groups to new API command implementations.
- Ported `session` command group to Glazed command implementations and switched canonical lifecycle CLI verb to `session close` (no `session delete` registration).
- Ported the full `exec` command group to Glazed command implementations while preserving JSON flag parsing and output formatting behavior.
- Ported the full `template` command group to Glazed command implementations (all existing template verbs).
