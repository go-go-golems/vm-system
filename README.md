# VM System: JavaScript VM with Goja

A Go implementation of a JavaScript VM system using goja that integrates with the dual-storage (Git+SQLite) filesystem subsystem.

## Overview

This project implements a VM subsystem that:
- Manages VM profiles (configurations) with module exposure control
- Creates and manages VM sessions (runtime instances)
- Executes code via REPL, run-file, and startup scripts
- Captures and persists complete execution logs
- Integrates with dual-storage workspaces for code access

## Architecture

### Core Components

1. **VM Store**: Database layer for VM profiles, capabilities, and startup files
2. **VM Session Manager**: Creates and manages VM runtime instances
3. **Execution Runner**: Executes REPL, run-file, and startup requests
4. **Module Resolver**: Enforces module allowlist and resolves imports
5. **Event Sink**: Captures and persists execution events

### Key Features

- **VM Profiles**: Template configurations for VM instances
- **Module Exposure Control**: Deny-by-default module allowlist
- **Session Isolation**: One execution at a time per session
- **Complete Event Log**: All stdout/stderr/console/value/exception events
- **Workspace Integration**: Reads code from dual-storage workspaces

## Installation

```bash
go install github.com/go-go-golems/vm-system/cmd/vm-system@latest
```

## Usage

See the CLI documentation for detailed usage instructions.

## Implementation Status

This is a reference implementation following the VM subsystem specification.

## License

MIT
