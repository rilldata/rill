import {
  formatRillTimeWallClock,
  parseRillTime,
} from "@rilldata/web-common/features/dashboards/url-state/time-ranges/parser.ts";
import { RillAllTimeInterval } from "@rilldata/web-common/features/dashboards/url-state/time-ranges/RillTime.ts";
import { TIME_COMPARISON } from "@rilldata/web-common/lib/time/config.ts";
import { DateTime } from "luxon";
import { isoDurationToFullTimeRange } from "@rilldata/web-common/lib/time/ranges/iso-ranges";
import {
  type DashboardTimeControls,
  TimeComparisonOption,
  TimeRangePreset,
} from "@rilldata/web-common/lib/time/types";
import {
  V1TimeGrain,
  type V1TimeRange,
  type V1TimeRangeSummary,
} from "@rilldata/web-common/runtime-client";
import type { TimeFilterManager } from "@rilldata/web-common/features/dashboards/time-controls/TimeFilterManager.svelte.ts";
import { ExploreStateURLParams } from "@rilldata/web-common/features/dashboards/url-state/url-params.ts";

// Temporary fix to split previous complete ranges to duration and round to grain to get it working on backend
// TODO: Eventually we should support this in the backend.
export const PreviousCompleteRangeMap: Partial<
  Record<TimeRangePreset, V1TimeRange>
> = {
  [TimeRangePreset.YESTERDAY_COMPLETE]: {
    isoDuration: "P1D",
    roundToGrain: V1TimeGrain.TIME_GRAIN_DAY,
  },
  [TimeRangePreset.PREVIOUS_WEEK_COMPLETE]: {
    isoDuration: "P1W",
    roundToGrain: V1TimeGrain.TIME_GRAIN_WEEK,
  },
  [TimeRangePreset.PREVIOUS_MONTH_COMPLETE]: {
    isoDuration: "P1M",
    roundToGrain: V1TimeGrain.TIME_GRAIN_MONTH,
  },
  [TimeRangePreset.PREVIOUS_QUARTER_COMPLETE]: {
    isoDuration: "P3M",
    roundToGrain: V1TimeGrain.TIME_GRAIN_QUARTER,
  },
  [TimeRangePreset.PREVIOUS_YEAR_COMPLETE]: {
    isoDuration: "P1Y",
    roundToGrain: V1TimeGrain.TIME_GRAIN_YEAR,
  },
};

// We are manually sending in duration, offset and round to grain for previous complete ranges.
// This is to map back that split
const PreviousCompleteRangeReverseMap: Record<string, TimeRangePreset> = {};
for (const preset in PreviousCompleteRangeMap) {
  const range: V1TimeRange = PreviousCompleteRangeMap[preset];
  PreviousCompleteRangeReverseMap[
    `${range.isoDuration}_${range.isoOffset ?? ""}_${range.roundToGrain}`
  ] = preset as TimeRangePreset;
}

export function mapV1TimeRangeToSelectedTimeRange(
  timeRange: V1TimeRange,
  timeRangeSummary: V1TimeRangeSummary,
  end: string,
) {
  let selectedTimeRange: DashboardTimeControls;
  let duration = timeRange.isoDuration;

  const fullRangeKey = `${timeRange.isoDuration ?? ""}_${timeRange.isoOffset ?? ""}_${timeRange.roundToGrain ?? ""}`;
  if (fullRangeKey in PreviousCompleteRangeReverseMap) {
    duration = PreviousCompleteRangeReverseMap[fullRangeKey];
  }

  if (timeRange.start && timeRange.end) {
    selectedTimeRange = {
      name: TimeRangePreset.CUSTOM,
      start: new Date(timeRange.start),
      end: new Date(timeRange.end),
    };
  } else if (timeRange.expression) {
    try {
      const rt = parseRillTime(timeRange.expression);
      selectedTimeRange = {
        name: rt.toString(),
        interval: rt.byGrain ?? rt.rangeGrain,
      } as DashboardTimeControls;
    } catch {
      return undefined;
    }
  } else if (duration && timeRangeSummary.min) {
    selectedTimeRange = isoDurationToFullTimeRange(
      duration,
      new Date(timeRangeSummary.min),
      new Date(end),
    );
  } else {
    return undefined;
  }

  if (!selectedTimeRange.interval) {
    selectedTimeRange.interval = timeRange.roundToGrain;
  }

  return selectedTimeRange;
}

