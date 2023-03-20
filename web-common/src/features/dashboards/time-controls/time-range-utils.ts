/** NOTE:
 *
 * this file should be deprecated in favor of the other time utils.
 *
 * */
import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import {
  lastXTimeRangeNames,
  TimeRangeName_DEPRECATE,
} from "./time-control-types";
import { TimeRange } from "./utils/time-types";

import { TIME_GRAIN } from "./utils/config";
import { durationToMillis } from "./utils/time-grain";

// May not need this anymore as using TimeGrain objects
export const supportedTimeGrainEnums = () => {
  const supportedEnums: string[] = [];
  const unsupportedTypes = [
    V1TimeGrain.TIME_GRAIN_UNSPECIFIED,
    V1TimeGrain.TIME_GRAIN_MILLISECOND,
    V1TimeGrain.TIME_GRAIN_SECOND,
  ];

  for (const timeGrain in V1TimeGrain) {
    if (unsupportedTypes.includes(V1TimeGrain[timeGrain])) {
      continue;
    }
    supportedEnums.push(timeGrain);
  }

  return supportedEnums;
};

// NOTE: we will need to keep this for the duration amounts in the runtime / config.
// let's plan to deprecate it later.
export function getRelativeTimeRangeOptions(
  allTimeRange: TimeRange,
  minTimeGrain: V1TimeGrain
): TimeRange[] {
  const allTimeRangeDurationMs = +allTimeRange.end - +allTimeRange.start;
  const timeRanges: TimeRange[] = [];

  for (const timeRangeName of lastXTimeRangeNames) {
    const timeRangeDurationMs = getLastXTimeRangeDurationMs(timeRangeName);

    // only show a time range if it is within the time range of the data and supports minTimeGrain
    const showTimeRange = timeRangeDurationMs <= allTimeRangeDurationMs;

    const allowedTimeGrains = getAllowedTimeGrains(timeRangeDurationMs);
    const allowedMaxGrain = allowedTimeGrains[allowedTimeGrains.length - 1];
    const isGrainPossible = !isGrainBigger(minTimeGrain, allowedMaxGrain);

    if (showTimeRange && isGrainPossible) {
      const timeRange = makeRelativeTimeRange(timeRangeName, allTimeRange);
      timeRanges.push(timeRange);
    }
  }

  // All time is always an option
  timeRanges.push({
    name: TimeRangeName_DEPRECATE.ALL_TIME,
    start: allTimeRange.start,
    end: allTimeRange.end,
  });

  return timeRanges;
}

//TODO: Co locate with TimeControls
// export function getDefaultTimeRange(allTimeRange: TimeRange): TimeRange {
//   // Use AllTime for now. When we go to production real-time datasets, we'll want to change this.
//   return allTimeRange;
// }

// Moved to time range and renamed to isTimeRangeValidForMinTimeGrain
export function isTimeRangeValidForTimeGrain(
  minTimeGrain: V1TimeGrain,
  timeRange: TimeRangeName_DEPRECATE
): boolean {
  const timeGrainEnums = supportedTimeGrainEnums();
  if (!timeGrainEnums.includes(minTimeGrain)) {
    return true;
  }
  if (!timeRange || timeRange === TimeRangeName_DEPRECATE.ALL_TIME) {
    return true;
  }

  const timeRangeDurationMs = getLastXTimeRangeDurationMs(timeRange);

  const allowedTimeGrains = getAllowedTimeGrains(timeRangeDurationMs);
  const maxAllowedTimeGrain = allowedTimeGrains[allowedTimeGrains.length - 1];
  return !isGrainBigger(minTimeGrain, maxAllowedTimeGrain);
}

// Maps a TimeRangeName to an ISO duration.
// This should eventually be deprecated once we have
// changed the runtime definition for default_time_range to be a preset string.
// see https://github.com/rilldata/rill-developer/issues/1961
export const timeRangeToISODuration = (
  timeRangeName: TimeRangeName_DEPRECATE
): string => {
  switch (timeRangeName) {
    case TimeRangeName_DEPRECATE.LAST_SIX_HOURS:
      return "PT6H";
    case TimeRangeName_DEPRECATE.LAST_24_HOURS:
      return "P1D";
    case TimeRangeName_DEPRECATE.LAST_7_DAYS:
      return "P7D";
    case TimeRangeName_DEPRECATE.LAST_4_WEEKS:
      return "P4W";
    case TimeRangeName_DEPRECATE.ALL_TIME:
      return "inf";
    default:
      return undefined;
  }
};

