import { V1TimeGrain } from "../../../../runtime-client";
import { getTimeWidth } from "./time-anchors";
import { Period, TIME, TimeGrain, TimeGrainOption } from "./time-types";

export const TIME_GRAIN: Record<string, TimeGrain> = {
  MINUTE: {
    grain: V1TimeGrain.TIME_GRAIN_MINUTE,
    label: "minute",
    prettyLabel: "minute",
    duration: Period.MINUTE,
    width: TIME.MINUTE,
  },
  HOUR: {
    grain: V1TimeGrain.TIME_GRAIN_HOUR,
    label: "hour",
    prettyLabel: "hourly",
    duration: Period.HOUR,
    width: TIME.HOUR,
  },
  DAY: {
    grain: V1TimeGrain.TIME_GRAIN_DAY,
    label: "day",
    prettyLabel: "daily",
    duration: Period.DAY,
    width: TIME.DAY,
  },
  WEEK: {
    grain: V1TimeGrain.TIME_GRAIN_WEEK,
    label: "week",
    prettyLabel: "weekly",
    duration: Period.WEEK,
    width: TIME.WEEK,
  },
  MONTH: {
    grain: V1TimeGrain.TIME_GRAIN_MONTH,
    label: "month",
    prettyLabel: "monthly",
    duration: Period.MONTH,
    width: TIME.MONTH,
  },
  YEAR: {
    grain: V1TimeGrain.TIME_GRAIN_YEAR,
    label: "year",
    prettyLabel: "yearly",
    duration: Period.YEAR,
    width: TIME.YEAR,
  },
};

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
    const timeGrainDurationMs = timeGrain.width;
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

  if (timeRangeDurationMs < 2 * TIME.HOUR) {
    return TIME_GRAIN.MINUTE;
  } else if (timeRangeDurationMs < 7 * TIME.DAY) {
    return TIME_GRAIN.HOUR;
  } else if (timeRangeDurationMs < 3 * TIME.MONTH) {
    return TIME_GRAIN.DAY;
  } else if (timeRangeDurationMs < 3 * TIME.YEAR) {
    return TIME_GRAIN.WEEK;
  } else {
    return TIME_GRAIN.MONTH;
  }
}

// Return time grains that are allowed for a given time range
export function getAllowedTimeGrains(start: Date, end: Date): TimeGrain[] {
  const timeRangeDurationMs = getTimeWidth(start, end);
  if (timeRangeDurationMs < 2 * TIME.HOUR) {
    return [TIME_GRAIN.MINUTE];
  } else if (timeRangeDurationMs < 6 * TIME.HOUR) {
    return [TIME_GRAIN.MINUTE, TIME_GRAIN.HOUR];
  } else if (timeRangeDurationMs < TIME.DAY) {
    return [TIME_GRAIN.HOUR];
  } else if (timeRangeDurationMs < 14 * TIME.DAY) {
    return [TIME_GRAIN.HOUR, TIME_GRAIN.DAY];
  } else if (timeRangeDurationMs < TIME.MONTH) {
    return [TIME_GRAIN.HOUR, TIME_GRAIN.DAY, TIME_GRAIN.WEEK];
  } else if (timeRangeDurationMs < 3 * TIME.MONTH) {
    return [TIME_GRAIN.DAY, TIME_GRAIN.WEEK];
  } else if (timeRangeDurationMs < 3 * TIME.YEAR) {
    return [TIME_GRAIN.DAY, TIME_GRAIN.WEEK, TIME_GRAIN.MONTH];
  } else {
    return [TIME_GRAIN.WEEK, TIME_GRAIN.MONTH, TIME_GRAIN.YEAR];
  }
}

// Check if minTimeGrain is bigger than provided grain
export function isMinGrainBigger(
  minTimeGrain: V1TimeGrain,
  grain: TimeGrain
): boolean {
  const minGrain = getTimeGrainFromRuntimeGrain(minTimeGrain);
  return minGrain?.width > grain.width;
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
  console.log("checkValidTimeGrain", timeGrain, minTimeGrain);
  const timeGrainOption = timeGrainOptions.find(
    (timeGrainOption) => timeGrainOption.grain === timeGrain
  );

  if (minTimeGrain === V1TimeGrain.TIME_GRAIN_UNSPECIFIED)
    return timeGrainOption?.enabled;

  const timeGrainObj = getTimeGrainFromRuntimeGrain(timeGrain);
  const isGrainPossible = !isMinGrainBigger(minTimeGrain, timeGrainObj);
  return timeGrainOption?.enabled && isGrainPossible;
}

export const formatDateByGrain = (
  interval: V1TimeGrain,
  date: string
): string => {
  if (!interval || !date) return "";
  switch (interval) {
    case V1TimeGrain.TIME_GRAIN_MINUTE:
      return new Date(date).toLocaleDateString(undefined, {
        year: "numeric",
        month: "short",
        day: "numeric",
        hour: "numeric",
        minute: "numeric",
      });
    case V1TimeGrain.TIME_GRAIN_HOUR:
      return new Date(date).toLocaleDateString(undefined, {
        year: "numeric",
        month: "short",
        day: "numeric",
        hour: "numeric",
      });
    case V1TimeGrain.TIME_GRAIN_DAY:
      return new Date(date).toLocaleDateString(undefined, {
        year: "numeric",
        month: "short",
        day: "numeric",
      });
    case V1TimeGrain.TIME_GRAIN_WEEK:
      return new Date(date).toLocaleDateString(undefined, {
        year: "numeric",
        month: "short",
        day: "numeric",
      });
    case V1TimeGrain.TIME_GRAIN_MONTH:
      return new Date(date).toLocaleDateString(undefined, {
        year: "numeric",
        month: "short",
      });
    case V1TimeGrain.TIME_GRAIN_YEAR:
      return new Date(date).toLocaleDateString(undefined, {
        year: "numeric",
      });
    default:
      throw new Error(`Unknown interval: ${interval}`);
  }
};