export function mapV1TimeRangeToSelectedComparisonTimeRange(
  timeRange: V1TimeRange,
  timeRangeSummary: V1TimeRangeSummary,
  end: string,
) {
  let selectedTimeRange: DashboardTimeControls;
  let duration = timeRange.isoOffset;

  const fullRangeKey = `${timeRange.isoDuration ?? ""}_${timeRange.isoOffset ?? ""}_${timeRange.roundToGrain ?? ""}`;
  if (fullRangeKey in PreviousCompleteRangeReverseMap) {
    duration = PreviousCompleteRangeReverseMap[fullRangeKey];
  }

  // A comparison sent as the base expression with an offset, such as "P1D offset -1W"
  const offsetOption = timeRange.expression
    ? getComparisonOptionFromExpression(timeRange.expression)
    : undefined;
  if (offsetOption) {
    duration = offsetOption;
  }

  if (timeRange.start && timeRange.end) {
    selectedTimeRange = {
      name: TimeComparisonOption.CUSTOM,
      start: new Date(timeRange.start),
      end: new Date(timeRange.end),
    };
  } else if (
    offsetOption === TimeComparisonOption.CONTIGUOUS ||
    (!offsetOption && timeRange.isoOffset === timeRange.isoDuration)
  ) {
    // Previous period is when offset = duration
    selectedTimeRange = {
      name: TimeComparisonOption.CONTIGUOUS,
    } as DashboardTimeControls;
  } else if (duration && timeRangeSummary.min) {
    let isoDuration = duration;
    if (duration in TIME_COMPARISON) {
      isoDuration = TIME_COMPARISON[duration].offsetIso;
    }
    selectedTimeRange = isoDurationToFullTimeRange(
      isoDuration,
      new Date(timeRangeSummary.min),
      new Date(end),
    );
    selectedTimeRange.name = duration as TimeComparisonOption;
  } else {
    return undefined;
  }

  selectedTimeRange.interval = timeRange.roundToGrain;

  return selectedTimeRange;
}

export function mapTimeManagerRangeToV1TimeRange(
  timeFilterManager: TimeFilterManager | undefined,
): V1TimeRange | undefined {
  // The requested range rather than the resolved `timeRange`:
  // it is what the user selected even before it resolves, or when it fails to resolve,
  // and the resolved range lags behind it while a new range resolves.
  const requestedTimeRange = timeFilterManager?.requestedTimeRange;
  if (!requestedTimeRange) return undefined;
  try {
    parseRillTime(requestedTimeRange);
    return {
      expression: requestedTimeRange,
      timeZone: timeFilterManager.timeZone,
    };
  } catch {
    // Legacy presets that are not rill time are handled below.
  }

  // Legacy rill-time handling
  const timeRange: V1TimeRange = {};
  switch (requestedTimeRange) {
    case TimeRangePreset.CUSTOM:
      // The full range, since `timeStart`/`timeEnd` follow a scrub.
      timeRange.start = timeFilterManager.interval?.start.toUTC().toISO();
      timeRange.end = timeFilterManager.interval?.end.toUTC().toISO();
      break;

    default:
      if (requestedTimeRange in PreviousCompleteRangeMap) {
        const prevCompleteTimeRange: V1TimeRange | undefined =
          PreviousCompleteRangeMap[requestedTimeRange];
        // Backend doesn't support previous complete ranges since it has offset built in.
        // We add the offset manually as a workaround for now
        timeRange.isoDuration = prevCompleteTimeRange?.isoDuration;
        timeRange.isoOffset = prevCompleteTimeRange?.isoOffset;
        timeRange.roundToGrain = prevCompleteTimeRange?.roundToGrain;
      } else {
        timeRange.isoDuration = requestedTimeRange;
      }
      break;
  }

  timeRange.timeZone = timeFilterManager.timeZone;

  return timeRange;
}

export function mapTimeManagerComparisonRangeToV1TimeRange(
  timeFilterManager: TimeFilterManager | undefined,
  timeRange: V1TimeRange | undefined,
): V1TimeRange | undefined {
  if (
    !timeRange ||
    !timeFilterManager ||
    !timeFilterManager.showComparison ||
    !timeFilterManager.comparisonTimeRange
  ) {
    return undefined;
  }

  // A custom comparison is stored as `start to end`, so it is sent as the resolved interval.
  // The full window, since `comparisonTimeStart`/`comparisonTimeEnd` follow a scrub.
  if (timeFilterManager.isCustomComparison) {
    return {
      start: timeFilterManager.comparisonInterval?.start.toUTC().toISO(),
      end: timeFilterManager.comparisonInterval?.end.toUTC().toISO(),
    };
  }

  let isoDuration = timeRange.isoDuration;
  const name = timeFilterManager.comparisonTimeRange;

  if (timeRange.expression && TIME_COMPARISON[name]?.rillTimeOffset) {
    const rt = parseRillTime(timeRange.expression);
    // The runtime can't offset all time, so all time keeps the legacy fields.
    if (!(rt.interval instanceof RillAllTimeInterval)) {
      return {
        expression:
          rt.toString() + " offset " + TIME_COMPARISON[name]?.rillTimeOffset,
      };
    }
    isoDuration = timeRange.expression;
  }

  const comparisonTimeRange: V1TimeRange = {};
  switch (name) {
    default:
      comparisonTimeRange.isoOffset = name;
      comparisonTimeRange.isoDuration = isoDuration;
      break;
    case TimeComparisonOption.CONTIGUOUS:
      comparisonTimeRange.isoOffset = comparisonTimeRange.isoDuration =
        isoDuration;
      break;
  }
  return comparisonTimeRange;
}

