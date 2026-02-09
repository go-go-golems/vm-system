import { createApi } from '@reduxjs/toolkit/query/react';
import { vmBaseQuery } from './vm/transport';
import { VM_TAG_TYPES } from './vm/endpoints/shared';
import { buildTemplateEndpoints } from './vm/endpoints/templates';
import { buildSessionEndpoints } from './vm/endpoints/sessions';
import { buildExecutionEndpoints } from './vm/endpoints/executions';

export const vmApi = createApi({
  reducerPath: 'vmApi',
  baseQuery: vmBaseQuery,
  tagTypes: VM_TAG_TYPES,
  endpoints: (builder) => ({
    ...buildTemplateEndpoints(builder),
    ...buildSessionEndpoints(builder),
    ...buildExecutionEndpoints(builder),
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
