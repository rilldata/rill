import { DateTime, type Interval } from "luxon";

/**
 * Formats a time range to a human readable time range.
 * TODO: update time picker to use this
 */
export function formatRange(interval: Interval) {
  const shouldShowSeconds =
    interval.start?.second !== 0 || interval.end?.second !== 0;

  const showShouldMinutes =
    interval.start?.minute !== 0 || interval.end?.minute !== 0;

  const shouldShowHours =
    interval.start?.hour !== 0 || interval.end?.hour !== 0;

  const intervalStartsAndEndsOnHour =
    interval.start?.minute === 0 && interval.end?.minute === 0;

  const showTime = shouldShowSeconds || showShouldMinutes || shouldShowHours;

  const inclusiveInterval = interval.set({
    end: interval.end?.minus({ millisecond: 1 }),
  });

  const displayedInterval = showTime ? interval : inclusiveInterval;

  const datePart = displayedInterval.toLocaleString(DateTime.DATE_MED);
  if (!showTime) return datePart;

  const timeFormat =
    getTimeFormat(
      intervalStartsAndEndsOnHour,
      showShouldMinutes,
      shouldShowSeconds,
    ) + " a";

  const timePart = displayedInterval.toFormat(timeFormat, { separator: "-" });

  return `${datePart} (${timePart})`;
}

export function formatTime(time: DateTime) {
  const shouldShowSeconds = time.second !== 0;

  const showShouldMinutes = time.minute !== 0;

  const shouldShowHours = time.hour !== 0;

  const showTime = shouldShowSeconds || showShouldMinutes || shouldShowHours;

  const datePart = time.toLocaleString(DateTime.DATE_MED);
  if (!showTime) return datePart;

  const timeFormat =
    getTimeFormat(shouldShowHours, showShouldMinutes, shouldShowSeconds) + " a";

  const timePart = time.toFormat(timeFormat);

  return `${datePart} (${timePart})`;
}

const fullTimeFormat = "h:mm:ss";
function getTimeFormat(hours: boolean, minutes: boolean, seconds: boolean) {
  if (seconds) {
    return fullTimeFormat.replace(/:SSS/, "");
  } else if (minutes) {
    return fullTimeFormat.replace(/:SSS/, "").replace(/:ss/, "");
  } else if (hours) {
    return "h";
  }
}