/**
 * Sets the time filter url params a {@link TimeFilterManager} is seeded with from API time ranges,
 * e.g. those saved in a report's or an alert's aggregation request.
 */
export function setTimeFilterParamsFromV1TimeRanges(
  params: URLSearchParams,
  timeRange: V1TimeRange | undefined,
  comparisonTimeRange: V1TimeRange | undefined,
) {
  if (!timeRange) return;

  const { timeRange: timeRangeParam, timeZone } =
    mapV1TimeRangeToTimeFilterParams(timeRange);
  if (!timeRangeParam) return;
  params.set(ExploreStateURLParams.TimeRange, timeRangeParam);
  if (timeZone) {
    params.set(ExploreStateURLParams.TimeZone, timeZone);
  }

  if (comparisonTimeRange) {
    const comparisonParam = mapV1TimeRangeToComparisonTimeOption(
      timeRange,
      comparisonTimeRange,
    );
    if (comparisonParam) {
      params.set(ExploreStateURLParams.ComparisonTimeRange, comparisonParam);
    }
  }
}

/**
 * Maps an API time range to the rill time range and zone a {@link TimeFilterManager} is seeded with.
 * The zone is undefined when the dashboard's default zone should apply.
 */
function mapV1TimeRangeToTimeFilterParams(timeRange: V1TimeRange): {
  timeRange: string | undefined;
  timeZone: string | undefined;
} {
  if (timeRange.start && timeRange.end) {
    // An absolute range without a saved zone is pinned to UTC instead of the dashboard's default zone,
    // since its endpoints are UTC instants.
    const timeZone = timeRange.timeZone ?? "UTC";
    // Rill time reads absolute times as wall-clock times in the zone.
    const start = DateTime.fromISO(timeRange.start, { zone: "UTC" });
    const end = DateTime.fromISO(timeRange.end, { zone: "UTC" });
    if (!start.isValid || !end.isValid) {
      return { timeRange: `${timeRange.start} to ${timeRange.end}`, timeZone };
    }
    return {
      timeRange: `${formatRillTimeWallClock(start, timeZone)} to ${formatRillTimeWallClock(end, timeZone)}`,
      timeZone,
    };
  }

  const fullRangeKey = `${timeRange.isoDuration ?? ""}_${timeRange.isoOffset ?? ""}_${timeRange.roundToGrain ?? ""}`;
  const duration =
    PreviousCompleteRangeReverseMap[fullRangeKey] ?? timeRange.isoDuration;
  return {
    timeRange: duration ?? timeRange.expression,
    timeZone: timeRange.timeZone,
  };
}

export function mapV1TimeRangeToComparisonTimeOption(
  primaryTimeRange: V1TimeRange,
  timeRange: V1TimeRange,
) {
  let duration = timeRange.isoOffset;

  const fullRangeKey = `${timeRange.isoDuration ?? ""}_${timeRange.isoOffset ?? ""}_${timeRange.roundToGrain ?? ""}`;
  if (fullRangeKey in PreviousCompleteRangeReverseMap) {
    duration = PreviousCompleteRangeReverseMap[fullRangeKey];
  }

  if (timeRange.start && timeRange.end) {
    return `${timeRange.start} to ${timeRange.end}`;
  } else if (
    timeRange.isoOffset &&
    timeRange.isoOffset === timeRange.isoDuration
  ) {
    return TimeComparisonOption.CONTIGUOUS;
  } else if (timeRange.expression) {
    if (
      primaryTimeRange.expression &&
      timeRange.expression.startsWith(primaryTimeRange.expression)
    ) {
      const offset = timeRange.expression.replace(
        primaryTimeRange.expression + " offset ",
        "",
      );
      for (const comparisonOption in TIME_COMPARISON) {
        if (TIME_COMPARISON[comparisonOption].rillTimeOffset === offset) {
          return comparisonOption;
        }
      }

      return timeRange.expression;
    }

    return timeRange.expression;
  }

  return duration;
}

// Returns the comparison option whose rill time offset ends the expression, such as "-1W" in "P7D offset -1W".
function getComparisonOptionFromExpression(
  expression: string,
): TimeComparisonOption | undefined {
  const offset = / offset (\S+)$/.exec(expression)?.[1];
  if (!offset) return undefined;
  return (Object.keys(TIME_COMPARISON) as TimeComparisonOption[]).find(
    (option) => TIME_COMPARISON[option].rillTimeOffset === offset,
  );
}
