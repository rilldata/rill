import { EntityType } from "$web-local/common/data-modeler-state-service/entity-state-service/EntityStateService";
import {
  addManyDimensions,
  addOneDimension,
  removeDimension,
  setDimensionValidationStatus,
  updateDimension,
} from "./dimension-definition-slice";
import type { DimensionDefinitionEntity } from "$web-local/common/data-modeler-state-service/entity-state-service/DimensionDefinitionStateService";
import { fetchWrapper } from "../../util/fetchWrapper";
import { generateApis } from "../utils/api-utils";
import type { ValidationConfig } from "../utils/validation-utils";
import { ValidationState } from "$web-local/common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";
import { handleErrorResponse } from "../utils/handleErrorResponse";
import { selectDimensionById } from "./dimension-definition-selectors";
import { invalidateExplorerThunk } from "../utils/invalidateExplorerThunk";
import { createAsyncThunk } from "../redux-toolkit-wrapper";
import type { RillReduxState } from "../store-root";

const DimensionColumnValidation: ValidationConfig<DimensionDefinitionEntity> = {
  field: "dimensionColumn",
  validate: async (entity, changes) => {
    try {
      return await fetchWrapper(
        "dimensions/validate-dimension-column",
        "POST",
        {
          metricsDefId: changes.metricsDefId ?? entity.metricsDefId,
          dimensionColumn: changes.dimensionColumn,
        }
      );
    } catch (err) {
      handleErrorResponse(err.response);
      return Promise.resolve({});
    }
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
export const updateDimensionsWrapperApi = invalidateExplorerThunk(
  EntityType.DimensionDefinition,
  updateDimensionsApi,
  ["dimensionColumn"],
  (state, id) => [selectDimensionById(state, id).metricsDefId]
);

export const validateDimensionColumnApi = createAsyncThunk(
  `${EntityType.DimensionDefinition}/validateDimensionColumnApi`,
  async (dimensionId: string, thunkAPI) => {
    const state = thunkAPI.getState() as RillReduxState;
    const dimension = selectDimensionById(state, dimensionId);

    const { dimensionIsValid } = await fetchWrapper(
      "dimensions/validate-dimension-column",
      "POST",
      {
        metricsDefId: dimension.metricsDefId,
        dimensionColumn: dimension.dimensionColumn,
      }
    );
    thunkAPI.dispatch(
      setDimensionValidationStatus(dimensionId, dimensionIsValid)
    );
  }
);
