---
Title: RTK Query migration analysis and implementation plan
Ticket: VM-011-RTK-QUERY-STATE-MIGRATION
Status: active
Topics:
    - frontend
    - state-management
DocType: analysis
Intent: long-term
Owners: []
RelatedFiles:
    - Path: /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system-ui/client/src/lib/vmService.ts
      Note: Monolithic VMService class — the entire data layer to be replaced
    - Path: /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system-ui/client/src/components/AppShell.tsx
      Note: Hand-rolled AppState context — to be replaced by Redux store + selectors
    - Path: /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system-ui/client/src/pages/Templates.tsx
      Note: Consumes AppState + raw fetch for template creation
    - Path: /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system-ui/client/src/pages/TemplateDetail.tsx
      Note: Consumes AppState + vmService for module/library mutations
    - Path: /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system-ui/client/src/pages/Sessions.tsx
      Note: Consumes AppState for session list + status filtering
    - Path: /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system-ui/client/src/pages/SessionDetail.tsx
      Note: Consumes AppState + vmService for REPL execution, session close
    - Path: /home/manuel/code/wesen/corporate-headquarters/vm-system/vm-system-ui/client/src/components/CreateSessionDialog.tsx
      Note: Consumes templates prop + vmService for session creation
ExternalSources: []
Summary: >
  Detailed analysis of every state/query location in vm-system-ui, what
  breaks today, and a phased plan to migrate to @reduxjs/toolkit + RTK Query.
LastUpdated: 2026-02-09T00:00:00Z
WhatFor: Audit current data layer and produce migration plan for RTK Query.
WhenToUse: Reference when implementing the RTK Query migration.
---

# RTK Query migration analysis and implementation plan

## 1. Why migrate

The current data layer is a single `VMService` class (≈650 lines) that owns:
- All API calls (fetch wrappers)
- All normalization (snake_case → camelCase, dates, defaults)
- All caching (in-memory Maps)
- All client-side state (current session, session names in localStorage)
- Bootstrap logic (auto-create default templates + first session)

This is consumed through a hand-rolled React context (`AppStateContext` in
`AppShell.tsx`) that exposes `refreshTemplates()` / `refreshSessions()` callbacks.
Pages call those imperatively after mutations, and bugs happen when they don't —
e.g., creating a template via raw `fetch()` in `Templates.tsx` because `vmService`
has no `createTemplate()` method, then hoping `refreshTemplates()` picks it up
through the singleton's private cache.

Problems today:

| Problem | Example |
|---------|---------|
| **Stale data after mutations** | `Templates.tsx` creates a template via raw fetch, calls `refreshTemplates()` which calls `vmService.initialize()` → `getVMs()`. But `initialize()` is guarded by an `initialized` flag, so the private `refreshTemplates()` inside `vmService` is never re-called. The list doesn't update. |
| **No loading/error states per query** | `AppShell` has a single `initialized` boolean. There's no per-page loading or error state. Pages show a spinner until the global init completes, then nothing if a subsequent fetch fails silently. |
| **Duplicate caching** | `vmService` maintains `Map<string, VMProfile>`, `Map<string, VMSession>`, `Map<string, Execution>`. The React context duplicates this into `useState` arrays. Two sources of truth that drift. |
| **Imperative refresh** | Every mutation handler must remember to call the right combination of `refreshTemplates()` / `refreshSessions()` / `setCurrentSession()`. Easy to forget, hard to test. |
| **Tangled bootstrap** | `ensureInitialized()` auto-creates default templates and a first session if none exist. This is side-effectful initialization mixed into the data layer. |
| **No cache invalidation** | Changing a template's modules calls `updateTemplateModules()` then `refreshTemplates()`, but the session list isn't refreshed, so `session.vm` objects go stale until the next manual refresh. |

RTK Query solves all of these with declarative cache tags, automatic
refetching after mutations, per-query loading/error states, and a single
normalized Redux store as the source of truth.

## 2. Current state inventory

### 2.1 VMService class — API methods

Each of these becomes an RTK Query endpoint.

