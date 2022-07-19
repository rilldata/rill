import {
  EntityStatus,
  EntityType,
} from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import type { MeasureDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
import type { BigNumberResponse } from "$common/database-service/DatabaseMetricsExploreActions";
import {
  setBigNumberStatus,
  updateBigNumber,
} from "$lib/redux-store/big-number/big-number-slice";
import { selectMetricsExploreParams } from "$lib/redux-store/explore/explore-selectors";
import type { ActiveValues } from "$lib/redux-store/explore/explore-slice";
import { createAsyncThunk } from "$lib/redux-store/redux-toolkit-wrapper";
import type { RillReduxState } from "$lib/redux-store/store-root";
import { streamingFetchWrapper } from "$lib/util/fetchWrapper";
import { isAnythingSelected } from "$lib/util/isAnythingSelected";

/**
 * Async-thunk to generate big numbers for given measures and filters.
 * Streams time series responses from backend  and updates it in the state.
 */
export const generateBigNumbersApi = createAsyncThunk(
  `${EntityType.MetricsExplore}/generateBigNumbers`,
  async (
    {
      id,
      measures,
      filters,
    }: {
      id: string;
      measures?: Array<MeasureDefinitionEntity>;
      filters?: ActiveValues;
    },
    thunkAPI
  ) => {
    const state = thunkAPI.getState() as RillReduxState;
    const { metricsExplore, prunedFilters, normalisedMeasures } =
      selectMetricsExploreParams(state, id, {
        measures,
        filters,
        dimensions: state.dimensionDefinition.entities,
      });
    const anythingSelected = isAnythingSelected(prunedFilters);

    thunkAPI.dispatch(setBigNumberStatus(id, EntityStatus.Running));

    const stream = streamingFetchWrapper<BigNumberResponse>(
      `metrics/${id}/big-number`,
      "POST",
      {
        measures: normalisedMeasures,
        filters: prunedFilters,
        timeRange: metricsExplore.selectedTimeRange,
      }
    );
    for await (const bigNumberEntity of stream) {
      thunkAPI.dispatch(
        updateBigNumber({
          id: bigNumberEntity.id,
          bigNumbers: bigNumberEntity.bigNumbers,
          ...(!anythingSelected
            ? { referenceValues: { ...bigNumberEntity.bigNumbers } }
            : {}),
          status: bigNumberEntity.error
            ? EntityStatus.Error
            : EntityStatus.Idle,
        })
      );
    }
  }
);
