import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import {
  lastXTimeRangeNames,
  TimeGrain,
  TimeRange,
  TimeRangeName,
  TimeSeriesTimeRange,
} from "./time-control-types";

export function getRelativeTimeRangeOptions(
  allTimeRange: TimeRange
): TimeRange[] {
  const allTimeRangeDurationMs = getAllTimeRangeDurationMs(allTimeRange);
  const timeRanges: TimeRange[] = [];

  for (const timeRangeName of lastXTimeRangeNames) {
    const timeRangeDurationMs = getLastXTimeRangeDurationMs(timeRangeName);

    // only show a time range if it is within the time range of the data
    const showTimeRange = timeRangeDurationMs <= allTimeRangeDurationMs;
    if (showTimeRange) {
      const timeRange = makeRelativeTimeRange(timeRangeName, allTimeRange.end);
      timeRanges.push(timeRange);
    }
  }

  return timeRanges;
}

export function getDefaultTimeRange(allTimeRange: TimeRange): TimeRange {
  // Use AllTime for now. When we go to production real-time datasets, we'll want to change this.
  return allTimeRange;
}

export interface TimeGrainOption {
  timeGrain: TimeGrain;
  enabled: boolean;
}

export function getTimeGrainOptions(start: Date, end: Date): TimeGrainOption[] {
  const timeRangeDurationMs = end.getTime() - start.getTime();

  const timeGrains: TimeGrainOption[] = [];
  for (const timeGrain in TimeGrain) {
    // only show a time grain if it results in a reasonable number of points on the line chart
    const MINIMUM_POINTS_ON_LINE_CHART = 2;
    const MAXIMUM_POINTS_ON_LINE_CHART = 2500;
    const timeGrainDurationMs = getTimeGrainDurationMs(TimeGrain[timeGrain]);
    const pointsOnLineChart = timeRangeDurationMs / timeGrainDurationMs;
    const showTimeGrain =
      pointsOnLineChart >= MINIMUM_POINTS_ON_LINE_CHART &&
      pointsOnLineChart <= MAXIMUM_POINTS_ON_LINE_CHART;
    timeGrains.push({
      timeGrain: TimeGrain[timeGrain],
      enabled: showTimeGrain,
    });
  }
  return timeGrains;
}

