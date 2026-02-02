/** Utility functions for using time grains within a Rill dashboard.
 * Most of these functions utilize the TIME_GRAIN object defined in config.ts
 * to generate either a subset of time grains or a single time grain.
 */

import { V1TimeGrain } from "@rilldata/web-common/runtime-client/gen/index.schemas";
import { Duration } from "luxon";
import { TIME_GRAIN } from "../config";
import type { AvailableTimeGrain, TimeGrain } from "../types";
import { allowedGrainsForInterval } from "@rilldata/web-common/lib/time/new-grains";
import { getRangePrecision } from "@rilldata/web-common/lib/time/rill-time-grains";
import type { RillTime } from "@rilldata/web-common/features/dashboards/url-state/time-ranges/RillTime";

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
    timeRangeDurationMs < durationToMillis(TIME_GRAIN.TIME_GRAIN_WEEK.duration)
  ) {
    return TIME_GRAIN.TIME_GRAIN_HOUR;
  } else if (
    timeRangeDurationMs <
    durationToMillis(TIME_GRAIN.TIME_GRAIN_QUARTER.duration)
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
// This should be deprecated in favor of using allowedGrainsForInterval directly
export function getAllowedTimeGrains(start: Date, end: Date): TimeGrain[] {
  const interval = Interval.fromDateTimes(start, end);
  return allowedGrainsForInterval(interval.isValid ? interval : undefined).map(
    (g) => TIME_GRAIN[g],
  );
}

const APITimeGrainOrder: V1TimeGrain[] = [
  V1TimeGrain.TIME_GRAIN_MILLISECOND,
  V1TimeGrain.TIME_GRAIN_SECOND,
  V1TimeGrain.TIME_GRAIN_MINUTE,
  V1TimeGrain.TIME_GRAIN_HOUR,
  V1TimeGrain.TIME_GRAIN_DAY,
  V1TimeGrain.TIME_GRAIN_WEEK,
  V1TimeGrain.TIME_GRAIN_MONTH,
  V1TimeGrain.TIME_GRAIN_QUARTER,
  V1TimeGrain.TIME_GRAIN_YEAR,
];

export function isGrainBigger(
  possiblyBiggerGrain: V1TimeGrain,
  possiblySmallerGrain: V1TimeGrain,
): boolean {
  const biggerIndex = APITimeGrainOrder.indexOf(possiblyBiggerGrain);
  const smallerIndex = APITimeGrainOrder.indexOf(possiblySmallerGrain);
  return biggerIndex > smallerIndex;
}

export function getMinGrain(...grains: V1TimeGrain[]) {
  let minGrain: V1TimeGrain = V1TimeGrain.TIME_GRAIN_UNSPECIFIED;
  let minGrainIndex = APITimeGrainOrder.length;

  for (const grain of grains) {
    if (grain === V1TimeGrain.TIME_GRAIN_UNSPECIFIED) continue;
    const grainIndex = APITimeGrainOrder.indexOf(grain);
    if (grainIndex < minGrainIndex) {
      minGrain = grain;
      minGrainIndex = grainIndex;
    }
  }

  return minGrain;
}

export function mapDurationToGrain(duration: string): V1TimeGrain {
  for (const g in TIME_GRAIN) {
    if (TIME_GRAIN[g].duration === duration) {
      return TIME_GRAIN[g].grain;
    }
  }
  return V1TimeGrain.TIME_GRAIN_UNSPECIFIED;
}

export function timeGrainToDuration(timeGrain: V1TimeGrain): string {
  if (isAvailableTimeGrain(timeGrain)) {
    const grainConfig = TIME_GRAIN[timeGrain];
    return grainConfig.duration;
  } else {
    console.warn("Requested duration for invalid time grain: ", timeGrain);
    // Default to 1 day if the time grain is invalid to fail gracefully
    return "P1D";
  }
}

export function isAvailableTimeGrain(
  grain: V1TimeGrain,
): grain is AvailableTimeGrain {
  return (
    grain !== V1TimeGrain.TIME_GRAIN_UNSPECIFIED &&
    grain !== V1TimeGrain.TIME_GRAIN_MILLISECOND &&
    grain !== V1TimeGrain.TIME_GRAIN_SECOND
  );
}

const grainOrder: AvailableTimeGrain[] = [
  V1TimeGrain.TIME_GRAIN_YEAR,
  V1TimeGrain.TIME_GRAIN_QUARTER,
  V1TimeGrain.TIME_GRAIN_MONTH,
  V1TimeGrain.TIME_GRAIN_WEEK,
  V1TimeGrain.TIME_GRAIN_DAY,
  V1TimeGrain.TIME_GRAIN_HOUR,
  V1TimeGrain.TIME_GRAIN_MINUTE,
];

/**
 * Get the largest grain from available grains
 */
export function getLargestGrain(
  grains: AvailableTimeGrain[],
): AvailableTimeGrain | undefined {
  for (const grain of grainOrder) {
    if (grains.includes(grain)) {
      return grain;
    }
  }
  return grains[0];
}

/**
 * Get the next smaller grain from the given grain
 */
export function getNextSmallerGrain(
  currentGrain: AvailableTimeGrain,
  availableGrains: AvailableTimeGrain[],
): AvailableTimeGrain | undefined {
  const currentIndex = grainOrder.indexOf(currentGrain);
  if (currentIndex === -1) return availableGrains[0];

  // Look for the next smaller grain that's available
  for (let i = currentIndex + 1; i < grainOrder.length; i++) {
    if (availableGrains.includes(grainOrder[i])) {
      return grainOrder[i];
    }
  }

  // If no smaller grain found, return the smallest available
  for (let i = grainOrder.length - 1; i >= 0; i--) {
    if (availableGrains.includes(grainOrder[i])) {
      return grainOrder[i];
    }
  }

  return availableGrains[0];
}

/**
 * Validates and adjusts the time grain for a given interval based on allowed grains.
 * Returns the validated grain, or undefined if validation cannot be performed.
 */
export function getValidatedTimeGrain(
  interval: Interval | undefined,
  minTimeGrain: V1TimeGrain,
  requestedPrecision: V1TimeGrain | undefined,
  parsed: RillTime | undefined,
): V1TimeGrain | undefined {
  if (!interval || !interval.isValid) {
    return undefined;
  }

  const allowedGrains = allowedGrainsForInterval(
    interval as Interval<true>,
    minTimeGrain,
  );

  const rangePrecision = parsed && getRangePrecision(parsed);

  const finalGrain =
    requestedPrecision && allowedGrains.includes(requestedPrecision)
      ? requestedPrecision
      : rangePrecision && allowedGrains.includes(rangePrecision)
        ? rangePrecision
        : allowedGrains[0];

  return finalGrain ?? minTimeGrain;
}
