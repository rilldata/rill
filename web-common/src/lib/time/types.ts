import type { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import type { DateTimeUnit } from "luxon";
import type { DEFAULT_TIME_RANGES } from "./config";

// Used for luxon's time units
export enum TimeUnit {
  PT1M = "minute",
  PT1H = "hour",
  P1D = "day",
  P1W = "week",
  P1M = "month",
  P3M = "quarter",
  P1Y = "year",
}

/** a Period is a natural duration of time that maps nicely to calendar time.
 * For instance, when we say a day period, we understand this means a 24-hour period
 * that starts at 00:00:00 and ends at 23:59:59.999. These periods are used for
 * time truncation functions.
 */
export enum Period {
  MINUTE = "PT1M",
  HOUR = "PT1H",
  DAY = "P1D",
  WEEK = "P1W",
  MONTH = "P1M",
  QUARTER = "P3M",
  YEAR = "P1Y",
}

export enum RangePresetType {
  PERIOD_ANCHORED = "PERIOD_ANCHORED",
  OFFSET_ANCHORED = "OFFSET_ANCHORED",
  ALL_TIME = "ALL_TIME",
  FIXED_RANGE = "FIXED_RANGE",
}

export enum ReferencePoint {
  LATEST_DATA = "LatestData",
  NOW = "Now",
}

/** An offset defines an operation on a point in time, primarily used to map from a
 * datetime to something we can pass into the dashboard APIs.
 * An offset on its own is just one operation; in a configuration, you create an array of operations,
 * and those map to indiidual sequential operations.
 *
 * Why are we defining these offsets as an interface / object rather than a function?
 * This will enable us to define wholly-custom time ranges in the configuration. Given that
 * there are really only four operations and one input – a duration – this is a fairly tractable and
 * elegant way to handle almost all of the basic time functions of interest in Rill.
 *
 */

export enum TimeOffsetType {
  /** Add the associated duration to this datetime.
   * @example 2020-05-02 12:22:53 -> ADD PT1H -> 2020-05-02 13:22:53
   */
  ADD = "ADD",
  /** Subtract the associted duration to this datetime.
   * @example 2020-05-02 12:22:53 -> SUBTRACT PT1H -> 2020-05-02 11:22:53
   */
  SUBTRACT = "SUBTRACT",
}

interface TimeOffset {
  duration: Period | string;
  operationType: TimeOffsetType;
}

interface TimeTruncation {
  period: Period;
  truncationType: TimeTruncationType;
}

/**
 * These types tell Rill to take the supplied duration, and map it to the beginning
 * or end of the period in which the datetime object is currently in. We utilize the ISO8601 duration
 * to specify when this duration should technically start; we will likely drastically limit the complexity
 * to a small subset of available values. For now, we'll be capitalizing on the Period enum to keep the set
 * of available periods to a normal amount.
 */
export enum TimeTruncationType {
  /**
   * @example 2020-05-02 12:23:53 -> START_OF_PERIOD PT1H -> 2020-05-02 12:00:00.000
   */
  START_OF_PERIOD = "START_OF_PERIOD",
  /**
   * @example 2020-05-02 12:22:53 -> END_OF_PERIOD PT1H -> 2020-05-02 12:59:59.999
   */
  END_OF_PERIOD = "END_OF_PERIOD",
}

/** An offset defines a transformation of an existing datetime into something more usable by our APIs
 * (and more coherent to humans). We only need to specify two types of offsets:
 * - operation, like subtracting or adding a time duration.
 * - truncation, which enables us to get the beginning of end of a period of interest.
 */
export type RelativeTimeTransformation = TimeOffset | TimeTruncation;

export interface RelativePointInTime {
  reference: ReferencePoint;
  transformation: RelativeTimeTransformation[];
}

export interface TimeRangeMeta {
  label: string;
  defaultGrain?: V1TimeGrain; // Affordance for future use
  rangePreset?: RangePresetType | string;
  defaultComparison?: TimeComparisonOption | string;
  start?: string | RelativePointInTime;
  end?: string | RelativePointInTime;
}

// FIXME: this will have to be relaxed when the dashboard time ranges
// are settable within a config.
export type TimeRangeType = keyof typeof DEFAULT_TIME_RANGES;

// FIXME: this is confusing. Why do we have RangePreset and TimeRangePreset?
// And why do we need to define this explicitly?
export const TimeRangePreset: { [K in TimeRangeType]: K } = {
  ALL_TIME: "ALL_TIME",
  LAST_SIX_HOURS: "LAST_SIX_HOURS",
  LAST_24_HOURS: "LAST_24_HOURS",
  LAST_7_DAYS: "LAST_7_DAYS",
  LAST_4_WEEKS: "LAST_4_WEEKS",
  LAST_YEAR: "LAST_YEAR",
  TODAY: "TODAY",
  THIS_WEEK: "THIS_WEEK",
  THIS_MONTH: "THIS_MONTH",
  THIS_QUARTER: "THIS_QUARTER",
  THIS_YEAR: "THIS_YEAR",
  CUSTOM: "CUSTOM",
};

export interface TimeRange {
  name: TimeRangeType;
  start: Date;
  end: Date;
}

export interface TimeRangeOption extends TimeRange {
  label: string;
}

export interface DashboardTimeControls extends TimeRange {
  interval?: V1TimeGrain;
  label?: string;
}

/** defines configuration for a time grain object in Rill. */
export interface TimeGrain {
  /** the grain defined by the runtime */
  grain: V1TimeGrain;
  /** a human-readable name, e.g. minute, second, etc. */
  label: DateTimeUnit | string;
  /** the ISO8601 duration, e.g. P1D, PT6H */
  duration: Period;
  /** the DateTimeFormatOptions of the Intl API that outputs
   * a human-readable representation of a timestamp based on
   * the time grain. This preserves locale-based formatting.
   */
  formatDate: Intl.DateTimeFormatOptions;
}

// FIXME: is this needed?
export interface TimeGrainOption extends TimeGrain {
  enabled: boolean;
}

// limit the set of available time grains to those supported
// by th dashboard.
export type AvailableTimeGrain = Exclude<
  V1TimeGrain,
  "TIME_GRAIN_UNSPECIFIED" | "TIME_GRAIN_MILLISECOND" | "TIME_GRAIN_SECOND"
>;

export enum TimeComparisonOption {
  CONTIGUOUS = "CONTIGUOUS",
  CUSTOM = "CUSTOM_COMPARISON_RANGE",
  DAY = "P1D",
  WEEK = "P1W",
  MONTH = "P1M",
  QUARTER = "P3M",
  YEAR = "P1Y",
}

export enum TimeRoundingStrategy {
  NEAREST = "NEAREST",
  PREVIOUS = "PREVIOUS",
}
