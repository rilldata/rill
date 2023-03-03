import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import {
  ComparisonRange,
  lastXTimeRangeNames,
  TimeRange,
  TimeRangeName,
  TimeSeriesTimeRange,
} from "./time-control-types";

const TIME = {
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

export function validateTimeRange(
  start: Date,
  end: Date,
  minTimeGrain: V1TimeGrain
): string {
  const timeRangeDurationMs = end.getTime() - start.getTime();

  const allowedTimeGrains = getAllowedTimeGrains(timeRangeDurationMs);
  const allowedMaxGrain = allowedTimeGrains[allowedTimeGrains.length - 1];

  const isGrainPossible = !isGrainBigger(minTimeGrain, allowedMaxGrain);

  if (start > end) {
    return "Start date must be before end date";
  } else if (!isGrainPossible) {
    return "Range is smaller than min time grain";
  } else {
    return undefined;
  }
}

export function getRelativeTimeRangeOptions(
  allTimeRange: TimeRange,
  minTimeGrain: V1TimeGrain
): TimeRange[] {
  const allTimeRangeDurationMs = getAllTimeRangeDurationMs(allTimeRange);
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
    name: TimeRangeName.AllTime,
    start: allTimeRange.start,
    end: allTimeRange.end,
  });

  return timeRanges;
}

export function getDefaultTimeRange(allTimeRange: TimeRange): TimeRange {
  // Use AllTime for now. When we go to production real-time datasets, we'll want to change this.
  return allTimeRange;
}

export interface TimeGrainOption {
  timeGrain: V1TimeGrain;
  enabled: boolean;
}

export function isTimeRangeValidForTimeGrain(
  minTimeGrain: V1TimeGrain,
  timeRange: TimeRangeName
): boolean {
  const timeGrainEnums = supportedTimeGrainEnums();
  if (!timeGrainEnums.includes(minTimeGrain)) {
    return true;
  }
  if (!timeRange || timeRange === TimeRangeName.AllTime) {
    return true;
  }

  const timeRangeDurationMs = getLastXTimeRangeDurationMs(timeRange);

  const allowedTimeGrains = getAllowedTimeGrains(timeRangeDurationMs);
  const maxAllowedTimeGrain = allowedTimeGrains[allowedTimeGrains.length - 1];
  return !isGrainBigger(minTimeGrain, maxAllowedTimeGrain);
}

export function getTimeGrainOptions(start: Date, end: Date): TimeGrainOption[] {
  const timeRangeDurationMs = end.getTime() - start.getTime();

  const timeGrains: TimeGrainOption[] = [];
  for (const timeGrain in V1TimeGrain) {
    const unsupportedTypes = [
      V1TimeGrain.TIME_GRAIN_UNSPECIFIED,
      V1TimeGrain.TIME_GRAIN_MILLISECOND,
      V1TimeGrain.TIME_GRAIN_SECOND,
    ];
    if (unsupportedTypes.includes(V1TimeGrain[timeGrain])) {
      continue;
    }
    // only show a time grain if it results in a reasonable number of points on the line chart
    const MINIMUM_POINTS_ON_LINE_CHART = 2;
    const MAXIMUM_POINTS_ON_LINE_CHART = 2500;
    const timeGrainDurationMs = getTimeGrainDurationMs(V1TimeGrain[timeGrain]);
    const pointsOnLineChart = timeRangeDurationMs / timeGrainDurationMs;
    const showTimeGrain =
      pointsOnLineChart >= MINIMUM_POINTS_ON_LINE_CHART &&
      pointsOnLineChart <= MAXIMUM_POINTS_ON_LINE_CHART;
    timeGrains.push({
      timeGrain: V1TimeGrain[timeGrain],
      enabled: showTimeGrain,
    });
  }
  return timeGrains;
}

export const timeRangeToISODuration = (
  timeRangeName: TimeRangeName
): string => {
  switch (timeRangeName) {
    case TimeRangeName.Last6Hours:
      return "PT6H";
    case TimeRangeName.LastDay:
      return "P1D";
    case TimeRangeName.LastWeek:
      return "P7D";
    case TimeRangeName.Last30Days:
      return "P30D";
    case TimeRangeName.AllTime:
      return "inf";
    default:
      return undefined;
  }
};

