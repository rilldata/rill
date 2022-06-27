import { generateApis } from "$lib/redux-store/slice-utils";
import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import {
  addManyDimensions,
  addOneDimension,
  removeDimension,
  updateDimension,
} from "$lib/redux-store/dimension-definition/dimension-definition-slice";

export const {
  fetchManyApi: fetchManyDimensionsApi,
  createApi: createDimensionsApi,
  updateApi: updateDimensionsApi,
  deleteApi: deleteDimensionsApi,
} = generateApis<
  EntityType.DimensionDefinition,
  { metricsDefId: string },
  { metricsDefId: string }
>(
  EntityType.DimensionDefinition,
  addManyDimensions,
  addOneDimension,
  updateDimension,
  removeDimension,
  "dimensions"
);
