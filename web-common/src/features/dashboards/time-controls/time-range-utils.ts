import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import {
  lastXTimeRanges,
  TimeRangeName,
  TimeSeriesTimeRange,
} from "./time-control-types";

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
  timeGrain: V1TimeGrain;
  enabled: boolean;
}

// This is for pre-set relative time ranges – where the start/end dates are not yet deterimined.
// For custom time ranges, we'll need another function with "breakpoint" logic that analyzes the user-determined start/end dates.
export const getSelectableTimeGrains = (
  timeRangeISO: string,
  allTimeRange: TimeSeriesTimeRange
): TimeGrainOption[] => {
  const timeRangeName = ISODurationToTimeRange(timeRangeISO);
  if (!timeRangeName || !allTimeRange) return [];
  const timeRangeDuration = getTimeRangeDuration(timeRangeName, allTimeRange);

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
    const timeGrainDuration = getTimeGrainDuration(V1TimeGrain[timeGrain]);
    const pointsOnLineChart = timeRangeDuration / timeGrainDuration;
    const showTimeGrain =
      pointsOnLineChart >= MINIMUM_POINTS_ON_LINE_CHART &&
      pointsOnLineChart <= MAXIMUM_POINTS_ON_LINE_CHART;
    timeGrains.push({
      timeGrain: V1TimeGrain[timeGrain],
      enabled: showTimeGrain,
    });
  }
  if (timeGrains.length === 0) {
    throw new Error(`No time grains generated for time range ${timeRangeName}`);
  }
  return timeGrains;
};

export const timeRangeToISODuration = (
  timeRangeName: TimeRangeName
): string => {
  switch (timeRangeName) {
    case TimeRangeName.LastHour:
      return "PT1H";
    case TimeRangeName.Last6Hours:
      return "PT6H";
    case TimeRangeName.LastDay:
      return "P1D";
    case TimeRangeName.Last2Days:
      return "P2D";
    case TimeRangeName.Last5Days:
      return "P5D";
    case TimeRangeName.LastWeek:
      return "P7D";
    case TimeRangeName.Last2Weeks:
      return "P14D";
    case TimeRangeName.Last30Days:
      return "P30D";
    case TimeRangeName.Last60Days:
      return "P60D";
    case TimeRangeName.AllTime:
      return "P9999Y";
    default:
      return undefined;
  }
};

export const ISODurationToTimeRange = (isoDuration: string): TimeRangeName => {
  switch (isoDuration) {
    case "PT1H":
      return TimeRangeName.LastHour;
    case "PT6H":
      return TimeRangeName.Last6Hours;
    case "P1D":
      return TimeRangeName.LastDay;
    case "P2D":
      return TimeRangeName.Last2Days;
    case "P5D":
      return TimeRangeName.Last5Days;
    case "P7D":
      return TimeRangeName.LastWeek;
    case "P14D":
      return TimeRangeName.Last2Weeks;
    case "P30D":
      return TimeRangeName.Last30Days;
    case "P60D":
      return TimeRangeName.Last60Days;
    case "P9999Y":
      return TimeRangeName.AllTime;
    default:
      return TimeRangeName.AllTime;
  }
};

export const getDefaultTimeGrain = (
  timeRangeName: TimeRangeName,
  allTimeRange: TimeSeriesTimeRange
): V1TimeGrain => {
  switch (timeRangeName) {
    case TimeRangeName.LastHour:
      return V1TimeGrain.TIME_GRAIN_MINUTE;
    case TimeRangeName.Last6Hours:
      return V1TimeGrain.TIME_GRAIN_HOUR;
    case TimeRangeName.LastDay:
      return V1TimeGrain.TIME_GRAIN_HOUR;
    case TimeRangeName.Last2Days:
      return V1TimeGrain.TIME_GRAIN_HOUR;
    case TimeRangeName.Last5Days:
      return V1TimeGrain.TIME_GRAIN_HOUR;
    case TimeRangeName.LastWeek:
      return V1TimeGrain.TIME_GRAIN_HOUR;
    case TimeRangeName.Last2Weeks:
      return V1TimeGrain.TIME_GRAIN_DAY;
    case TimeRangeName.Last30Days:
      return V1TimeGrain.TIME_GRAIN_DAY;
    case TimeRangeName.Last60Days:
      return V1TimeGrain.TIME_GRAIN_DAY;
    case TimeRangeName.AllTime: {
      if (!allTimeRange) return V1TimeGrain.TIME_GRAIN_DAY;
      const allTimeRangeDuration = getTimeRangeDuration(
        TimeRangeName.AllTime,
        allTimeRange
      );
      if (allTimeRangeDuration <= 2 * 60 * 60 * 1000) {
        return V1TimeGrain.TIME_GRAIN_MINUTE;
      }
      if (allTimeRangeDuration <= 14 * 24 * 60 * 60 * 1000) {
        return V1TimeGrain.TIME_GRAIN_HOUR;
      }
      if (allTimeRangeDuration <= 60 * 24 * 60 * 60 * 1000) {
        return V1TimeGrain.TIME_GRAIN_DAY;
      }
      if (allTimeRangeDuration <= 365 * 24 * 60 * 60 * 1000) {
        return V1TimeGrain.TIME_GRAIN_WEEK;
      }
      if (allTimeRangeDuration <= 20 * 365 * 24 * 60 * 60 * 1000) {
        return V1TimeGrain.TIME_GRAIN_MONTH;
      }
      return V1TimeGrain.TIME_GRAIN_YEAR;
    }
    default:
      throw new Error(`No default time grain for time range ${timeRangeName}`);
  }
};

