import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import { Duration } from "luxon";
import { getTimeWidth } from "./anchors";
import { Period, TimeGrain, TimeGrainOption } from "./time-types";

export type AvailableTimeGrain = Exclude<
  V1TimeGrain,
  "TIME_GRAIN_UNSPECIFIED" | "TIME_GRAIN_MILLISECOND" | "TIME_GRAIN_SECOND"
>;

// HAMILTON: make this the only source of truth for all time grain related functionality.
export const TIME_GRAIN: Record<AvailableTimeGrain, TimeGrain> = {
  TIME_GRAIN_MINUTE: {
    grain: V1TimeGrain.TIME_GRAIN_MINUTE,
    label: "minute",
    duration: Period.MINUTE,
    width: Duration.fromISO(Period.MINUTE).toMillis(),
    formatDate: {
      year: "numeric",
      month: "short",
      day: "numeric",
      hour: "numeric",
      minute: "numeric",
    },
  },
  TIME_GRAIN_HOUR: {
    grain: V1TimeGrain.TIME_GRAIN_HOUR,
    label: "hour",
    duration: Period.HOUR,
    width: Duration.fromISO(Period.HOUR).toMillis(),
    formatDate: {
      year: "numeric",
      month: "short",
      day: "numeric",
      hour: "numeric",
    },
  },
  TIME_GRAIN_DAY: {
    grain: V1TimeGrain.TIME_GRAIN_DAY,
    label: "day",
    duration: Period.DAY,
    width: Duration.fromISO(Period.DAY).toMillis(),
    formatDate: {
      year: "numeric",
      month: "short",
      day: "numeric",
    },
  },
  TIME_GRAIN_WEEK: {
    grain: V1TimeGrain.TIME_GRAIN_WEEK,
    label: "week",
    duration: Period.WEEK,
    width: Duration.fromISO(Period.WEEK).toMillis(),
    formatDate: {
      year: "numeric",
      month: "short",
      day: "numeric",
    },
  },
  TIME_GRAIN_MONTH: {
    grain: V1TimeGrain.TIME_GRAIN_MONTH,
    label: "month",
    duration: Period.MONTH,
    // note: this will not always be accurate.
    width: Duration.fromISO("P1M").toMillis(),
    formatDate: {
      year: "numeric",
      month: "short",
    },
  },
  TIME_GRAIN_YEAR: {
    grain: V1TimeGrain.TIME_GRAIN_YEAR,
    label: "year",
    duration: Period.YEAR,
    width: Duration.fromISO("P1Y").toMillis(),
    formatDate: {
      year: "numeric",
    },
  },
};

export function durationToMillis(duration: string): number {
  return Duration.fromISO(duration).toMillis();
}

export function getTimeGrain(grain: V1TimeGrain): TimeGrain {
  return TIME_GRAIN[grain];
}

export function supportedTimeGrainEnums(): V1TimeGrain[] {
  return Object.values(TIME_GRAIN).map((timeGrain) => timeGrain.grain);
}

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

// Get the default grain for a given time range
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

// Return time grains that are allowed for a given time range
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
  const minGrain = TIME_GRAIN[possiblyBiggerGrain];
  const comparingGrain = TIME_GRAIN[possiblySmallerGrain];
  return (
    durationToMillis(minGrain?.duration) >
    durationToMillis(comparingGrain.duration)
  );
}

export function getTimeGrainFromRuntimeGrain(grain: V1TimeGrain): TimeGrain {
  for (const timeGrain of Object.values(TIME_GRAIN)) {
    if (timeGrain.grain === grain) {
      return timeGrain;
    }
  }
  // Do nothing when grain is not found
  return undefined;
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
