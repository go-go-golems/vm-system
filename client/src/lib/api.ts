import { createApi, type BaseQueryFn } from '@reduxjs/toolkit/query/react';
import {
  type VMProfile,
  type VMSession,
  type Execution,
  type ExecutionEvent,
  type RawTemplate,
  type RawTemplateDetail,
  type RawSession,
  type RawExecution,
  type RawExecutionEvent,
  type CreateSessionArgs,
  API_BASE_URL,
  DEFAULT_TEMPLATE_SPECS,
} from './types';
import {
  mapTemplateDetail,
  mapSession,
  mapExecution,
  mapEvents,
  asArray,
} from './normalize';
// NOTE: We can't import RootState from store.ts due to circular dependency.
// Instead we define the shape we need inline.
interface UiSliceState {
  sessionNames: Record<string, string>;
  currentSessionId: string | null;
}
interface ApiRootState {
  ui: UiSliceState;
}

// ---------------------------------------------------------------------------
// Custom baseQuery (mirrors the old request() helper)
// ---------------------------------------------------------------------------

const ABSOLUTE_HTTP_URL_RE = /^https?:\/\//i;

function getURL(path: string, query?: Record<string, string | number | undefined>): string {
  const hasAbsoluteBase = ABSOLUTE_HTTP_URL_RE.test(API_BASE_URL);
  const baseURL = hasAbsoluteBase
    ? new URL(API_BASE_URL.endsWith('/') ? API_BASE_URL : `${API_BASE_URL}/`)
    : new URL(window.location.origin);
  const url = hasAbsoluteBase
    ? new URL(path, baseURL)
    : new URL(`${API_BASE_URL}${path}`, baseURL);
  if (query) {
    Object.entries(query).forEach(([key, value]) => {
      if (value === undefined || value === '') return;
      url.searchParams.set(key, String(value));
    });
  }
  return hasAbsoluteBase ? url.toString() : `${url.pathname}${url.search}`;
}

interface ApiRequestArgs {
  url: string;
  method?: string;
  body?: unknown;
  query?: Record<string, string | number | undefined>;
}

const vmBaseQuery: BaseQueryFn<ApiRequestArgs | string, unknown, { status: number; message: string }> = async (
  args,
) => {
  const { url, method = 'GET', body, query } =
    typeof args === 'string' ? { url: args } : args;

  try {
    const response = await fetch(getURL(url, query), {
      method,
      headers: { 'Content-Type': 'application/json' },
      body: body !== undefined ? JSON.stringify(body) : undefined,
    });

    if (!response.ok) {
      let message = `HTTP ${response.status}`;
      try {
        const envelope = await response.json();
        if (envelope?.error?.message) message = envelope.error.message;
      } catch { /* ignore */ }
      return { error: { status: response.status, message } };
    }

    if (response.status === 204) return { data: undefined };
    const data = await response.json();
    return { data };
  } catch (err: any) {
    return { error: { status: 0, message: err.message || 'Network error' } };
  }
};

// ---------------------------------------------------------------------------
// RTK Query API
// ---------------------------------------------------------------------------

