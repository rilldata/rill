import {
  getAvailableComparisonsForTimeRange,
  getComparisonRange,
} from "@rilldata/web-common/lib/time/comparisons";
import {
  LATEST_WINDOW_TIME_RANGES,
  PERIOD_TO_DATE_RANGES,
  PREVIOUS_COMPLETE_DATE_RANGES,
} from "@rilldata/web-common/lib/time/config";
import { getChildTimeRanges } from "@rilldata/web-common/lib/time/ranges";
import {
  TimeComparisonOption,
  TimeRangePreset,
  type DashboardTimeControls,
} from "@rilldata/web-common/lib/time/types";
import { V1TimeGrain } from "@rilldata/web-common/runtime-client";

export function getTimeRangeForCanvas(
  selectedTimezone: string,
  defaultTimeRange?: string,
) {
  const allTimeRange = {
    name: TimeRangePreset.ALL_TIME,
    start: new Date(0),
    end: new Date(),
  };

  const latestWindowTimeRanges = LATEST_WINDOW_TIME_RANGES;
  const periodToDateRanges = PERIOD_TO_DATE_RANGES;
  const previousCompleteDateRanges = PREVIOUS_COMPLETE_DATE_RANGES;
  const hasDefaultInRanges =
    !!defaultTimeRange &&
    (defaultTimeRange in LATEST_WINDOW_TIME_RANGES ||
      defaultTimeRange in PERIOD_TO_DATE_RANGES ||
      defaultTimeRange in PREVIOUS_COMPLETE_DATE_RANGES);

  return {
    latestWindowTimeRanges: getChildTimeRanges(
      allTimeRange.start,
      allTimeRange.end,
      latestWindowTimeRanges,
      V1TimeGrain.TIME_GRAIN_UNSPECIFIED,
      selectedTimezone,
    ),
    periodToDateRanges: getChildTimeRanges(
      allTimeRange.start,
      allTimeRange.end,
      periodToDateRanges,
      V1TimeGrain.TIME_GRAIN_UNSPECIFIED,
      selectedTimezone,
    ),
    previousCompleteDateRanges: getChildTimeRanges(
      allTimeRange.start,
      allTimeRange.end,
      previousCompleteDateRanges,
      V1TimeGrain.TIME_GRAIN_UNSPECIFIED,
      selectedTimezone,
    ),
    showDefaultItem: !!defaultTimeRange && !hasDefaultInRanges,
  };
}

export function getComparisonOptionsForCanvas(
  selectedTimeRange: DashboardTimeControls | undefined,
) {
  if (!selectedTimeRange) {
    return [];
  }
  const allTimeRange = {
    name: TimeRangePreset.ALL_TIME,
    start: new Date(0),
    end: new Date(),
  };

  let allOptions = [...Object.values(TimeComparisonOption)];

  if (
    selectedTimeRange?.name &&
    selectedTimeRange?.name in PREVIOUS_COMPLETE_DATE_RANGES
  ) {
    // Previous complete ranges should only have previous period.
    // Other options dont make sense with our current wording of the comparison ranges.
    allOptions = [TimeComparisonOption.CONTIGUOUS, TimeComparisonOption.CUSTOM];
  }

  const timeComparisonOptions = getAvailableComparisonsForTimeRange(
    allTimeRange.start,
    allTimeRange.end,
    selectedTimeRange.start,
    selectedTimeRange.end,
    allOptions,
  );

  return timeComparisonOptions.map((co, i) => {
    const comparisonTimeRange = getComparisonRange(
      selectedTimeRange.start,
      selectedTimeRange.end,
      co,
    );
    return {
      name: co,
      key: i,
      start: comparisonTimeRange.start,
      end: comparisonTimeRange.end,
    };
  });
}