// This should eventually be deprecated once we have
// changed the runtime definition for default_time_range to be a preset string.
// see https://github.com/rilldata/rill-developer/issues/1961
export const ISODurationToTimeRange = (
  isoDuration: string,
  defaultToAllTime = true
): TimeRangeName_DEPRECATE => {
  switch (isoDuration) {
    case "PT6H":
      return TimeRangeName_DEPRECATE.LAST_SIX_HOURS;
    case "P1D":
      return TimeRangeName_DEPRECATE.LAST_24_HOURS;
    case "P7D":
      return TimeRangeName_DEPRECATE.LAST_7_DAYS;
    case "P4W":
      return TimeRangeName_DEPRECATE.LAST_4_WEEKS;
    case "inf":
      return TimeRangeName_DEPRECATE.ALL_TIME;
    default:
      return defaultToAllTime ? TimeRangeName_DEPRECATE.ALL_TIME : undefined;
  }
};

// Moved to time-grain and renamed
export function isGrainBigger(
  grain1: V1TimeGrain,
  grain2: V1TimeGrain
): boolean {
  if (grain1 === V1TimeGrain.TIME_GRAIN_UNSPECIFIED) return false;
  return getTimeGrainDurationMs(grain1) > getTimeGrainDurationMs(grain2);
}

// Moved
export function getAllowedTimeGrains(timeRangeDurationMs) {
  if (
    timeRangeDurationMs <
    2 * durationToMillis(TIME_GRAIN.TIME_GRAIN_HOUR.duration)
  ) {
    return [V1TimeGrain.TIME_GRAIN_MINUTE];
  } else if (
    timeRangeDurationMs <
    6 * durationToMillis(TIME_GRAIN.TIME_GRAIN_HOUR.duration)
  ) {
    return [V1TimeGrain.TIME_GRAIN_MINUTE, V1TimeGrain.TIME_GRAIN_HOUR];
  } else if (
    timeRangeDurationMs < durationToMillis(TIME_GRAIN.TIME_GRAIN_DAY.duration)
  ) {
    return [V1TimeGrain.TIME_GRAIN_HOUR];
  } else if (
    timeRangeDurationMs <
    14 * durationToMillis(TIME_GRAIN.TIME_GRAIN_DAY.duration)
  ) {
    return [V1TimeGrain.TIME_GRAIN_HOUR, V1TimeGrain.TIME_GRAIN_DAY];
  } else if (
    timeRangeDurationMs <
    durationToMillis(TIME_GRAIN.TIME_GRAIN_DAY.duration) * 30
  ) {
    return [
      V1TimeGrain.TIME_GRAIN_HOUR,
      V1TimeGrain.TIME_GRAIN_DAY,
      V1TimeGrain.TIME_GRAIN_WEEK,
    ];
  } else if (
    timeRangeDurationMs <
    3 * durationToMillis(TIME_GRAIN.TIME_GRAIN_DAY.duration) * 30
  ) {
    return [V1TimeGrain.TIME_GRAIN_DAY, V1TimeGrain.TIME_GRAIN_WEEK];
  } else if (
    timeRangeDurationMs <
    3 * durationToMillis(TIME_GRAIN.TIME_GRAIN_YEAR.duration)
  ) {
    return [
      V1TimeGrain.TIME_GRAIN_DAY,
      V1TimeGrain.TIME_GRAIN_WEEK,
      V1TimeGrain.TIME_GRAIN_MONTH,
    ];
  } else {
    return [
      V1TimeGrain.TIME_GRAIN_WEEK,
      V1TimeGrain.TIME_GRAIN_MONTH,
      V1TimeGrain.TIME_GRAIN_YEAR,
    ];
  }
}

// Moved
export function getDefaultTimeGrain(start: Date, end: Date): V1TimeGrain {
  const timeRangeDurationMs = end.getTime() - start.getTime();

  if (
    timeRangeDurationMs <
    2 * durationToMillis(TIME_GRAIN.TIME_GRAIN_HOUR.duration)
  ) {
    return V1TimeGrain.TIME_GRAIN_MINUTE;
  } else if (
    timeRangeDurationMs <
    7 * durationToMillis(TIME_GRAIN.TIME_GRAIN_DAY.duration)
  ) {
    return V1TimeGrain.TIME_GRAIN_HOUR;
  } else if (
    timeRangeDurationMs <
    3 * durationToMillis(TIME_GRAIN.TIME_GRAIN_DAY.duration) * 30
  ) {
    return V1TimeGrain.TIME_GRAIN_DAY;
  } else if (
    timeRangeDurationMs <
    3 * durationToMillis(TIME_GRAIN.TIME_GRAIN_YEAR.duration)
  ) {
    return V1TimeGrain.TIME_GRAIN_WEEK;
  } else {
    return V1TimeGrain.TIME_GRAIN_MONTH;
  }
}

