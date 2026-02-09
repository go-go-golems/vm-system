---
Title: Template-first UI redesign — object model, layout, and implementation plan
Ticket: VM-010-REDESIGN-VM-SYSTEM-UI
Status: active
Topics:
    - frontend
    - docs
DocType: design-doc
Intent: long-term
Owners: []
RelatedFiles:
    - Path: ../../../../../../../vm-system-ui/client/src/components/SessionManager.tsx
      Note: Current session list/actions — to be replaced by new Sessions page
    - Path: ../../../../../../../vm-system-ui/client/src/components/VMConfig.tsx
      Note: Current template config controls — to move into Template Detail view
    - Path: ../../../../../../../vm-system-ui/client/src/lib/vmService.ts
      Note: Data layer and API client — mostly stable, minor additions needed
    - Path: ../../../../../../../vm-system-ui/client/src/pages/Docs.tsx
      Note: Docs page — to be folded into inline help and a lighter reference page
    - Path: ../../../../../../../vm-system-ui/client/src/pages/Home.tsx
      Note: Current editor-first landing — to be replaced by template-first shell
    - Path: ../../../../../../../vm-system-ui/client/src/pages/SystemOverview.tsx
      Note: Architecture page — to be condensed into System status view
    - Path: ../../../../../../../vm-system/pkg/vmmodels/models.go
      Note: Canonical Go struct definitions for Template, Session, Execution, Event
ExternalSources: []
Summary: >
  Pragmatic redesign of vm-system-ui. Restructures navigation around the
  Template → Session → Execution object hierarchy so the UI matches the
  backend data model instead of hiding it behind an editor-first layout.
LastUpdated: 2026-02-09T00:00:00Z
WhatFor: Define scope, layout, and phased plan for the vm-system-ui redesign.
WhenToUse: Reference when implementing or reviewing VM-010 changes.
---

# Template-first UI redesign

## 1. Why redesign

The current UI works but its layout obscures the backend's data model. A user
lands on an editor with tabs for Sessions / Execution Log / VM Config, but the
relationship between those concepts is never made explicit. Consequences:

- **Templates are buried.** The "VM Config" tab shows one template's modules
  and libraries but there is no list of templates, no way to create or duplicate
  one, and no visual connection between a template and the sessions it spawns.

- **Sessions lack context.** The session list shows status and name but not
  which template a session came from. Creating a session does not surface the
  inherited configuration. Users cannot tell why two sessions behave differently.

- **Execution history is disconnected.** Executions are shown for "the current
  session" but there is no way to browse executions across sessions or correlate
  an execution with its parent session/template without reading IDs.

- **The Docs and System Overview pages are heavy.** They duplicate content and
  read like marketing copy instead of quick-reference material for a backend
  tool.

The fix is straightforward: restructure navigation around the object hierarchy
the backend already enforces — **Template → Session → Execution → Events** —
and let each screen show exactly one level of that hierarchy with links to its
parent and children.

## 2. Object model

The backend defines four core objects. The UI should mirror this hierarchy.

```
Template (VM)
│  id, name, engine, exposed_modules[], libraries[],
│  settings {limits, resolver, runtime},
│  capabilities[], startup_files[]
│
└─► Session
    │  id, template_id, workspace_id, base_commit_oid,
    │  worktree_path, status, created_at, closed_at
    │
    └─► Execution
        │  id, session_id, kind (repl | run_file | startup),
        │  input, path, status, started_at, ended_at,
        │  result, error
        │
        └─► ExecutionEvent
              execution_id, seq, ts,
              type (console | value | exception | stdout |
                    stderr | system | input_echo),
              payload
```

**Key scope rules the UI must communicate:**

1. Template changes affect **new** sessions only. Running sessions keep their
   snapshot.
2. A session owns its runtime state. Closing the daemon destroys in-memory
   runtimes even though records persist in SQLite.
3. Executions belong to exactly one session. Events belong to exactly one
   execution.

## 3. Proposed layout

### 3.1 Information architecture

```
VM System
├── Templates            (list → detail with tabs)
├── Sessions             (list → detail with REPL + executions)
├── System               (health, runtime summary)
└── Reference            (condensed object model + API cheatsheet)
```

Four top-level sections. No "Learn" wizard — this is a backend tool. Instead,
every list and detail view gets a one-line description sentence and a small
"scope" badge showing the object hierarchy breadcrumb.

### 3.2 App shell

