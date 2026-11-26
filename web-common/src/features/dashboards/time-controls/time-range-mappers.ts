import {
  formatRillTimeWallClock,
  parseRillTime,
} from "@rilldata/web-common/features/dashboards/url-state/time-ranges/parser.ts";
import { RillAllTimeInterval } from "@rilldata/web-common/features/dashboards/url-state/time-ranges/RillTime.ts";
import { TIME_COMPARISON } from "@rilldata/web-common/lib/time/config.ts";
import { DateTime } from "luxon";
import {
  TimeComparisonOption,
  TimeRangePreset,
} from "@rilldata/web-common/lib/time/types";
import {
  V1TimeGrain,
  type V1TimeRange,
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

export function convertTimeManagerRangeToV1TimeRange(
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

export function convertTimeManagerComparisonRangeToV1TimeRange(
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
 * Converts API time ranges to the time filter url params a {@link TimeFilterManager} is seeded with,
 * e.g. those saved in a report's or an alert's aggregation request.
 */
export function convertV1TimeRangesToTimeFilterUrlParams(
  timeRange: V1TimeRange | undefined,
  comparisonTimeRange: V1TimeRange | undefined,
) {
  const urlParams = new URLSearchParams();
  if (!timeRange) return urlParams;

  const { timeRange: timeRangeParam, timeZone } =
    convertV1TimeRangeToTimeFilterParams(timeRange);
  if (!timeRangeParam) return urlParams;
  urlParams.set(ExploreStateURLParams.TimeRange, timeRangeParam);
  if (timeZone) {
    urlParams.set(ExploreStateURLParams.TimeZone, timeZone);
  }

  if (comparisonTimeRange) {
    const comparisonParam =
      convertV1TimeRangeToComparisonTimeOption(comparisonTimeRange);
    if (comparisonParam) {
      urlParams.set(ExploreStateURLParams.ComparisonTimeRange, comparisonParam);
    }
  }

  return urlParams;
}

/**
 * Maps an API time range to the rill time range and zone a {@link TimeFilterManager} is seeded with.
 * The zone is undefined when the dashboard's default zone should apply.
 */
export function convertV1TimeRangeToTimeFilterParams(timeRange: V1TimeRange): {
  timeRange: string | undefined;
  timeZone: string | undefined;
} {
  if (timeRange.start && timeRange.end) {
    // An absolute range without a saved zone is pinned to UTC instead of the dashboard's default zone,
    // since its endpoints are UTC instants.
    const timeZone = timeRange.timeZone ?? "UTC";
    return {
      timeRange: convertAbsoluteTimeRangeToRillTime(
        timeRange.start,
        timeRange.end,
        timeZone,
      ),
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

/**
 * Maps an API comparison time range to the explore url's comparison param.
 */
export function convertV1TimeRangeToComparisonTimeOption(
  timeRange: V1TimeRange,
) {
  if (timeRange.start && timeRange.end) {
    return convertAbsoluteTimeRangeToUrlParam(timeRange.start, timeRange.end);
  }

  if (
    timeRange.isoDuration &&
    // Older saved queries set only the duration for the previous period.
    (!timeRange.isoOffset || timeRange.isoOffset === timeRange.isoDuration)
  ) {
    return TimeComparisonOption.CONTIGUOUS;
  }

  if (timeRange.expression) {
    // A comparison sent as the base expression with an offset, such as "P1D offset -1W".
    return (
      convertExpressionToComparisonOption(timeRange.expression) ??
      timeRange.expression
    );
  }

  return timeRange.isoOffset;
}

/**
 * Formats an absolute range of UTC instants as a rill time range, `start,end`.
 * Rill time reads absolute times as wall-clock times in the request's zone,
 * so both ends are written as wall-clock times in `timeZone` to keep the same instants.
 */
export function convertAbsoluteTimeRangeToRillTime(
  start: string | Date,
  end: string | Date,
  timeZone: string,
) {
  const [startDt, endDt] = [start, end].map((t) =>
    typeof t === "string"
      ? DateTime.fromISO(t, { zone: "UTC" })
      : DateTime.fromJSDate(t, { zone: "UTC" }),
  );
  if (!startDt.isValid || !endDt.isValid)
    return `${String(start)},${String(end)}`;
  return `${formatRillTimeWallClock(startDt, timeZone)},${formatRillTimeWallClock(endDt, timeZone)}`;
}

/**
 * Formats an absolute range as the explore url's custom range, `start,end`.
 * Both ends are normalized to ISO strings with milliseconds, which the comparison and highlighted range params require.
 */
export function convertAbsoluteTimeRangeToUrlParam(
  start: string | Date,
  end: string | Date,
) {
  return `${new Date(start).toISOString()},${new Date(end).toISOString()}`;
}

// Returns the comparison option whose rill time offset ends the expression, such as "-1W" in "P7D offset -1W".
function convertExpressionToComparisonOption(
  expression: string,
): TimeComparisonOption | undefined {
  const offset = / offset (\S+)$/.exec(expression)?.[1];
  if (!offset) return undefined;
  return (Object.keys(TIME_COMPARISON) as TimeComparisonOption[]).find(
    (option) => TIME_COMPARISON[option].rillTimeOffset === offset,
  );
}
