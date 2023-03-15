import { V1TimeGrain } from "../../../../runtime-client";
import { TIME } from "./time-types";

// Filter out time grains not used in the UI
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

// Return time grains that are allowed for a given time range
export function getAllowedTimeGrains(timeRangeDurationMs) {
  if (timeRangeDurationMs < 2 * TIME.HOUR) {
    return [V1TimeGrain.TIME_GRAIN_MINUTE];
  } else if (timeRangeDurationMs < 6 * TIME.HOUR) {
    return [V1TimeGrain.TIME_GRAIN_MINUTE, V1TimeGrain.TIME_GRAIN_HOUR];
  } else if (timeRangeDurationMs < TIME.DAY) {
    return [V1TimeGrain.TIME_GRAIN_HOUR];
  } else if (timeRangeDurationMs < 14 * TIME.DAY) {
    return [V1TimeGrain.TIME_GRAIN_HOUR, V1TimeGrain.TIME_GRAIN_DAY];
  } else if (timeRangeDurationMs < TIME.MONTH) {
    return [
      V1TimeGrain.TIME_GRAIN_HOUR,
      V1TimeGrain.TIME_GRAIN_DAY,
      V1TimeGrain.TIME_GRAIN_WEEK,
    ];
  } else if (timeRangeDurationMs < 3 * TIME.MONTH) {
    return [V1TimeGrain.TIME_GRAIN_DAY, V1TimeGrain.TIME_GRAIN_WEEK];
  } else if (timeRangeDurationMs < 3 * TIME.YEAR) {
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

// Check if grain1 is bigger than grain2
export function isGrainBigger(
  grain1: V1TimeGrain,
  grain2: V1TimeGrain
): boolean {
  if (grain1 === V1TimeGrain.TIME_GRAIN_UNSPECIFIED) return false;
  return getTimeGrainDurationMs(grain1) > getTimeGrainDurationMs(grain2);
}