```
┌──────────────────────────────────────────────────────────────┐
│  ■ VM System          Templates  Sessions  System  Reference │
├──────────────────────────────────────────────────────────────┤
│  Breadcrumb: Templates / Utility Playground / Modules        │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│                    [ Page content ]                           │
│                                                              │
├──────────────────────────────────────────────────────────────┤
│  Session: Workshop #1 (ready)  ·  Template: Utility  ·  3 ex│
└──────────────────────────────────────────────────────────────┘
```

- **Top bar**: logo + nav links. No sidebar — the app has only four sections.
- **Breadcrumb**: always visible, shows where you are in the object hierarchy.
- **Footer bar**: persistent status showing current session, its template, and
  execution count. Clicking any item navigates to that object.

### 3.3 Templates page

**List view** — the default landing page.

```
┌──────────────────────────────────────────────────────────────┐
│  Templates                                   [+ New Template]│
│  Configuration blueprints for VM sessions.                   │
├──────────────────────────────────────────────────────────────┤
│  Name                Engine  Modules  Libs  Sessions  Actions│
│  ─────────────────── ──────  ───────  ────  ────────  ──────│
│  Default JavaScript  goja    4        0     1 ready   [→] [▶]│
│  Utility Playground  goja    5        1     2 ready   [→] [▶]│
│  Library Sandbox     goja    2        2     0         [→] [▶]│
└──────────────────────────────────────────────────────────────┘
```

- Each row shows the template name, engine, module/library counts, and a
  count of active sessions spawned from it.
- **[→]** opens template detail. **[▶]** opens the "Create Session" dialog
  pre-filled with that template.
- **[+ New Template]** opens a create form (name + engine, defaults for
  everything else).

**Detail view** — tabs for the template's sub-resources.

```
┌──────────────────────────────────────────────────────────────┐
│  ← Templates / Utility Playground                            │
│  Engine: goja  ·  5 modules  ·  1 library  ·  2 sessions    │
├──────────────────────────────────────────────────────────────┤
│  [Overview] [Modules] [Libraries] [Startup Files] [Settings] │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  Tab content: checkbox lists for modules/libraries,          │
│  ordered file list for startup, limits/resolver/runtime      │
│  for settings.                                               │
│                                                              │
│  ┌────────────────────────────────────────────────────────┐  │
│  │ ⓘ Changes apply to new sessions only. Running sessions│  │
│  │   keep their original configuration snapshot.          │  │
│  └────────────────────────────────────────────────────────┘  │
│                                                              │
│  Sessions using this template              [+ New Session]   │
│  ─────────────────────────────────────────────────────────   │
│  READY  Workshop #1    created 2m ago             [Open]     │
│  READY  Workshop #2    created 10s ago            [Open]     │
└──────────────────────────────────────────────────────────────┘
```

- The bottom of every template detail shows sessions derived from it, with
  direct links. This reinforces the Template → Session relationship.
- The info banner about scope is always visible on edit tabs.

### 3.4 Sessions page

**List view**

```
┌──────────────────────────────────────────────────────────────┐
│  Sessions                                    [+ New Session] │
│  Running VM instances created from templates.                │
├──────────────────────────────────────────────────────────────┤
│  Status  Name              Template             Created   Act│
│  ──────  ────────────────  ───────────────────  ────────  ───│
│  READY   Workshop #1       Utility Playground   2m ago    [→]│
│  READY   Default Session   Default JavaScript   5m ago    [→]│
│  CLOSED  old-run           Library Sandbox      1d ago    [→]│
└──────────────────────────────────────────────────────────────┘
```

- Grouped or filterable by status (ready / starting / crashed / closed).
- Template name is a link to the template detail.
- **[+ New Session]** opens a dialog with a template picker dropdown,
  inherited config summary, optional name field, and collapsed advanced
  fields (workspace_id, base_commit_oid, worktree_path).

**Detail view** — the primary workspace for interacting with a session.

