import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config.ts";
import { durationToMillis } from "@rilldata/web-common/lib/time/grains";
import { V1TimeGrainToOrder } from "@rilldata/web-common/lib/time/new-grains.ts";
import { getDateMonthYearForTimezone } from "@rilldata/web-common/lib/time/timezone";
import { getTimeWidth } from "@rilldata/web-common/lib/time/transforms";
import { TimeRangePreset } from "@rilldata/web-common/lib/time/types.ts";
import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import { DateTime, type DateTimeFormatOptions } from "luxon";

/**
 * Formats a start and end for usage in the application.
 * NOTE: this is primarily used for the time range picker. We might want to
 * colocate the code w/ the component.
 */
export const prettyFormatTimeRange = (
  start: Date,
  end: Date | undefined,
  timePreset: string | undefined,
  timeZone: string,
): string => {
  const isAllTime = timePreset === TimeRangePreset.ALL_TIME;

  if (!end) {
    return prettyFormatTimestamp(start, timeZone);
  }

  const {
    day: startDate,
    month: startMonth,
    year: startYear,
  } = getDateMonthYearForTimezone(start, timeZone);

  let {
    day: endDate,
    month: endMonth,
    year: endYear,
  } = getDateMonthYearForTimezone(end, timeZone);

  if (
    startDate === endDate &&
    startMonth === endMonth &&
    startYear === endYear
  ) {
    return `${start.toLocaleDateString(undefined, {
      month: "short",
      timeZone,
    })} ${startDate}, ${startYear} (${start
      .toLocaleString(undefined, {
        hour12: true,
        hour: "numeric",
        minute: "numeric",
        timeZone,
      })
      .replace(/\s/g, "")}-${end
      .toLocaleString(undefined, {
        hour12: true,
        hour: "numeric",
        minute: "numeric",
        timeZone,
      })
      .replace(/\s/g, "")})`;
  }

  const timeRangeDurationMs = getTimeWidth(start, end);
  if (
    timeRangeDurationMs <= durationToMillis(TIME_GRAIN.TIME_GRAIN_DAY.duration)
  ) {
    return `${start.toLocaleDateString(undefined, {
      month: "short",
      timeZone,
    })} ${startDate}-${endDate}, ${startYear} (${start
      .toLocaleString(undefined, {
        hour12: true,
        hour: "numeric",
        minute: "numeric",
        timeZone,
      })
      .replace(/\s/g, "")}-${end
      .toLocaleString(undefined, {
        hour12: true,
        hour: "numeric",
        minute: "numeric",
        timeZone,
      })
      .replace(/\s/g, "")})`;
  }

  let inclusiveEndDate;

  let timeString = "";

  const startTime = start.toLocaleTimeString(undefined, { timeZone });
  const endTime = end.toLocaleTimeString(undefined, { timeZone });

  if (isAllTime) {
    inclusiveEndDate = new Date(end);
  } else if (startTime === "12:00:00 am" && endTime === "12:00:00 am") {
    // beyond this point, we're dealing with time ranges that are full day periods
    // since time range is exclusive at the end, we need to subtract a day
    inclusiveEndDate = new Date(
      end.getTime() - durationToMillis(TIME_GRAIN.TIME_GRAIN_DAY.duration),
    );

    const inclusiveEndDateWithTimeZone = getDateMonthYearForTimezone(
      inclusiveEndDate,
      timeZone,
    );

    endDate = inclusiveEndDateWithTimeZone.day;
    endMonth = inclusiveEndDateWithTimeZone.month;
    endYear = inclusiveEndDateWithTimeZone.year;
  } else {
    // display full time when the hours are not at 00:00
    inclusiveEndDate = end;

    timeString = `(${start
      .toLocaleString(undefined, {
        hour12: true,
        hour: "numeric",
        minute: "numeric",
        timeZone,
      })
      .replace(/\s/g, "")}-${end
      .toLocaleString(undefined, {
        hour12: true,
        hour: "numeric",
        minute: "numeric",
        timeZone,
      })
      .replace(/\s/g, "")})`;
  }

  // month is the same
  if (startMonth === endMonth && startYear === endYear) {
    return `${start.toLocaleDateString(undefined, {
      month: "short",
      timeZone,
    })} ${startDate}-${endDate}, ${startYear} ${timeString}`;
  }

  // year is the same
  if (startYear === endYear) {
    return `${start.toLocaleDateString(undefined, {
      month: "short",
      day: "numeric",
      timeZone,
    })} - ${inclusiveEndDate.toLocaleDateString(undefined, {
      month: "short",
      day: "numeric",
      timeZone,
    })}, ${startYear} ${timeString}`;
  }
  // year is different
  const dateFormatOptions: Intl.DateTimeFormatOptions = {
    year: "numeric",
    month: "short",
    day: "numeric",
    timeZone,
  };
  return `${start.toLocaleDateString(
    undefined,
    dateFormatOptions,
  )} - ${inclusiveEndDate.toLocaleDateString(undefined, dateFormatOptions)}`;
};