export const ISODurationToTimeRange = (
  isoDuration: string,
  defaultToAllTime = true
): TimeRangeName => {
  switch (isoDuration) {
    case "PT6H":
      return TimeRangeName.Last6Hours;
    case "P1D":
      return TimeRangeName.LastDay;
    case "P7D":
      return TimeRangeName.LastWeek;
    case "P30D":
      return TimeRangeName.Last30Days;
    case "inf":
      return TimeRangeName.AllTime;
    default:
      return defaultToAllTime ? TimeRangeName.AllTime : undefined;
  }
};

export function isGrainBigger(
  grain1: V1TimeGrain,
  grain2: V1TimeGrain
): boolean {
  if (grain1 === V1TimeGrain.TIME_GRAIN_UNSPECIFIED) return false;
  return getTimeGrainDurationMs(grain1) > getTimeGrainDurationMs(grain2);
}

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

export const prettyFormatTimeRange = (
  timeRange: TimeSeriesTimeRange
): string => {
  if (!timeRange?.start && timeRange?.end) {
    return `- ${timeRange.end}`;
  }

  if (timeRange?.start && !timeRange?.end) {
    return `${timeRange.start} -`;
  }

  if (!timeRange?.start && !timeRange?.end) {
    return "";
  }

  const start = new Date(timeRange.start);
  // timeRange.end is exclusive. We subtract one ms to render the last inclusive value.
  const end = new Date(new Date(timeRange.end).getTime() - 1);

  const TIMEZONE = "UTC";
  // const TIMEZONE = Intl.DateTimeFormat().resolvedOptions().timeZone; // the user's local timezone

  const startDate = start.getUTCDate(); // use start.getDate() for local timezone
  const startMonth = start.getUTCMonth();
  const startYear = start.getUTCFullYear();
  const endDate = end.getUTCDate();
  const endMonth = end.getUTCMonth();
  const endYear = end.getUTCFullYear();

  // day is the same
  if (
    startDate === endDate &&
    startMonth === endMonth &&
    startYear === endYear
  ) {
    return `${start.toLocaleDateString(undefined, {
      month: "long",
      timeZone: TIMEZONE,
    })} ${startDate}, ${startYear} (${start
      .toLocaleString(undefined, {
        hour12: true,
        hour: "numeric",
        minute: "numeric",
        timeZone: TIMEZONE,
      })
      .replace(/\s/g, "")}-${end
      .toLocaleString(undefined, {
        hour12: true,
        hour: "numeric",
        minute: "numeric",
        timeZone: TIMEZONE,
      })
      .replace(/\s/g, "")})`;
  }

  // month is the same
  if (startMonth === endMonth && startYear === endYear) {
    return `${start.toLocaleDateString(undefined, {
      month: "long",
      timeZone: TIMEZONE,
    })} ${startDate}-${endDate}, ${startYear} (${start
      .toLocaleString(undefined, {
        hour12: true,
        hour: "numeric",
        minute: "numeric",
        timeZone: TIMEZONE,
      })
      .replace(/\s/g, "")}-${end
      .toLocaleString(undefined, {
        hour12: true,
        hour: "numeric",
        minute: "numeric",
        timeZone: TIMEZONE,
      })
      .replace(/\s/g, "")})`;
  }
  // year is the same
  if (startYear === endYear) {
    return `${start.toLocaleDateString(undefined, {
      month: "long",
      day: "numeric",
      timeZone: TIMEZONE,
    })} - ${end.toLocaleDateString(undefined, {
      month: "long",
      day: "numeric",
      timeZone: TIMEZONE,
    })}, ${startYear}`;
  }
  // year is different
  const dateFormatOptions: Intl.DateTimeFormatOptions = {
    year: "numeric",
    month: "long",
    day: "numeric",
    timeZone: TIMEZONE,
  };
  return `${start.toLocaleDateString(
    undefined,
    dateFormatOptions
  )} - ${end.toLocaleDateString(undefined, dateFormatOptions)}`;
};