| VMService method | HTTP call | RTK Query type | Cache tag(s) |
|---|---|---|---|
| `refreshTemplates()` (private) | `GET /api/v1/templates` then `GET /api/v1/templates/:id` for each | **query**: `listTemplates` | `{type:'Template', id:'LIST'}` |
| `fetchTemplateProfile(id)` (private) | `GET /api/v1/templates/:id` | **query**: `getTemplate(id)` | `{type:'Template', id}` |
| (raw fetch in Templates.tsx) | `POST /api/v1/templates` | **mutation**: `createTemplate` | invalidates `{type:'Template', id:'LIST'}` |
| `updateTemplateModules(id, modules)` | `GET` + `POST` + `DELETE` modules | **mutation**: `updateTemplateModules` | invalidates `{type:'Template', id}`, `{type:'Template', id:'LIST'}` |
| `updateTemplateLibraries(id, libs)` | `GET` + `POST` + `DELETE` libraries | **mutation**: `updateTemplateLibraries` | invalidates `{type:'Template', id}`, `{type:'Template', id:'LIST'}` |
| `loadSessions(status?)` (private) | `GET /api/v1/sessions` | **query**: `listSessions` | `{type:'Session', id:'LIST'}` |
| `getSession(id)` | `GET /api/v1/sessions/:id` | **query**: `getSession(id)` | `{type:'Session', id}` |
| `createSession(templateId, name?)` → `createSessionInternal` | `POST /api/v1/sessions` | **mutation**: `createSession` | invalidates `{type:'Session', id:'LIST'}` |
| `closeSession(id)` | `POST /api/v1/sessions/:id/close` | **mutation**: `closeSession` | invalidates `{type:'Session', id}`, `{type:'Session', id:'LIST'}` |
| `deleteSession(id)` | `DELETE /api/v1/sessions/:id` | **mutation**: `deleteSession` | invalidates `{type:'Session', id}`, `{type:'Session', id:'LIST'}` |
| `getExecutionsBySession(sessionId)` | `GET /api/v1/executions?session_id=` | **query**: `listExecutions(sessionId)` | `{type:'Execution', id:'LIST'}` |
| `getExecution(id)` | `GET /api/v1/executions/:id` | **query**: `getExecution(id)` | `{type:'Execution', id}` |
| `getExecutionEvents(execId)` (private) | `GET /api/v1/executions/:id/events` | **query**: `getExecutionEvents(id)` | `{type:'Event', id:execId}` |
| `executeREPL(code, sessionId?)` | `POST /api/v1/executions/repl` | **mutation**: `executeREPL` | invalidates `{type:'Execution', id:'LIST'}` |

**Tag types**: `Template`, `Session`, `Execution`, `Event`

### 2.2 Client-only state (Redux slice, not RTK Query)

These are not server-fetched; they're UI state that lives in localStorage
or in-memory.

| State | Current location | Redux slice |
|---|---|---|
| `currentSessionId` | `vmService.currentSessionId` + localStorage | `uiSlice.currentSessionId` |
| `sessionNames` (aliases) | `vmService.sessionNames` + localStorage | `uiSlice.sessionNames` |
| REPL editor code | `SessionDetail` local `useState` | stays local (component state) — not global |
| Execution list for REPL console | `SessionDetail` local `useState` | stays local — populated from `listExecutions` query |
| Active tab per page | local `useState` | stays local |
| Create dialog open/templateId | local `useState` | stays local |

### 2.3 Normalization helpers

These functions stay (extracted to a `lib/normalize.ts`):

- `toDate`, `toRecord`, `parseNumber`, `parseBool`, `parseStringArray`, `asArray`
- `normalizeLimits`, `normalizeResolver`, `normalizeRuntime`
- `normalizeExecutionResult`, `normalizeExecutionError`
- `mapExecutionKind`

They'll be used inside RTK Query's `transformResponse` callbacks.

### 2.4 Bootstrap logic

`ensureInitialized()` auto-creates default templates and a first session.
This should become an explicit `useBootstrap()` hook called once in `AppShell`,
using RTK Query mutations, not baked into the data layer.

### 2.5 Static data

These stay as plain exports (no API involved):

