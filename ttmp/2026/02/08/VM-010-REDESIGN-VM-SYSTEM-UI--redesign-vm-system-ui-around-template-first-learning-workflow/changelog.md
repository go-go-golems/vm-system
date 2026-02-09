# Changelog

## 2026-02-08

- Initial workspace created

## 2026-02-09

- Added detailed redesign plan focused on learner mental model and template-first UX flow.
- Documented canonical object model (Template -> Session -> Execution -> ExecutionEvent).
- Added multi-screen ASCII wireframes for shell, templates, session creation, session detail, onboarding, and mobile.
- Added phased implementation plan and execution task backlog for follow-on engineering.

## 2026-02-08

Rewrote design doc: trimmed learner/onboarding fluff, restructured around pragmatic backend-tool layout, tightened wireframes, replaced verbose Docs/SystemOverview with compact Reference and System pages, simplified implementation phases from 6 to 5, added change-from-current comparison table

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system/ttmp/2026/02/08/VM-010-REDESIGN-VM-SYSTEM-UI--redesign-vm-system-ui-around-template-first-learning-workflow/design-doc/01-template-first-learner-ui-redesign-plan-with-object-model-and-wireframes.md — Rewritten design document


## 2026-02-08

Implemented full UI redesign: new AppShell with top nav/breadcrumb/footer status bar, Templates list+detail pages with tabs (overview/modules/libraries/startup/settings), Sessions list+detail pages with REPL workspace+execution log, compact System status page, Reference cheatsheet page, CreateSessionDialog with template picker, new routing (/ redirects to /templates), removed old Home/Docs/SystemOverview pages from routing

### Related Files

- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system-ui/client/src/App.tsx — Rewired routing to new pages
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system-ui/client/src/components/AppShell.tsx — New shared app shell with nav
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system-ui/client/src/components/CreateSessionDialog.tsx — Reusable create session dialog with template picker and config summary
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system-ui/client/src/pages/Reference.tsx — Compact reference with object model
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system-ui/client/src/pages/SessionDetail.tsx — Session workspace with REPL panel and execution log
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system-ui/client/src/pages/Sessions.tsx — Session list with status filtering
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system-ui/client/src/pages/System.tsx — Compact system status page
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system-ui/client/src/pages/TemplateDetail.tsx — Template detail with 5 tabs and derived sessions list
- /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system-ui/client/src/pages/Templates.tsx — Template list page (new landing page)


## 2026-02-09

Closed per consolidation pass before VM-014 implementation

