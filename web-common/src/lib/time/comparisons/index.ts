import { TIME_COMPARISON } from "@rilldata/web-common/lib/time/config";
import { prettyFormatTimeRange } from "@rilldata/web-common/lib/time/ranges/formatter.ts";
import { humaniseISODuration } from "@rilldata/web-common/lib/time/ranges/iso-ranges";
import {
  V1TimeGrain,
  type V1TimeRange,
} from "@rilldata/web-common/runtime-client/gen/index.schemas";
import { DateTime, Duration, Interval } from "luxon";
import { getTimeWidth } from "../transforms";
import {
  type RelativeTimeTransformation,
  TimeComparisonOption,
  TimeOffsetType,
  TimeRangePreset,
} from "../types";

export function getComparisonTransform(
  start: Date,
  end: Date,
  comparison: TimeComparisonOption,
): RelativeTimeTransformation {
  if (
    comparison === TimeComparisonOption.CONTIGUOUS ||
    comparison === TimeComparisonOption.CUSTOM ||
    // @ts-expect-error Hack for `comparison` not being the correct type
    comparison === TimeRangePreset.CUSTOM
  ) {
    /** for custom & otherwise-contiguous comparisons,
     * we will calculate the width of the time range
     * and then subtract it from the start and end dates. */

    return {
      operationType: TimeOffsetType.SUBTRACT,
      duration: Interval.fromDateTimes(start, end).toDuration().toISO(),
    };
  } else {
    // map to a distinct Period-like TimeComparisonOption (e.g. "P1D")
    return {
      operationType: TimeOffsetType.SUBTRACT,
      duration: TIME_COMPARISON[comparison].offsetIso,
    };
  }
}

/**
 * get the comparison range for a scrub such that
 * it aligns with the scrub start and end.
 */
export function getComparionRangeForScrub(
  start: Date,
  end: Date,
  comparisonStart: Date,
  comparisonEnd: Date,
  scrubStart: Date,
  scrubEnd: Date,
) {
  // validate if selected range and comparison range are of equal width
  if (
    getTimeWidth(start, end) !== getTimeWidth(comparisonStart, comparisonEnd)
  ) {
    // TODO: Have better handling on uneven widths caused by custom comparisons
    return { start: comparisonStart, end: comparisonEnd };
  } else {
    const startMOffset = getTimeWidth(start, scrubStart);
    const endMOffset = getTimeWidth(scrubEnd, end);

    return {
      start: new Date(comparisonStart.getTime() + startMOffset),
      end: new Date(comparisonEnd.getTime() - endMOffset),
    };
  }
}

// Returns true if the comparison range is inside the bounds.
export function isComparisonInsideBounds(
  // the earliest possible Date for this data set
  boundStart: Date,
  // the latest possible Date for this data set
  boundEnd: Date,
  start: Date,
  end: Date,
  comparison: TimeComparisonOption,
  timeZone: string,
) {
  const interval = Interval.fromDateTimes(
    DateTime.fromJSDate(start),
    DateTime.fromJSDate(end),
  );
  if (!interval.isValid) {
    return false;
  }
  const maxInterval = Interval.fromDateTimes(
    DateTime.fromJSDate(boundStart),
    DateTime.fromJSDate(boundEnd),
  );
  // compute comparison start and ends.
  const comparisonInterval = getComparisonInterval(
    interval,
    comparison,
    timeZone,
  );

  if (!comparisonInterval || !comparisonInterval.isValid) {
    return false;
  }
  // check if comparison bounds are inside the bounds.
  return (
    maxInterval.contains(comparisonInterval.start) &&
    maxInterval.contains(comparisonInterval.end)
  );
}

export function isRangeLargerThanDuration(
  start: Date,
  end: Date,
  duration: string,
) {
  if (duration === TimeComparisonOption.CONTIGUOUS.toString()) {
    return false;
  }

  // To account for possible leap years
  if (duration === "P1Y") {
    return end.getFullYear() - start.getFullYear() > 1;
  }

  return (
    Interval.fromDateTimes(start, end).toDuration().toMillis() >
    Duration.fromISO(duration).toMillis()
  );
}

// Checks if last period is a duplicate comparison.
function isLastPeriodDuplicate(
  start: Date,
  end: Date,
  comparisonOptions: Array<TimeComparisonOption>,
  zone: string,
) {
  const interval = Interval.fromDateTimes(
    DateTime.fromJSDate(start, { zone }),
    DateTime.fromJSDate(end, { zone }),
  );

  if (!interval.isValid) {
    return false;
  }
  const lastPeriod = getComparisonInterval(
    interval,
    TimeComparisonOption.CONTIGUOUS,
    zone,
  );

  if (!lastPeriod || !lastPeriod.isValid) {
    return false;
  }

  comparisonOptions = comparisonOptions.filter(
    (option) =>
      option !== TimeComparisonOption.CUSTOM &&
      option !== TimeComparisonOption.CONTIGUOUS,
  );

  return comparisonOptions.some((option) => {
    const periodComparison = getComparisonInterval(interval, option, zone);

    return (
      lastPeriod &&
      lastPeriod.isValid &&
      periodComparison &&
      lastPeriod.start.equals(periodComparison?.start) &&
      lastPeriod.end.equals(periodComparison?.end)
    );
  });
}

/** get the available comparison options for a selected time range + the boundary range.
 * This is used to populate the comparison dropdown.
 * We need to check boundary conditions on all sides, but ultimately the two checks per comparison option are:
 * 1. is the comparison range inside the bounds?
 * 2. is the comparison range larger than the selected time range?
 */