export function getDefaultTimeGrain(start: Date, end: Date): TimeGrain {
  const timeRangeDurationMs = end.getTime() - start.getTime();
  if (timeRangeDurationMs <= 2 * 60 * 60 * 1000) {
    return TimeGrain.OneMinute;
  } else if (timeRangeDurationMs <= 14 * 24 * 60 * 60 * 1000) {
    return TimeGrain.OneHour;
  } else if (timeRangeDurationMs <= 60 * 24 * 60 * 60 * 1000) {
    return TimeGrain.OneDay;
  } else if (timeRangeDurationMs <= 365 * 24 * 60 * 60 * 1000) {
    return TimeGrain.OneWeek;
  } else if (timeRangeDurationMs <= 20 * 365 * 24 * 60 * 60 * 1000) {
    return TimeGrain.OneMonth;
  } else {
    return TimeGrain.OneYear;
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
  interval: string, // DuckDB interval
  date: string
): string => {
  if (!interval || !date) return "";
  switch (interval) {
    case "minute":
      return new Date(date).toLocaleDateString(undefined, {
        year: "numeric",
        month: "short",
        day: "numeric",
        hour: "numeric",
        minute: "numeric",
      });
    case "hour":
      return new Date(date).toLocaleDateString(undefined, {
        year: "numeric",
        month: "short",
        day: "numeric",
        hour: "numeric",
      });
    case "day":
      return new Date(date).toLocaleDateString(undefined, {
        year: "numeric",
        month: "short",
        day: "numeric",
      });
    case "week":
      return new Date(date).toLocaleDateString(undefined, {
        year: "numeric",
        month: "short",
        day: "numeric",
      });
    case "month":
      return new Date(date).toLocaleDateString(undefined, {
        year: "numeric",
        month: "short",
      });
    case "year":
      return new Date(date).toLocaleDateString(undefined, {
        year: "numeric",
      });
    default:
      throw new Error(`Unknown interval: ${interval}`);
  }
};

export const prettyTimeGrain = (timeGrain: TimeGrain): string => {
  if (!timeGrain) return "";
  switch (timeGrain) {
    case TimeGrain.OneMinute:
      return "minute";
    // case TimeGrain.FiveMinutes:
    //   return "5 minute";
    // case TimeGrain.FifteenMinutes:
    //   return "15 minute";
    case TimeGrain.OneHour:
      return "hourly";
    case TimeGrain.OneDay:
      return "daily";
    case TimeGrain.OneWeek:
      return "weekly";
    case TimeGrain.OneMonth:
      return "monthly";
    case TimeGrain.OneYear:
      return "yearly";
    default:
      throw new Error(`Unknown time grain: ${timeGrain}`);
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
    case TimeRangeName.LastHour:
      return 60 * 60 * 1000;
    case TimeRangeName.Last6Hours:
      return 6 * 60 * 60 * 1000;
    case TimeRangeName.LastDay:
      return 24 * 60 * 60 * 1000;
    case TimeRangeName.Last2Days:
      return 2 * 24 * 60 * 60 * 1000;
    case TimeRangeName.Last5Days:
      return 5 * 24 * 60 * 60 * 1000;
    case TimeRangeName.LastWeek:
      return 7 * 24 * 60 * 60 * 1000;
    case TimeRangeName.Last2Weeks:
      return 2 * 7 * 24 * 60 * 60 * 1000;
    case TimeRangeName.Last30Days:
      return 30 * 24 * 60 * 60 * 1000;
    case TimeRangeName.Last60Days:
      return 60 * 24 * 60 * 60 * 1000;
    default:
      throw new Error(`Unknown last X time range name: ${name}`);
  }
};

const getTimeGrainDurationMs = (timeGrain: TimeGrain): number => {
  switch (timeGrain) {
    case TimeGrain.OneMinute:
      return 60 * 1000;
    // case TimeGrain.FiveMinutes:
    //   return 5 * 60 * 1000;
    // case TimeGrain.FifteenMinutes:
    //   return 15 * 60 * 1000;
    case TimeGrain.OneHour:
      return 60 * 60 * 1000;
    case TimeGrain.OneDay:
      return 24 * 60 * 60 * 1000;
    case TimeGrain.OneWeek:
      return 7 * 24 * 60 * 60 * 1000;
    case TimeGrain.OneMonth:
      return 30 * 24 * 60 * 60 * 1000;
    case TimeGrain.OneYear:
      return 365 * 24 * 60 * 60 * 1000;
    default:
      throw new Error(`Unknown time grain: ${timeGrain}`);
  }
};

export const toV1TimeGrain = (timeGrain: TimeGrain): V1TimeGrain => {
  switch (timeGrain) {
    case TimeGrain.OneMinute:
      return V1TimeGrain.TIME_GRAIN_MINUTE;
    case TimeGrain.OneHour:
      return V1TimeGrain.TIME_GRAIN_HOUR;
    case TimeGrain.OneDay:
      return V1TimeGrain.TIME_GRAIN_DAY;
    case TimeGrain.OneWeek:
      return V1TimeGrain.TIME_GRAIN_WEEK;
    case TimeGrain.OneMonth:
      return V1TimeGrain.TIME_GRAIN_MONTH;
    case TimeGrain.OneYear:
      return V1TimeGrain.TIME_GRAIN_YEAR;
    default:
      throw new Error(`Unknown time grain: ${timeGrain}`);
  }
};

export const floorDate = (
  date: Date | undefined,
  timeGrain: TimeGrain
): Date => {
  if (!date) return new Date();
  switch (timeGrain) {
    case TimeGrain.OneMinute: {
      const interval = 60 * 1000;
      return new Date(Math.floor(date.getTime() / interval) * interval);
    }
    // case TimeGrain.FiveMinutes: {
    //   const interval = 5 * 60 * 1000;
    //   return new Date(Math.round(date.getTime() / interval) * interval);
    // }
    // case TimeGrain.FifteenMinutes: {
    //   const interval = 15 * 60 * 1000;
    //   return new Date(Math.floor(date.getTime() / interval) * interval);
    // }
    case TimeGrain.OneHour: {
      const interval = 60 * 60 * 1000;
      return new Date(Math.floor(date.getTime() / interval) * interval);
    }
    case TimeGrain.OneDay: {
      const interval = 24 * 60 * 60 * 1000;
      return new Date(Math.floor(date.getTime() / interval) * interval);
    }
    case TimeGrain.OneWeek: {
      // rounds to the most recent Monday
      const day = date.getUTCDay();
      const dateRoundedDownByDay = floorDate(date, TimeGrain.OneDay);
      const timeFromMonday = (day === 0 ? 6 : day - 1) * 24 * 60 * 60 * 1000;
      return new Date(dateRoundedDownByDay.getTime() - timeFromMonday);
    }
    case TimeGrain.OneMonth: {
      // rounds to the 1st of the current month
      return new Date(
        Date.UTC(date.getUTCFullYear(), date.getUTCMonth(), 1, 0, 0, 0, 0)
      );
    }
    case TimeGrain.OneYear: {
      // rounds to January 1st of the current year
      return new Date(Date.UTC(date.getUTCFullYear(), 1, 1));
    }
    default:
      throw new Error(`Unknown time grain: ${timeGrain}`);
  }
};

export function ceilDate(date: Date, timeGrain: TimeGrain): Date {
  const floor = floorDate(date, timeGrain);
  return addGrains(floor, 1, timeGrain);
}

export function addGrains(date: Date, units: number, grain: TimeGrain): Date {
  switch (grain) {
    case TimeGrain.OneMinute:
      return new Date(date.getTime() + units * 60 * 1000);
    case TimeGrain.OneHour:
      return new Date(date.getTime() + units * 60 * 60 * 1000);
    case TimeGrain.OneDay:
      return new Date(date.getTime() + units * 24 * 60 * 60 * 1000);
    case TimeGrain.OneWeek:
      return new Date(date.getTime() + units * 7 * 24 * 60 * 60 * 1000);
    case TimeGrain.OneMonth:
      return new Date(
        Date.UTC(date.getUTCFullYear(), date.getUTCMonth() + units, 1)
      );
    case TimeGrain.OneYear:
      return new Date(Date.UTC(date.getUTCFullYear() + units, 1, 1));
    default:
      throw new Error(`Unknown time grain: ${grain}`);
  }
}

export function checkValidTimeGrain(
  timeGrain: TimeGrain,
  timeGrainOptions: TimeGrainOption[]
): boolean {
  const timeGrainOption = timeGrainOptions.find(
    (timeGrainOption) => timeGrainOption.timeGrain === timeGrain
  );
  return timeGrainOption?.enabled;
}

export function makeRelativeTimeRange(
  timeRangeName: TimeRangeName,
  endTime: Date
): TimeRange {
  const startTime = new Date(
    endTime.getTime() - getLastXTimeRangeDurationMs(timeRangeName)
  );
  return {
    name: timeRangeName,
    start: startTime,
    end: endTime,
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
