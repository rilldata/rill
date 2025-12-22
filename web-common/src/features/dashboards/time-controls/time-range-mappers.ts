import {
  parseRillTime,
  validateRillTime,
} from "@rilldata/web-common/features/dashboards/url-state/time-ranges/parser.ts";
import { TIME_COMPARISON } from "@rilldata/web-common/lib/time/config.ts";
import { isoDurationToFullTimeRange } from "@rilldata/web-common/lib/time/ranges/iso-ranges";
import {
  type DashboardTimeControls,
  TimeComparisonOption,
  TimeRangePreset,
} from "@rilldata/web-common/lib/time/types";
import {
  type V1ExploreSpec,
  V1TimeGrain,
  type V1TimeRange,
  type V1TimeRangeSummary,
} from "@rilldata/web-common/runtime-client";

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

export function mapSelectedTimeRangeToV1TimeRange(
  selectedTimeRange: DashboardTimeControls | undefined,
  timeZone: string,
  explore: V1ExploreSpec,
): V1TimeRange | undefined {
  if (!selectedTimeRange?.name) return undefined;
  if (!validateRillTime(selectedTimeRange.name)) {
    return {
      expression: selectedTimeRange.name,
      timeZone,
    };
  }

  const timeRange: V1TimeRange = {};
  switch (selectedTimeRange.name) {
    case TimeRangePreset.DEFAULT:
      timeRange.isoDuration = explore?.defaultPreset?.timeRange;
      break;

    case TimeRangePreset.CUSTOM:
      timeRange.start = selectedTimeRange.start.toISOString();
      timeRange.end = selectedTimeRange.end.toISOString();
      break;

    default:
      if (selectedTimeRange.name in PreviousCompleteRangeMap) {
        const prevCompleteTimeRange: V1TimeRange | undefined =
          PreviousCompleteRangeMap[selectedTimeRange.name];
        // Backend doesn't support previous complete ranges since it has offset built in.
        // We add the offset manually as a workaround for now
        timeRange.isoDuration = prevCompleteTimeRange?.isoDuration;
        timeRange.isoOffset = prevCompleteTimeRange?.isoOffset;
        timeRange.roundToGrain = prevCompleteTimeRange?.roundToGrain;
      } else {
        timeRange.isoDuration = selectedTimeRange.name;
      }
      break;
  }

  timeRange.timeZone = timeZone;

  return timeRange;
}

export function mapSelectedComparisonTimeRangeToV1TimeRange(
  selectedComparisonTimeRange: DashboardTimeControls | undefined,
  showTimeComparison: boolean,
  timeRange: V1TimeRange | undefined,
) {
  if (!timeRange || !showTimeComparison || !selectedComparisonTimeRange?.name) {
    return undefined;
  }

  let isoDuration = timeRange.isoDuration;
  const name = selectedComparisonTimeRange.name;

  if (
    timeRange.expression &&
    TIME_COMPARISON[selectedComparisonTimeRange.name]?.rillTimeOffset
  ) {
    const rt = parseRillTime(timeRange.expression);
    if (!rt.isOldFormat) {
      return {
        expression:
          rt.toString() +
          " offset " +
          TIME_COMPARISON[selectedComparisonTimeRange.name]?.rillTimeOffset,
      };
    } else {
      // Handle old syntax differently until we have the backend parser updated.
      isoDuration = timeRange.expression;
    }
  }

  const comparisonTimeRange: V1TimeRange = {};
  switch (name) {
    default:
      comparisonTimeRange.isoOffset = selectedComparisonTimeRange.name;
      comparisonTimeRange.isoDuration = isoDuration;
      break;
    case TimeComparisonOption.CONTIGUOUS:
      comparisonTimeRange.isoOffset = comparisonTimeRange.isoDuration =
        isoDuration;
      break;

    case TimeComparisonOption.CUSTOM:
      comparisonTimeRange.start =
        selectedComparisonTimeRange.start.toISOString();
      comparisonTimeRange.end = selectedComparisonTimeRange.end.toISOString();
      break;
  }
  return comparisonTimeRange;
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

  if (timeRange.start && timeRange.end) {
    selectedTimeRange = {
      name: TimeComparisonOption.CUSTOM,
      start: new Date(timeRange.start),
      end: new Date(timeRange.end),
    };
  } else if (timeRange.isoOffset === timeRange.isoDuration) {
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