export const formatDateByInterval = (
  interval: V1TimeGrain, // DuckDB interval
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

export const prettyTimeGrain = (timeGrain: V1TimeGrain): string => {
  if (!timeGrain) return "";
  switch (timeGrain) {
    case V1TimeGrain.TIME_GRAIN_MINUTE:
      return "minute";
    case V1TimeGrain.TIME_GRAIN_HOUR:
      return "hourly";
    case V1TimeGrain.TIME_GRAIN_DAY:
      return "daily";
    case V1TimeGrain.TIME_GRAIN_WEEK:
      return "weekly";
    case V1TimeGrain.TIME_GRAIN_MONTH:
      return "monthly";
    case V1TimeGrain.TIME_GRAIN_YEAR:
      return "yearly";
    default:
      return timeGrain;
  }
};

function getAllTimeRangeDurationMs(allTimeRange: TimeRange): number {
  return (
    new Date(allTimeRange.end).getTime() -
    new Date(allTimeRange.start).getTime()
  );
}

const getLastXTimeRangeDurationMs = (name: TimeRangeName): number => {
  switch (name) {
    case TimeRangeName.Last6Hours:
      return 6 * TIME.HOUR;
    case TimeRangeName.LastDay:
      return TIME.DAY;
    case TimeRangeName.LastWeek:
      return TIME.WEEK;
    case TimeRangeName.Last30Days:
      return TIME.MONTH;

    default:
      throw new Error(`Unknown last X time range name: ${name}`);
  }
};

const getTimeGrainDurationMs = (timeGrain: V1TimeGrain): number => {
  switch (timeGrain) {
    case V1TimeGrain.TIME_GRAIN_MINUTE:
      return TIME.MINUTE;
    case V1TimeGrain.TIME_GRAIN_HOUR:
      return TIME.HOUR;
    case V1TimeGrain.TIME_GRAIN_DAY:
      return TIME.DAY;
    case V1TimeGrain.TIME_GRAIN_WEEK:
      return TIME.WEEK;
    case V1TimeGrain.TIME_GRAIN_MONTH:
      return TIME.MONTH;
    case V1TimeGrain.TIME_GRAIN_YEAR:
      return TIME.YEAR;
    default:
      throw new Error(`Unknown time grain: ${timeGrain}`);
  }
};

export const floorDate = (
  date: Date | undefined,
  timeGrain: V1TimeGrain
): Date => {
  if (!date) return new Date();
  switch (timeGrain) {
    case V1TimeGrain.TIME_GRAIN_MINUTE: {
      const interval = TIME.MINUTE;
      return new Date(Math.floor(date.getTime() / interval) * interval);
    }
    case V1TimeGrain.TIME_GRAIN_HOUR: {
      const interval = TIME.HOUR;
      return new Date(Math.floor(date.getTime() / interval) * interval);
    }
    case V1TimeGrain.TIME_GRAIN_DAY: {
      const interval = TIME.DAY;
      return new Date(Math.floor(date.getTime() / interval) * interval);
    }
    case V1TimeGrain.TIME_GRAIN_WEEK: {
      // rounds to the most recent Monday
      const day = date.getUTCDay();
      const dateRoundedDownByDay = floorDate(date, V1TimeGrain.TIME_GRAIN_DAY);
      const timeFromMonday = (day === 0 ? 6 : day - 1) * TIME.DAY;
      return new Date(dateRoundedDownByDay.getTime() - timeFromMonday);
    }
    case V1TimeGrain.TIME_GRAIN_MONTH: {
      // rounds to the 1st of the current month
      return new Date(
        Date.UTC(date.getUTCFullYear(), date.getUTCMonth(), 1, 0, 0, 0, 0)
      );
    }
    case V1TimeGrain.TIME_GRAIN_YEAR: {
      // rounds to January 1st of the current year
      return new Date(Date.UTC(date.getUTCFullYear(), 1, 1));
    }
    default:
      throw new Error(`Unknown time grain: ${timeGrain}`);
  }
};

export function ceilDate(date: Date, timeGrain: V1TimeGrain): Date {
  const floor = floorDate(date, timeGrain);
  return addGrains(floor, 1, timeGrain);
}

export function addGrains(date: Date, units: number, grain: V1TimeGrain): Date {
  switch (grain) {
    case V1TimeGrain.TIME_GRAIN_MINUTE:
      return new Date(date.getTime() + units * TIME.MINUTE);
    case V1TimeGrain.TIME_GRAIN_HOUR:
      return new Date(date.getTime() + units * TIME.HOUR);
    case V1TimeGrain.TIME_GRAIN_DAY:
      return new Date(date.getTime() + units * TIME.DAY);
    case V1TimeGrain.TIME_GRAIN_WEEK:
      return new Date(date.getTime() + units * TIME.WEEK);
    case V1TimeGrain.TIME_GRAIN_MONTH:
      return new Date(
        Date.UTC(date.getUTCFullYear(), date.getUTCMonth() + units, 1)
      );
    case V1TimeGrain.TIME_GRAIN_YEAR:
      return new Date(Date.UTC(date.getUTCFullYear() + units, 1, 1));
    default:
      throw new Error(`Unknown time grain: ${grain}`);
  }
}

export function checkValidTimeGrain(
  timeGrain: V1TimeGrain,
  timeGrainOptions: TimeGrainOption[],
  minTimeGrain: V1TimeGrain
): boolean {
  const timeGrainOption = timeGrainOptions.find(
    (timeGrainOption) => timeGrainOption.timeGrain === timeGrain
  );
  const isGrainPossible = !isGrainBigger(minTimeGrain, timeGrain);
  return timeGrainOption?.enabled && isGrainPossible;
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

export function exclusiveToInclusiveEndISOString(exclusiveEnd: string): string {
  const date = new Date(exclusiveEnd);
  date.setDate(date.getDate() - 1);
  return date.toISOString();
}

export function getDateFromISOString(isoString: string): string {
  return isoString.split("T")[0];
}

export function getISOStringFromDate(date: string): string {
  return date + "T00:00:00.000Z";
}

// Comparison Time Utils

const pointInTimeComparisons = [
  ComparisonRange.DayOverDay,
  ComparisonRange.WeekOverWeek,
  ComparisonRange.MonthOverMonth,
  ComparisonRange.YearOverYear,
];

export function getComparisonOptionsForTimeRange(
  timeRange: TimeSeriesTimeRange
): ComparisonRange[] {
  const alwaysAllowed = [...pointInTimeComparisons, ComparisonRange.Custom];

  switch (timeRange.name) {
    case TimeRangeName.Last6Hours:
      return [ComparisonRange.Previous6Hours, ...alwaysAllowed];
    case TimeRangeName.LastDay:
      return [ComparisonRange.PreviousDay, ...alwaysAllowed];
    case TimeRangeName.LastWeek:
      return [ComparisonRange.PreviousWeek, ...alwaysAllowed];
    case TimeRangeName.Last30Days:
      return [ComparisonRange.Previous30Days, ...alwaysAllowed];
    case TimeRangeName.Custom:
      return alwaysAllowed;
    case TimeRangeName.AllTime:
      return [];
    default:
      throw new Error(`Unknown time range: ${timeRange.name}`);
  }
}

function getPointInTimeComparisonDurations(
  comparisonRange: ComparisonRange
): number {
  switch (comparisonRange) {
    case ComparisonRange.DayOverDay:
      return TIME.DAY;
    case ComparisonRange.WeekOverWeek:
      return TIME.WEEK;
    case ComparisonRange.MonthOverMonth:
      return TIME.MONTH;
    case ComparisonRange.YearOverYear:
      return TIME.YEAR;
    default:
      throw new Error(`Unknown comparison range: ${comparisonRange}`);
  }
}

export function getComparisonTimeRange(
  timeRange: TimeSeriesTimeRange,
  comparisonRange: ComparisonRange
): TimeSeriesTimeRange {
  const currentStartDate = new Date(timeRange.start).getTime();
  const currentEndDate = new Date(timeRange.end).getTime();

  // TODO:  Work on All time and custom comparison later
  if (
    timeRange.name === TimeRangeName.Custom ||
    timeRange.name === TimeRangeName.AllTime
  ) {
    return timeRange;
  }

  let startDate;
  let endDate;

  // Handle Point in Time comparisons
  if (pointInTimeComparisons.includes(comparisonRange)) {
    startDate =
      currentStartDate - getPointInTimeComparisonDurations(comparisonRange);
    endDate =
      currentEndDate - getPointInTimeComparisonDurations(comparisonRange);
  }
  // Handle Previous X comparisons
  else {
    startDate = currentStartDate - getLastXTimeRangeDurationMs(timeRange.name);
    endDate = currentEndDate - getLastXTimeRangeDurationMs(timeRange.name);
  }

  return {
    name: timeRange.name,
    start: new Date(startDate).toISOString(),
    end: new Date(endDate).toISOString(),
  };
}
