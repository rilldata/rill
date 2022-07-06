import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import type { RillReduxState } from "$lib/redux-store/store-root";
import { prune } from "../../../routes/_surfaces/workspace/explore/utils";
import { fetchWrapper, streamingFetchWrapper } from "$lib/util/fetchWrapper";
import {
  clearMetricsExplore,
  initMetricsExplore,
  MetricsExploreEntity,
  setBigNumber,
  setLeaderboardDimensionValues,
  setMeasureId,
  toggleExploreMeasure,
  toggleLeaderboardActiveValue,
} from "$lib/redux-store/explore/explore-slice";
import { createAsyncThunk } from "$lib/redux-store/redux-toolkit-wrapper";
import { generateTimeSeriesApi } from "$lib/redux-store/timeseries/timeseries-apis";
import type { DimensionDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/DimensionDefinitionStateService";
import type { MeasureDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";

export const initMeasureAndUpdateExplore = (
  dispatch,
  id: string,
  dimensions: Array<DimensionDefinitionEntity>,
  measures: Array<MeasureDefinitionEntity>
) => {
  dispatch(initMetricsExplore(id, dimensions, measures));
  dispatch(updateLeaderboardApi(id));
  dispatch(
    generateTimeSeriesApi({
      id,
    })
  );
};

export const toggleMeasureAndUpdateExplore = (
  dispatch,
  id: string,
  measureId: string
) => {
  dispatch(toggleExploreMeasure(id, measureId));
  dispatch(
    generateTimeSeriesApi({
      id,
    })
  );
};

export const setMeasureIdAndUpdateLeaderboard = (
  dispatch,
  id: string,
  measureId: string
) => {
  dispatch(setMeasureId(id, measureId));
  dispatch(updateLeaderboardApi(id));
};

export const toggleValueAndUpdateLeaderboard = (
  dispatch,
  id: string,
  dimensionName: string,
  dimensionValue: unknown,
  include: boolean
) => {
  dispatch(
    toggleLeaderboardActiveValue(id, dimensionName, dimensionValue, include)
  );
  dispatch(updateLeaderboardApi(id));
  dispatch(
    generateTimeSeriesApi({
      id,
    })
  );
};

export const clearLeaderboardAndUpdate = (dispatch, id: string) => {
  dispatch(clearMetricsExplore(id));
  dispatch(updateLeaderboardApi(id));
  dispatch(
    generateTimeSeriesApi({
      id,
    })
  );
};

export const updateLeaderboardApi = createAsyncThunk(
  `${EntityType.MetricsLeaderboard}/updateLeaderboard`,
  async (id: string, thunkAPI) => {
    const metricsLeaderboard: MetricsExploreEntity = (
      thunkAPI.getState() as RillReduxState
    ).metricsLeaderboard.entities[id];
    const filters = prune(metricsLeaderboard.activeValues);
    const requestBody = {
      measureId: metricsLeaderboard.measureId,
      filters,
    };

    thunkAPI.dispatch(
      setBigNumber(
        metricsLeaderboard.id,
        await fetchWrapper(
          `metrics/${metricsLeaderboard.id}/bigNumber`,
          "POST",
          requestBody
        )
      )
    );
    const stream = streamingFetchWrapper<{
      dimensionName: string;
      values: Array<unknown>;
    }>(`metrics/${metricsLeaderboard.id}/leaderboards`, "POST", requestBody);
    for await (const dimensionData of stream) {
      thunkAPI.dispatch(
        setLeaderboardDimensionValues(
          metricsLeaderboard.id,
          dimensionData.dimensionName,
          dimensionData.values
        )
      );
    }
  }
);
