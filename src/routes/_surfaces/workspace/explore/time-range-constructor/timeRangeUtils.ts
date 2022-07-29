import {
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

// This is for pre-set relative time ranges – where the start/end dates are not yet deterimined.
// For custom time ranges, we'll need another function with "breakpoint" logic that analyzes the user-determined start/end dates.
export const getSelectableTimeGrains = (
  timeRangeName: TimeRangeName,
  allTimeRange: TimeSeriesTimeRange
): TimeGrain[] => {
  if (!allTimeRange) return [];
  const timeRangeDuration = getTimeRangeDuration(timeRangeName, allTimeRange);

  const timeGrains: TimeGrain[] = [];
  for (const timeGrain in TimeGrain) {
    // only show a time grain if it results in a reasonable number of points on the line chart
    const MINIMUM_POINTS_ON_LINE_CHART = 2;
    const MAXIMUM_POINTS_ON_LINE_CHART = 2500;
    const timeGrainDuration = getTimeGrainDuration(TimeGrain[timeGrain]);
    const pointsOnLineChart = timeRangeDuration / timeGrainDuration;
    const showTimeGrain =
      pointsOnLineChart >= MINIMUM_POINTS_ON_LINE_CHART &&
      pointsOnLineChart <= MAXIMUM_POINTS_ON_LINE_CHART;
    if (showTimeGrain) {
      timeGrains.push(TimeGrain[timeGrain]);
    }
  }
  if (timeGrains.length === 0) {
    throw new Error(`No time grains generated for time range ${timeRangeName}`);
  }
  return timeGrains;
};

export const getDefaultTimeGrain = (
  timeRangeName: TimeRangeName
): TimeGrain => {
  switch (timeRangeName) {
    case TimeRangeName.LastHour:
      return TimeGrain.FifteenMinutes;
    case TimeRangeName.Last6Hours:
      return TimeGrain.FifteenMinutes;
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
    case TimeRangeName.AllTime:
      // TODO: this needs breakpoint logic using start/end time.
      return TimeGrain.OneDay;
    default:
      throw new Error(`No default time grain for time range ${timeRangeName}`);
  }
};

export const makeTimeRange = (
  timeRangeName: TimeRangeName,
  timeGrain: TimeGrain,
  allTimeRange: TimeSeriesTimeRange
): TimeSeriesTimeRange => {
  switch (timeRangeName) {
    case TimeRangeName.LastHour: {
      const endDate = roundDateUp(new Date(allTimeRange?.end), timeGrain);
      const startDate = new Date(endDate.getTime() - 60 * 60 * 1000);
      return {
        name: timeRangeName,
        start: startDate.toISOString(),
        end: endDate.toISOString(),
        interval: timeGrain.toString(),
      };
    }
    case TimeRangeName.Last6Hours: {
      const endDate = roundDateUp(new Date(allTimeRange?.end), timeGrain);
      const startDate = new Date(endDate.getTime() - 6 * 60 * 60 * 1000);
      return {
        name: timeRangeName,
        start: startDate.toISOString(),
        end: endDate.toISOString(),
        interval: timeGrain.toString(),
      };
    }
    case TimeRangeName.LastDay: {
      const endDate = roundDateUp(new Date(allTimeRange?.end), timeGrain);
      const startDate = new Date(endDate.getTime() - 24 * 60 * 60 * 1000);
      return {
        name: timeRangeName,
        start: startDate.toISOString(),
        end: endDate.toISOString(),
        interval: timeGrain.toString(),
      };
    }
    case TimeRangeName.Last2Days: {
      const endDate = roundDateUp(new Date(allTimeRange?.end), timeGrain);
      const startDate = new Date(endDate.getTime() - 2 * 24 * 60 * 60 * 1000);
      return {
        name: timeRangeName,
        start: startDate.toISOString(),
        end: endDate.toISOString(),
        interval: timeGrain.toString(),
      };
    }
    case TimeRangeName.Last5Days: {
      const endDate = roundDateUp(new Date(allTimeRange?.end), timeGrain);
      const startDate = new Date(endDate.getTime() - 5 * 24 * 60 * 60 * 1000);
      return {
        name: timeRangeName,
        start: startDate.toISOString(),
        end: endDate.toISOString(),
        interval: timeGrain.toString(),
      };
    }
    case TimeRangeName.LastWeek: {
      const endDate = roundDateUp(new Date(allTimeRange?.end), timeGrain);
      const startDate = new Date(endDate.getTime() - 7 * 24 * 60 * 60 * 1000);
      return {
        name: timeRangeName,
        start: startDate.toISOString(),
        end: endDate.toISOString(),
        interval: timeGrain.toString(),
      };
    }
    case TimeRangeName.Last2Weeks: {
      const endDate = roundDateUp(new Date(allTimeRange?.end), timeGrain);
      const startDate = new Date(
        endDate.getTime() - 2 * 7 * 24 * 60 * 60 * 1000
      );
      return {
        name: timeRangeName,
        start: startDate.toISOString(),
        end: endDate.toISOString(),
        interval: timeGrain.toString(),
      };
    }
    case TimeRangeName.Last30Days: {
      const endDate = roundDateUp(new Date(allTimeRange?.end), timeGrain);
      const startDate = new Date(endDate.getTime() - 30 * 24 * 60 * 60 * 1000);
      return {
        name: timeRangeName,
        start: startDate.toISOString(),
        end: endDate.toISOString(),
        interval: timeGrain.toString(),
      };
    }
    case TimeRangeName.Last60Days: {
      const endDate = roundDateUp(new Date(allTimeRange?.end), timeGrain);
      const startDate = new Date(endDate.getTime() - 60 * 24 * 60 * 60 * 1000);
      return {
        name: timeRangeName,
        start: startDate.toISOString(),
        end: endDate.toISOString(),
        interval: timeGrain.toString(),
      };
    }
    case TimeRangeName.AllTime: {
      const startDate = roundDateDown(new Date(allTimeRange?.start), timeGrain);
      const endDate = roundDateUp(new Date(allTimeRange?.end), timeGrain);
      return {
        name: timeRangeName,
        start: startDate.toISOString(),
        end: endDate.toISOString(),
        interval: timeGrain.toString(),
      };
    }
    default:
      throw new Error(`Unknown time range name: ${timeRangeName}`);
  }
};

export const makeTimeRanges = (
  timeRangeNames: TimeRangeName[],
  allTimeRangeInDataset: TimeSeriesTimeRange
): TimeSeriesTimeRange[] => {
  if (!timeRangeNames || !allTimeRangeInDataset) return [];

  const timeRanges: TimeSeriesTimeRange[] = [];
  for (const timeRangeName of timeRangeNames) {
    const defaultTimeGrain = getDefaultTimeGrain(timeRangeName);
    const timeRange = makeTimeRange(
      timeRangeName,
      defaultTimeGrain,
      allTimeRangeInDataset
    );
    timeRanges.push(timeRange);
  }
  return timeRanges;
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

  const TIMEZONE = Intl.DateTimeFormat().resolvedOptions().timeZone; // the user's local timezone
  // const TIMEZONE = "UTC";

  const start = new Date(timeRange.start);
  const end = new Date(timeRange.end);

  // day is the same
  if (
    start.getDate() === end.getDate() &&
    start.getMonth() === end.getMonth() &&
    start.getFullYear() === end.getFullYear()
  ) {
    return `${start.toLocaleDateString(undefined, {
      month: "long",
      timeZone: TIMEZONE,
    })} ${start.getDate()}, ${start.getFullYear()} (${start
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
  if (
    start.getMonth() === end.getMonth() &&
    start.getFullYear() === end.getFullYear()
  ) {
    return `${start.toLocaleDateString(undefined, {
      month: "long",
      timeZone: TIMEZONE,
    })} ${start.getDate()}-${end.getDate()}, ${start.getFullYear()} (${start
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
  if (start.getFullYear() === end.getFullYear()) {
    return `${start.toLocaleDateString(undefined, {
      month: "long",
      day: "numeric",
      timeZone: TIMEZONE,
    })} - ${end.toLocaleDateString(undefined, {
      month: "long",
      day: "numeric",
      timeZone: TIMEZONE,
    })}, ${start.getFullYear()}`;
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

export const prettyTimeGrain = (timeGrain: TimeGrain): string => {
  if (!timeGrain) return "";
  switch (timeGrain) {
    case TimeGrain.FiveMinutes:
      return "5 minutes";
    case TimeGrain.FifteenMinutes:
      return "15 minutes";
    case TimeGrain.OneHour:
      return "1 hour";
    case TimeGrain.OneDay:
      return "1 day";
    case TimeGrain.OneWeek:
      return "1 week";
    case TimeGrain.OneMonth:
      return "1 month";
    case TimeGrain.OneYear:
      return "1 year";
    default:
      throw new Error(`Unknown time grain: ${timeGrain}`);
  }
};

const getTimeRangeDuration = (
  timeRangeName: TimeRangeName,
  allTimeRange: TimeSeriesTimeRange
): number => {
  switch (timeRangeName) {
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
    case TimeRangeName.AllTime:
      return (
        new Date(allTimeRange.end).getTime() -
        new Date(allTimeRange.start).getTime()
      );
    default:
      throw new Error(`Unknown time range name: ${timeRangeName}`);
  }
};

const getTimeGrainDuration = (timeGrain: TimeGrain): number => {
  switch (timeGrain) {
    case TimeGrain.FiveMinutes:
      return 5 * 60 * 1000;
    case TimeGrain.FifteenMinutes:
      return 15 * 60 * 1000;
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

const roundDateDown = (date: Date | undefined, timeGrain: TimeGrain): Date => {
  if (!date) return new Date();
  switch (timeGrain) {
    case TimeGrain.FiveMinutes: {
      const interval = 5 * 60 * 1000;
      return new Date(Math.round(date.getTime() / interval) * interval);
    }
    case TimeGrain.FifteenMinutes: {
      const interval = 15 * 60 * 1000;
      return new Date(Math.floor(date.getTime() / interval) * interval);
    }
    case TimeGrain.OneHour: {
      const interval = 60 * 60 * 1000;
      return new Date(Math.floor(date.getTime() / interval) * interval);
    }
    case TimeGrain.OneDay: {
      const interval = 24 * 60 * 60 * 1000;
      return new Date(Math.floor(date.getTime() / interval) * interval);
    }
    case TimeGrain.OneWeek: {
      const interval = 7 * 24 * 60 * 60 * 1000;
      return new Date(Math.floor(date.getTime() / interval) * interval);
    }
    case TimeGrain.OneMonth: {
      const interval = 30 * 24 * 60 * 60 * 1000;
      return new Date(Math.floor(date.getTime() / interval) * interval);
    }
    case TimeGrain.OneYear: {
      const interval = 365 * 24 * 60 * 60 * 1000;
      return new Date(Math.floor(date.getTime() / interval) * interval);
    }
    default:
      throw new Error(`Unknown time grain: ${timeGrain}`);
  }
};

const roundDateUp = (date: Date | undefined, timeGrain: TimeGrain): Date => {
  if (!date) return new Date();
  switch (timeGrain) {
    case TimeGrain.FiveMinutes: {
      const interval = 5 * 60 * 1000;
      return new Date(Math.ceil(date.getTime() / interval) * interval);
    }
    case TimeGrain.FifteenMinutes: {
      const interval = 15 * 60 * 1000;
      return new Date(Math.ceil(date.getTime() / interval) * interval);
    }
    case TimeGrain.OneHour: {
      const interval = 60 * 60 * 1000;
      return new Date(Math.ceil(date.getTime() / interval) * interval);
    }
    case TimeGrain.OneDay: {
      const interval = 24 * 60 * 60 * 1000;
      return new Date(Math.ceil(date.getTime() / interval) * interval);
    }
    case TimeGrain.OneWeek: {
      const interval = 7 * 24 * 60 * 60 * 1000;
      return new Date(Math.ceil(date.getTime() / interval) * interval);
    }
    case TimeGrain.OneMonth: {
      const interval = 30 * 24 * 60 * 60 * 1000;
      return new Date(Math.ceil(date.getTime() / interval) * interval);
    }
    case TimeGrain.OneYear: {
      const interval = 365 * 24 * 60 * 60 * 1000;
      return new Date(Math.ceil(date.getTime() / interval) * interval);
    }
    default:
      throw new Error(`Unknown time grain: ${timeGrain}`);
  }
};
