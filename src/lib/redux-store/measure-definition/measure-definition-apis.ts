import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import {
  addManyMeasures,
  addOneMeasure,
  removeMeasure,
  setMeasureExpressionValidation,
  updateMeasure,
} from "$lib/redux-store/measure-definition/measure-definition-slice";
import { fetchWrapper } from "$lib/util/fetchWrapper";
import type { MeasureDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
import { generateApis } from "$lib/redux-store/utils/api-utils";
import type { ValidationConfig } from "$lib/redux-store/utils/validation-utils";
import { ValidationState } from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";
import { createAsyncThunk } from "$lib/redux-store/redux-toolkit-wrapper";
import { getMessageFromParseError } from "$common/expression-parser/getMessageFromParseError";
import { Debounce } from "$common/utils/Debounce";
import { handleErrorResponse } from "$lib/redux-store/utils/handleErrorResponse";
import { selectMeasureById } from "$lib/redux-store/measure-definition/measure-definition-selectors";
import { invalidateExplorerThunk } from "$lib/redux-store/utils/invalidateExplorerThunk";

const MeasureExpressionValidation: ValidationConfig<MeasureDefinitionEntity> = {
  field: "expression",
  validate: async (entity, changes) => {
    try {
      const resp = await fetchWrapper("measures/validate-expression", "POST", {
        metricsDefId: changes.metricsDefId ?? entity.metricsDefId,
        expression: changes.expression,
      });
      return {
        expressionIsValid: resp.expressionIsValid,
        expressionValidationError: resp.expressionValidationError
          ? getMessageFromParseError(
              changes.expression,
              resp.expressionValidationError
            )
          : "",
      };
    } catch (err) {
      handleErrorResponse(err.response);
      return Promise.resolve({});
    }
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
export const updateMeasuresWrapperApi = invalidateExplorerThunk(
  EntityType.MeasureDefinition,
  updateMeasuresApi,
  ["expression", "sqlName"],
  (state, id) => [selectMeasureById(state, id).metricsDefId]
);

const validationDebounce = new Debounce();
export const validateMeasureExpression = (
  dispatch,
  metricsDefId: string,
  measureId: string,
  expression: string
) => {
  validationDebounce.debounce(
    measureId,
    () => {
      dispatch(
        validateMeasureExpressionApi({ metricsDefId, measureId, expression })
      );
    },
    250
  );
};

const validateMeasureExpressionApi = createAsyncThunk(
  `${EntityType.MeasureDefinition}/validate-expression`,
  async (
    {
      metricsDefId,
      measureId,
      expression,
    }: { metricsDefId: string; measureId: string; expression: string },
    thunkAPI
  ) => {
    try {
      const resp = await fetchWrapper("measures/validate-expression", "POST", {
        metricsDefId,
        expression,
      });
      thunkAPI.dispatch(
        setMeasureExpressionValidation(
          measureId,
          resp.expressionIsValid,
          resp.expressionValidationError
            ? getMessageFromParseError(
                expression,
                resp.expressionValidationError
              )
            : ""
        )
      );
    } catch (err) {
      handleErrorResponse(err.response);
    }
  }
);
