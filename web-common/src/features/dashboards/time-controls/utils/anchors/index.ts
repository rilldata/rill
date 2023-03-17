import { DateTime, Duration } from "luxon";
import type { TimeRange } from "../../time-control-types";
import {
  Period,
  ReferencePoint,
  RelativePointInTime,
  RelativeTimeTransformation,
  TimeOffsetType,
  TimeTruncationType,
  TimeUnit,
} from "../time-types";

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
  return date.startOf(TimeUnit[period]).toJSDate();
}

export function getEndOfPeriod(period: Period, referenceTime: Date) {
  const date = DateTime.fromJSDate(referenceTime);
  return date.endOf(TimeUnit[period]).toJSDate();
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

/**
 * @returns the width of the time range defined by start and end in milliseconds
 */
export function getTimeWidth(start: Date, end: Date) {
  return end.getTime() - start.getTime();
}

export function ISOToMilliseconds(duration: string) {
  return Duration.fromISO(duration).as("milliseconds");
}

/**
 * Returns true if the range defined by start and end is completely
 * inside the range defined by otherStart and otherEnd.
 */
export function isRangeInsideOther(
  start: Date,
  end: Date,
  otherStart: Date,
  otherEnd: Date
) {
  return start >= otherStart && end <= otherEnd;
}

/** Loops through all of the offset transformations and applies each of them
 * to the supplied referenceTime. The transformations are applied in the orer
 * they appear; we define these in a way that can later be serialized in
 * a configuration file.
 */
export function transformDate(
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

// Move to time-range as not pure?
export function relativePointInTimeToAbsolute(
  referenceTime: Date,
  start: string | RelativePointInTime,
  end: string | RelativePointInTime
) {
  let startDate: Date;
  let endDate: Date;

  if (typeof start === "string") startDate = new Date(start);
  else {
    if (start.reference === ReferencePoint.NOW)
      referenceTime = getPresentTime();
    startDate = transformDate(referenceTime, start.transformation);
  }

  if (typeof end === "string") endDate = new Date(end);
  else {
    if (end.reference === ReferencePoint.NOW) referenceTime = getPresentTime();
    endDate = transformDate(referenceTime, end.transformation);
  }

  return {
    startDate,
    endDate,
  };
}
