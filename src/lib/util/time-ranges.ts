import {
  TimeRangeName,
  TimeSeriesTimeRange,
} from "$common/database-service/DatabaseTimeSeriesActions";

const makeTimeRange = (name: TimeRangeName): TimeSeriesTimeRange => {
  switch (name) {
    case TimeRangeName.LastHour:
      return {
        name,
        start: new Date(Date.now() - 60 * 60 * 1000).toISOString(),
        end: new Date().toISOString(),
      };
    case TimeRangeName.Last6Hours:
      return {
        name,
        start: new Date(Date.now() - 6 * 60 * 60 * 1000).toISOString(),
        end: new Date().toISOString(),
      };
    case TimeRangeName.LastDay:
      return {
        name,
        start: new Date(Date.now() - 24 * 60 * 60 * 1000).toISOString(),
        end: new Date().toISOString(),
      };
    case TimeRangeName.Last2Days:
      return {
        name,
        start: new Date(Date.now() - 2 * 24 * 60 * 60 * 1000).toISOString(),
        end: new Date().toISOString(),
      };
    case TimeRangeName.Last5Days:
      return {
        name,
        start: new Date(Date.now() - 5 * 24 * 60 * 60 * 1000).toISOString(),
        end: new Date().toISOString(),
      };
    case TimeRangeName.LastWeek:
      return {
        name,
        start: new Date(Date.now() - 7 * 24 * 60 * 60 * 1000).toISOString(),
        end: new Date().toISOString(),
      };
    case TimeRangeName.Last2Weeks:
      return {
        name,
        start: new Date(Date.now() - 14 * 24 * 60 * 60 * 1000).toISOString(),
        end: new Date().toISOString(),
      };
    case TimeRangeName.Last30Days:
      return {
        name,
        start: new Date(Date.now() - 30 * 24 * 60 * 60 * 1000).toISOString(),
        end: new Date().toISOString(),
      };
    case TimeRangeName.Last60Days:
      return {
        name,
        start: new Date(Date.now() - 60 * 24 * 60 * 60 * 1000).toISOString(),
        end: new Date().toISOString(),
      };
    case TimeRangeName.Today:
      return {
        name,
        start: new Date(new Date().setHours(0, 0, 0, 0)).toISOString(),
        end: new Date().toISOString(),
      };
    case TimeRangeName.MonthToDate:
      return {
        name,
        start: new Date(
          new Date(new Date().setDate(1)).setHours(0, 0, 0, 0)
        ).toISOString(),
        end: new Date().toISOString(),
      };
    // case TimeRangeName.LastMonth:
    //   return {
    //     name,
    //     start: new Date(new Date().setMonth(new Date().getMonth() - 1)),
    //     end: new Date(),
    //   };
    //   // const lastMonth = new Date(new Date().setMonth(new Date().getMonth() - 1));
    //   return {
    //     name,
    //     start: new Date(lastMonth.setDate(1)),
    //     end: new Date(lastMonth.setMonth(lastMonth.getMonth() + 1)),
    //   };
    // case TimeRangeName.CustomRange:
    //   return {
    //     name,
    //     start: new Date(),
    //     end: new Date(),
    //   };
    default:
      throw new Error(`Unknown time range name: ${name}`);
  }
};

export const timeRanges: TimeSeriesTimeRange[] = Object.keys(TimeRangeName).map(
  (name) => makeTimeRange(TimeRangeName[name])
);
