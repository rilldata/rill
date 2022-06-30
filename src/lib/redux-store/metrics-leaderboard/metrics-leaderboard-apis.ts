import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import type { RillReduxState } from "$lib/redux-store/store-root";
import { prune } from "../../../routes/_surfaces/workspace/leaderboard/utils";
import { fetchWrapper, streamingFetchWrapper } from "$lib/util/fetchWrapper";
import {
  clearLeaderboard,
  MetricsLeaderboardEntity,
  setBigNumber,
  setDimensionLeaderboard,
  setMeasureId,
  toggleLeaderboardActiveValue,
} from "$lib/redux-store/metrics-leaderboard/metrics-leaderboard-slice";
import { createAsyncThunk } from "$lib/redux-store/redux-toolkit-wrapper";
import { generateTimeSeriesApi } from "$lib/redux-store/timeseries/timeseries-apis";

export const updateLeaderboardMeasure = (
  dispatch,
  id: string,
  measureId: string,
  expression: string
) => {
  dispatch(setMeasureId(id, measureId));
  dispatch(updateLeaderboardApi(id));
  dispatch(
    generateTimeSeriesApi({
      metricsDefId: id,
      measures: [{ id, expression } as any],
    })
  );
};

export const toggleValueAndUpdateLeaderboard = (
  dispatch,
  id: string,
  dimensionName: string,
  dimensionValue: unknown,
  include: boolean,
  expression: string
) => {
  dispatch(
    toggleLeaderboardActiveValue(id, dimensionName, dimensionValue, include)
  );
  dispatch(updateLeaderboardApi(id));
  dispatch(
    generateTimeSeriesApi({
      metricsDefId: id,
      measures: [{ id, expression } as any],
    })
  );
};

export const clearLeaderboardAndUpdate = (
  dispatch,
  id: string,
  expression: string
) => {
  dispatch(clearLeaderboard(id));
  dispatch(updateLeaderboardApi(id));
  dispatch(
    generateTimeSeriesApi({
      metricsDefId: id,
      measures: [{ id, expression } as any],
    })
  );
};

export const updateLeaderboardApi = createAsyncThunk(
  `${EntityType.MetricsLeaderboard}/updateLeaderboard`,
  async (id: string, thunkAPI) => {
    const metricsLeaderboard: MetricsLeaderboardEntity = (
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
        setDimensionLeaderboard(
          metricsLeaderboard.id,
          dimensionData.dimensionName,
          dimensionData.values
        )
      );
    }
  },
  {
    condition: (id: string, { getState }) => {
      const metricsLeaderboard: MetricsLeaderboardEntity = (
        getState() as RillReduxState
      ).metricsLeaderboard.entities[id];
      return metricsLeaderboard.measureId !== "";
    },
  }
);
