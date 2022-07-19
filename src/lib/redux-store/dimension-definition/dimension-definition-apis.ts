import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import {
  addManyDimensions,
  addOneDimension,
  removeDimension,
  updateDimension,
} from "$lib/redux-store/dimension-definition/dimension-definition-slice";
import type { DimensionDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/DimensionDefinitionStateService";
import { fetchWrapper } from "$lib/util/fetchWrapper";
import { generateApis } from "$lib/redux-store/utils/api-utils";
import type { ValidationConfig } from "$lib/redux-store/utils/validation-utils";
import { ValidationState } from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";

const DimensionColumnValidation: ValidationConfig<DimensionDefinitionEntity> = {
  field: "dimensionColumn",
  validate: (entity, changes) => {
    return fetchWrapper("dimensions/validate-dimension-column", "POST", {
      metricsDefId: changes.metricsDefId ?? entity.metricsDefId,
      dimensionColumn: changes.dimensionColumn,
    });
  },
  validationPassed: (changes) =>
    changes.dimensionIsValid === ValidationState.OK,
};

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
  [EntityType.DimensionDefinition, "dimensionDefinition", "dimensions"],
  [addManyDimensions, addOneDimension, updateDimension, removeDimension],
  [DimensionColumnValidation]
);
