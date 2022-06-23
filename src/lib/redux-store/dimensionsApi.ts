import {
  api,
  defaultArrayProvidesFunction,
  defaultInvalidatesFunction,
  defaultProvidesFunction,
  defaultTransformFunction,
} from "$lib/redux-store/api";
import type { DimensionDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/DimensionDefinitionStateService";
import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";

interface UpdateDimensionRequest {
  id: string;
  dimension: Partial<DimensionDefinitionEntity>;
}

const dimensionProvidesFunction = defaultTransformFunction(
  EntityType.DimensionDefinition
);
const dimensionArrayProvidesFunction = defaultArrayProvidesFunction(
  EntityType.DimensionDefinition
);
const dimensionInvalidatesFunction = defaultInvalidatesFunction(
  EntityType.DimensionDefinition
);

export const dimensionsApi = api.injectEndpoints({
  endpoints: (build) => ({
    getAllDimensions: build.query<DimensionDefinitionEntity[], string>({
      query: (metricsDefId: string) => ({
        url: `dimensions/?metricsDefId=${metricsDefId}`,
      }),
      transformResponse: defaultTransformFunction,
      providesTags: dimensionArrayProvidesFunction,
    }),
    getOneDimension: build.query<DimensionDefinitionEntity, string>({
      query: (id: string) => ({
        url: `dimensions/${id}`,
      }),
      transformResponse: defaultTransformFunction,
      providesTags: dimensionProvidesFunction,
    }),
    createDimension: build.mutation<DimensionDefinitionEntity, string>({
      query: (metricsDefId: string) => ({
        url: "dimensions",
        method: "PUT",
        body: { metricsDefId },
      }),
      transformResponse: defaultTransformFunction,
      invalidatesTags: [EntityType.DimensionDefinition],
    }),
    updateDimension: build.mutation<
      DimensionDefinitionEntity,
      UpdateDimensionRequest
    >({
      query: ({ id, dimension }: UpdateDimensionRequest) => ({
        url: `dimensions/${id}`,
        method: "POST",
        body: dimension,
      }),
      transformResponse: defaultTransformFunction,
      invalidatesTags: dimensionInvalidatesFunction,
    }),
    deleteDimension: build.mutation<DimensionDefinitionEntity, string>({
      query: (id: string) => ({
        url: `dimensions/${id}`,
        method: "DELETE",
      }),
      invalidatesTags: dimensionInvalidatesFunction,
    }),
  }),
  overrideExisting: false,
});