- `BUILTIN_MODULES`, `BUILTIN_LIBRARIES`, `PRESET_EXAMPLES`
- `DEFAULT_LIMITS`, `DEFAULT_RESOLVER`, `DEFAULT_RUNTIME`

## 3. File-by-file migration map

### 3.1 `lib/vmService.ts` → split into 3 files

| New file | Content |
|----------|---------|
| `lib/api.ts` | RTK Query `createApi` with `baseQuery` (reuse existing `getURL` + `request` pattern as a custom `fetchBaseQuery`), all endpoint definitions |
| `lib/normalize.ts` | Pure normalization functions (extracted from vmService) |
| `lib/types.ts` | TypeScript interfaces (`VMProfile`, `VMSession`, `Execution`, `ExecutionEvent`, etc.) + static data (`BUILTIN_MODULES`, `BUILTIN_LIBRARIES`, `PRESET_EXAMPLES`) |

`vmService.ts` is deleted after migration. The `VMService` class is entirely
replaced by RTK Query endpoints + a Redux slice.

### 3.2 `lib/store.ts` (new)

```
configureStore({
  reducer: {
    [vmApi.reducerPath]: vmApi.reducer,
    ui: uiReducer,
  },
  middleware: (getDefault) => getDefault().concat(vmApi.middleware),
})
```

### 3.3 `lib/uiSlice.ts` (new)

```
createSlice({
  name: 'ui',
  initialState: {
    currentSessionId: localStorage.getItem(...) || null,
    sessionNames: JSON.parse(localStorage.getItem(...)) || {},
  },
  reducers: {
    setCurrentSessionId,
    setSessionName,
  },
})
```

With a localStorage middleware (or `listenerMiddleware`) that persists on change.

### 3.4 `components/AppShell.tsx`

**Before**: owns `useState` for templates/sessions/currentSession, provides
`refreshTemplates`/`refreshSessions` via context.

**After**: wraps children in `<Provider store={store}>`. Remove `AppStateContext`.
The `useAppState()` hook is replaced by individual hooks:

| Old | New |
|-----|-----|
| `useAppState().templates` | `useListTemplatesQuery()` → `data` |
| `useAppState().sessions` | `useListSessionsQuery()` → `data` |
| `useAppState().currentSession` | `useSelector(selectCurrentSession)` (derived from `ui.currentSessionId` + session cache) |
| `useAppState().initialized` | `isLoading` from queries |
| `useAppState().refreshTemplates()` | not needed (automatic) |
| `useAppState().refreshSessions()` | not needed (automatic) |
| `useAppState().setCurrentSession()` | `dispatch(setCurrentSessionId(id))` |

Breadcrumbs and footer status bar become selector-driven — they subscribe
to the RTK Query cache directly.

### 3.5 `pages/Templates.tsx`

**Before**: reads `templates` from context, creates template via raw `fetch()`,
calls `refreshTemplates()`.

**After**:
```tsx
const { data: templates, isLoading } = useListTemplatesQuery();
const { data: sessions } = useListSessionsQuery();
const [createTemplate] = useCreateTemplateMutation();

const handleCreate = async () => {
  await createTemplate({ name, engine: 'goja' });
  // no manual refresh needed — tag invalidation does it
};
```

The raw `fetch()` call and `vmService` import are removed.

### 3.6 `pages/TemplateDetail.tsx`

**Before**: calls `vmService.getVM(id)` synchronously from cache, falls back to
`refreshTemplates()`. Calls `vmService.updateTemplateModules/Libraries` and
then `refreshTemplates()` manually.

**After**:
```tsx
const { data: template, isLoading } = useGetTemplateQuery(templateId);
const { data: sessions } = useListSessionsQuery();
const [updateModules] = useUpdateTemplateModulesMutation();
const [updateLibraries] = useUpdateTemplateLibrariesMutation();

const handleModuleToggle = async (moduleId) => {
  const next = /* compute new set */;
  await updateModules({ templateId, modules: next });
  // cache auto-invalidates Template + Template LIST
};
```

Local `selectedModules`/`selectedLibraries` `useState` can be replaced by
optimistic updates in the mutation, or kept as local UI state for instant
checkbox feedback with rollback on error.

