import {
  createSlice,
  createEntityAdapter,
} from "$lib/redux-store/redux-toolkit-wrapper";
import type { TimeSeriesTimeRange } from "$common/database-service/DatabaseTimeSeriesActions";
import type { EntityStatus } from "$common/data-modeler-state-service/entity-state-service/EntityStateService";
import {
  setStatusPrepare,
  setStatusReducer,
} from "$lib/redux-store/utils/loading-utils";

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
  status: EntityStatus;
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

    setTimeSeriesStatus: {
      reducer: setStatusReducer,
      prepare: setStatusPrepare,
    },
  },
});

export const { updateTimeSeries, setTimeSeriesStatus } =
  timeSeriesSlice.actions;

export const timeSeriesReducer = timeSeriesSlice.reducer;
