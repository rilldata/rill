import {
  EntityStatus,
  EntityType,
} from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import type {
  BasicMeasureDefinition,
  MeasureDefinitionEntity,
} from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
import type { BigNumberResponse } from "$common/database-service/DatabaseMetricsExplorerActions";
import {
  setBigNumber,
  setBigNumberStatus,
  setReferenceValues,
} from "$lib/redux-store/big-number/big-number-slice";
import { selectMetricsExplorerParams } from "$lib/redux-store/explore/explore-selectors";
import type {
  ActiveValues,
  MetricsExplorerEntity,
} from "$lib/redux-store/explore/explore-slice";
import { createAsyncThunk } from "$lib/redux-store/redux-toolkit-wrapper";
import type { RillReduxState } from "$lib/redux-store/store-root";
import { streamingFetchWrapper } from "$lib/util/fetchWrapper";
import { isAnythingSelected } from "$lib/util/isAnythingSelected";
import { selectBigNumberById } from "$lib/redux-store/big-number/big-number-selectors";

/**
 * Async-thunk to generate big numbers for given measures and filters.
 * Streams time series responses from backend  and updates it in the state.
 */
export const generateBigNumbersApi = createAsyncThunk(
  `${EntityType.MetricsExplorer}/generateBigNumbers`,
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
    const { metricsExplorer, prunedFilters, normalisedMeasures } =
      selectMetricsExplorerParams(state, id, {
        measures,
        filters,
        dimensions: state.dimensionDefinition.entities,
      });
    const anythingSelected = isAnythingSelected(prunedFilters);

    thunkAPI.dispatch(setBigNumberStatus(id, EntityStatus.Running));

    const { payload: bigNumbers } = (await thunkAPI.dispatch(
      getBigNumberApi({
        id,
        metricsExplorer,
        normalisedMeasures,
        prunedFilters,
      })
    )) as { payload: Record<string, number> };
    if (!anythingSelected) {
      thunkAPI.dispatch(setReferenceValues(id, bigNumbers));
    } else {
      const bigNumbers = selectBigNumberById(state, id);
      if (!bigNumbers?.referenceValues) {
        const { payload: referenceValues } = (await thunkAPI.dispatch(
          getBigNumberApi({
            id,
            metricsExplorer,
            normalisedMeasures,
            prunedFilters: {},
          })
        )) as { payload: Record<string, number> };
        thunkAPI.dispatch(setReferenceValues(id, referenceValues));
      }
    }

    thunkAPI.dispatch(setBigNumber(id, bigNumbers));
    thunkAPI.dispatch(setBigNumberStatus(id, EntityStatus.Idle));
  }
);

const getBigNumberApi = createAsyncThunk(
  `${EntityType.MetricsExplorer}/getBigNumberApi`,
  async ({
    id,
    metricsExplorer,
    normalisedMeasures,
    prunedFilters,
  }: {
    id: string;
    metricsExplorer: MetricsExplorerEntity;
    normalisedMeasures: Array<BasicMeasureDefinition>;
    prunedFilters: ActiveValues;
  }) => {
    const stream = streamingFetchWrapper<BigNumberResponse>(
      `metrics/${id}/big-number`,
      "POST",
      {
        measures: normalisedMeasures,
        filters: prunedFilters,
        timeRange: metricsExplorer.selectedTimeRange,
      }
    );
    let bigNumbers: Record<string, number> = {};
    for await (const bigNumberEntity of stream) {
      // TODO: handle multiple big numbers' return
      bigNumbers = bigNumberEntity.bigNumbers;
    }
    return bigNumbers;
  }
);
