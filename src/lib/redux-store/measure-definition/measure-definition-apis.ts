import { generateApis } from "$lib/redux-store/slice-utils";
import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import {
  addManyMeasures,
  addOneMeasure,
  removeMeasure,
  updateMeasure,
} from "$lib/redux-store/measure-definition/measure-definition-slice";

export const {
  fetchManyApi: fetchManyMeasuresApi,
  createApi: createMeasuresApi,
  updateApi: updateMeasuresApi,
  deleteApi: deleteMeasuresApi,
} = generateApis<
  EntityType.MeasureDefinition,
  { metricsDefId: string },
  { metricsDefId: string }
>(
  EntityType.MeasureDefinition,
  addManyMeasures,
  addOneMeasure,
  updateMeasure,
  removeMeasure,
  "measures"
);
