import {
  TimeRangeName,
  TimeSeriesTimeRange,
} from "$common/database-service/DatabaseTimeSeriesActions";

const makeSelectableTimeRange = (
  name: TimeRangeName,
  datasetTimeRange: TimeSeriesTimeRange
): TimeSeriesTimeRange => {
  const start = new Date(datasetTimeRange?.start);
  const end = new Date(datasetTimeRange?.end);
  switch (name) {
    case TimeRangeName.LastHour:
      return {
        name,
        start: new Date(end.getTime() - 60 * 60 * 1000).toISOString(),
        end: end.toISOString(),
      };
    case TimeRangeName.Last6Hours:
      return {
        name,
        start: new Date(end.getTime() - 6 * 60 * 60 * 1000).toISOString(),
        end: end.toISOString(),
      };
    case TimeRangeName.LastDay:
      return {
        name,
        start: new Date(end.getTime() - 24 * 60 * 60 * 1000).toISOString(),
        end: end.toISOString(),
      };
    case TimeRangeName.Last2Days:
      return {
        name,
        start: new Date(end.getTime() - 2 * 24 * 60 * 60 * 1000).toISOString(),
        end: end.toISOString(),
      };
    case TimeRangeName.Last5Days:
      return {
        name,
        start: new Date(end.getTime() - 5 * 24 * 60 * 60 * 1000).toISOString(),
        end: end.toISOString(),
      };
    case TimeRangeName.LastWeek:
      return {
        name,
        start: new Date(end.getTime() - 7 * 24 * 60 * 60 * 1000).toISOString(),
        end: end.toISOString(),
      };
    case TimeRangeName.Last2Weeks:
      return {
        name,
        start: new Date(end.getTime() - 14 * 24 * 60 * 60 * 1000).toISOString(),
        end: end.toISOString(),
      };
    case TimeRangeName.Last30Days:
      return {
        name,
        start: new Date(end.getTime() - 30 * 24 * 60 * 60 * 1000).toISOString(),
        end: end.toISOString(),
      };
    case TimeRangeName.Last60Days:
      return {
        name,
        start: new Date(end.getTime() - 60 * 24 * 60 * 60 * 1000).toISOString(),
        end: end.toISOString(),
      };
    case TimeRangeName.AllTime:
      return {
        name,
        start: start.toISOString(),
        end: end.toISOString(),
      };
    default:
      throw new Error(`Unknown time range name: ${name}`);
  }
};

export const makeSelectableTimeRanges = (
  fullTimeRangeInDataset: TimeSeriesTimeRange
): TimeSeriesTimeRange[] => {
  return Object.keys(TimeRangeName).map((name) =>
    makeSelectableTimeRange(TimeRangeName[name], fullTimeRangeInDataset)
  );
};
