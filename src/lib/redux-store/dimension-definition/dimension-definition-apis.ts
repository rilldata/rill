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
import { handleErrorResponse } from "$lib/redux-store/utils/handleErrorResponse";
import { createAsyncThunk } from "$lib/redux-store/redux-toolkit-wrapper";
import { setExplorerIsStale } from "$lib/redux-store/explore/explore-slice";
import type { RillReduxState } from "$lib/redux-store/store-root";
import { selectDimensionById } from "$lib/redux-store/dimension-definition/dimension-definition-selectors";

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
export const updateDimensionsWrapperApi = createAsyncThunk(
  `${EntityType.DimensionDefinition}/updateDimensionsWrapperApi`,
  async (
    {
      id,
      changes,
    }: { id: string; changes: Partial<DimensionDefinitionEntity> },
    thunkAPI
  ) => {
    await thunkAPI.dispatch(updateDimensionsApi({ id, changes }));
    if ("dimensionColumn" in changes) {
      await thunkAPI.dispatch(
        setExplorerIsStale(
          selectDimensionById(thunkAPI.getState() as RillReduxState, id)
            .metricsDefId,
          true
        )
      );
    }
  }
);
