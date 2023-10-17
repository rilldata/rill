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
  PeriodAndUnits,
  PeriodToUnitsMap,
} from "@rilldata/web-common/lib/time/config";
import type { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import {
  ReferencePoint,
  RelativePointInTime,
  RelativeTimeTransformation,
  TimeGrain,
  TimeTruncationType,
} from "../types";
import { DateTime, Duration, DurationLike } from "luxon";
import { Period, TimeOffsetType, TimeUnit } from "../types";

/** Returns the current time */
export function getPresentTime() {
  return new Date();
}

/** Returns the start of the period for the given reference time. */
export function getStartOfPeriod(
  referenceTime: Date,
  period: Period,
  zone = "Etc/UTC"
) {
  const date = DateTime.fromJSDate(referenceTime, { zone });
  return date.startOf(TimeUnit[period]).toJSDate();
}

/** Returns the end of the period that the reference time is in. */
export function getEndOfPeriod(
  referenceTime: Date,
  period: Period,
  zone = "Etc/UTC"
) {
  const date = DateTime.fromJSDate(referenceTime, { zone });
  return date.endOf(TimeUnit[period]).toJSDate();
}

/** Offsets a date by a certain ISO duration amount. */
export function getOffset(
  referenceTime: Date,
  duration: string,
  direction: TimeOffsetType,
  zone = "Etc/UTC"
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
  transformations: RelativeTimeTransformation[],
  zone = "Etc/UTC"
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
      absoluteTime = getStartOfPeriod(
        absoluteTime,
        transformation.period,
        zone
      );
    } else if (
      transformation.truncationType === TimeTruncationType.END_OF_PERIOD
    ) {
      absoluteTime = getEndOfPeriod(absoluteTime, transformation.period, zone);
    }
  }

  return absoluteTime;
}

// FIXME: we might end up deprecating this function.
export function relativePointInTimeToAbsolute(
  referenceTime: Date,
  start: string | RelativePointInTime,
  end: string | RelativePointInTime,
  zone: string
) {
  let startDate: Date;
  let endDate: Date;
  if (typeof start === "string") startDate = new Date(start);
  else {
    if (start.reference === ReferencePoint.NOW) {
      referenceTime = getPresentTime();
    } else if (start.reference === ReferencePoint.MIN_OF_LATEST_DATA_AND_NOW) {
      referenceTime = new Date(
        Math.min(referenceTime.getTime(), getPresentTime().getTime())
      );
    }

    startDate = transformDate(referenceTime, start.transformation, zone);
  }

  if (typeof end === "string") endDate = new Date(end);
  else {
    if (end.reference === ReferencePoint.NOW) {
      referenceTime = getPresentTime();
    } else if (end.reference === ReferencePoint.MIN_OF_LATEST_DATA_AND_NOW) {
      referenceTime = new Date(
        Math.min(referenceTime.getTime(), getPresentTime().getTime())
      );
    }
    endDate = transformDate(referenceTime, end.transformation, zone);
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

export function subtractFromPeriod(duration: Duration, period: Period) {
  if (!PeriodToUnitsMap[period]) return duration;
  return duration.minus({ [PeriodToUnitsMap[period]]: 1 });
}

export function getAdditionalOffset(
  isoDuration: string,
  smallestGrain: TimeGrain,
  multiple: number
) {
  isoDuration = Duration.fromISO(isoDuration).shiftToAll().toISO(); // normalise the range

  const smallerDuration = stripLargerUnits(isoDuration, smallestGrain.grain);
  const largerDuration = Duration.fromISO(smallestGrain.duration);
  const offset =
    largerDuration.as("milliseconds") * multiple -
    smallerDuration.as("milliseconds");
  if (offset < 0) return "";

  return Duration.fromMillis(offset)
    .shiftTo("days", "hours", "minutes", "seconds")
    .toISO();
}

function stripLargerUnits(isoDuration: string, smallestGrain: V1TimeGrain) {
  const duration = Duration.fromISO(isoDuration);
  const newDurationLike: DurationLike = {};

  for (const { grain, unit } of PeriodAndUnits) {
    if (grain === smallestGrain) {
      break;
    }
    newDurationLike[unit] = duration[unit];
  }

  return Duration.fromDurationLike(newDurationLike);
}