### 3.7 `pages/Sessions.tsx`

**Before**: reads `sessions` from context.

**After**:
```tsx
const { data: sessions, isLoading } = useListSessionsQuery();
const { data: templates } = useListTemplatesQuery();
```

Minimal change — just swap the data source.

### 3.8 `pages/SessionDetail.tsx`

**Before**: calls `vmService.getSession(id)`, `vmService.getExecutionsBySession(id)`,
`vmService.executeREPL()`, `vmService.closeSession()`. Manually manages
`session`/`executions` state and calls `refreshSessions()`.

**After**:
```tsx
const { data: session } = useGetSessionQuery(sessionId);
const { data: executions } = useListExecutionsQuery(sessionId);
const [executeREPL] = useExecuteREPLMutation();
const [closeSession] = useCloseSessionMutation();
const dispatch = useDispatch();

useEffect(() => {
  if (session) dispatch(setCurrentSessionId(session.id));
}, [session]);

const handleExecute = async () => {
  await executeREPL({ sessionId, input: code });
  // Execution LIST tag invalidated → executions query refetches
};

const handleClose = async () => {
  await closeSession(sessionId);
  // Session tag invalidated → session query refetches
};
```

Local `executions` state is removed — the query result is the source of truth.
The REPL console still renders from `executions.data`. For the "clear console"
feature, we can either keep a local "hidden IDs" set or just re-render from
query data.

### 3.9 `components/CreateSessionDialog.tsx`

**Before**: receives `templates` as prop, calls `vmService.createSession()` +
`vmService.setCurrentSession()`.

**After**: uses `useListTemplatesQuery()` directly (no prop drilling), uses
`useCreateSessionMutation()` + `dispatch(setCurrentSessionId(id))`.

### 3.10 `pages/System.tsx`

**Before**: reads `templates`/`sessions` from context.

**After**: reads from `useListTemplatesQuery()` / `useListSessionsQuery()`.

### 3.11 `pages/Reference.tsx`

No data layer changes — only uses `PRESET_EXAMPLES` (static).

### 3.12 Other components (no changes needed)

- `CodeEditor.tsx` — pure controlled component
- `ExecutionConsole.tsx` — receives `executions` as prop
- `ExecutionLogViewer.tsx` — receives `executions` as prop
- `PresetSelector.tsx` — uses `PRESET_EXAMPLES` (static)

## 4. RTK Query API definition sketch

