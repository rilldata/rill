import { Duration, Interval } from "luxon";
import { transformDate } from "../transforms";
import {
  RelativeTimeTransformation,
  TimeComparisonOption,
  TimeOffsetType,
} from "../types";

export function getComparisonTransform(
  start: Date,
  end: Date,
  comparison: TimeComparisonOption
): RelativeTimeTransformation {
  if (
    comparison === TimeComparisonOption.CONTIGUOUS ||
    comparison === TimeComparisonOption.CUSTOM
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
      duration: comparison as TimeComparisonOption,
    };
  }
}

/** takes a start and end, and performs transformDate accordingly.
 * Contiguous periods (for instance "last 6 hours" or a custom range) is handled
 * a bit differently.
 */
export function getComparisonRange(
  start: Date,
  end: Date,
  comparison: TimeComparisonOption
) {
  const transform = getComparisonTransform(start, end, comparison);
  return {
    start: transformDate(start, [transform]),
    end: transformDate(end, [transform]),
  };
}

// Returns true if the comparison range is inside the bounds.
export function isComparisonInsideBounds(
  // the earliest possible Date for this data set
  boundStart: Date,
  // the latest possible Date for this data set
  boundEnd: Date,
  start: Date,
  end: Date,
  comparison: TimeComparisonOption
) {
  // compute comparison start and ends.
  const { start: comparisonStart, end: comparisonEnd } = getComparisonRange(
    start,
    end,
    comparison
  );
  // check if comparison bounds are inside the bounds.
  return comparisonStart >= boundStart && comparisonEnd <= boundEnd;
}

export function isRangeLargerThanDuration(
  start: Date,
  end: Date,
  duration: string
) {
  if (duration === TimeComparisonOption.CONTIGUOUS) {
    return false;
  }
  return (
    Interval.fromDateTimes(start, end).toDuration().toMillis() >
    Duration.fromISO(duration).toMillis()
  );
}

// Checks if last period is a duplicate comparison.
function isLastPeriodDuplicate(start: Date, end: Date) {
  const lastPeriod = getComparisonRange(
    start,
    end,
    TimeComparisonOption.CONTIGUOUS
  );

  const comparisonOptions = [...Object.values(TimeComparisonOption)].filter(
    (option) =>
      option !== TimeComparisonOption.CUSTOM &&
      option !== TimeComparisonOption.CONTIGUOUS
  );

  return comparisonOptions.some((option) => {
    const { start: comparisonStart, end: comparisonEnd } = getComparisonRange(
      start,
      end,
      option
    );
    return (
      comparisonStart.getTime() === lastPeriod.start.getTime() &&
      comparisonEnd.getTime() === lastPeriod.end.getTime()
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
  // the set of additional comparisons we should keep in mind, but not
  // necessarily the right widt.
  acceptedComparisons: TimeComparisonOption[] = []
) {
  let comparisons = comparisonOptions.filter((comparison) => {
    if (comparison === TimeComparisonOption.CUSTOM) {
      return false;
    }

    return (
      acceptedComparisons.includes(comparison) ||
      (isComparisonInsideBounds(
        boundStart,
        boundEnd,
        start,
        end,
        // treat a custom comparison as contiguous.
        comparison
      ) &&
        !isRangeLargerThanDuration(start, end, comparison))
    );
  });

  if (isLastPeriodDuplicate(start, end)) {
    comparisons = comparisons.filter(
      (comparison) => comparison !== TimeComparisonOption.CONTIGUOUS
    );
  }
  return comparisons;
}

/** A convenience function that gets comparison range and states whether it is within bounds. */
export function getTimeComparisonParametersForComponent(
  comparisonOption: TimeComparisonOption,
  boundStart,
  boundEnd,
  currentStart,
  currentEnd
) {
  const { start, end } = getComparisonRange(
    currentStart,
    currentEnd,
    comparisonOption
  );

  const isComparisonRangeAvailable = isComparisonInsideBounds(
    boundStart,
    boundEnd,
    start,
    end,
    comparisonOption
  );

  return {
    start,
    end,
    isComparisonRangeAvailable,
  };
}
