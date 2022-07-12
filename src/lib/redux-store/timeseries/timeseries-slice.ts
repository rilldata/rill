import {
  createSlice,
  createEntityAdapter,
} from "$lib/redux-store/redux-toolkit-wrapper";
import type { TimeSeriesTimeRange } from "$common/database-service/DatabaseTimeSeriesActions";

export type TimeSeriesValue = {
  ts: string;
  bin?: number;
} & Record<string, number>;

export interface TimeSeriesEntity {
  id: string;
  // time range for the time series.
  // this could be different from the one selected by the User
  timeRange: TimeSeriesTimeRange;
  values: Array<TimeSeriesValue>;
  spark: Array<TimeSeriesValue>;
}

const timeSeriesAdapter = createEntityAdapter<TimeSeriesEntity>();

const timeSeriesSlice = createSlice({
  name: "timeSeries",
  initialState: timeSeriesAdapter.getInitialState(),
  reducers: {
    updateTimeSeries: {
      reducer: timeSeriesAdapter.upsertOne,
      prepare: (timeSeriesEntity: TimeSeriesEntity) => ({
        payload: timeSeriesEntity,
      }),
    },
  },
});

export const { updateTimeSeries } = timeSeriesSlice.actions;

export const timeSeriesReducer = timeSeriesSlice.reducer;
