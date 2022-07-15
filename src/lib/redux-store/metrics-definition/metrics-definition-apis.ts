import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import {
  addManyMetricsDefs,
  addOneMetricsDef,
  removeMetricsDef,
  updateMetricsDef,
} from "$lib/redux-store/metrics-definition/metrics-definition-slice";
import { generateApis } from "$lib/redux-store/utils/api-utils";
import { createAsyncThunk } from "$lib/redux-store/redux-toolkit-wrapper";
import { streamingFetchWrapper } from "$lib/util/fetchWrapper";
import type { MeasureDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
import type { DimensionDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/DimensionDefinitionStateService";
import {
  addOneMeasure,
  clearMeasuresForMetricsDefId,
} from "$lib/redux-store/measure-definition/measure-definition-slice";
import {
  addOneDimension,
  clearDimensionsForMetricsDefId,
} from "$lib/redux-store/dimension-definition/dimension-definition-slice";
import { asyncWait } from "$common/utils/waitUtils";
import { dataModelerService } from "$lib/application-state-stores/application-store";
import type { MetricsDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";

export const {
  fetchManyApi: fetchManyMetricsDefsApi,
  createApi: createMetricsDefsApi,
  updateApi: updateMetricsDefsApi,
  deleteApi: deleteMetricsDefsApi,
} = generateApis<
  EntityType.MetricsDefinition,
  Partial<MetricsDefinitionEntity>
>(
  [EntityType.MetricsDefinition, "metricsDefinition", "metrics"],
  [addManyMetricsDefs, addOneMetricsDef, updateMetricsDef, removeMetricsDef],
  [],
  {
    createHook: async (createdMetricsDef: MetricsDefinitionEntity) => {
      await dataModelerService.dispatch("setActiveAsset", [
        EntityType.MetricsDefinition,
        createdMetricsDef.id,
      ]);
    },
  }
);

export const generateMeasuresAndDimensionsApi = createAsyncThunk(
  `${EntityType.MetricsDefinition}/generateMeasuresAndDimensions`,
  async (id: string, thunkAPI) => {
    const stream = streamingFetchWrapper<
      MeasureDefinitionEntity | DimensionDefinitionEntity
    >(`metrics/${id}/generate-measures-dimensions`, "POST");
    thunkAPI.dispatch(clearMeasuresForMetricsDefId(id));
    thunkAPI.dispatch(clearDimensionsForMetricsDefId(id));
    await asyncWait(10);
    for await (const measureOrDimension of stream) {
      if (measureOrDimension.type === EntityType.MeasureDefinition) {
        thunkAPI.dispatch(
          addOneMeasure(measureOrDimension as MeasureDefinitionEntity)
        );
      } else {
        thunkAPI.dispatch(
          addOneDimension(measureOrDimension as DimensionDefinitionEntity)
        );
      }
    }
  }
);
