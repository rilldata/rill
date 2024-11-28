import type { StateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import { getOrderedStartEnd } from "@rilldata/web-common/features/dashboards/time-series/utils";
import {
  getComparionRangeForScrub,
  getComparisonRange,
  getTimeComparisonParametersForComponent,
} from "@rilldata/web-common/lib/time/comparisons";
import { DEFAULT_TIME_RANGES } from "@rilldata/web-common/lib/time/config";
import {
  checkValidTimeGrain,
  findValidTimeGrain,
  getAllowedTimeGrains,
  getDefaultTimeGrain,
} from "@rilldata/web-common/lib/time/grains";
import {
  convertTimeRangePreset,
  getAdjustedFetchTime,
} from "@rilldata/web-common/lib/time/ranges";
import { isoDurationToFullTimeRange } from "@rilldata/web-common/lib/time/ranges/iso-ranges";
import type { DashboardTimeControls } from "@rilldata/web-common/lib/time/types";
import {
  TimeComparisonOption,
  type TimeRange,
  TimeRangePreset,
} from "@rilldata/web-common/lib/time/types";
import {
  type V1ExploreSpec,
  type V1MetricsViewSpec,
  type V1MetricsViewTimeRangeResponse,
  V1TimeGrain,
} from "@rilldata/web-common/runtime-client";
import type { QueryObserverResult } from "@tanstack/svelte-query";
import type { Readable } from "svelte/store";
import { derived } from "svelte/store";
import { memoizeMetricsStore } from "../state-managers/memoize-metrics-store";

export type TimeRangeState = {
  // Selected ranges with start and end filled based on time range type
  selectedTimeRange?: DashboardTimeControls;
  // In all of our queries we do a check on hasTime and pass in undefined for start and end if false.
  // Using these directly will simplify those usages since this store will take care of marking them undefined.
  timeStart?: string;
  adjustedStart?: string;
  timeEnd?: string;
  adjustedEnd?: string;
};
export type ComparisonTimeRangeState = {
  showTimeComparison?: boolean;
  selectedComparisonTimeRange?: DashboardTimeControls;
  comparisonTimeStart?: string;
  comparisonAdjustedStart?: string;
  comparisonTimeEnd?: string;
  comparisonAdjustedEnd?: string;
};
export type TimeControlState = {
  isFetching: boolean;

  // Computed properties from all time range query
  minTimeGrain?: V1TimeGrain;
  allTimeRange?: TimeRange;
  defaultTimeRange?: TimeRange;
  timeDimension?: string;

  ready?: boolean;
} & TimeRangeState &
  ComparisonTimeRangeState;
export type TimeControlStore = Readable<TimeControlState>;

export const timeControlStateSelector = ([
  metricsView,
  explore,
  timeRangeResponse,
  metricsExplorer,
]: [
  V1MetricsViewSpec | undefined,
  V1ExploreSpec | undefined,
  QueryObserverResult<V1MetricsViewTimeRangeResponse, unknown>,
  MetricsExplorerEntity,
]): TimeControlState => {
  const hasTimeSeries = Boolean(metricsView?.timeDimension);
  const timeDimension = metricsView?.timeDimension;
  if (
    !metricsView ||
    !explore ||
    !metricsExplorer ||
    !timeRangeResponse ||
    !timeRangeResponse.isSuccess ||
    !timeRangeResponse.data.timeRangeSummary?.min ||
    !timeRangeResponse.data.timeRangeSummary?.max
  ) {
    return {
      isFetching: timeRangeResponse.isRefetching,
      ready: !metricsExplorer || !hasTimeSeries,
    } as TimeControlState;
  }

  const allTimeRange = {
    name: TimeRangePreset.ALL_TIME,
    start: new Date(timeRangeResponse.data.timeRangeSummary.min),
    end: new Date(timeRangeResponse.data.timeRangeSummary.max),
  };
  const minTimeGrain =
    (metricsView.smallestTimeGrain as V1TimeGrain) ||
    V1TimeGrain.TIME_GRAIN_UNSPECIFIED;
  const defaultTimeRange = isoDurationToFullTimeRange(
    explore?.defaultPreset?.timeRange,
    allTimeRange.start,
    allTimeRange.end,
    metricsExplorer.selectedTimezone,
  );

  const timeRangeState = calculateTimeRangePartial(
    metricsExplorer,
    allTimeRange,
    defaultTimeRange,
    minTimeGrain,
  );
  if (!timeRangeState) {
    return {
      ready: false,
      isFetching: false,
    };
  }

  const comparisonTimeRangeState = calculateComparisonTimeRangePartial(
    explore,
    metricsExplorer,
    allTimeRange,
    timeRangeState,
  );

  return {
    isFetching: false,
    minTimeGrain,
    allTimeRange,
    defaultTimeRange,
    timeDimension,
    ready: true,

    ...timeRangeState,

    ...comparisonTimeRangeState,
  } as TimeControlState;
};

export function createTimeControlStore(ctx: StateManagers) {
  return derived(
    [ctx.validSpecStore, ctx.timeRangeSummaryStore, ctx.dashboardStore],
    ([validSpecResp, timeRangeSummaryResp, dashboardStore]) =>
      timeControlStateSelector([
        validSpecResp.data?.metricsView,
        validSpecResp.data?.explore,
        timeRangeSummaryResp,
        dashboardStore,
      ]),
  );
}

/**
 * Memoized version of the store. Currently, memoized by metrics view name.
 */
export const useTimeControlStore = memoizeMetricsStore<TimeControlStore>(
  (ctx: StateManagers) => createTimeControlStore(ctx),
);

/**
 * Calculates time range and grain from all time range and selected time range name.
 * Also adds start, end and their adjusted counterparts as strings ready to use in requests.
 */
function calculateTimeRangePartial(
  metricsExplorer: MetricsExplorerEntity,
  allTimeRange: DashboardTimeControls,
  defaultTimeRange: DashboardTimeControls,
  minTimeGrain: V1TimeGrain,
): TimeRangeState | undefined {
  if (!metricsExplorer.selectedTimeRange) return undefined;

  const selectedTimeRange = getTimeRange(
    metricsExplorer,
    allTimeRange,
    defaultTimeRange,
  );
  if (!selectedTimeRange) return undefined;

  selectedTimeRange.interval = getTimeGrain(
    metricsExplorer,
    selectedTimeRange,
    minTimeGrain,
  );
  const { start: adjustedStart, end: adjustedEnd } = getAdjustedFetchTime(
    selectedTimeRange.start,
    selectedTimeRange.end,
    metricsExplorer.selectedTimezone,
    selectedTimeRange.interval,
  );

  let timeStart = selectedTimeRange.start;
  let timeEnd = selectedTimeRange.end;
  if (metricsExplorer.lastDefinedScrubRange) {
    const { start, end } = getOrderedStartEnd(
      metricsExplorer.lastDefinedScrubRange.start,
      metricsExplorer.lastDefinedScrubRange.end,
    );
    timeStart = start;
    timeEnd = end;
  }

  return {
    selectedTimeRange,
    timeStart: timeStart.toISOString(),
    adjustedStart,
    timeEnd: timeEnd.toISOString(),
    adjustedEnd,
  };
}

/**
 * Calculates time range and grain for comparison based on time range and comparison selection.
 * Also adds start, end and their adjusted counterparts as strings ready to use in requests.
 */
function calculateComparisonTimeRangePartial(
  explore: V1ExploreSpec,
  metricsExplorer: MetricsExplorerEntity,
  allTimeRange: DashboardTimeControls,
  timeRangeState: TimeRangeState,
): ComparisonTimeRangeState {
  const selectedComparisonTimeRange = getComparisonTimeRange(
    explore,
    allTimeRange,
    timeRangeState.selectedTimeRange,
    metricsExplorer.selectedComparisonTimeRange,
  );

  let comparisonAdjustedStart: string | undefined = undefined;
  let comparisonAdjustedEnd: string | undefined = undefined;
  if (selectedComparisonTimeRange) {
    const adjustedComparisonTime = getAdjustedFetchTime(
      selectedComparisonTimeRange.start,
      selectedComparisonTimeRange.end,
      metricsExplorer.selectedTimezone,
      timeRangeState.selectedTimeRange?.interval,
    );
    comparisonAdjustedStart = adjustedComparisonTime.start;
    comparisonAdjustedEnd = adjustedComparisonTime.end;
  }

  let comparisonTimeStart = selectedComparisonTimeRange?.start;
  let comparisonTimeEnd = selectedComparisonTimeRange?.end;
  if (selectedComparisonTimeRange && metricsExplorer.lastDefinedScrubRange) {
    const { start, end } = getOrderedStartEnd(
      metricsExplorer.lastDefinedScrubRange.start,
      metricsExplorer.lastDefinedScrubRange.end,
    );

    if (!timeRangeState.selectedTimeRange?.start) {
      throw new Error("No time range");
    }

    const comparisonRange = getComparionRangeForScrub(
      timeRangeState.selectedTimeRange?.start,
      timeRangeState.selectedTimeRange?.end,
      selectedComparisonTimeRange.start,
      selectedComparisonTimeRange.end,
      start,
      end,
    );
    comparisonTimeStart = comparisonRange.start;
    comparisonTimeEnd = comparisonRange.end;
  }

  return {
    showTimeComparison: metricsExplorer.showTimeComparison,
    selectedComparisonTimeRange,
    comparisonTimeStart: comparisonTimeStart?.toISOString(),
    comparisonAdjustedStart,
    comparisonTimeEnd: comparisonTimeEnd?.toISOString(),
    comparisonAdjustedEnd,
  };
}

function getTimeRange(
  metricsExplorer: MetricsExplorerEntity,
  allTimeRange: DashboardTimeControls,
  defaultTimeRange: DashboardTimeControls,
) {
  if (!metricsExplorer.selectedTimeRange) return undefined;

  let timeRange: DashboardTimeControls;

  if (metricsExplorer.selectedTimeRange?.name === TimeRangePreset.CUSTOM) {
    /** set the time range to the fixed custom time range */
    timeRange = {
      name: TimeRangePreset.CUSTOM,
      start: new Date(metricsExplorer.selectedTimeRange.start),
      end: new Date(metricsExplorer.selectedTimeRange.end),
    };
  } else if (metricsExplorer.selectedTimeRange?.name) {
    if (metricsExplorer.selectedTimeRange?.name in DEFAULT_TIME_RANGES) {
      /** rebuild off of relative time range */
      timeRange = convertTimeRangePreset(
        metricsExplorer.selectedTimeRange?.name ?? TimeRangePreset.ALL_TIME,
        allTimeRange.start,
        allTimeRange.end,
        metricsExplorer.selectedTimezone,
      );
    } else {
      timeRange = isoDurationToFullTimeRange(
        metricsExplorer.selectedTimeRange?.name,
        allTimeRange.start,
        allTimeRange.end,
        metricsExplorer.selectedTimezone,
      );
    }
  } else {
    /** set the time range to the fixed custom time range */
    timeRange = {
      name: defaultTimeRange.name,
      start: defaultTimeRange.start,
      end: defaultTimeRange.end,
    };
  }

  return timeRange;
}

function getTimeGrain(
  metricsExplorer: MetricsExplorerEntity,
  timeRange: DashboardTimeControls,
  minTimeGrain: V1TimeGrain,
) {
  const timeGrainOptions = getAllowedTimeGrains(timeRange.start, timeRange.end);
  const isValidTimeGrain = checkValidTimeGrain(
    metricsExplorer.selectedTimeRange?.interval,
    timeGrainOptions,
    minTimeGrain,
  );

  let timeGrain: V1TimeGrain | undefined;
  if (isValidTimeGrain) {
    timeGrain = metricsExplorer.selectedTimeRange?.interval;
  } else {
    const defaultTimeGrain = getDefaultTimeGrain(
      timeRange.start,
      timeRange.end,
    ).grain;
    timeGrain = findValidTimeGrain(
      defaultTimeGrain,
      timeGrainOptions,
      minTimeGrain,
    );
  }

  return timeGrain;
}

function getComparisonTimeRange(
  explore: V1ExploreSpec,
  allTimeRange: DashboardTimeControls | undefined,
  timeRange: DashboardTimeControls | undefined,
  comparisonTimeRange: DashboardTimeControls | undefined,
) {
  if (!timeRange || !timeRange.name || !allTimeRange) return undefined;

  if (!comparisonTimeRange?.name) {
    const comparisonOption =
      (DEFAULT_TIME_RANGES[timeRange.name as TimeComparisonOption]
        ?.defaultComparison as TimeComparisonOption) ??
      explore.timeRanges?.find((tr) => tr.range === timeRange.name)
        ?.comparisonTimeRanges?.[0]?.offset ??
      TimeComparisonOption.CONTIGUOUS;
    const range = getTimeComparisonParametersForComponent(
      comparisonOption,
      allTimeRange.start,
      allTimeRange.end,
      timeRange.start,
      timeRange.end,
    );

    if (range.isComparisonRangeAvailable && range.start && range.end) {
      return {
        start: range.start,
        end: range.end,
        name: comparisonOption,
      };
    }
  } else if (comparisonTimeRange.name === TimeComparisonOption.CUSTOM) {
    return comparisonTimeRange;
  } else {
    // variable time range of some kind.
    const comparisonOption = comparisonTimeRange.name as TimeComparisonOption;
    const range = getComparisonRange(
      timeRange.start,
      timeRange.end,
      comparisonOption,
    );

    return {
      ...range,
      name: comparisonOption,
    };
  }
}

/**
 * Fills in start and end dates based on selected time range and all time range.
 */
export function selectedTimeRangeSelector([
  exploreSpec,
  timeRangeResponse,
  explorer,
]: [
  V1ExploreSpec | undefined,
  QueryObserverResult<V1MetricsViewTimeRangeResponse, unknown>,
  MetricsExplorerEntity,
]) {
  if (
    !exploreSpec ||
    !timeRangeResponse.data?.timeRangeSummary ||
    !timeRangeResponse.data.timeRangeSummary.min ||
    !timeRangeResponse.data.timeRangeSummary.max
  ) {
    return undefined;
  }

  const allTimeRange = {
    name: TimeRangePreset.ALL_TIME,
    start: new Date(timeRangeResponse.data.timeRangeSummary.min),
    end: new Date(timeRangeResponse.data.timeRangeSummary.max),
  };
  const defaultTimeRange = isoDurationToFullTimeRange(
    exploreSpec?.defaultPreset?.timeRange,
    allTimeRange.start,
    allTimeRange.end,
    explorer.selectedTimezone,
  );

  return getTimeRange(explorer, allTimeRange, defaultTimeRange);
}
