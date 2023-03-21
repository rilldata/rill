import type { V1TimeGrain } from "../../../runtime-client";
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

export enum ComparisonRange {
  // Previous period for the given TimeRangeName
  Previous6Hours = "Previous 6 hours",
  PreviousDay = "Previous day",
  PreviousWeek = "Previous week",
  Previous30Days = "Previous 30 days",
  PreviousYear = "Previous year",

  // Custom Range which the user can select
  Custom = "Custom range",
}

export const lastXTimeRangeNames: TimeRangeName[] = [
  TimeRangeName.Last6Hours,
  TimeRangeName.LastDay,
  TimeRangeName.LastWeek,
  TimeRangeName.Last30Days,
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

export interface ComparisonWithTimeRange {
  name?: ComparisonRange;
  start?: string;
  end?: string;
}
