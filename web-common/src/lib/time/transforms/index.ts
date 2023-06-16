/**
 * Utility functions around transforming Date objects in ways
 * that are useful to the Rill dashboard. The core function to use here is
 * transformDate, which takes a reference time and a list of transformations
 * to apply to that reference time. The transformations are applied in the order
 * they appear in the list.
 *
 * We are opting to define transformations in a way that can be serialized
 * in a configuration file.
 */
import {
  ReferencePoint,
  RelativePointInTime,
  RelativeTimeTransformation,
  TimeTruncationType,
} from "../types";

import { DateTime, Duration } from "luxon";
import { Period, TimeOffsetType, TimeUnit } from "../types";

/** Returns the current time. Might be deprecated later. */
export function getPresentTime() {
  return new Date();
}

/** Returns the start of the period for the given reference time. */
export function getStartOfPeriod(
  referenceTime: Date,
  period: Period,
  zone = "utc"
) {
  const date = DateTime.fromJSDate(referenceTime, { zone });
  return date.startOf(TimeUnit[period]).toJSDate();
}

/** Returns the end of the period that the reference time is in. */
export function getEndOfPeriod(
  referenceTime: Date,
  period: Period,
  zone = "utc"
) {
  const date = DateTime.fromJSDate(referenceTime, { zone });
  return date.endOf(TimeUnit[period]).toJSDate();
}

/** Offsets a date by a certain ISO duration amount. */
export function getOffset(
  referenceTime: Date,
  duration: string,
  direction: TimeOffsetType,
  zone = "utc"
) {
  const durationObj = Duration.fromISO(duration);
  return DateTime.fromJSDate(referenceTime, { zone })
    [direction === TimeOffsetType.ADD ? "plus" : "minus"](durationObj)
    .toJSDate();
}

/** The width of the time range defined by start and end in milliseconds */
export function getTimeWidth(start: Date, end: Date) {
  return end.getTime() - start.getTime();
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
      absoluteTime = getStartOfPeriod(absoluteTime, transformation.period);
    } else if (
      transformation.truncationType === TimeTruncationType.END_OF_PERIOD
    ) {
      absoluteTime = getEndOfPeriod(absoluteTime, transformation.period);
    }
  }

  return absoluteTime;
}

// FIXME: we might end up deprecating this function.
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

/** Returns the ISO Duration as a multiple of given duration  */
export function getDurationMultiple(duration: string, multiple: number) {
  const durationObj = Duration.fromISO(duration);
  const totalDuration = durationObj.as("milliseconds");
  const newDuration = totalDuration * multiple;
  return Duration.fromMillis(newDuration)
    .shiftTo("days", "hours", "minutes", "seconds")
    .toISO();
}
