import type { VMProfile, RawTemplate, RawTemplateDetail } from '../../types';
import { DEFAULT_TEMPLATE_SPECS } from '../../types';
import { asArray, mapTemplateDetail } from '../../normalize';
import type { VmEndpointBuilder } from './shared';

export function buildTemplateEndpoints(builder: VmEndpointBuilder) {
  return {
    listTemplates: builder.query<VMProfile[], void>({
      async queryFn(_arg, _api, _extraOpts, baseQuery) {
        const listRes = await baseQuery('/api/v1/templates');
        if (listRes.error) return { error: listRes.error };

        const templates = asArray(listRes.data as RawTemplate[] | null);

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
            const libraries =
              ('libraries' in spec ? spec.libraries : undefined) || [];

            for (const moduleName of modules) {
              await baseQuery({
                url: `/api/v1/templates/${created.id}/modules`,
                method: 'POST',
                body: { name: moduleName },
              });
            }
            for (const libraryName of libraries) {
              await baseQuery({
                url: `/api/v1/templates/${created.id}/libraries`,
                method: 'POST',
                body: { name: libraryName },
              });
            }
            templates.push(created);
          }
        }

        const profiles: VMProfile[] = [];
        for (const template of templates) {
          const detailRes = await baseQuery(`/api/v1/templates/${template.id}`);
          if (detailRes.error) return { error: detailRes.error };
          profiles.push(mapTemplateDetail(detailRes.data as RawTemplateDetail));
        }

        return { data: profiles.sort((a, b) => a.name.localeCompare(b.name)) };
      },
      providesTags: (result) =>
        result
          ? [
              ...result.map((template) => ({
                type: 'Template' as const,
                id: template.id,
              })),
              { type: 'Template', id: 'LIST' },
            ]
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

    createTemplate: builder.mutation<RawTemplate, { name: string; engine: string }>(
      {
        query: (body) => ({ url: '/api/v1/templates', method: 'POST', body }),
        invalidatesTags: [{ type: 'Template', id: 'LIST' }],
      },
    ),

    updateTemplateModules: builder.mutation<
      VMProfile,
      { templateId: string; modules: string[] }
    >({
      async queryFn({ templateId, modules }, _api, _extraOpts, baseQuery) {
        const currentRes = await baseQuery(`/api/v1/templates/${templateId}/modules`);
        if (currentRes.error) return { error: currentRes.error };
        const currentModules = asArray(currentRes.data as string[] | null);

        const wanted = new Set(modules);
        const existing = new Set(currentModules);

        const toAdd = modules.filter((moduleName) => !existing.has(moduleName));
        const toRemove = currentModules.filter(
          (moduleName) => !wanted.has(moduleName),
        );

        for (const moduleName of toAdd) {
          const res = await baseQuery({
            url: `/api/v1/templates/${templateId}/modules`,
            method: 'POST',
            body: { name: moduleName },
          });
          if (res.error) return { error: res.error };
        }

        for (const moduleName of toRemove) {
          const res = await baseQuery({
            url: `/api/v1/templates/${templateId}/modules/${encodeURIComponent(moduleName)}`,
            method: 'DELETE',
          });
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

    updateTemplateLibraries: builder.mutation<
      VMProfile,
      { templateId: string; libraries: string[] }
    >({
      async queryFn({ templateId, libraries }, _api, _extraOpts, baseQuery) {
        const currentRes = await baseQuery(
          `/api/v1/templates/${templateId}/libraries`,
        );
        if (currentRes.error) return { error: currentRes.error };
        const currentLibraries = asArray(currentRes.data as string[] | null);

        const wanted = new Set(libraries);
        const existing = new Set(currentLibraries);

        const toAdd = libraries.filter(
          (libraryName) => !existing.has(libraryName),
        );
        const toRemove = currentLibraries.filter(
          (libraryName) => !wanted.has(libraryName),
        );

        for (const libraryName of toAdd) {
          const res = await baseQuery({
            url: `/api/v1/templates/${templateId}/libraries`,
            method: 'POST',
            body: { name: libraryName },
          });
          if (res.error) return { error: res.error };
        }

        for (const libraryName of toRemove) {
          const res = await baseQuery({
            url: `/api/v1/templates/${templateId}/libraries/${encodeURIComponent(libraryName)}`,
            method: 'DELETE',
          });
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
  };
}