```ts
// lib/api.ts
import { createApi, fetchBaseQuery } from '@reduxjs/toolkit/query/react';

const API_BASE_URL = ...;

export const vmApi = createApi({
  reducerPath: 'vmApi',
  baseQuery: fetchBaseQuery({
    baseUrl: API_BASE_URL,
    // custom error parsing for APIErrorEnvelope
  }),
  tagTypes: ['Template', 'Session', 'Execution', 'Event'],
  endpoints: (builder) => ({

    // ── Templates ──────────────────────────────────────
    listTemplates: builder.query<VMProfile[], void>({
      async queryFn(_arg, _api, _extraOpts, baseQuery) {
        // 1) GET /api/v1/templates
        // 2) For each, GET /api/v1/templates/:id (detail)
        // 3) Normalize with transformResponse
        // 4) If empty, bootstrap defaults
      },
      providesTags: (result) =>
        result
          ? [...result.map(t => ({ type: 'Template' as const, id: t.id })),
             { type: 'Template', id: 'LIST' }]
          : [{ type: 'Template', id: 'LIST' }],
    }),

    getTemplate: builder.query<VMProfile, string>({
      query: (id) => `/api/v1/templates/${id}`,
      transformResponse: (raw: RawTemplateDetail) => mapTemplateDetail(raw),
      providesTags: (_result, _err, id) => [{ type: 'Template', id }],
    }),

    createTemplate: builder.mutation<RawTemplate, { name: string; engine: string }>({
      query: (body) => ({ url: '/api/v1/templates', method: 'POST', body }),
      invalidatesTags: [{ type: 'Template', id: 'LIST' }],
    }),

    updateTemplateModules: builder.mutation<VMProfile, { templateId: string; modules: string[] }>({
      async queryFn({ templateId, modules }, _api, _extraOpts, baseQuery) {
        // diff current vs desired, POST/DELETE as needed
        // return updated template detail
      },
      invalidatesTags: (_result, _err, { templateId }) => [
        { type: 'Template', id: templateId },
        { type: 'Template', id: 'LIST' },
      ],
    }),

    updateTemplateLibraries: builder.mutation<VMProfile, { templateId: string; libraries: string[] }>({
      // same pattern as modules
      invalidatesTags: (_result, _err, { templateId }) => [
        { type: 'Template', id: templateId },
        { type: 'Template', id: 'LIST' },
      ],
    }),

    // ── Sessions ───────────────────────────────────────
    listSessions: builder.query<VMSession[], void>({
      query: () => '/api/v1/sessions',
      transformResponse: (raw: RawSession[]) => /* normalize */,
      providesTags: (result) =>
        result
          ? [...result.map(s => ({ type: 'Session' as const, id: s.id })),
             { type: 'Session', id: 'LIST' }]
          : [{ type: 'Session', id: 'LIST' }],
    }),

    getSession: builder.query<VMSession, string>({
      query: (id) => `/api/v1/sessions/${id}`,
      transformResponse: (raw: RawSession) => /* normalize */,
      providesTags: (_result, _err, id) => [{ type: 'Session', id }],
    }),

    createSession: builder.mutation<VMSession, CreateSessionArgs>({
      query: (body) => ({ url: '/api/v1/sessions', method: 'POST', body }),
      invalidatesTags: [{ type: 'Session', id: 'LIST' }],
    }),

    closeSession: builder.mutation<VMSession, string>({
      query: (id) => ({ url: `/api/v1/sessions/${id}/close`, method: 'POST', body: {} }),
      invalidatesTags: (_result, _err, id) => [
        { type: 'Session', id },
        { type: 'Session', id: 'LIST' },
      ],
    }),

    deleteSession: builder.mutation<void, string>({
      query: (id) => ({ url: `/api/v1/sessions/${id}`, method: 'DELETE' }),
      invalidatesTags: (_result, _err, id) => [
        { type: 'Session', id },
        { type: 'Session', id: 'LIST' },
      ],
    }),

    // ── Executions ─────────────────────────────────────
    listExecutions: builder.query<Execution[], string>({
      query: (sessionId) => `/api/v1/executions?session_id=${sessionId}&limit=50`,
      transformResponse: /* normalize + hydrate events */,
      providesTags: (_result, _err, sessionId) => [
        { type: 'Execution', id: `LIST-${sessionId}` },
      ],
    }),

    executeREPL: builder.mutation<Execution, { sessionId: string; input: string }>({
      query: ({ sessionId, input }) => ({
        url: '/api/v1/executions/repl',
        method: 'POST',
        body: { session_id: sessionId, input },
      }),
      invalidatesTags: (_result, _err, { sessionId }) => [
        { type: 'Execution', id: `LIST-${sessionId}` },
      ],
    }),

    getExecutionEvents: builder.query<ExecutionEvent[], string>({
      query: (execId) => `/api/v1/executions/${execId}/events?after_seq=0`,
      transformResponse: (raw: RawExecutionEvent[]) => /* normalize */,
      providesTags: (_result, _err, execId) => [{ type: 'Event', id: execId }],
    }),
  }),
});
```

## 5. Handling `session.vm` (template snapshot on session)

Currently, `mapSession()` fetches the full template detail for every session
to attach `session.vm`. With RTK Query, sessions and templates are cached
separately. Components that need both just call both queries:

```tsx
const { data: session } = useGetSessionQuery(id);
const { data: template } = useGetTemplateQuery(session?.vmId, { skip: !session });
```

This is cleaner — no N+1 inside the session list query. The session list
query returns raw session data; template data comes from the separately
cached template queries (which are likely already loaded from `listTemplates`).

## 6. Bootstrap strategy

The current `ensureInitialized()` method auto-creates default templates and a
first session if none exist. This becomes a `useBootstrap` hook:

```tsx
function useBootstrap() {
  const { data: templates, isSuccess: tplReady } = useListTemplatesQuery();
  const { data: sessions, isSuccess: sessReady } = useListSessionsQuery();
  const [createTemplate] = useCreateTemplateMutation();
  const [createSession] = useCreateSessionMutation();

  useEffect(() => {
    if (tplReady && templates.length === 0) {
      // create default templates
    }
  }, [tplReady, templates]);

  useEffect(() => {
    if (sessReady && tplReady && sessions.length === 0 && templates.length > 0) {
      // create default session
    }
  }, [sessReady, tplReady, sessions, templates]);
}
```

Called once in `AppShell`. Mutations invalidate tags, so the queries
automatically refetch after bootstrap creates data.

## 7. Implementation phases

### Phase 0: Install dependencies

```bash
pnpm add @reduxjs/toolkit react-redux
```

### Phase 1: Foundation (no UI changes)

- Create `lib/types.ts` — extract interfaces + static data from `vmService.ts`
- Create `lib/normalize.ts` — extract normalization functions
- Create `lib/api.ts` — define RTK Query API with all endpoints
- Create `lib/uiSlice.ts` — currentSessionId + sessionNames + localStorage persistence
- Create `lib/store.ts` — configure store
- Wrap `<App>` in `<Provider store={store}>`
- Write `useBootstrap` hook
- **Verify**: app still compiles, old code untouched, store available in devtools

### Phase 2: Migrate AppShell + Templates page

- Replace `AppStateContext` with Redux hooks in `AppShell.tsx`
  - Breadcrumbs: `useListTemplatesQuery()` + `useListSessionsQuery()`
  - Footer: `useSelector(selectCurrentSession)`
- Replace `Templates.tsx`: `useListTemplatesQuery()` + `useCreateTemplateMutation()`
- Remove `refreshTemplates` from context
- **Verify**: template list loads, new template appears immediately after creation

### Phase 3: Migrate TemplateDetail

- Replace `vmService.getVM()` / `refreshTemplates()` with `useGetTemplateQuery(id)`
- Replace `vmService.updateTemplateModules/Libraries` with mutations
- Derived sessions: `useListSessionsQuery()` + filter by `vmId`
- **Verify**: module/library toggles work, derived sessions list stays current

### Phase 4: Migrate Sessions + SessionDetail

- `Sessions.tsx`: `useListSessionsQuery()` + `useListTemplatesQuery()`
- `SessionDetail.tsx`: `useGetSessionQuery(id)` + `useListExecutionsQuery(sessionId)` + `useExecuteREPLMutation()` + `useCloseSessionMutation()`
- Wire `dispatch(setCurrentSessionId)` on session open
- **Verify**: REPL works, execution list updates after run, close session updates list

### Phase 5: Migrate CreateSessionDialog + System page

- `CreateSessionDialog`: `useListTemplatesQuery()` (remove `templates` prop) + `useCreateSessionMutation()`
- `System.tsx`: `useListTemplatesQuery()` + `useListSessionsQuery()`
- **Verify**: create session flow works end-to-end, system page shows counts

### Phase 6: Cleanup

- Delete `vmService.ts`
- Delete `AppStateContext` and `useAppState` export
- Remove `vmService` imports from all files
- Remove unused `refreshTemplates`/`refreshSessions` callbacks
- **Verify**: `tsc --noEmit` clean, `vite build` clean, manual smoke test

## 8. Risk notes

- **Execution event hydration**: The current code does N+1 queries (fetch
  events for each execution in a list). RTK Query can do the same, but it's
  worth considering lazy-loading events only when an execution is expanded,
  rather than eagerly hydrating all of them.

- **Optimistic updates for checkboxes**: Module/library toggles update
  immediately in the UI via local state, then fire the API call. With RTK
  Query, we either keep local state for instant feedback (with rollback on
  error) or use RTK Query's `onQueryStarted` for optimistic cache updates.
  Recommendation: keep the local `selectedModules`/`selectedLibraries` state
  pattern — it's simple and already works.

- **Session name aliases**: These are localStorage-only (not persisted to
  the backend). The `uiSlice` handles them. When rendering session names,
  a selector combines `uiSlice.sessionNames[id]` with the raw session data.
