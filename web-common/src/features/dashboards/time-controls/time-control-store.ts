import type { StateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import { ALL_TIME_RANGE_ALIAS } from "@rilldata/web-common/features/dashboards/time-controls/new-time-controls";
import { getOrderedStartEnd } from "@rilldata/web-common/features/dashboards/time-series/utils";
import {
  getComparionRangeForScrub,
  getComparisonRange,
  getTimeComparisonParametersForComponent,
  inferCompareTimeRange,
} from "@rilldata/web-common/lib/time/comparisons";
import {
  checkValidTimeGrain,
  findValidTimeGrain,
  getAllowedTimeGrains,
  getDefaultTimeGrain,
} from "@rilldata/web-common/lib/time/grains";
import { getAdjustedFetchTime } from "@rilldata/web-common/lib/time/ranges";
import { isoDurationToFullTimeRange } from "@rilldata/web-common/lib/time/ranges/iso-ranges";
import {
  type DashboardTimeControls,
  TimeComparisonOption,
  type TimeRange,
  TimeRangePreset,
} from "@rilldata/web-common/lib/time/types";
import {
  type RpcStatus,
  type V1ExploreSpec,
  type V1MetricsViewSpec,
  type V1MetricsViewTimeRangeResponse,
  type V1MetricsViewTimeRangesResponse,
  V1TimeGrain,
  type V1TimeRange,
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

/**
 * Returns a TimeControlState. Calls getTimeControlState internally.
 *
 * Consumers of this will have a QueryObserverResult for time range summary.
 * They will need `isFetching` to wait for this response.
 */
export const timeControlStateSelector = ([
  metricsView,
  explore,
  timeRanges,
  metricsExplorer,
]: [
  V1MetricsViewSpec | undefined,
  V1ExploreSpec | undefined,
  QueryObserverResult<V1MetricsViewTimeRangesResponse, RpcStatus>,
  MetricsExplorerEntity,
]): TimeControlState => {
  const hasTimeSeries = Boolean(metricsView?.timeDimension);
  if (
    !metricsView ||
    !explore ||
    !metricsExplorer ||
    !timeRanges.isSuccess ||
    !timeRanges?.data?.timeRanges ||
    !hasTimeSeries
  ) {
    return {
      isFetching: timeRanges.isRefetching,
      ready: !metricsExplorer || !hasTimeSeries,
    } as TimeControlState;
  }

  const state = getTimeControlState(
    metricsView,
    explore,
    timeRanges.data.timeRanges,
    metricsExplorer,
  );
  if (!state) {
    return {
      ready: false,
      isFetching: false,
    };
  }

  return {
    ...state,
    isFetching: false,
    ready: true,
  } as TimeControlState;
};

/**
 * Generates TimeControlState
 *
 * Consumers of this will already have a V1TimeRangeSummary.
 */
export function getTimeControlState(
  metricsViewSpec: V1MetricsViewSpec,
  exploreSpec: V1ExploreSpec,
  timeRanges: V1TimeRange[],
  exploreState: MetricsExplorerEntity,
) {
  const hasTimeSeries = Boolean(metricsViewSpec.timeDimension);
  const timeDimension = metricsViewSpec.timeDimension;
  if (!hasTimeSeries) return undefined;

  const allTimeRange = findTimeRange(
    ALL_TIME_RANGE_ALIAS,
    timeRanges,
  ) as DashboardTimeControls;
  if (!allTimeRange) return undefined;

  const minTimeGrain =
    (metricsViewSpec.smallestTimeGrain as V1TimeGrain) ||
    V1TimeGrain.TIME_GRAIN_UNSPECIFIED;
  const defaultTimeRange = isoDurationToFullTimeRange(
    exploreSpec.defaultPreset?.timeRange,
    allTimeRange.start,
    allTimeRange.end,
    exploreState.selectedTimezone,
  );

  const timeRangeState = calculateTimeRangePartial(
    exploreState,
    defaultTimeRange,
    minTimeGrain,
    timeRanges,
  );
  if (!timeRangeState) {
    return undefined;
  }

  const comparisonTimeRangeState = calculateComparisonTimeRangePartial(
    exploreSpec,
    exploreState,
    allTimeRange,
    timeRangeState,
  );

  return {
    minTimeGrain,
    allTimeRange,
    defaultTimeRange,
    timeDimension,

    ...timeRangeState,

    ...comparisonTimeRangeState,
  } as TimeControlState;
}

export function createTimeControlStore(ctx: StateManagers) {
  return derived(
    [ctx.validSpecStore, ctx.timeRanges, ctx.dashboardStore],
    ([validSpecResp, timeRangeResp, dashboardStore]) =>
      timeControlStateSelector([
        validSpecResp.data?.metricsView,
        validSpecResp.data?.explore,
        timeRangeResp,
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
  defaultTimeRange: DashboardTimeControls,
  minTimeGrain: V1TimeGrain,
  timeRanges: V1TimeRange[],
): TimeRangeState | undefined {
  if (!metricsExplorer.selectedTimeRange) return undefined;

  const selectedTimeRange = getTimeRange(
    metricsExplorer,
    defaultTimeRange,
    timeRanges,
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
  defaultTimeRange: DashboardTimeControls,
  timeRanges: V1TimeRange[],
) {
  if (!metricsExplorer.selectedTimeRange) return undefined;
  if (!metricsExplorer.selectedTimeRange?.name) {
    return defaultTimeRange;
  }
  if (metricsExplorer.selectedTimeRange.name === TimeRangePreset.CUSTOM) {
    /** set the time range to the fixed custom time range */
    return <DashboardTimeControls>{
      name: TimeRangePreset.CUSTOM,
      start: new Date(metricsExplorer.selectedTimeRange.start),
      end: new Date(metricsExplorer.selectedTimeRange.end),
    };
  }

  return findTimeRange(metricsExplorer.selectedTimeRange?.name, timeRanges);
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
    const comparisonOption = inferCompareTimeRange(
      explore.timeRanges,
      timeRange.name,
    );
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
  } else if (
    comparisonTimeRange.name === TimeComparisonOption.CUSTOM ||
    // 1st step towards using a single `Custom` variable
    // TODO: replace the usage of TimeComparisonOption.CUSTOM with TimeRangePreset.CUSTOM
    comparisonTimeRange.name === TimeRangePreset.CUSTOM
  ) {
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

  return getTimeRange(explorer, defaultTimeRange, []);
}

export function findTimeRange(
  name: string,
  timeRanges: V1TimeRange[],
): DashboardTimeControls | undefined {
  const tr = timeRanges.find((tr) => tr.expression === name);
  if (!tr) return undefined;
  return {
    name: name as TimeRangePreset,
    start: new Date(tr.start ?? ""),
    end: new Date(tr.end ?? ""),
  };
}
