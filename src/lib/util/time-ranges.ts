enum TimeRangeName {
  LastHour = "Last hour",
  Last6Hours = "Last 6 hours",
  LastDay = "Last day",
  Last2Days = "Last 2 days",
  Last5Days = "Last 5 days",
  LastWeek = "Last week",
  Last2Weeks = "Last 2 weeks",
  Last30Days = "Last 30 days",
  Last60Days = "Last 60 days",
  Today = "Today",
  MonthToDate = "Month to date",
  // CustomRange = "Custom range",
}

export interface TimeRange {
  name: TimeRangeName;
  start: Date;
  end: Date;
}

const makeTimeRange = (name: TimeRangeName): TimeRange => {
  switch (name) {
    case TimeRangeName.LastHour:
      return {
        name,
        start: new Date(Date.now() - 60 * 60 * 1000),
        end: new Date(),
      };
    case TimeRangeName.Last6Hours:
      return {
        name,
        start: new Date(Date.now() - 6 * 60 * 60 * 1000),
        end: new Date(),
      };
    case TimeRangeName.LastDay:
      return {
        name,
        start: new Date(Date.now() - 24 * 60 * 60 * 1000),
        end: new Date(),
      };
    case TimeRangeName.Last2Days:
      return {
        name,
        start: new Date(Date.now() - 2 * 24 * 60 * 60 * 1000),
        end: new Date(),
      };
    case TimeRangeName.Last5Days:
      return {
        name,
        start: new Date(Date.now() - 5 * 24 * 60 * 60 * 1000),
        end: new Date(),
      };
    case TimeRangeName.LastWeek:
      return {
        name,
        start: new Date(Date.now() - 7 * 24 * 60 * 60 * 1000),
        end: new Date(),
      };
    case TimeRangeName.Last2Weeks:
      return {
        name,
        start: new Date(Date.now() - 14 * 24 * 60 * 60 * 1000),
        end: new Date(),
      };
    case TimeRangeName.Last30Days:
      return {
        name,
        start: new Date(Date.now() - 30 * 24 * 60 * 60 * 1000),
        end: new Date(),
      };
    case TimeRangeName.Last60Days:
      return {
        name,
        start: new Date(Date.now() - 60 * 24 * 60 * 60 * 1000),
        end: new Date(),
      };
    case TimeRangeName.Today:
      return {
        name,
        start: new Date(new Date().setHours(0, 0, 0, 0)),
        end: new Date(),
      };
    case TimeRangeName.MonthToDate:
      return {
        name,
        start: new Date(new Date(new Date().setDate(1)).setHours(0, 0, 0, 0)),
        end: new Date(),
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

export const timeRanges: TimeRange[] = Object.keys(TimeRangeName).map((name) =>
  makeTimeRange(TimeRangeName[name])
);
