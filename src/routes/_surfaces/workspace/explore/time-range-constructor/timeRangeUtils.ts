import {
  TimeGrain,
  TimeRangeName,
  TimeSeriesTimeRange,
} from "$common/database-service/DatabaseTimeSeriesActions";

export interface TimeGrainOption {
  grain: TimeGrain;
  default?: boolean;
}

export interface TimeOption {
  timeRangeName: TimeRangeName;
  timeGrains: TimeGrainOption[];
  default?: boolean;
}

// TODO: replace this with a call to the `/meta?metricsDefId={metricsDefId}` endpoint, once it's available
// TODO: split this into getTimeRanges (which will come from the meta API) and getTimeGrain (which will be client-side API)
export const getTimeOptions = (
  allTimeRange: TimeSeriesTimeRange
): TimeOption[] => {
  if (!allTimeRange) return [];

  const allTimeRangeDuration =
    new Date(allTimeRange.end).getTime() -
    new Date(allTimeRange.start).getTime();

  let timeOptions: TimeOption[] = [];
  for (const timeRangeName in TimeRangeName) {
    // only show a time range if it is within the time range of the data
    const timeRangeDuration = getTimeRangeDuration(
      TimeRangeName[timeRangeName],
      allTimeRangeDuration
    );
    const showTimeRange = allTimeRangeDuration >= timeRangeDuration;

    if (showTimeRange) {
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
        if (showTimeGrain) {
          timeGrains.push({ grain: TimeGrain[timeGrain] });
        }
      }
      if (timeGrains.length === 0) {
        throw new Error(
          `No time grains generated for time range ${timeRangeName}`
        );
      }
      let timeOption = {
        timeRangeName: TimeRangeName[timeRangeName],
        timeGrains,
      };
      timeOption = setDefaultTimeGrain(timeOption);
      timeOptions.push(timeOption);
    }
  }

  timeOptions = setDefaultTimeRange(timeOptions);
  return timeOptions;
};

