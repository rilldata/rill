import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import {
  addManyMeasures,
  addOneMeasure,
  removeMeasure,
  updateMeasure,
} from "$lib/redux-store/measure-definition/measure-definition-slice";
import { fetchWrapper } from "$lib/util/fetchWrapper";
import type { MeasureDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
import { generateApis } from "$lib/redux-store/utils/api-utils";
import type { ValidationConfig } from "$lib/redux-store/utils/validation-utils";
import { ValidationState } from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";

const MeasureExpressionValidation: ValidationConfig<MeasureDefinitionEntity> = {
  field: "expression",
  validate: (entity, changes) => {
    return fetchWrapper("measures/validate-expression", "POST", {
      metricsDefId: changes.metricsDefId ?? entity.metricsDefId,
      expression: changes.expression,
    });
  },
  validationPassed: (changes) =>
    changes.expressionIsValid === ValidationState.OK,
};

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
  [EntityType.MeasureDefinition, "measureDefinition", "measures"],
  [addManyMeasures, addOneMeasure, updateMeasure, removeMeasure],
  [MeasureExpressionValidation]
);
