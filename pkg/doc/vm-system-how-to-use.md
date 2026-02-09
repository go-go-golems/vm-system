---
Title: vm-system how to use
Slug: how-to-use
Short: Quick start for vm-system CLI
Topics:
- vm-system
- cli
IsTopLevel: true
ShowPerDefault: true
SectionType: Tutorial
---

# vm-system quick start

## Daemon

```bash
vm-system serve --db vm-system.db --listen 127.0.0.1:3210
```

## Template flow

```bash
vm-system template create --name demo --engine goja
vm-system template list
```

## Session flow

```bash
vm-system session create --template-id <template-id> --workspace-id ws --base-commit deadbeef --worktree-path /abs/path
vm-system session list
vm-system session close <session-id>
```

## Execution flow

```bash
vm-system exec repl <session-id> '1+2'
vm-system exec run-file <session-id> app.js
```

## Operations

```bash
vm-system ops health
vm-system ops runtime-summary
```