```
┌──────────────────────────────────────────────────────────────┐
│  ← Sessions / Workshop #1                                    │
│  Template: Utility Playground  ·  Status: READY  ·  id: a3f…│
├──────────────────────────────────────────────────────────────┤
│  [REPL]  [Executions]                                        │
├──────────────────────────┬───────────────────────────────────┤
│  Editor                  │  Output                           │
│  ─────────────────────── │  ──────────────────────────────── │
│  > const x = 21;         │  12:04:10  Hello, VM!             │
│  > x * 2;                │  12:04:10  → 42                   │
│                          │                                   │
│                          │  12:04:20  Error: oops            │
│                          │  12:04:20  ✗ oops                 │
│  ┌──────────────────┐    │                                   │
│  │ Examples ▾       │    │                                   │
│  └──────────────────┘    │                                   │
│  [Run]  [Clear]          │                                   │
├──────────────────────────┴───────────────────────────────────┤
│  Recent executions                                           │
│  #3 repl ok  42     12:04:20  1ms                            │
│  #2 repl err oops   12:04:15  2ms                            │
│  #1 repl ok  "Hi"   12:04:10  1ms                            │
└──────────────────────────────────────────────────────────────┘
```

- The REPL tab is a two-panel layout: editor on the left, console output on
  the right. Below both panels is a compact execution history strip for the
  current session (click to expand details or re-load input).
- The Executions tab shows the full filterable execution log with event
  details (existing `ExecutionLogViewer` component, mostly unchanged).
- The template name in the header links back to the template detail.
- When the session is not ready, the editor is disabled and a status banner
  explains why.

### 3.5 Create Session dialog

```
┌──────────────────────────────────────────────┐
│  Create Session                              │
├──────────────────────────────────────────────┤
│  Template *                                  │
│  [ Utility Playground           ▾ ]          │
│                                              │
│  Inherited: 5 modules, 1 library,            │
│  cpu 2000ms, wall 5000ms, mem 128MB          │
│                                              │
│  Name (optional)                             │
│  [ ________________________________ ]        │
│                                              │
│  ▸ Advanced                                  │
│    workspace_id:    [ ws-web-ui           ]  │
│    base_commit_oid: [ web-ui              ]  │
│    worktree_path:   [ /tmp                ]  │
│                                              │
│                    [Cancel]  [Create Session] │
└──────────────────────────────────────────────┘
```

- Template selection is mandatory and always first. Changing the template
  updates the "Inherited" summary instantly.
- Advanced fields are collapsed by default and pre-filled with env defaults.

### 3.6 System page

A single page replacing the current `SystemOverview.tsx` — much lighter.

```
┌──────────────────────────────────────────────────────────────┐
│  System                                                      │
│  Runtime health and daemon status.                           │
├──────────────────────────────────────────────────────────────┤
│  Daemon          ● connected     API: /api/v1                │
│  Templates       3 active                                    │
│  Sessions        2 ready / 1 closed                          │
│  Executions      47 total (last 24h)                         │
├──────────────────────────────────────────────────────────────┤
│  Runtime                                                     │
│  Engine: goja (ES5.1)  ·  Storage: Git + SQLite              │
│  Default limits: cpu 2000ms, wall 5000ms, mem 128MB          │
├──────────────────────────────────────────────────────────────┤
│  ⓘ Runtime sessions are daemon-owned. Restarting the daemon  │
│    destroys in-memory state. Persisted records remain in DB. │
└──────────────────────────────────────────────────────────────┘
```

No marketing copy. Just facts and numbers.

### 3.7 Reference page

Replaces the current verbose Docs page with a compact reference:

- **Object model** — the hierarchy diagram from §2 plus one-paragraph
  descriptions of each object.
- **API cheatsheet** — table of endpoints grouped by resource
  (templates, sessions, executions, events) with method, path, and
  one-line description.
- **Available globals** — list of JS globals available in the runtime
  (console, Math, Date, JSON, etc.).
- **Preset examples** — the existing code snippets, rendered as
  collapsible cards.

No "Getting Started" wizard. No three-column feature grids. If someone needs
a walkthrough, the template list page is self-explanatory: pick a template,
create a session, run code.

## 4. What changes from the current UI

| Current | Proposed | Why |
|---------|----------|-----|
| Editor-first homepage with tabs | Templates list as landing page | Matches object hierarchy; makes templates discoverable |
| Sessions tab inside Home | Dedicated `/sessions` route | Sessions deserve their own list + detail views |
| VM Config tab inside Home | Template detail tabs | Config belongs to the template, not the editor |
| Execution Log tab inside Home | Executions tab inside Session detail | Executions belong to a session |
| REPL editor as primary view | REPL panel inside Session detail | Editing happens *in* a session, not before one exists |
| Separate Docs page (800+ lines) | Compact Reference page | Backend tool users want a cheatsheet, not a textbook |
| Separate SystemOverview page | Compact System status page | Show live stats, not architecture prose |
| No breadcrumbs | Always-visible breadcrumb + footer status | Makes current location and object scope obvious |

