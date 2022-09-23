import {
  lastXTimeRanges,
  TimeGrain,
  TimeRangeName,
  TimeSeriesTimeRange,
} from "$common/database-service/DatabaseTimeSeriesActions";

// TODO: replace this with a call to the `/meta?metricsDefId={metricsDefId}` endpoint, once it's available
export const getSelectableTimeRangeNames = (
  allTimeRange: TimeSeriesTimeRange
): TimeRangeName[] => {
  if (!allTimeRange) return [];

  const allTimeRangeDuration = getTimeRangeDuration(
    TimeRangeName.AllTime,
    allTimeRange
  );

  const selectableTimeRangeNames: TimeRangeName[] = [];
  for (const timeRangeName in TimeRangeName) {
    const timeRangeDuration = getTimeRangeDuration(
      TimeRangeName[timeRangeName],
      allTimeRange
    );
    // only show a time range if it is within the time range of the data
    const showTimeRange = allTimeRangeDuration >= timeRangeDuration;
    if (showTimeRange) {
      selectableTimeRangeNames.push(TimeRangeName[timeRangeName]);
    }
  }

  return selectableTimeRangeNames;
};

// TODO: replace this with a call to the `/meta?metricsDefId={metricsDefId}` endpoint, once it's available
export const getDefaultTimeRangeName = (): TimeRangeName => {
  // Use AllTime for now. When we go to production real-time datasets, we'll want to change this.
  return TimeRangeName.AllTime;
};

export interface TimeGrainOption {
  timeGrain: TimeGrain;
  enabled: boolean;
}

// This is for pre-set relative time ranges – where the start/end dates are not yet deterimined.
// For custom time ranges, we'll need another function with "breakpoint" logic that analyzes the user-determined start/end dates.
export const getSelectableTimeGrains = (
  timeRangeName: TimeRangeName,
  allTimeRange: TimeSeriesTimeRange
): TimeGrainOption[] => {
  if (!timeRangeName || !allTimeRange) return [];
  const timeRangeDuration = getTimeRangeDuration(timeRangeName, allTimeRange);

  const timeGrains: TimeGrainOption[] = [];
  for (const timeGrain in TimeGrain) {
    // only show a time grain if it results in a reasonable number of points on the line chart
    const MINIMUM_POINTS_ON_LINE_CHART = 2;
    const MAXIMUM_POINTS_ON_LINE_CHART = 2500;
    const timeGrainDuration = getTimeGrainDuration(TimeGrain[timeGrain]);
    const pointsOnLineChart = timeRangeDuration / timeGrainDuration;
    const showTimeGrain =
      pointsOnLineChart >= MINIMUM_POINTS_ON_LINE_CHART &&
      pointsOnLineChart <= MAXIMUM_POINTS_ON_LINE_CHART;
    timeGrains.push({
      timeGrain: TimeGrain[timeGrain],
      enabled: showTimeGrain,
    });
  }
  if (timeGrains.length === 0) {
    throw new Error(`No time grains generated for time range ${timeRangeName}`);
  }
  return timeGrains;
};

export const getDefaultTimeGrain = (
  timeRangeName: TimeRangeName,
  allTimeRange: TimeSeriesTimeRange
): TimeGrain => {
  switch (timeRangeName) {
    case TimeRangeName.LastHour:
      return TimeGrain.OneMinute;
    case TimeRangeName.Last6Hours:
      return TimeGrain.OneHour;
    case TimeRangeName.LastDay:
      return TimeGrain.OneHour;
    case TimeRangeName.Last2Days:
      return TimeGrain.OneHour;
    case TimeRangeName.Last5Days:
      return TimeGrain.OneHour;
    case TimeRangeName.LastWeek:
      return TimeGrain.OneHour;
    case TimeRangeName.Last2Weeks:
      return TimeGrain.OneDay;
    case TimeRangeName.Last30Days:
      return TimeGrain.OneDay;
    case TimeRangeName.Last60Days:
      return TimeGrain.OneDay;
    case TimeRangeName.AllTime: {
      if (!allTimeRange) return TimeGrain.OneDay;
      const allTimeRangeDuration = getTimeRangeDuration(
        TimeRangeName.AllTime,
        allTimeRange
      );
      if (allTimeRangeDuration <= 2 * 60 * 60 * 1000) {
        return TimeGrain.OneMinute;
      }
      if (allTimeRangeDuration <= 14 * 24 * 60 * 60 * 1000) {
        return TimeGrain.OneHour;
      }
      if (allTimeRangeDuration <= 60 * 24 * 60 * 60 * 1000) {
        return TimeGrain.OneDay;
      }
      if (allTimeRangeDuration <= 365 * 24 * 60 * 60 * 1000) {
        return TimeGrain.OneWeek;
      }
      if (allTimeRangeDuration <= 20 * 365 * 24 * 60 * 60 * 1000) {
        return TimeGrain.OneMonth;
      }
      return TimeGrain.OneYear;
    }
    default:
      throw new Error(`No default time grain for time range ${timeRangeName}`);
  }
};