export function prettyFormatTimestamp(date: Date, timeZone: string): string {
  const dateTime = DateTime.fromJSDate(date).setZone(timeZone);
  if (!dateTime.isValid) return "";
  return dateTime.toLocaleString(DateTime.DATETIME_MED);
}

export function prettyFormatTimeRangeV2(
  start: Date,
  end: Date,
  grain: V1TimeGrain,
  timeZone: string,
) {
  const startDateTime = DateTime.fromJSDate(start).setZone(timeZone);
  const endDateTime = DateTime.fromJSDate(end).setZone(timeZone);
  const datePart = formatDatePartOfTimeRange(startDateTime, endDateTime, grain);
  const timePart = formatTimePartOfTimeRange(startDateTime, endDateTime, grain);
  return `${datePart}${timePart}`;
}

const yearGrainOrder = V1TimeGrainToOrder[V1TimeGrain.TIME_GRAIN_YEAR];
const monthGrainOrder = V1TimeGrainToOrder[V1TimeGrain.TIME_GRAIN_MONTH];
function formatDatePartOfTimeRange(
  start: DateTime,
  end: DateTime,
  grain: V1TimeGrain,
) {
  const grainOrder = V1TimeGrainToOrder[grain] ?? 0;
  const hasSameYear = start.year === end.year;
  const showMonth = grainOrder < yearGrainOrder || start.month !== 1;
  const hasSameMonth = start.month === end.month;
  const showDay = grainOrder < monthGrainOrder || start.day !== 1;
  const hasSameDay = hasSameMonth && start.day === end.day;
  const hasSameDate = hasSameYear && hasSameDay;

  const startFormat: DateTimeFormatOptions = {};
  const endFormat: DateTimeFormatOptions = {
    year: "numeric",
  };

  if (!hasSameYear) {
    startFormat.year = "numeric";
  }

  if (showMonth) {
    startFormat.month = "short";
    if (!hasSameMonth) {
      endFormat.month = "short";
    }
  }

  if (showDay) {
    startFormat.day = "numeric";
    if (!hasSameDay) {
      endFormat.day = "numeric";
    }
  }

  if (hasSameDate) {
    startFormat.year = "numeric";
    return start.toLocaleString(startFormat);
  }

  const showYearAsSuffix = endFormat.day && !endFormat.month;
  let suffix = "";
  if (showYearAsSuffix) {
    delete endFormat.year;
    suffix = `${startFormat.day ? "," : ""} ${end.toLocaleString({ year: "numeric" })}`;
  }

  const formattedStart = start.toLocaleString(startFormat);
  const formattedEnd = end.toLocaleString(endFormat);
  return `${formattedStart} - ${formattedEnd}${suffix}`;
}

const hourGrainOrder = V1TimeGrainToOrder[V1TimeGrain.TIME_GRAIN_HOUR];
const minuteGrainOrder = V1TimeGrainToOrder[V1TimeGrain.TIME_GRAIN_MINUTE];
function formatTimePartOfTimeRange(
  start: DateTime,
  end: DateTime,
  grain: V1TimeGrain,
) {
  const hasSameHour = start.hour === end.hour;
  const hasSameMinute = start.minute === end.minute;
  const hasSameTime = hasSameHour && hasSameMinute;
  const grainOrder = V1TimeGrainToOrder[grain] ?? 0;

  const format: DateTimeFormatOptions = {
    hour12: true,
    hour: "numeric",
  };
  const showMinute =
    !hasSameMinute || grainOrder <= minuteGrainOrder || start.minute !== 0;
  if (showMinute) format.minute = "2-digit";

  if (hasSameTime) {
    const onDayBoundary = start.startOf("day").equals(start);
    const showTimePart = !onDayBoundary || grainOrder < hourGrainOrder;
    const formattedTime = start.toLocaleString(format).replace(/\s/g, "");
    return showTimePart ? ` (${formattedTime})` : "";
  }

  const formattedStart = start.toLocaleString(format).replace(/\s/g, "");
  const formattedEnd = end.toLocaleString(format).replace(/\s/g, "");
  return ` (${formattedStart}-${formattedEnd})`;
}
