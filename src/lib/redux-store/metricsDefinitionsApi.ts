import {
  api,
  defaultInvalidatesFunction,
  defaultArrayProvidesFunction,
  defaultTransformFunction,
} from "$lib/redux-store/api";
import type { MetricsDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";
import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";

interface UpdateMetricsDefinitionRequest {
  id: string;
  metricsDef: Partial<MetricsDefinitionEntity>;
}

const metricsDefProvidesFunction = defaultTransformFunction(
  EntityType.MetricsDefinition
);
const dimensionArrayProvidesFunction = defaultArrayProvidesFunction(
  EntityType.MetricsDefinition
);
const metricsDefInvalidatesFunction = defaultInvalidatesFunction(
  EntityType.MetricsDefinition
);

export const metricsDefinitionsApi = api.injectEndpoints({
  endpoints: (build) => ({
    getAllMetricsDefinitions: build.query<MetricsDefinitionEntity[], void>({
      query: () => ({
        url: "metrics",
      }),
      transformResponse: defaultTransformFunction,
      providesTags: dimensionArrayProvidesFunction,
    }),
    getOneMetricsDefinition: build.query<MetricsDefinitionEntity, string>({
      query: (id: string) => ({
        url: `metrics/${id}`,
      }),
      transformResponse: defaultTransformFunction,
      providesTags: metricsDefProvidesFunction,
    }),
    createMetricsDefinition: build.mutation<MetricsDefinitionEntity, void>({
      query: () => ({
        url: "metrics",
        method: "PUT",
      }),
      transformResponse: defaultTransformFunction,
      invalidatesTags: [EntityType.MetricsDefinition],
    }),
    updateMetricsDefinition: build.mutation<
      MetricsDefinitionEntity,
      UpdateMetricsDefinitionRequest
    >({
      query: ({ id, metricsDef }: UpdateMetricsDefinitionRequest) => ({
        url: `metrics/${id}`,
        method: "POST",
        body: metricsDef,
      }),
      transformResponse: defaultTransformFunction,
      invalidatesTags: metricsDefInvalidatesFunction,
    }),
    deleteMetricsDefinition: build.mutation<MetricsDefinitionEntity, string>({
      query: (id: string) => ({
        url: `metrics/${id}`,
        method: "DELETE",
      }),
      invalidatesTags: metricsDefInvalidatesFunction,
    }),
  }),
  overrideExisting: false,
});
