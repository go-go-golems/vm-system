# VM System UI Design

## Design Philosophy: Technical Precision

**Movement**: Swiss Design meets Developer Tools  
**Core Principles**:
- Functional clarity with precise typography
- High-contrast monospace code display
- Structured grid-based layouts with clear hierarchy
- Minimal color palette focused on code syntax highlighting

**Color Philosophy**:
- Background: Deep slate (#0f172a) for reduced eye strain
- Foreground: Crisp white (#f8fafc) for maximum readability
- Accent: Electric blue (#3b82f6) for interactive elements
- Success: Emerald (#10b981) for successful executions
- Error: Rose (#f43f5e) for errors and warnings

**Layout Paradigm**: Split-pane IDE-style interface
- Left: Code editor with Monaco (VSCode engine)
- Right: Execution console and VM controls
- Top: Toolbar with presets and VM configuration

**Signature Elements**:
- Monospace typography throughout (JetBrains Mono)
- Syntax-highlighted code blocks
- Real-time execution feedback with event streaming

**Interaction Philosophy**:
- Keyboard-first navigation (Cmd+Enter to execute)
- Instant feedback on code changes
- Collapsible panels for focus mode

**Typography System**:
- Display: JetBrains Mono Bold for headers
- Code: JetBrains Mono Regular for editor
- UI: Inter for controls and labels

## Runtime Integration

- The UI is a real client of the daemon HTTP API:
  - `GET/POST /api/v1/templates...`
  - `GET/POST/DELETE /api/v1/sessions...`
  - `GET/POST /api/v1/executions...`
- Browser code does not execute user snippets; all execution happens in daemon-owned goja sessions.
- In dev mode, Vite proxies `/api/v1` to `VM_SYSTEM_API_PROXY_TARGET` (default `http://127.0.0.1:3210`).
- Optional client environment variables:
  - `VITE_VM_SYSTEM_API_BASE_URL` (absolute or path prefix for API requests)
  - `VITE_VM_SYSTEM_WORKSPACE_ID`
  - `VITE_VM_SYSTEM_BASE_COMMIT_OID`
  - `VITE_VM_SYSTEM_WORKTREE_PATH`
