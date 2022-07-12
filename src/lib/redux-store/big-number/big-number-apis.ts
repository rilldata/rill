import { createAsyncThunk } from "$lib/redux-store/redux-toolkit-wrapper";
import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import type { ActiveValues } from "$lib/redux-store/explore/explore-slice";
import type { RillReduxState } from "$lib/redux-store/store-root";
import type { MeasureDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
import { streamingFetchWrapper } from "$lib/util/fetchWrapper";
import { selectMetricsExploreParams } from "$lib/redux-store/explore/explore-selectors";
import { updateBigNumber } from "$lib/redux-store/big-number/big-number-slice";
import { isAnythingSelected } from "$lib/util/isAnythingSelected";
import type { BigNumberResponse } from "$common/database-service/DatabaseMetricsExploreActions";

/**
 * Async-thunk to generate big numbers for given measures and filters.
 * Streams time series responses from backend  and updates it in the state.
 */
export const generateBigNumbersApi = createAsyncThunk(
  `${EntityType.MetricsLeaderboard}/generateBigNumbers`,
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
        })
      );
    }
  }
);
