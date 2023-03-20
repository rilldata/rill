/** Utility functions for using time grains within a Rill dashboard.
 * Most of these functions utilize the TIME_GRAIN object defined in config.ts
 * to generate either a subset of time grains or a single time grain.
 */

import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import { Duration } from "luxon";
import { getTimeWidth } from "./anchors";
import { TIME_GRAIN } from "./config";
import type { TimeGrain, TimeGrainOption } from "./types";

export function unitToTimeGrain(unit: string): V1TimeGrain {
  return (
    Object.values(TIME_GRAIN).find((timeGrain) => timeGrain.label === unit)
      ?.grain || V1TimeGrain.TIME_GRAIN_UNSPECIFIED
  );
}

export function durationToMillis(duration: string): number {
  return Duration.fromISO(duration).toMillis();
}

// FIXME: what is the difference between this and getAllowedTimeGrains?
// It appears that we're using this instead of getAllowedTimeGrains.
// I think we should deprecate this.
export function getTimeGrainOptions(start: Date, end: Date): TimeGrainOption[] {
  const timeGrains: TimeGrainOption[] = [];
  const timeRangeDurationMs = getTimeWidth(start, end);

  for (const timeGrain of Object.values(TIME_GRAIN)) {
    // only show a time grain if it results in a reasonable number of points on the line chart
    const MINIMUM_POINTS_ON_LINE_CHART = 2;
    const MAXIMUM_POINTS_ON_LINE_CHART = 2500;
    const timeGrainDurationMs = durationToMillis(timeGrain.duration);
    const pointsOnLineChart = timeRangeDurationMs / timeGrainDurationMs;
    const showTimeGrain =
      pointsOnLineChart >= MINIMUM_POINTS_ON_LINE_CHART &&
      pointsOnLineChart <= MAXIMUM_POINTS_ON_LINE_CHART;
    timeGrains.push({
      ...timeGrain,
      enabled: showTimeGrain,
    });
  }
  return timeGrains;
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

// Return time grains that are allowed for a given time range. Note that
// this function is similar to getTimeGrainOptions. We should deprecate getTimeGrainOptions
// in favor of this logic.
export function getAllowedTimeGrains(start: Date, end: Date): TimeGrain[] {
  const timeRangeDurationMs = getTimeWidth(start, end);
  if (
    timeRangeDurationMs <
    2 * durationToMillis(TIME_GRAIN.TIME_GRAIN_HOUR.duration)
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
    14 * durationToMillis(TIME_GRAIN.TIME_GRAIN_DAY.duration)
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
    timeRangeDurationMs <
    3 * durationToMillis(TIME_GRAIN.TIME_GRAIN_DAY.duration) * 30
  ) {
    return [TIME_GRAIN.TIME_GRAIN_DAY, TIME_GRAIN.TIME_GRAIN_WEEK];
  } else if (
    timeRangeDurationMs <
    3 * durationToMillis(TIME_GRAIN.TIME_GRAIN_YEAR.duration)
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

//TODO: Simplify use of this method
export function checkValidTimeGrain(
  timeGrain: V1TimeGrain,
  timeGrainOptions: TimeGrainOption[],
  minTimeGrain: V1TimeGrain
): boolean {
  const timeGrainOption = timeGrainOptions.find(
    (timeGrainOption) => timeGrainOption.grain === timeGrain
  );

  if (minTimeGrain === V1TimeGrain.TIME_GRAIN_UNSPECIFIED)
    return timeGrainOption?.enabled;

  const isGrainPossible = !isGrainBigger(minTimeGrain, timeGrain);
  return timeGrainOption?.enabled && isGrainPossible;
}
