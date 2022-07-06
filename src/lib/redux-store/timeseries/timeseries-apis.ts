import type { ActiveValues } from "$lib/redux-store/explore/explore-slice";
import type {
  BasicMeasureDefinition,
  MeasureDefinitionEntity,
} from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
import { createAsyncThunk } from "$lib/redux-store/redux-toolkit-wrapper";
import { EntityType } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import { streamingFetchWrapper } from "$lib/util/fetchWrapper";
import type { TimeSeriesResponse } from "$common/database-service/DatabaseTimeSeriesActions";
import { updateTimeSeries } from "$lib/redux-store/timeseries/timeseries-slice";
import type { MetricsExploreEntity } from "$lib/redux-store/explore/explore-slice";
import type { RillReduxState } from "$lib/redux-store/store-root";
import { prune } from "../../../routes/_surfaces/workspace/explore/utils";
import type { RollupInterval } from "$common/database-service/DatabaseColumnActions";
import { selectMeasureById } from "$lib/redux-store/measure-definition/measure-definition-selectors";

export const generateTimeSeriesApi = createAsyncThunk(
  `${EntityType.MetricsLeaderboard}/generateTimeSeries`,
  async (
    {
      id,
      measures,
      filters,
      pixels,
      rollupInterval,
    }: {
      id: string;
      measures?: Array<MeasureDefinitionEntity>;
      filters?: ActiveValues;
      pixels?: number;
      rollupInterval?: RollupInterval;
    },
    thunkAPI
  ) => {
    const state = thunkAPI.getState() as RillReduxState;
    const metricsLeaderboard: MetricsExploreEntity =
      state.metricsLeaderboard.entities[id];
    if (!filters) {
      filters = prune(metricsLeaderboard.activeValues);
    }
    if (!measures) {
      measures = metricsLeaderboard.measureIds.map((measureId) =>
        selectMeasureById(state, measureId)
      );
    }

    const stream = streamingFetchWrapper<TimeSeriesResponse>(
      `metrics/${id}/time-series`,
      "POST",
      {
        measures: measures.map(
          (measure) =>
            ({
              id: measure.id,
              expression: measure.expression,
              sqlName: measure.sqlName,
            } as BasicMeasureDefinition)
        ),
        filters,
        pixels,
        rollupInterval,
      }
    );
    for await (const timeSeriesResponse of stream) {
      thunkAPI.dispatch(
        updateTimeSeries({
          id: timeSeriesResponse.id,
          values: timeSeriesResponse.results,
          rollupInterval: timeSeriesResponse.rollupInterval,
          spark: timeSeriesResponse.spark,
        })
      );
    }
  }
);