export const makeTimeRange = (
  timeRangeName: TimeRangeName,
  timeGrain: TimeGrain,
  allTimeRange: TimeSeriesTimeRange
): TimeSeriesTimeRange => {
  switch (timeRangeName) {
    case TimeRangeName.LastHour: {
      const endDate = roundDateDown(new Date(allTimeRange?.end), timeGrain);
      const startDate = new Date(endDate.getTime() - 60 * 60 * 1000);
      return {
        name: timeRangeName,
        start: startDate.toISOString(),
        end: endDate.toISOString(),
        interval: timeGrain.toString(),
      };
    }
    case TimeRangeName.Last6Hours: {
      const endDate = roundDateDown(new Date(allTimeRange?.end), timeGrain);
      const startDate = new Date(endDate.getTime() - 6 * 60 * 60 * 1000);
      return {
        name: timeRangeName,
        start: startDate.toISOString(),
        end: endDate.toISOString(),
        interval: timeGrain.toString(),
      };
    }
    case TimeRangeName.LastDay: {
      const endDate = roundDateDown(new Date(allTimeRange?.end), timeGrain);
      const startDate = new Date(endDate.getTime() - 24 * 60 * 60 * 1000);
      return {
        name: timeRangeName,
        start: startDate.toISOString(),
        end: endDate.toISOString(),
        interval: timeGrain.toString(),
      };
    }
    case TimeRangeName.Last2Days: {
      const endDate = roundDateDown(new Date(allTimeRange?.end), timeGrain);
      const startDate = new Date(endDate.getTime() - 2 * 24 * 60 * 60 * 1000);
      return {
        name: timeRangeName,
        start: startDate.toISOString(),
        end: endDate.toISOString(),
        interval: timeGrain.toString(),
      };
    }
    case TimeRangeName.Last5Days: {
      const endDate = roundDateDown(new Date(allTimeRange?.end), timeGrain);
      const startDate = new Date(endDate.getTime() - 5 * 24 * 60 * 60 * 1000);
      return {
        name: timeRangeName,
        start: startDate.toISOString(),
        end: endDate.toISOString(),
        interval: timeGrain.toString(),
      };
    }
    case TimeRangeName.LastWeek: {
      const endDate = roundDateDown(new Date(allTimeRange?.end), timeGrain);
      const startDate = new Date(endDate.getTime() - 7 * 24 * 60 * 60 * 1000);
      return {
        name: timeRangeName,
        start: startDate.toISOString(),
        end: endDate.toISOString(),
        interval: timeGrain.toString(),
      };
    }
    case TimeRangeName.Last2Weeks: {
      const endDate = roundDateDown(new Date(allTimeRange?.end), timeGrain);
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
      const endDate = roundDateDown(new Date(allTimeRange?.end), timeGrain);
      const startDate = new Date(endDate.getTime() - 30 * 24 * 60 * 60 * 1000);
      return {
        name: timeRangeName,
        start: startDate.toISOString(),
        end: endDate.toISOString(),
        interval: timeGrain.toString(),
      };
    }
    case TimeRangeName.Last60Days: {
      const endDate = roundDateDown(new Date(allTimeRange?.end), timeGrain);
      const startDate = new Date(endDate.getTime() - 60 * 24 * 60 * 60 * 1000);
      return {
        name: timeRangeName,
        start: startDate.toISOString(),
        end: endDate.toISOString(),
        interval: timeGrain.toString(),
      };
    }
    case TimeRangeName.AllTime: {
      const startDate = roundDateUp(new Date(allTimeRange?.start), timeGrain);
      const endDate = roundDateDown(new Date(allTimeRange?.end), timeGrain);
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
  timeOptions: TimeOption[],
  allTimeRangeInDataset: TimeSeriesTimeRange
): TimeSeriesTimeRange[] => {
  if (!timeOptions || !allTimeRangeInDataset) return [];

  const timeRanges: TimeSeriesTimeRange[] = [];
  for (const timeOption of timeOptions) {
    const defaultTimeGrain = timeOption.timeGrains.find(
      (timeGrainOption) => timeGrainOption.default
    ).grain;
    const timeRange = makeTimeRange(
      timeOption.timeRangeName,
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
  allTimeRangeDuration: number
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
      return allTimeRangeDuration;
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

const setDefaultTimeRange = (timeOptions: TimeOption[]): TimeOption[] => {
  // Use AllTime for now. When we go to production real-time datasets, we'll want to change this.
  const allTimeIdx = timeOptions.findIndex(
    (timeOption) => timeOption.timeRangeName === TimeRangeName.AllTime
  );
  if (!allTimeIdx) {
    throw new Error("AllTime time range not found");
  }
  timeOptions[allTimeIdx].default = true;
  return timeOptions;
};

const setDefaultTimeGrain = (timeOption: TimeOption): TimeOption => {
  switch (timeOption.timeRangeName) {
    case TimeRangeName.LastHour:
      timeOption.timeGrains.find(
        (timeGrainOption) => timeGrainOption.grain === TimeGrain.FifteenMinutes
      ).default = true;
      break;
    case TimeRangeName.Last6Hours:
      timeOption.timeGrains.find(
        (timeGrainOption) => timeGrainOption.grain === TimeGrain.FifteenMinutes
      ).default = true;
      break;
    case TimeRangeName.LastDay:
      timeOption.timeGrains.find(
        (timeGrainOption) => timeGrainOption.grain === TimeGrain.OneHour
      ).default = true;
      break;
    case TimeRangeName.Last2Days:
      timeOption.timeGrains.find(
        (timeGrainOption) => timeGrainOption.grain === TimeGrain.OneHour
      ).default = true;
      break;
    case TimeRangeName.Last5Days:
      timeOption.timeGrains.find(
        (timeGrainOption) => timeGrainOption.grain === TimeGrain.OneHour
      ).default = true;
      break;
    case TimeRangeName.LastWeek:
      timeOption.timeGrains.find(
        (timeGrainOption) => timeGrainOption.grain === TimeGrain.OneHour
      ).default = true;
      break;
    case TimeRangeName.Last2Weeks:
      timeOption.timeGrains.find(
        (timeGrainOption) => timeGrainOption.grain === TimeGrain.OneDay
      ).default = true;
      break;
    case TimeRangeName.Last30Days:
      timeOption.timeGrains.find(
        (timeGrainOption) => timeGrainOption.grain === TimeGrain.OneDay
      ).default = true;
      break;
    case TimeRangeName.Last60Days:
      timeOption.timeGrains.find(
        (timeGrainOption) => timeGrainOption.grain === TimeGrain.OneDay
      ).default = true;
      break;
    case TimeRangeName.AllTime:
      timeOption.timeGrains.find(
        (timeGrainOption) => timeGrainOption.grain === TimeGrain.OneDay
      ).default = true;
      break;
    default:
      throw new Error(`Unknown time range name: ${timeOption.timeRangeName}`);
  }
  return timeOption;
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
