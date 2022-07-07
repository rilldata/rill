import {
  createSlice,
  createEntityAdapter,
} from "$lib/redux-store/redux-toolkit-wrapper";

export interface TimeSeriesValue {
  ts: string;
  count: number;
  bin?: number;
}

export interface TimeSeriesEntity {
  id: string;
  rollupInterval: string;
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
