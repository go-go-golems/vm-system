import type {
  Execution,
  RawExecution,
  RawExecutionEvent,
} from '../../types';
import { asArray, mapEvents, mapExecution } from '../../normalize';
import type { VmEndpointBuilder } from './shared';

export function buildExecutionEndpoints(builder: VmEndpointBuilder) {
  return {
    listExecutions: builder.query<Execution[], string>({
      async queryFn(sessionId, _api, _extraOpts, baseQuery) {
        const res = await baseQuery({
          url: '/api/v1/executions',
          query: { session_id: sessionId, limit: 50 },
        });
        if (res.error) return { error: res.error };

        const rawExecutions = asArray(res.data as RawExecution[] | null);
        const executions: Execution[] = [];

        for (const execution of rawExecutions) {
          const eventsRes = await baseQuery({
            url: `/api/v1/executions/${execution.id}/events`,
            query: { after_seq: 0 },
          });
          const events = eventsRes.error
            ? []
            : mapEvents(asArray(eventsRes.data as RawExecutionEvent[] | null));
          executions.push(mapExecution(execution, events));
        }

        return {
          data: executions.sort((a, b) => a.startedAt.localeCompare(b.startedAt)),
        };
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
        const eventsRes = await baseQuery({
          url: `/api/v1/executions/${raw.id}/events`,
          query: { after_seq: 0 },
        });
        const events = eventsRes.error
          ? []
          : mapEvents(asArray(eventsRes.data as RawExecutionEvent[] | null));

        return { data: mapExecution(raw, events) };
      },
      invalidatesTags: (_result, _err, { sessionId }) => [
        { type: 'Execution', id: `LIST-${sessionId}` },
      ],
    }),
  };
}
