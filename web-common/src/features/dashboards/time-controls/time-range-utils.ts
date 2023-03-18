/** NOTE:
 *
 * this file should be deprecated in favor of the other time utils.
 *
 * */
import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import {
  lastXTimeRangeNames,
  TimeRange,
  TimeRangeName,
  TimeSeriesTimeRange,
} from "./time-control-types";

// Moved to time-types
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

//TODO: Co locate with application
// export function validateTimeRange(
//   start: Date,
//   end: Date,
//   minTimeGrain: V1TimeGrain
// ): string {
//   const timeRangeDurationMs = end.getTime() - start.getTime();

//   const allowedTimeGrains = getAllowedTimeGrains(timeRangeDurationMs);
//   const allowedMaxGrain = allowedTimeGrains[allowedTimeGrains.length - 1];

//   const isGrainPossible = !isGrainBigger(minTimeGrain, allowedMaxGrain);

//   if (start > end) {
//     return "Start date must be before end date";
//   } else if (!isGrainPossible) {
//     return "Range is smaller than min time grain";
//   } else {
//     return undefined;
//   }
// }

// NOTE: we will need to keep this for the duration amounts in the runtime / config.
// let's plan to deprecate it later.
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
    name: TimeRangeName.ALL_TIME,
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

// Moved to time-types
export interface TimeGrainOption {
  timeGrain: V1TimeGrain;
  enabled: boolean;
}

// Moved to time range and renamed to isTimeRangeValidForMinTimeGrain
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

// Moved to time-grain
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

// Remove
export const timeRangeToISODuration = (
  timeRangeName: TimeRangeName
): string => {
  switch (timeRangeName) {
    case TimeRangeName.LAST_SIX_HOURS:
      return "PT6H";
    case TimeRangeName.LAST_24_HOURS:
      return "P1D";
    case TimeRangeName.LAST_7_DAYS:
      return "P7D";
    case TimeRangeName.LAST_4_WEEKS:
      return "P4W";
    case TimeRangeName.ALL_TIME:
      return "inf";
    default:
      return undefined;
  }
};

// Remove
export const ISODurationToTimeRange = (
  isoDuration: string,
  defaultToAllTime = true
): TimeRangeName => {
  switch (isoDuration) {
    case "PT6H":
      return TimeRangeName.LAST_SIX_HOURS;
    case "P1D":
      return TimeRangeName.LAST_24_HOURS;
    case "P7D":
      return TimeRangeName.LAST_7_DAYS;
    case "P4W":
      return TimeRangeName.LAST_4_WEEKS;
    case "inf":
      return TimeRangeName.ALL_TIME;
    default:
      return defaultToAllTime ? TimeRangeName.ALL_TIME : undefined;
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

// Moved
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

// Moved to time-range
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

// Moved to time-grain and renamed to formatDateByGrain
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

// Not needed
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

// Not needed
function getAllTimeRangeDurationMs(allTimeRange: TimeRange): number {
  return (
    new Date(allTimeRange.end).getTime() -
    new Date(allTimeRange.start).getTime()
  );
}

// Not needed
const getLastXTimeRangeDurationMs = (name: TimeRangeName): number => {
  switch (name) {
    case TimeRangeName.LAST_SIX_HOURS:
      return 6 * TIME.HOUR;
    case TimeRangeName.LAST_24_HOURS:
      return TIME.DAY;
    case TimeRangeName.LAST_7_DAYS:
      return TIME.WEEK;
    case TimeRangeName.LAST_4_WEEKS:
      return TIME.DAY * 28;

    default:
      throw new Error(`Unknown last X time range name: ${name}`);
  }
};

// Not needed
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
// moved to time-grain
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

// might not need it
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

// Do we need it?
export function exclusiveToInclusiveEndISOString(exclusiveEnd: string): string {
  const date = new Date(exclusiveEnd);
  date.setDate(date.getDate() - 1);
  return date.toISOString();
}

// moved to time-range
export function getDateFromISOString(isoString: string): string {
  return isoString.split("T")[0];
}

// moved to time-range
export function getISOStringFromDate(date: string): string {
  return date + "T00:00:00.000Z";
}
