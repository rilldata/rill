import { V1TimeGrain } from "../../../../runtime-client";
import { TimeRangeName } from "../time-control-types";
import { getAllowedTimeGrains } from "../time-range-utils";
import {
  getTimeWidth,
  ISOToMilliseconds,
  relativePointInTimeToAbsolute,
} from "./time-anchors";
import { isGrainBigger } from "./time-grain";
import { Period, TIME } from "./time-types";

// enum - screaming snake case
// interface - Camel case

enum RangePreset {
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

export interface TimeRange {
  label: string;
  defaultGrain?: V1TimeGrain; // Affordance for future use
  rangePreset?: RangePreset | string;
  start?: string | RelativePointInTime;
  end?: string | RelativePointInTime;
}

export const TIME_RANGES: TimeRange[] = [
  {
    label: "Last 6 Hours",
    rangePreset: RangePreset.OFFSET_ANCHORED,
    start: {
      reference: ReferencePoint.LATEST_DATA,
      transformation: [
        { duration: "PT6H", operationType: TimeOffsetType.SUBTRACT }, // operation
        {
          period: Period.HOUR,
          truncationType: TimeTruncationType.START_OF_PERIOD,
        }, // truncation
      ],
    },
    end: {
      reference: ReferencePoint.LATEST_DATA,
      transformation: [
        { duration: "PT1H", operationType: TimeOffsetType.SUBTRACT },
        {
          period: Period.HOUR,
          truncationType: TimeTruncationType.START_OF_PERIOD,
        },
      ],
    },
  },
  {
    label: "All time data",
    rangePreset: RangePreset.ALL_TIME,
  },
];

// Get the default grain for a given time range
export function getDefaultTimeGrain(start: Date, end: Date): V1TimeGrain {
  const timeRangeDurationMs = end.getTime() - start.getTime();

  if (timeRangeDurationMs < 2 * TIME.HOUR) {
    return V1TimeGrain.TIME_GRAIN_MINUTE;
  } else if (timeRangeDurationMs < 7 * TIME.DAY) {
    return V1TimeGrain.TIME_GRAIN_HOUR;
  } else if (timeRangeDurationMs < 3 * TIME.MONTH) {
    return V1TimeGrain.TIME_GRAIN_DAY;
  } else if (timeRangeDurationMs < 3 * TIME.YEAR) {
    return V1TimeGrain.TIME_GRAIN_WEEK;
  } else {
    return V1TimeGrain.TIME_GRAIN_MONTH;
  }
}

// Loop through all preset to check if they can be a part of subset of given start and end date
export function getChildTimeRanges(
  start: Date,
  end: Date,
  minTimeGrain: V1TimeGrain
) {
  const timeRanges = [];

  for (const timeRange of TIME_RANGES) {
    const timeRangeDates = relativePointInTimeToAbsolute(
      start,
      timeRange.start,
      timeRange.end
    );

    // check if valid
    timeRanges.push({
      label: timeRange.label,
      start: timeRangeDates.startDate,
      end: timeRangeDates.endDate,
    });

    // if (timeRange.rangePreset == RangePreset.ALL_TIME) {
    //   // All time is always an option
    //   timeRanges.push({
    //     label: timeRange.label,
    //     start,
    //     end,
    //   });
    // }

    // const timeRangeDurationMs = ISOToMilliseconds(timeRangeName.duration);
    // // only show a time range if it is within the time range of the data and supports minTimeGrain
    // const showTimeRange = timeRangeDurationMs <= durationMs;

    // const allowedTimeGrains = getAllowedTimeGrains(timeRangeDurationMs);
    // const allowedMaxGrain = allowedTimeGrains[allowedTimeGrains.length - 1];
    // const isGrainPossible = !isGrainBigger(minTimeGrain, allowedMaxGrain);

    // if (showTimeRange && isGrainPossible) {
    //   const timeRange = makeRelativeTimeRange(timeRangeName, allTimeRange);
    //   timeRanges.push(timeRange);
    // }

    // const timeRangeDurationMs = ISOToMilliseconds(timeRangeName.duration);
  }

  return timeRanges;
}

export function makeRelativeTimeRange(
  timeRangeName: TimeRangeName,
  allTimeRange: TimeRange
): TimeRange {
  if (timeRangeName === TimeRangeName.AllTime) return allTimeRange;
  const startTime = new Date(
    allTimeRange.end.getTime() - getLastXTimeRangeDurationMs(timeRangeName)
  );
  return {
    name: timeRangeName,
    start: startTime,
    end: allTimeRange.end,
  };
}
