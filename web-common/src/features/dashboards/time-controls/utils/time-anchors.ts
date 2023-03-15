import type { TimeRange } from "../time-control-types";
import { Duration, DateTime } from "luxon";
import { Period, TimeUnit } from "./time-types";
import {
  RelativePointInTime,
  TimeOffsetType,
  TimeTruncationType,
  RelativeTimeTransformation,
} from "./time-range";

// reference timestamp method
export function getPresentTime() {
  return new Date();
}

export function getLatestDataTimestamp(allTimeRange: TimeRange) {
  return allTimeRange.end;
}

// Period anchor methods
export function getStartOfPeriod(period: Period, referenceTime: Date) {
  const date = DateTime.fromJSDate(referenceTime);
  return date.startOf(TIME_UNIT[period]).toJSDate();
}

export function getEndOfPeriod(period: Period, referenceTime: Date) {
  const date = DateTime.fromJSDate(referenceTime);
  return date.startOf(TIME_UNIT[period]).toJSDate();
}

// offset methods
export function getOffset(
  referenceTime: Date,
  duration: string,
  direction: TimeOffsetType
) {
  const durationObj = Duration.fromISO(duration);
  return DateTime.fromJSDate(referenceTime)
    [direction === TimeOffsetType.ADD ? "plus" : "minus"](durationObj)
    .toJSDate();
}

export function getTimeWidth(start: Date, end: Date) {
  return DateTime.fromJSDate(end).diff(
    DateTime.fromJSDate(start),
    "milliseconds"
  ).milliseconds;
}

export function ISOToMilliseconds(duration: string) {
  return Duration.fromISO(duration).as("milliseconds");
}

/** Loops through all of the offset transformations and applies each of them
 * to the supplied referenceTime.
 * FIXME: write tests for this function
 */
export function getAbsoluteDateFromTransformations(
  referenceTime: Date,
  transformations: RelativeTimeTransformation[]
) {
  let absoluteTime = referenceTime;
  for (const transformation of transformations) {
    /** add or subtract an offset duration from the datetime. Otherwise, perform a truncation transformation. */
    if ("operationType" in transformation) {
      absoluteTime = getOffset(
        absoluteTime,
        transformation.duration,
        transformation.operationType
      );
    } else if (
      transformation.truncationType === TimeTruncationType.START_OF_PERIOD
    ) {
      absoluteTime = getStartOfPeriod(transformation.period, absoluteTime);
    } else if (
      transformation.truncationType === TimeTruncationType.END_OF_PERIOD
    ) {
      absoluteTime = getEndOfPeriod(transformation.period, absoluteTime);
    }
  }

  return absoluteTime;
}

export function relativePointInTimeToAbsolute(
  referenceTime: Date,
  start: string | RelativePointInTime,
  end: string | RelativePointInTime
) {
  let startDate: Date;
  let endDate: Date;

  if (typeof start === "string") startDate = new Date(start);
  else
    startDate = getAbsoluteDateFromTransformations(
      referenceTime,
      start.transformation
    );

  if (typeof end === "string") endDate = new Date(end);
  else
    endDate = getAbsoluteDateFromTransformations(
      referenceTime,
      end.transformation
    );

  return {
    startDate,
    endDate,
  };
}