// Not needed
export const timeGrainStringToEnum = (timeGrain: string): V1TimeGrain => {
  switch (timeGrain) {
    case "minute":
      return V1TimeGrain.TIME_GRAIN_MINUTE;
    case "hour":
      return V1TimeGrain.TIME_GRAIN_HOUR;
    case "day":
      return V1TimeGrain.TIME_GRAIN_DAY;
    case "week":
      return V1TimeGrain.TIME_GRAIN_WEEK;
    case "month":
      return V1TimeGrain.TIME_GRAIN_MONTH;
    case "year":
      return V1TimeGrain.TIME_GRAIN_YEAR;
    default:
      return V1TimeGrain.TIME_GRAIN_UNSPECIFIED;
  }
};

// Not needed
export const timeGrainEnumToYamlString = (timeGrain: V1TimeGrain): string => {
  if (!timeGrain) return "";
  switch (timeGrain) {
    case V1TimeGrain.TIME_GRAIN_MINUTE:
      return "minute";
    case V1TimeGrain.TIME_GRAIN_HOUR:
      return "hour";
    case V1TimeGrain.TIME_GRAIN_DAY:
      return "day";
    case V1TimeGrain.TIME_GRAIN_WEEK:
      return "week";
    case V1TimeGrain.TIME_GRAIN_MONTH:
      return "month";
    case V1TimeGrain.TIME_GRAIN_YEAR:
      return "year";
    default:
      return timeGrain;
  }
};

// This is the wrong way to deal with this. We should be (1) calculating the time range first
// then (2) getting the exact duration.
const getLastXTimeRangeDurationMs = (name: TimeRangeName_DEPRECATE): number => {
  switch (name) {
    case TimeRangeName_DEPRECATE.LAST_SIX_HOURS:
      return durationToMillis(TIME_GRAIN.TIME_GRAIN_HOUR.duration) * 6;
    case TimeRangeName_DEPRECATE.LAST_24_HOURS:
      return durationToMillis(TIME_GRAIN.TIME_GRAIN_DAY.duration);
    case TimeRangeName_DEPRECATE.LAST_7_DAYS:
      return durationToMillis(TIME_GRAIN.TIME_GRAIN_DAY.duration) * 7;
    case TimeRangeName_DEPRECATE.LAST_4_WEEKS:
      return durationToMillis(TIME_GRAIN.TIME_GRAIN_DAY.duration) * 28;

    default:
      throw new Error(`Unknown last X time range name: ${name}`);
  }
};

// map from time grain to duration in ms.
const getTimeGrainDurationMs = (timeGrain: V1TimeGrain): number => {
  switch (timeGrain) {
    case V1TimeGrain.TIME_GRAIN_MINUTE:
      return durationToMillis(TIME_GRAIN.TIME_GRAIN_MINUTE.duration);
    case V1TimeGrain.TIME_GRAIN_HOUR:
      return durationToMillis(TIME_GRAIN.TIME_GRAIN_HOUR.duration);
    case V1TimeGrain.TIME_GRAIN_DAY:
      return durationToMillis(TIME_GRAIN.TIME_GRAIN_DAY.duration);
    case V1TimeGrain.TIME_GRAIN_WEEK:
      return durationToMillis(TIME_GRAIN.TIME_GRAIN_DAY.duration) * 7;
    case V1TimeGrain.TIME_GRAIN_MONTH:
      return durationToMillis(TIME_GRAIN.TIME_GRAIN_DAY.duration) * 30;
    case V1TimeGrain.TIME_GRAIN_YEAR:
      return durationToMillis(TIME_GRAIN.TIME_GRAIN_YEAR.duration);
    default:
      throw new Error(`Unknown time grain: ${timeGrain}`);
  }
};

// might not need it
export function makeRelativeTimeRange(
  timeRangeName: TimeRangeName_DEPRECATE,
  allTimeRange: TimeRange
): TimeRange {
  if (timeRangeName === TimeRangeName_DEPRECATE.ALL_TIME) return allTimeRange;
  const startTime = new Date(
    allTimeRange.end.getTime() - getLastXTimeRangeDurationMs(timeRangeName)
  );
  return {
    name: timeRangeName,
    start: startTime,
    end: allTimeRange.end,
  };
}