export const makeTimeRange = (
  timeRangeName: TimeRangeName,
  timeGrain: TimeGrain,
  allTimeRange: TimeSeriesTimeRange
): TimeSeriesTimeRange => {
  if (lastXTimeRanges.includes(timeRangeName)) {
    return makeLastXTimeRange(
      timeRangeName,
      new Date(allTimeRange?.end),
      timeGrain
    );
  }

  if (timeRangeName === TimeRangeName.AllTime) {
    return {
      name: timeRangeName,
      start: allTimeRange?.start,
      end: allTimeRange?.end,
      interval: timeGrain.toString(),
    };
  }

  throw new Error(`Unknown time range name: ${timeRangeName}`);
};

export const makeTimeRanges = (
  timeRangeNames: TimeRangeName[],
  allTimeRangeInDataset: TimeSeriesTimeRange
): TimeSeriesTimeRange[] => {
  if (!timeRangeNames || !allTimeRangeInDataset) return [];

  const timeRanges: TimeSeriesTimeRange[] = [];
  for (const timeRangeName of timeRangeNames) {
    const defaultTimeGrain = getDefaultTimeGrain(
      timeRangeName,
      allTimeRangeInDataset
    );
    const timeRange = makeTimeRange(
      timeRangeName,
      defaultTimeGrain,
      allTimeRangeInDataset
    );
    timeRanges.push(timeRange);
  }
  return timeRanges;
};

export const getSelectableTimeRanges = (
  allTimeRangeInDataset: TimeSeriesTimeRange
) => {
  // TODO: replace this with a call to the `/meta` endpoint, once available.
  const selectableTimeRangeNames = getSelectableTimeRangeNames(
    allTimeRangeInDataset
  );
  return makeTimeRanges(selectableTimeRangeNames, allTimeRangeInDataset);
};

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
  const end = new Date(timeRange.end);

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
    case "1 minute":
      return new Date(date).toLocaleDateString(undefined, {
        year: "numeric",
        month: "short",
        day: "numeric",
        hour: "numeric",
        minute: "numeric",
      });
    case "5 minute":
      return new Date(date).toLocaleDateString(undefined, {
        year: "numeric",
        month: "short",
        day: "numeric",
        hour: "numeric",
        minute: "numeric",
      });
    case "15 minute":
      return new Date(date).toLocaleDateString(undefined, {
        year: "numeric",
        month: "short",
        day: "numeric",
        hour: "numeric",
        minute: "numeric",
      });
    case "1 hour":
      return new Date(date).toLocaleDateString(undefined, {
        year: "numeric",
        month: "short",
        day: "numeric",
        hour: "numeric",
      });
    case "1 day":
      return new Date(date).toLocaleDateString(undefined, {
        year: "numeric",
        month: "short",
        day: "numeric",
      });
    case "7 day":
      return new Date(date).toLocaleDateString(undefined, {
        year: "numeric",
        month: "short",
        day: "numeric",
      });
    case "1 month":
      return new Date(date).toLocaleDateString(undefined, {
        year: "numeric",
        month: "short",
      });
    case "1 year":
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

