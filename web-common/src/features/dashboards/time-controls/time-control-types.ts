import type { V1TimeGrain } from "../../../runtime-client";
export interface TimeRange {
  name: TimeRangeName;
  start: Date;
  end: Date;
}

/** NOTE: this file should be deprecated once we've
 * resolved https://github.com/rilldata/rill-developer/issues/1961.
 */

// TODO: we should deprecate this as soon as its not needed.
// We primarily use this in the DefaultTimeRangeSelector component.
// see https://github.com/rilldata/rill-developer/issues/1961 for progress.
export enum TimeRangeName {
  LAST_SIX_HOURS = "Last 6 hours", // hour
  LAST_24_HOURS = "Last 24 hours", // hour
  LAST_7_DAYS = "Last 7 days", // day
  LAST_4_WEEKS = "Last 4 weeks", // Make last 4 weeks and truncate with week
  ALL_TIME = "All time",
  CUSTOM = "Custom range",
}

export const lastXTimeRangeNames: TimeRangeName[] = [
  TimeRangeName.LAST_SIX_HOURS,
  TimeRangeName.LAST_24_HOURS,
  TimeRangeName.LAST_7_DAYS,
  TimeRangeName.LAST_4_WEEKS,
];

// TODO: we should deprecate this as soon as its not needed.

export const supportedTimeRangeEnums: TimeRangeName[] = [
  ...lastXTimeRangeNames,
  TimeRangeName.ALL_TIME,
];

// The start and end times are rounded to the time grain (interval) such that start is inclusive and end is exclusive.
export interface TimeSeriesTimeRange {
  name?: TimeRangeName;
  start?: string;
  end?: string;
  interval?: V1TimeGrain;
}
