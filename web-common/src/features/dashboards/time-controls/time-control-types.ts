export interface TimeRange {
  name: TimeRangeName;
  start: Date;
  end: Date;
}

export enum TimeRangeName {
  Last6Hours = "Last 6 hours",
  LastDay = "Last day",
  LastWeek = "Last week",
  Last30Days = "Last 30 days",
  AllTime = "All time",
  // Today = "Today",
  // MonthToDate = "Month to date",
  Custom = "Custom range",
}

export const lastXTimeRangeNames: TimeRangeName[] = [
  TimeRangeName.Last6Hours,
  TimeRangeName.LastDay,
  TimeRangeName.LastWeek,
  TimeRangeName.Last30Days,
];

// The string values must adhere to DuckDB INTERVAL syntax, since, in some places, we interpolate an SQL queries with these values.
export enum TimeGrain {
  OneMinute = "minute",
  // FiveMinutes = "5 minute",
  // FifteenMinutes = "15 minute",
  OneHour = "hour",
  OneDay = "day",
  OneWeek = "week",
  OneMonth = "month",
  OneYear = "year",
}

// The start and end times are rounded to the time grain (interval) such that start is inclusive and end is exclusive.
export interface TimeSeriesTimeRange {
  name?: TimeRangeName;
  start?: string;
  end?: string;
  interval?: TimeGrain;
}
