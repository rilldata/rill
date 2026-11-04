import { DateTime, type Interval } from "luxon";

/**
 * The instants a protobuf Timestamp can carry.
 * The runtime fails to marshal anything outside this range and the request encoder refuses to send it,
 * so every time range the UI derives has to stay within it.
 * Placeholder dates such as 0001-01-01 in the data otherwise push a previous period or an adjusted fetch window into year 0.
 */
// Date.UTC treats years below 100 as 1900-based, so the bounds are parsed from ISO strings instead.
export const MIN_TIMESTAMP_MS = Date.parse("0001-01-01T00:00:00.000Z");
// The request encoder rejects anything after 9999-12-31T23:59:59Z, including the rest of that second.
export const MAX_TIMESTAMP_MS = Date.parse("9999-12-31T23:59:59.000Z");

export function isWithinTimestampBounds(time: Date | DateTime): boolean {
  const ms = time instanceof Date ? time.getTime() : time.toMillis();
  return ms >= MIN_TIMESTAMP_MS && ms <= MAX_TIMESTAMP_MS;
}

export function isIntervalWithinTimestampBounds(
  interval: Interval<true>,
): boolean {
  return (
    isWithinTimestampBounds(interval.start) &&
    isWithinTimestampBounds(interval.end)
  );
}

export function clampToTimestampBounds(time: DateTime): DateTime {
  if (time.toMillis() < MIN_TIMESTAMP_MS) {
    return DateTime.fromMillis(MIN_TIMESTAMP_MS, { zone: time.zone });
  }
  if (time.toMillis() > MAX_TIMESTAMP_MS) {
    return DateTime.fromMillis(MAX_TIMESTAMP_MS, { zone: time.zone });
  }
  return time;
}
