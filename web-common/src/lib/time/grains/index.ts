/** Utility functions for using time grains within a Rill dashboard.
 * Most of these functions utilize the TIME_GRAIN object defined in config.ts
 * to generate either a subset of time grains or a single time grain.
 */

import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import { Duration } from "luxon";
import { TIME_GRAIN } from "../config";
import { getTimeWidth } from "../transforms";
import type { TimeGrain } from "../types";

export function unitToTimeGrain(unit: string): V1TimeGrain {
  return (
    Object.values(TIME_GRAIN).find((timeGrain) => timeGrain.label === unit)
      ?.grain || V1TimeGrain.TIME_GRAIN_UNSPECIFIED
  );
}

export function durationToMillis(duration: string): number {
  return Duration.fromISO(duration).toMillis();
}

// Get the default grain for a given time range.
export function getDefaultTimeGrain(start: Date, end: Date): TimeGrain {
  const timeRangeDurationMs = end.getTime() - start.getTime();
  if (
    timeRangeDurationMs <
    2 * durationToMillis(TIME_GRAIN.TIME_GRAIN_HOUR.duration)
  ) {
    return TIME_GRAIN.TIME_GRAIN_MINUTE;
  } else if (
    timeRangeDurationMs <
    7 * durationToMillis(TIME_GRAIN.TIME_GRAIN_DAY.duration)
  ) {
    return TIME_GRAIN.TIME_GRAIN_HOUR;
  } else if (
    timeRangeDurationMs <
    3 * durationToMillis(TIME_GRAIN.TIME_GRAIN_DAY.duration) * 30
  ) {
    return TIME_GRAIN.TIME_GRAIN_DAY;
  } else if (
    timeRangeDurationMs <
    3 * durationToMillis(TIME_GRAIN.TIME_GRAIN_YEAR.duration)
  ) {
    return TIME_GRAIN.TIME_GRAIN_WEEK;
  } else {
    return TIME_GRAIN.TIME_GRAIN_MONTH;
  }
}

// Return time grains that are allowed for a given time range.
export function getAllowedTimeGrains(start: Date, end: Date): TimeGrain[] {
  const timeRangeDurationMs = getTimeWidth(start, end);
  if (
    timeRangeDurationMs < durationToMillis(TIME_GRAIN.TIME_GRAIN_HOUR.duration)
  ) {
    return [TIME_GRAIN.TIME_GRAIN_MINUTE];
  } else if (
    timeRangeDurationMs <
    6 * durationToMillis(TIME_GRAIN.TIME_GRAIN_HOUR.duration)
  ) {
    return [TIME_GRAIN.TIME_GRAIN_MINUTE, TIME_GRAIN.TIME_GRAIN_HOUR];
  } else if (
    timeRangeDurationMs < durationToMillis(TIME_GRAIN.TIME_GRAIN_DAY.duration)
  ) {
    return [TIME_GRAIN.TIME_GRAIN_HOUR];
  } else if (
    timeRangeDurationMs <
    7 * durationToMillis(TIME_GRAIN.TIME_GRAIN_DAY.duration)
  ) {
    return [TIME_GRAIN.TIME_GRAIN_HOUR, TIME_GRAIN.TIME_GRAIN_DAY];
  } else if (
    timeRangeDurationMs <
    durationToMillis(TIME_GRAIN.TIME_GRAIN_DAY.duration) * 30
  ) {
    return [
      TIME_GRAIN.TIME_GRAIN_HOUR,
      TIME_GRAIN.TIME_GRAIN_DAY,
      TIME_GRAIN.TIME_GRAIN_WEEK,
    ];
  } else if (
    timeRangeDurationMs < durationToMillis(TIME_GRAIN.TIME_GRAIN_YEAR.duration)
  ) {
    return [
      TIME_GRAIN.TIME_GRAIN_DAY,
      TIME_GRAIN.TIME_GRAIN_WEEK,
      TIME_GRAIN.TIME_GRAIN_MONTH,
    ];
  } else {
    return [
      TIME_GRAIN.TIME_GRAIN_WEEK,
      TIME_GRAIN.TIME_GRAIN_MONTH,
      TIME_GRAIN.TIME_GRAIN_YEAR,
    ];
  }
}

export function isGrainBigger(
  possiblyBiggerGrain: V1TimeGrain,
  possiblySmallerGrain: V1TimeGrain
): boolean {
  const biggerGrainConfig = TIME_GRAIN[possiblyBiggerGrain];
  const smallerGrainConfig = TIME_GRAIN[possiblySmallerGrain];
  return (
    durationToMillis(biggerGrainConfig?.duration) >
    durationToMillis(smallerGrainConfig.duration)
  );
}

export function checkValidTimeGrain(
  timeGrain: V1TimeGrain,
  timeGrainOptions: TimeGrain[],
  minTimeGrain: V1TimeGrain
): boolean {
  if (!timeGrainOptions.find((t) => t.grain === timeGrain)) return false;

  // If minTimeGrain is not specified, then all available timeGrains are valid
  if (minTimeGrain === V1TimeGrain.TIME_GRAIN_UNSPECIFIED) return true;

  const isGrainPossible = !isGrainBigger(minTimeGrain, timeGrain);
  return isGrainPossible;
}

export function findValidTimeGrain(
  timeGrain: V1TimeGrain,
  timeGrainOptions: TimeGrain[],
  minTimeGrain: V1TimeGrain
): V1TimeGrain {
  const timeGrains = Object.values(TIME_GRAIN).map(
    (timeGrain) => timeGrain.grain
  );

  const defaultIndex = timeGrains.indexOf(timeGrain);

  // Loop through the timeGrains starting from the default value
  for (let i = defaultIndex; i < timeGrains.length; i++) {
    const currentGrain = timeGrains[i];

    if (checkValidTimeGrain(currentGrain, timeGrainOptions, minTimeGrain)) {
      return currentGrain;
    }
  }
  // If no valid timeGrain is found, loop from the beginning of the array
  for (let i = 0; i < defaultIndex; i++) {
    const currentGrain = timeGrains[i];

    if (checkValidTimeGrain(currentGrain, timeGrainOptions, minTimeGrain)) {
      return currentGrain;
    }
  }

  // If no valid timeGrain is found, return the default timeGrain as fallback
  return timeGrain;
}