## 5. Implementation plan

### Phase 1: Shell and routing ✅

- ✅ Added shared app shell (`AppShell.tsx`) with top nav, breadcrumb, and footer status bar.
- ✅ Added routes: `/templates`, `/templates/:id`, `/sessions`, `/sessions/:id`,
  `/system`, `/reference`.
- ✅ Redirect `/` to `/templates`.
- ✅ Removed old `/docs` and `/overview` routes from App.tsx routing.

### Phase 2: Templates ✅

- ✅ Built template list page (`Templates.tsx`) with table view, session counts,
  and quick actions (Open, Create Session).
- ✅ Built template detail page (`TemplateDetail.tsx`) with Overview / Modules /
  Libraries / Startup Files / Settings tabs.
  - Modules and Libraries tabs: adapted from existing `VMConfig` checkbox UI.
  - Settings tab: read-only display of limits, resolver, runtime.
- ✅ Added "Create Template" action (name + engine).
- ⬜ "Duplicate Template" action (deferred — needs backend clone endpoint).
- ✅ Derived sessions list at bottom of detail view.

### Phase 3: Sessions ✅

- ✅ Built session list page (`Sessions.tsx`) with status filtering.
- ✅ Built session detail page (`SessionDetail.tsx`) with two tabs: REPL and Executions.
  - REPL tab: two-panel layout (editor + console) plus compact execution strip.
    Reuses `CodeEditor`, `ExecutionConsole`, `PresetSelector`.
  - Executions tab: reuses existing `ExecutionLogViewer` as-is.
- ✅ Built `CreateSessionDialog` with template picker, config summary, and
  advanced fields.
- ✅ Wired "Create Session from Template" flow (template list/detail → dialog →
  redirect to new session detail).

### Phase 4: System and Reference ✅

- ✅ Built System page (`System.tsx`) with live stats and architecture summary.
- ✅ Built Reference page (`Reference.tsx`) with object model, API endpoint table,
  globals list, collapsible code examples.
- ⬜ Old Docs.tsx and SystemOverview.tsx files still exist but are unreachable
  (removed from routing). Can be deleted in cleanup.

### Phase 5: Polish (remaining)

- ⬜ Keyboard shortcuts (Cmd/Ctrl+Enter to run).
- ⬜ Accessibility pass (focus management, ARIA labels).
- ⬜ Delete old unreachable page files (Home.tsx, Docs.tsx, SystemOverview.tsx).

## 6. API surface

The existing REST API covers all core flows. No new endpoints are strictly
required, but a few additions would improve the UX:

| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/api/v1/templates/:id/clone` | POST | Duplicate a template (avoids client-side re-creation) |
| `/api/v1/stats` | GET | Return aggregate counts for System page (optional — client can compute from list endpoints) |

Everything else — template CRUD, module/library management, session
lifecycle, execution + events — already exists and the `vmService` client
handles it.

## 7. Design decisions

1. **Templates as landing page.** The object hierarchy starts with templates.
   Making them the first thing a user sees establishes the mental model without
   a tutorial.

2. **No onboarding wizard.** This is a developer backend tool. The UI should
   be self-documenting through clear labels, scope badges, and one-line
   descriptions — not step-by-step wizards.

3. **REPL lives inside Session detail.** Running code only makes sense in the
   context of a session. Putting the editor inside the session view eliminates
   the "which session am I running in?" confusion.

4. **Compact Reference instead of heavy Docs.** Developers scan reference
   pages; they don't read documentation websites for internal tools. A single
   page with the object model, API table, and code examples is enough.

5. **Footer status bar.** A persistent indicator of the "current" session and
   template provides context without requiring the user to navigate away from
   their current view.

## 8. Open questions

1. **Template editing granularity.** Should limits/resolver/runtime be editable
   from the UI, or only via API/CLI? Current recommendation: display as
   read-only in the UI, edit via API. Add UI editing later if there's demand.

2. **Session auto-select.** When a user creates a session, should it
   automatically become the "current" session in the footer bar? Current
   recommendation: yes.

3. **Execution event streaming.** The current implementation fetches events
   after execution completes. For long-running executions, SSE or polling would
   give better feedback. Not in scope for this redesign — add later.
