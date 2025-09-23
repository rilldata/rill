import { V1TimeGrainToOrder } from "@rilldata/web-common/lib/time/new-grains.ts";
import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import { DateTime, type DateTimeFormatOptions, Interval } from "luxon";

// Formats a Luxon interval for human readable display throughout the application.
export function prettyFormatTimeRange(
  interval: Interval | undefined,
  grain: V1TimeGrain,
) {
  if (!interval?.isValid || !interval.start || !interval.end)
    return "Invalid interval";

  const datePart = formatDatePartOfTimeRange(interval, grain);
  const timePart = formatTimePartOfTimeRange(
    interval.start,
    interval.end,
    grain,
  );
  return `${datePart}${timePart}`;
}

const yearGrainOrder = V1TimeGrainToOrder[V1TimeGrain.TIME_GRAIN_YEAR];
const monthGrainOrder = V1TimeGrainToOrder[V1TimeGrain.TIME_GRAIN_MONTH];
const dayGrainOrder = V1TimeGrainToOrder[V1TimeGrain.TIME_GRAIN_DAY];

function formatDatePartOfTimeRange(interval: Interval, grain: V1TimeGrain) {
  if (!interval.start?.isValid || !interval.end?.isValid) return ""; // type safety

  const grainOrder = V1TimeGrainToOrder[grain] ?? 0;
  const hasSameMonth = interval.start.month === interval.end.month;
  const showMonth =
    !hasSameMonth || grainOrder < yearGrainOrder || interval.start.month !== 1;
  const hasSameDay = interval.start.day === interval.end.day;
  const showDay =
    !hasSameDay || grainOrder < monthGrainOrder || interval.start.day !== 1;

  const format: DateTimeFormatOptions = {
    year: "numeric",
  };

  if (showMonth) format.month = "short";

  if (showDay) format.day = "numeric";

  const displayAsInclusiveEnd =
    grainOrder >= dayGrainOrder && interval.end > interval.start;

  return displayAsInclusiveEnd
    ? Interval.fromDateTimes(
        interval.start,
        interval.end.minus({ millisecond: 1 }),
      ).toLocaleString(format)
    : interval.toLocaleString(format);
}

const hourGrainOrder = V1TimeGrainToOrder[V1TimeGrain.TIME_GRAIN_HOUR];
const minuteGrainOrder = V1TimeGrainToOrder[V1TimeGrain.TIME_GRAIN_MINUTE];
function formatTimePartOfTimeRange(
  start: DateTime,
  end: DateTime,
  grain: V1TimeGrain,
) {
  const grainOrder = getCorrectGrainOrder(grain);
  const hasSameHour = start.hour === end.hour;
  const hasSameMinute = start.minute === end.minute;
  const showMinute =
    !hasSameMinute || grainOrder <= minuteGrainOrder || start.minute !== 0;
  const hasSameSecond = start.second === end.second;
  const showSeconds = !hasSameSecond || start.second !== 0;
  const hasSameTime = hasSameHour && hasSameMinute && hasSameSecond;

  const format: DateTimeFormatOptions = {
    hour12: true,
    hour: "numeric",
  };
  if (showMinute || showSeconds) format.minute = "2-digit";
  if (showSeconds) format.second = "2-digit";

  if (hasSameTime) {
    const onDayBoundary = start.startOf("day").equals(start);
    const showTimePart = !onDayBoundary || grainOrder <= hourGrainOrder;
    const formattedTime = start.toLocaleString(format).replace(/\s/g, "");
    return showTimePart ? ` (${formattedTime})` : "";
  }

  const formattedStart = start.toLocaleString(format).replace(/\s/g, "");
  const formattedEnd = end.toLocaleString(format).replace(/\s/g, "");
  return ` (${formattedStart}-${formattedEnd})`;
}

function getCorrectGrainOrder(grain: V1TimeGrain) {
  if (
    grain === V1TimeGrain.TIME_GRAIN_UNSPECIFIED ||
    !(grain in V1TimeGrainToOrder)
  )
    return yearGrainOrder + 1;
  return V1TimeGrainToOrder[grain];
}
