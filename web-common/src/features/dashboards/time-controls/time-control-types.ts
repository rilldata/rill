/** NOTE: this file should be deprecated once we've
 * resolved https://github.com/rilldata/rill-developer/issues/1961.
 */

export enum TimeRangeName_DEPRECATE {
  LAST_SIX_HOURS = "Last 6 hours", // hour
  LAST_24_HOURS = "Last 24 hours", // hour
  LAST_7_DAYS = "Last 7 days", // day
  LAST_4_WEEKS = "Last 4 weeks", // Make last 4 weeks and truncate with week
  ALL_TIME = "All time",
  CUSTOM = "Custom range",
}

export const lastXTimeRangeNames: TimeRangeName_DEPRECATE[] = [
  TimeRangeName_DEPRECATE.LAST_SIX_HOURS,
  TimeRangeName_DEPRECATE.LAST_24_HOURS,
  TimeRangeName_DEPRECATE.LAST_7_DAYS,
  TimeRangeName_DEPRECATE.LAST_4_WEEKS,
];

export const supportedTimeRangeEnums: TimeRangeName_DEPRECATE[] = [
  ...lastXTimeRangeNames,
  TimeRangeName_DEPRECATE.ALL_TIME,
];
