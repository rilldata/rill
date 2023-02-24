import type { V1TimeGrain } from "../../../runtime-client";
export interface TimeRange {
  name: TimeRangeName;
  start: Date;
  end: Date;
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
  Custom = "Custom range",
}

export const lastXTimeRangeNames: TimeRangeName[] = [
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

export const supportedTimeRangeEnums: TimeRangeName[] = [
  ...lastXTimeRangeNames,
  TimeRangeName.AllTime,
];

// The start and end times are rounded to the time grain (interval) such that start is inclusive and end is exclusive.
export interface TimeSeriesTimeRange {
  name?: TimeRangeName;
  start?: string;
  end?: string;
  interval?: V1TimeGrain;
}
