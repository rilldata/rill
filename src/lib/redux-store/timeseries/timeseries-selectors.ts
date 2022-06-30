import { generateEntitySelectors } from "$lib/redux-store/utils/selector-utils";
import type { TimeSeriesEntity } from "$lib/redux-store/timeseries/timeseries-slice";

export const { singleSelector: selectTimeSeriesById } =
  generateEntitySelectors<TimeSeriesEntity>("timeSeries");