export const makeTimeRange = (
  timeRangeName: TimeRangeName,
  timeGrain: V1TimeGrain,
  allTimeRange: TimeSeriesTimeRange
): TimeSeriesTimeRange => {
  // Compute actual start time
  let start: Date;
  if (timeRangeName === TimeRangeName.AllTime) {
    start = new Date(allTimeRange.start);
  } else if (lastXTimeRanges.includes(timeRangeName)) {
    const allTimeEnd = new Date(allTimeRange?.end);
    start = new Date(
      allTimeEnd.getTime() - getLastXTimeRangeDuration(timeRangeName)
    );
  } else {
    throw new Error(`Unknown time range name: ${timeRangeName}`);
  }

  // Round start time to nearest lower time grain
  start = floorDate(start, timeGrain);

  // Round end time to start of next grain, since end times are exclusive
  let end = addGrains(new Date(allTimeRange?.end), 1, timeGrain);
  end = floorDate(end, timeGrain);

  return {
    name: timeRangeName,
    start: start.toISOString(),
    end: end.toISOString(),
    interval: timeGrain,
  };
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

export const prettyTimeGrain = (timeGrain: V1TimeGrain): string => {
  if (!timeGrain) return "";
  switch (timeGrain) {
    case V1TimeGrain.TIME_GRAIN_MINUTE:
      return "minute";
    // case TimeGrain.FiveMinutes:
    //   return "5 minute";
    // case TimeGrain.FifteenMinutes:
    //   return "15 minute";
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
      throw new Error(`Unknown time grain: ${timeGrain}`);
  }
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

const getTimeGrainDuration = (timeGrain: V1TimeGrain): number => {
  switch (timeGrain) {
    case V1TimeGrain.TIME_GRAIN_MINUTE:
      return 60 * 1000;
    case V1TimeGrain.TIME_GRAIN_HOUR:
      return 60 * 60 * 1000;
    case V1TimeGrain.TIME_GRAIN_DAY:
      return 24 * 60 * 60 * 1000;
    case V1TimeGrain.TIME_GRAIN_WEEK:
      return 7 * 24 * 60 * 60 * 1000;
    case V1TimeGrain.TIME_GRAIN_MONTH:
      return 30 * 24 * 60 * 60 * 1000;
    case V1TimeGrain.TIME_GRAIN_YEAR:
      return 365 * 24 * 60 * 60 * 1000;
    default:
      throw new Error(`Unknown time grain: ${timeGrain}`);
  }
};

const floorDate = (date: Date | undefined, timeGrain: V1TimeGrain): Date => {
  if (!date) return new Date();
  switch (timeGrain) {
    case V1TimeGrain.TIME_GRAIN_MINUTE: {
      const interval = 60 * 1000;
      return new Date(Math.floor(date.getTime() / interval) * interval);
    }
    case V1TimeGrain.TIME_GRAIN_HOUR: {
      const interval = 60 * 60 * 1000;
      return new Date(Math.floor(date.getTime() / interval) * interval);
    }
    case V1TimeGrain.TIME_GRAIN_DAY: {
      const interval = 24 * 60 * 60 * 1000;
      return new Date(Math.floor(date.getTime() / interval) * interval);
    }
    case V1TimeGrain.TIME_GRAIN_WEEK: {
      // rounds to the most recent Monday
      const day = date.getUTCDay();
      const dateRoundedDownByDay = floorDate(date, V1TimeGrain.TIME_GRAIN_DAY);
      const timeFromMonday = (day === 0 ? 6 : day - 1) * 24 * 60 * 60 * 1000;
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

export const addGrains = (
  date: Date,
  units: number,
  grain: V1TimeGrain
): Date => {
  switch (grain) {
    case V1TimeGrain.TIME_GRAIN_MINUTE:
      return new Date(date.getTime() + units * 60 * 1000);
    case V1TimeGrain.TIME_GRAIN_HOUR:
      return new Date(date.getTime() + units * 60 * 60 * 1000);
    case V1TimeGrain.TIME_GRAIN_DAY:
      return new Date(date.getTime() + units * 24 * 60 * 60 * 1000);
    case V1TimeGrain.TIME_GRAIN_WEEK:
      return new Date(date.getTime() + units * 7 * 24 * 60 * 60 * 1000);
    case V1TimeGrain.TIME_GRAIN_MONTH:
      return new Date(
        Date.UTC(date.getUTCFullYear(), date.getUTCMonth() + units, 1)
      );
    case V1TimeGrain.TIME_GRAIN_YEAR:
      return new Date(Date.UTC(date.getUTCFullYear() + units, 1, 1));
    default:
      throw new Error(`Unknown time grain: ${grain}`);
  }
};
