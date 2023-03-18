import type { DateTimeUnit } from "luxon";
import { V1TimeGrain } from "../../../../runtime-client";
import type { DEFAULT_TIME_RANGES } from "./defaults";

export const TIME = {
  MILLISECOND: 1,
  get SECOND() {
    return 1000 * this.MILLISECOND;
  },
  get MINUTE() {
    return 60 * this.SECOND;
  },
  get HOUR() {
    return 60 * this.MINUTE;
  },
  get DAY() {
    return 24 * this.HOUR;
  },
  get WEEK() {
    return 7 * this.DAY;
  },
  get MONTH() {
    return 30 * this.DAY;
  },
  get YEAR() {
    return 365 * this.DAY;
  },
};

// Used for luxon's time units
export const TimeUnit: Record<string, DateTimeUnit> = {
  PT1M: "minute",
  PT1H: "hour",
  P1D: "day",
  P1W: "week",
  P1M: "month",
  P3M: "quarter",
  P1Y: "year",
};

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

export const TimeGrainTriplets = [
  {
    timeGrain: V1TimeGrain.TIME_GRAIN_MINUTE,
    duration: Period.MINUTE,
    timeGrainUnit: TimeUnit.PT1M,
  },
  {
    timeGrain: V1TimeGrain.TIME_GRAIN_HOUR,
    duration: Period.HOUR,
    timeGrainUnit: TimeUnit.PT1H,
  },
  {
    timeGrain: V1TimeGrain.TIME_GRAIN_DAY,
    duration: Period.DAY,
    timeGrainUnit: TimeUnit.P1D,
  },
  {
    timeGrain: V1TimeGrain.TIME_GRAIN_WEEK,
    duration: Period.WEEK,
    timeGrainUnit: TimeUnit.P1W,
  },
  {
    timeGrain: V1TimeGrain.TIME_GRAIN_MONTH,
    duration: Period.MONTH,
    timeGrainUnit: TimeUnit.P1M,
  },
  {
    timeGrain: V1TimeGrain.TIME_GRAIN_YEAR,
    duration: Period.YEAR,
    timeGrainUnit: TimeUnit.P1Y,
  },
];

export const grainToDuration = (grain: V1TimeGrain) => {
  const triplet = TimeGrainTriplets.find((t) => t.timeGrain === grain);
  return triplet?.duration;
};

export const grainToUnit = (grain: V1TimeGrain) => {
  const triplet = TimeGrainTriplets.find((t) => t.timeGrain === grain);
  return triplet?.timeGrainUnit;
};

export enum RangePreset {
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
  rangePreset?: RangePreset | string;
  start?: string | RelativePointInTime;
  end?: string | RelativePointInTime;
}

// FIXME: this will have to be relaxed when the dashboard time ranges
// are settable within a config.
export type TimeRangeType = keyof typeof DEFAULT_TIME_RANGES;

// HAMILTON: I left off here. Get TimeRangePreset to just use the DEFAULT_TIME_RANGES keys.
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
  WEEK_TO_DATE: "WEEK_TO_DATE",
  MONTH_TO_DATE: "MONTH_TO_DATE",
  YEAR_TO_DATE: "YEAR_TO_DATE",
  CUSTOM: "CUSTOM",
  // ...Object.keys(DEFAULT_TIME_RANGES).reduce((obj, key) => {
  //   obj[key] = key;
  //   return obj;
  // }, {}),
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

export interface TimeGrain {
  grain: V1TimeGrain;
  label: DateTimeUnit;
  duration: Period;
  width: number;
}

export interface TimeGrainOption extends TimeGrain {
  enabled: boolean;
}
