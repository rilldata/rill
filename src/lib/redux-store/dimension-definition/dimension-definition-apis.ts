import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import {
  addManyDimensions,
  addOneDimension,
  removeDimension,
  setDimensionValidationStatus,
  updateDimension,
} from "$lib/redux-store/dimension-definition/dimension-definition-slice";
import type { DimensionDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/DimensionDefinitionStateService";
import { fetchWrapper } from "$lib/util/fetchWrapper";
import { generateApis } from "$lib/redux-store/utils/api-utils";
import type { ValidationConfig } from "$lib/redux-store/utils/validation-utils";
import { ValidationState } from "$common/data-modeler-state-service/entity-state-service/MetricsDefinitionEntityService";
import { handleErrorResponse } from "$lib/redux-store/utils/handleErrorResponse";
import { selectDimensionById } from "$lib/redux-store/dimension-definition/dimension-definition-selectors";
import { invalidateExplorerThunk } from "$lib/redux-store/utils/invalidateExplorerThunk";
import { createAsyncThunk } from "$lib/redux-store/redux-toolkit-wrapper";
import type { RillReduxState } from "$lib/redux-store/store-root";

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