const makeLastXTimeRange = (
  name: TimeRangeName,
  anchorDate: Date,
  timeGrain: TimeGrain
): TimeSeriesTimeRange => {
  const roundedUpAnchorDate = roundDateUp(anchorDate, timeGrain);
  const duration = getLastXTimeRangeDuration(name);
  const startDate = new Date(roundedUpAnchorDate.getTime() - duration);
  const endDate = subtractOneUnit(roundedUpAnchorDate, TimeUnit.Second);
  return {
    name: name,
    start: startDate.toISOString(),
    end: endDate.toISOString(),
    interval: timeGrain.toString(),
  };
};

const getTimeRangeDuration = (
  timeRangeName: TimeRangeName,
  allTimeRange: TimeSeriesTimeRange
): number => {
  if (lastXTimeRanges.includes(timeRangeName)) {
    return getLastXTimeRangeDuration(timeRangeName);
  }
  if (timeRangeName === TimeRangeName.AllTime) {
    return (
      new Date(allTimeRange.end).getTime() -
      new Date(allTimeRange.start).getTime()
    );
  }
  throw new Error(`Unknown time range: ${timeRangeName}`);
};

const getLastXTimeRangeDuration = (name: TimeRangeName): number => {
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

const getTimeGrainDuration = (timeGrain: TimeGrain): number => {
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

const roundDateUp = (date: Date | undefined, timeGrain: TimeGrain): Date => {
  if (!date) return new Date();
  switch (timeGrain) {
    case TimeGrain.OneMinute: {
      const interval = 60 * 1000;
      return new Date(Math.ceil(date.getTime() / interval) * interval);
    }
    // case TimeGrain.FiveMinutes: {
    //   const interval = 5 * 60 * 1000;
    //   return new Date(Math.ceil(date.getTime() / interval) * interval);
    // }
    // case TimeGrain.FifteenMinutes: {
    //   const interval = 15 * 60 * 1000;
    //   return new Date(Math.ceil(date.getTime() / interval) * interval);
    // }
    case TimeGrain.OneHour: {
      const interval = 60 * 60 * 1000;
      return new Date(Math.ceil(date.getTime() / interval) * interval);
    }
    case TimeGrain.OneDay: {
      const interval = 24 * 60 * 60 * 1000;
      return new Date(Math.ceil(date.getTime() / interval) * interval);
    }
    case TimeGrain.OneWeek: {
      // rounds to the next Monday
      const day = date.getUTCDay();
      const dateRoundedDownByDay = roundDateDown(date, TimeGrain.OneDay);
      const timeUntilMonday = (day === 0 ? 1 : 8 - day) * 24 * 60 * 60 * 1000;
      return new Date(dateRoundedDownByDay.getTime() + timeUntilMonday);
    }
    case TimeGrain.OneMonth: {
      // rounds to the 1st of the next month
      return new Date(
        Date.UTC(date.getUTCFullYear(), date.getUTCMonth() + 1, 1)
      );
    }
    case TimeGrain.OneYear: {
      // rounds to Jan 1st of the next year
      return new Date(Date.UTC(date.getUTCFullYear() + 1, 1, 1));
    }
    default:
      throw new Error(`Unknown time grain: ${timeGrain}`);
  }
};

const roundDateDown = (date: Date | undefined, timeGrain: TimeGrain): Date => {
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
      const dateRoundedDownByDay = roundDateDown(date, TimeGrain.OneDay);
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

enum TimeUnit {
  Second = "second",
  Minute = "minute",
  Hour = "hour",
  Day = "day",
  Week = "week",
  Month = "month",
  Year = "year",
}

const subtractOneUnit = (date: Date, unit: TimeUnit): Date => {
  switch (unit) {
    case TimeUnit.Second:
      return new Date(date.getTime() - 1000);
    case TimeUnit.Minute:
      return new Date(date.getTime() - 60 * 1000);
    case TimeUnit.Hour:
      return new Date(date.getTime() - 60 * 60 * 1000);
    case TimeUnit.Day:
      return new Date(date.getTime() - 24 * 60 * 60 * 1000);
    case TimeUnit.Week:
      return new Date(date.getTime() - 7 * 24 * 60 * 60 * 1000);
    case TimeUnit.Month:
      return new Date(date.getUTCFullYear(), date.getUTCMonth() - 1, 1);
    case TimeUnit.Year:
      return new Date(date.getUTCFullYear() - 1, 1, 1);
    default:
      throw new Error(`Unknown time unit: ${unit}`);
  }
};
