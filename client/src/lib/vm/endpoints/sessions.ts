import type {
  CreateSessionArgs,
  RawSession,
  VMSession,
} from '../../types';
import { asArray, mapSession } from '../../normalize';
import type { ApiRootState, VmEndpointBuilder } from './shared';

export function buildSessionEndpoints(builder: VmEndpointBuilder) {
  return {
    listSessions: builder.query<VMSession[], void>({
      async queryFn(_arg, api, _extraOpts, baseQuery) {
        const res = await baseQuery('/api/v1/sessions');
        if (res.error) return { error: res.error };

        const rawSessions = asArray(res.data as RawSession[] | null);
        const state = api.getState() as ApiRootState;
        const sessionNames = state.ui.sessionNames;
        const sessions = rawSessions.map((session) =>
          mapSession(session, sessionNames),
        );

        return {
          data: sessions.sort((a, b) => b.createdAt.localeCompare(a.createdAt)),
        };
      },
      providesTags: (result) =>
        result
          ? [
              ...result.map((session) => ({
                type: 'Session' as const,
                id: session.id,
              })),
              { type: 'Session', id: 'LIST' },
            ]
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
        const res = await baseQuery({
          url: '/api/v1/sessions',
          method: 'POST',
          body,
        });
        if (res.error) return { error: res.error };

        const raw = res.data as RawSession;
        const state = api.getState() as ApiRootState;
        const sessionNames = { ...state.ui.sessionNames };
        if (name?.trim()) {
          sessionNames[raw.id] = name.trim();
        }

        return { data: mapSession(raw, sessionNames) };
      },
      invalidatesTags: [{ type: 'Session', id: 'LIST' }],
    }),

    closeSession: builder.mutation<void, string>({
      query: (id) => ({
        url: `/api/v1/sessions/${id}/close`,
        method: 'POST',
        body: {},
      }),
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
  };
}
