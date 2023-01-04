export type TimeSeriesValue = {
  ts: string;
  bin?: number;
} & Record<string, number>;

export interface TimeSeriesResponse {
  id?: string;
  results: Array<TimeSeriesValue>;
  spark?: Array<TimeSeriesValue>;
  timeRange?: TimeSeriesTimeRange;
  sampleSize?: number;
  error?: string;
}

export enum TimeRangeName {
  LastHour = "Last hour",
  Last6Hours = "Last 6 hours",
  LastDay = "Last day",
  Last2Days = "Last 2 days",
  Last5Days = "Last 5 days",
  LastWeek = "Last week",
  Last2Weeks = "Last 2 weeks",
  Last30Days = "Last 30 days",
  Last60Days = "Last 60 days",
  AllTime = "All time",
  // Today = "Today",
  // MonthToDate = "Month to date",
  // CustomRange = "Custom range",
}

export const lastXTimeRanges: TimeRangeName[] = [
  TimeRangeName.LastHour,
  TimeRangeName.Last6Hours,
  TimeRangeName.LastDay,
  TimeRangeName.Last2Days,
  TimeRangeName.Last5Days,
  TimeRangeName.LastWeek,
  TimeRangeName.Last2Weeks,
  TimeRangeName.Last30Days,
  TimeRangeName.Last60Days,
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
  interval?: TimeGrain; // TODO: switch this to TimeGrain
}
