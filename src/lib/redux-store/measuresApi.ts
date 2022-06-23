import {
  api,
  defaultInvalidatesFunction,
  defaultArrayProvidesFunction,
  defaultTransformFunction,
} from "$lib/redux-store/api";
import type { MeasureDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";

interface UpdateMeasureRequest {
  id: string;
  measure: Partial<MeasureDefinitionEntity>;
}

const measureProvidesFunction = defaultArrayProvidesFunction(
  EntityType.MeasureDefinition
);
const dimensionArrayProvidesFunction = defaultArrayProvidesFunction(
  EntityType.MeasureDefinition
);
const measureInvalidatesFunction = defaultInvalidatesFunction(
  EntityType.MeasureDefinition
);

export const measuresApi = api.injectEndpoints({
  endpoints: (build) => ({
    getAllMeasures: build.query<MeasureDefinitionEntity[], string>({
      query: (metricsDefId: string) => ({
        url: `measures/?metricsDefId=${metricsDefId}`,
      }),
      transformResponse: defaultTransformFunction,
      providesTags: dimensionArrayProvidesFunction,
    }),
    getOneMeasure: build.query<MeasureDefinitionEntity, string>({
      query: (id: string) => ({
        url: `measures/${id}`,
      }),
      transformResponse: defaultTransformFunction,
      providesTags: measureProvidesFunction,
    }),
    createMeasure: build.mutation<MeasureDefinitionEntity, string>({
      query: (metricsDefId: string) => ({
        url: "measures",
        method: "PUT",
        body: { metricsDefId },
      }),
      transformResponse: defaultTransformFunction,
      invalidatesTags: [EntityType.MeasureDefinition],
    }),
    updateMeasure: build.mutation<
      MeasureDefinitionEntity,
      UpdateMeasureRequest
    >({
      query: ({ id, measure }: UpdateMeasureRequest) => ({
        url: `measures/${id}`,
        method: "POST",
        body: measure,
      }),
      transformResponse: defaultTransformFunction,
      invalidatesTags: measureInvalidatesFunction,
    }),
    deleteMeasure: build.mutation<MeasureDefinitionEntity, string>({
      query: (id: string) => ({
        url: `measures/${id}`,
        method: "DELETE",
      }),
      invalidatesTags: measureInvalidatesFunction,
    }),
  }),
  overrideExisting: false,
});