export function getAvailableComparisonsForTimeRange(
  // the earliest and latest possible Dates for this data set
  boundStart: Date,
  boundEnd: Date,
  // start and end Dates for the currently focused time range
  start: Date,
  end: Date,
  comparisonOptions: TimeComparisonOption[],
  timezone: string,
) {
  let comparisons = comparisonOptions.filter((comparison) => {
    if (comparison === TimeComparisonOption.CUSTOM) {
      return false;
    }

    return (
      isComparisonInsideBounds(
        boundStart,
        boundEnd,
        start,
        end,
        // treat a custom comparison as contiguous.
        comparison,
        timezone,
      ) &&
      !isRangeLargerThanDuration(
        start,
        end,
        TIME_COMPARISON[comparison].offsetIso,
      )
    );
  });

  if (isLastPeriodDuplicate(start, end, comparisonOptions, timezone)) {
    comparisons = comparisons.filter(
      (comparison) => comparison !== TimeComparisonOption.CONTIGUOUS,
    );
  }

  return comparisons;
}

/** A convenience function that gets comparison range and states whether it is within bounds. */
export function getTimeComparisonParametersForComponent(
  comparisonOption: TimeComparisonOption | undefined,
  boundStart: Date | null | undefined,
  boundEnd: Date | null | undefined,
  currentStart: Date | null | undefined,
  currentEnd: Date | null | undefined,
  timezone: string,
) {
  if (
    !comparisonOption ||
    !boundStart ||
    !currentStart ||
    !currentEnd ||
    !boundEnd
  ) {
    return {
      start: undefined,
      end: undefined,
      isComparisonRangeAvailable: false,
    };
  }

  const interval = Interval.fromDateTimes(
    DateTime.fromJSDate(currentStart).setZone(timezone),
    DateTime.fromJSDate(currentEnd).setZone(timezone),
  );

  if (!interval.isValid) {
    return {
      start: undefined,
      end: undefined,
      isComparisonRangeAvailable: false,
    };
  }

  const comparisonInterval = getComparisonInterval(
    interval,
    comparisonOption,
    timezone,
  );

  const isComparisonRangeAvailable = isComparisonInsideBounds(
    boundStart,
    boundEnd,
    interval.start.toJSDate(),
    interval.end.toJSDate(),
    comparisonOption,
    timezone,
  );
  const start = comparisonInterval?.start.toJSDate();
  const end = comparisonInterval?.end.toJSDate();
  return {
    start,
    end,
    isComparisonRangeAvailable,
  };
}

export function getComparisonLabel(comparisonTimeRange: V1TimeRange) {
  if (
    (!comparisonTimeRange.isoOffset && !comparisonTimeRange.expression) ||
    comparisonTimeRange.isoOffset === TimeRangePreset.CUSTOM
  ) {
    return prettyFormatTimeRange(
      Interval.fromDateTimes(
        DateTime.fromISO(comparisonTimeRange.start ?? "").setZone("UTC"),
        DateTime.fromISO(comparisonTimeRange.end ?? "").setZone("UTC"),
      ),
      V1TimeGrain.TIME_GRAIN_UNSPECIFIED,
    );
  }
  switch (true) {
    case comparisonTimeRange.isoOffset === TimeRangePreset.ALL_TIME:
      return "All time";
    case comparisonTimeRange.isoDuration === comparisonTimeRange.isoOffset ||
      comparisonTimeRange.expression?.toLowerCase()?.endsWith("offset pp"):
      return "Previous period";
    case comparisonTimeRange.isoOffset &&
      comparisonTimeRange.isoOffset in TIME_COMPARISON:
      return TIME_COMPARISON[comparisonTimeRange.isoOffset].label;
    case comparisonTimeRange.expression &&
      comparisonTimeRange.expression in TIME_COMPARISON:
      return TIME_COMPARISON[comparisonTimeRange.expression].label;
    default:
      return `Last ${humaniseISODuration(comparisonTimeRange.isoOffset ?? comparisonTimeRange.expression ?? "")}`;
  }
}

export function getComparisonInterval(
  interval: Interval<true> | undefined,
  comparisonRange: string | undefined,
  activeTimeZone: string,
): Interval<true> | undefined {
  if (!interval || !comparisonRange) return undefined;

  let comparisonInterval: Interval | undefined = undefined;

  const COMPARISON_DURATIONS = {
    "rill-PP": interval.toDuration(),
    "rill-PD": { days: 1 },
    "rill-PW": { weeks: 1 },
    "rill-PM": { months: 1 },
    "rill-PQ": { quarter: 1 },
    "rill-PY": { years: 1 },
  };

  const duration =
    COMPARISON_DURATIONS[comparisonRange as keyof typeof COMPARISON_DURATIONS];

  if (duration) {
    comparisonInterval = Interval.fromDateTimes(
      interval.start.minus(duration),
      interval.end.minus(duration),
    );
    // If this didn't work, it's likely because we fell on a boundary case
    // such as looking at March 31st and subtracting a month
    // We can fall back to adding the duration to the start date
    if (!comparisonInterval.isValid) {
      comparisonInterval = Interval.fromDateTimes(
        interval.start.minus(duration),
        interval.start.minus(duration).plus(interval.toDuration()),
      );
    }
  } else {
    const normalizedRange = comparisonRange.replace(",", "/");
    comparisonInterval = Interval.fromISO(normalizedRange).mapEndpoints((dt) =>
      dt.setZone(activeTimeZone),
    );
  }

  return comparisonInterval.isValid ? comparisonInterval : undefined;
}
