import {
  EntityStatus,
  EntityType,
} from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import type { MeasureDefinitionEntity } from "$common/data-modeler-state-service/entity-state-service/MeasureDefinitionStateService";
import type {
  TimeSeriesResponse,
  TimeSeriesTimeRange,
} from "$common/database-service/DatabaseTimeSeriesActions";
import { selectMetricsExplorerParams } from "$lib/redux-store/explore/explore-selectors";
import type { ActiveValues } from "$lib/redux-store/explore/explore-slice";
import { createAsyncThunk } from "$lib/redux-store/redux-toolkit-wrapper";
import type { RillReduxState } from "$lib/redux-store/store-root";
import {
  setTimeSeriesStatus,
  updateTimeSeries,
} from "$lib/redux-store/timeseries/timeseries-slice";
import { streamingFetchWrapper } from "$lib/util/fetchWrapper";

/**
 * Async-thunk for generating time series for given measures and filters.
 * Streams time series responses from backend  and updates it in the state.
 */
export const generateTimeSeriesApi = createAsyncThunk(
  `${EntityType.MetricsExplorer}/generateTimeSeries`,
  async (
    {
      id,
      measures,
      filters,
      pixels,
      timeRange,
    }: {
      id: string;
      measures?: Array<MeasureDefinitionEntity>;
      filters?: ActiveValues;
      pixels?: number;
      timeRange?: TimeSeriesTimeRange;
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

    thunkAPI.dispatch(setTimeSeriesStatus(id, EntityStatus.Running));

    const stream = streamingFetchWrapper<TimeSeriesResponse>(
      `metrics/${id}/time-series`,
      "POST",
      {
        measures: normalisedMeasures,
        filters: prunedFilters,
        pixels,
        timeRange: timeRange ?? metricsExplorer.selectedTimeRange,
      }
    );
    for await (const timeSeriesResponse of stream) {
      thunkAPI.dispatch(
        updateTimeSeries({
          id: timeSeriesResponse.id,
          values: timeSeriesResponse.results,
          timeRange: timeSeriesResponse.timeRange,
          spark: timeSeriesResponse.spark,
          status: timeSeriesResponse.error
            ? EntityStatus.Error
            : EntityStatus.Idle,
        })
      );
    }
  }
);
