/** NOTE:
 *
 * this file should be deprecated in favor of the other time utils.
 *
 * */
import type { TimeRange } from "@rilldata/web-common/lib/time/types";
import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import { TimeRangeName_DEPRECATE } from "./time-control-types";

import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";
import { durationToMillis } from "@rilldata/web-common/lib/time/grains";

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

// Moved to time range and renamed to isTimeRangeValidForMinTimeGrain
export function isTimeRangeValidForTimeGrain(
  minTimeGrain: V1TimeGrain,
  timeRange: TimeRangeName_DEPRECATE,
): boolean {
  const timeGrainEnums = Object.values(TIME_GRAIN).map(
    (timeGrain) => timeGrain.grain,
  );
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

// Moved to time-grain and renamed
export function isGrainBigger(
  grain1: V1TimeGrain,
  grain2: V1TimeGrain,
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
  allTimeRange: TimeRange,
): TimeRange {
  if (timeRangeName === TimeRangeName_DEPRECATE.ALL_TIME) return allTimeRange;
  const startTime = new Date(
    allTimeRange.end.getTime() - getLastXTimeRangeDurationMs(timeRangeName),
  );
  return {
    name: timeRangeName,
    start: startTime,
    end: allTimeRange.end,
  };
}