export const vmApi = createApi({
  reducerPath: 'vmApi',
  baseQuery: vmBaseQuery,
  tagTypes: ['Template', 'Session', 'Execution'],
  endpoints: (builder) => ({

    // ── Templates ──────────────────────────────────────────────────────

    listTemplates: builder.query<VMProfile[], void>({
      async queryFn(_arg, _api, _extraOpts, baseQuery) {
        // 1) GET list
        const listRes = await baseQuery('/api/v1/templates');
        if (listRes.error) return { error: listRes.error };

        let templates = asArray(listRes.data as RawTemplate[] | null);

        // 2) Auto-bootstrap if empty
        if (templates.length === 0) {
          for (const spec of DEFAULT_TEMPLATE_SPECS) {
            const createRes = await baseQuery({
              url: '/api/v1/templates',
              method: 'POST',
              body: { name: spec.name, engine: spec.engine },
            });
            if (createRes.error) return { error: createRes.error };
            const created = createRes.data as RawTemplate;

            const modules = spec.modules || [];
            const libraries = ('libraries' in spec ? spec.libraries : undefined) || [];
            for (const m of modules) {
              await baseQuery({ url: `/api/v1/templates/${created.id}/modules`, method: 'POST', body: { name: m } });
            }
            for (const l of libraries) {
              await baseQuery({ url: `/api/v1/templates/${created.id}/libraries`, method: 'POST', body: { name: l } });
            }
            templates.push(created);
          }
        }

        // 3) Fetch full detail for each
        const profiles: VMProfile[] = [];
        for (const tpl of templates) {
          const detailRes = await baseQuery(`/api/v1/templates/${tpl.id}`);
          if (detailRes.error) return { error: detailRes.error };
          profiles.push(mapTemplateDetail(detailRes.data as RawTemplateDetail));
        }

        return { data: profiles.sort((a, b) => a.name.localeCompare(b.name)) };
      },
      providesTags: (result) =>
        result
          ? [...result.map(t => ({ type: 'Template' as const, id: t.id })), { type: 'Template', id: 'LIST' }]
          : [{ type: 'Template', id: 'LIST' }],
    }),

    getTemplate: builder.query<VMProfile, string>({
      async queryFn(id, _api, _extraOpts, baseQuery) {
        const res = await baseQuery(`/api/v1/templates/${id}`);
        if (res.error) return { error: res.error };
        return { data: mapTemplateDetail(res.data as RawTemplateDetail) };
      },
      providesTags: (_result, _err, id) => [{ type: 'Template', id }],
    }),

    createTemplate: builder.mutation<RawTemplate, { name: string; engine: string }>({
      query: (body) => ({ url: '/api/v1/templates', method: 'POST', body }),
      invalidatesTags: [{ type: 'Template', id: 'LIST' }],
    }),

    updateTemplateModules: builder.mutation<VMProfile, { templateId: string; modules: string[] }>({
      async queryFn({ templateId, modules }, _api, _extraOpts, baseQuery) {
        // Get current modules
        const currentRes = await baseQuery(`/api/v1/templates/${templateId}/modules`);
        if (currentRes.error) return { error: currentRes.error };
        const currentModules = asArray(currentRes.data as string[] | null);

        const wanted = new Set(modules);
        const existing = new Set(currentModules);

        const toAdd = modules.filter(m => !existing.has(m));
        const toRemove = currentModules.filter(m => !wanted.has(m));

        for (const m of toAdd) {
          const res = await baseQuery({ url: `/api/v1/templates/${templateId}/modules`, method: 'POST', body: { name: m } });
          if (res.error) return { error: res.error };
        }
        for (const m of toRemove) {
          const res = await baseQuery({ url: `/api/v1/templates/${templateId}/modules/${encodeURIComponent(m)}`, method: 'DELETE' });
          if (res.error) return { error: res.error };
        }

        // Fetch updated detail
        const detailRes = await baseQuery(`/api/v1/templates/${templateId}`);
        if (detailRes.error) return { error: detailRes.error };
        return { data: mapTemplateDetail(detailRes.data as RawTemplateDetail) };
      },
      invalidatesTags: (_result, _err, { templateId }) => [
        { type: 'Template', id: templateId },
        { type: 'Template', id: 'LIST' },
      ],
    }),

    updateTemplateLibraries: builder.mutation<VMProfile, { templateId: string; libraries: string[] }>({
      async queryFn({ templateId, libraries }, _api, _extraOpts, baseQuery) {
        const currentRes = await baseQuery(`/api/v1/templates/${templateId}/libraries`);
        if (currentRes.error) return { error: currentRes.error };
        const currentLibraries = asArray(currentRes.data as string[] | null);

        const wanted = new Set(libraries);
        const existing = new Set(currentLibraries);

        const toAdd = libraries.filter(l => !existing.has(l));
        const toRemove = currentLibraries.filter(l => !wanted.has(l));

        for (const l of toAdd) {
          const res = await baseQuery({ url: `/api/v1/templates/${templateId}/libraries`, method: 'POST', body: { name: l } });
          if (res.error) return { error: res.error };
        }
        for (const l of toRemove) {
          const res = await baseQuery({ url: `/api/v1/templates/${templateId}/libraries/${encodeURIComponent(l)}`, method: 'DELETE' });
          if (res.error) return { error: res.error };
        }

        const detailRes = await baseQuery(`/api/v1/templates/${templateId}`);
        if (detailRes.error) return { error: detailRes.error };
        return { data: mapTemplateDetail(detailRes.data as RawTemplateDetail) };
      },
      invalidatesTags: (_result, _err, { templateId }) => [
        { type: 'Template', id: templateId },
        { type: 'Template', id: 'LIST' },
      ],
    }),

    // ── Sessions ───────────────────────────────────────────────────────

    listSessions: builder.query<VMSession[], void>({
      async queryFn(_arg, api, _extraOpts, baseQuery) {
        const res = await baseQuery('/api/v1/sessions');
        if (res.error) return { error: res.error };
        const raw = asArray(res.data as RawSession[] | null);
        const state = api.getState() as ApiRootState;
        const sessionNames = state.ui.sessionNames;
        const sessions = raw.map(r => mapSession(r, sessionNames));
        return { data: sessions.sort((a, b) => b.createdAt.localeCompare(a.createdAt)) };
      },
      providesTags: (result) =>
        result
          ? [...result.map(s => ({ type: 'Session' as const, id: s.id })), { type: 'Session', id: 'LIST' }]
          : [{ type: 'Session', id: 'LIST' }],
    }),

    getSession: builder.query<VMSession, string>({
      async queryFn(id, api, _extraOpts, baseQuery) {
        const res = await baseQuery(`/api/v1/sessions/${id}`);
        if (res.error) return { error: res.error };
        const state = api.getState() as ApiRootState;
        return { data: mapSession(res.data as RawSession, state.ui.sessionNames) };
      },
      providesTags: (_result, _err, id) => [{ type: 'Session', id }],
    }),

    createSession: builder.mutation<VMSession, CreateSessionArgs & { name?: string }>({
      async queryFn(args, api, _extraOpts, baseQuery) {
        const { name, ...body } = args;
        const res = await baseQuery({ url: '/api/v1/sessions', method: 'POST', body });
        if (res.error) return { error: res.error };
        const raw = res.data as RawSession;
        const state = api.getState() as ApiRootState;
        const sessionNames = { ...state.ui.sessionNames };
        if (name?.trim()) sessionNames[raw.id] = name.trim();
        return { data: mapSession(raw, sessionNames) };
      },
      invalidatesTags: [{ type: 'Session', id: 'LIST' }],
    }),

    closeSession: builder.mutation<void, string>({
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

    // ── Executions ─────────────────────────────────────────────────────

    listExecutions: builder.query<Execution[], string>({
      async queryFn(sessionId, _api, _extraOpts, baseQuery) {
        const res = await baseQuery({ url: '/api/v1/executions', query: { session_id: sessionId, limit: 50 } });
        if (res.error) return { error: res.error };
        const rawExecs = asArray(res.data as RawExecution[] | null);

        // Hydrate events for each execution
        const executions: Execution[] = [];
        for (const raw of rawExecs) {
          const eventsRes = await baseQuery({ url: `/api/v1/executions/${raw.id}/events`, query: { after_seq: 0 } });
          const events = eventsRes.error ? [] : mapEvents(asArray(eventsRes.data as RawExecutionEvent[] | null));
          executions.push(mapExecution(raw, events));
        }

        return { data: executions.sort((a, b) => a.startedAt.localeCompare(b.startedAt)) };
      },
      providesTags: (_result, _err, sessionId) => [
        { type: 'Execution', id: `LIST-${sessionId}` },
      ],
    }),

    executeREPL: builder.mutation<Execution, { sessionId: string; input: string }>({
      async queryFn({ sessionId, input }, _api, _extraOpts, baseQuery) {
        const res = await baseQuery({
          url: '/api/v1/executions/repl',
          method: 'POST',
          body: { session_id: sessionId, input },
        });
        if (res.error) return { error: res.error };
        const raw = res.data as RawExecution;

        // Fetch events for the new execution
        const eventsRes = await baseQuery({ url: `/api/v1/executions/${raw.id}/events`, query: { after_seq: 0 } });
        const events = eventsRes.error ? [] : mapEvents(asArray(eventsRes.data as RawExecutionEvent[] | null));

        return { data: mapExecution(raw, events) };
      },
      invalidatesTags: (_result, _err, { sessionId }) => [
        { type: 'Execution', id: `LIST-${sessionId}` },
      ],
    }),
  }),
});

export const {
  useListTemplatesQuery,
  useGetTemplateQuery,
  useCreateTemplateMutation,
  useUpdateTemplateModulesMutation,
  useUpdateTemplateLibrariesMutation,
  useListSessionsQuery,
  useGetSessionQuery,
  useCreateSessionMutation,
  useCloseSessionMutation,
  useDeleteSessionMutation,
  useListExecutionsQuery,
  useExecuteREPLMutation,
} = vmApi;
