import {
  createTimeRangeSummary,
  useMetricsView,
} from "@rilldata/web-common/features/dashboards/selectors/index";
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
  TimeRange,
  TimeRangePreset,
} from "@rilldata/web-common/lib/time/types";
import {
  RpcStatus,
  V1MetricsViewSpec,
  V1MetricsViewTimeRangeResponse,
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
  showComparison?: boolean;
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

  ready?: boolean;
} & TimeRangeState &
  ComparisonTimeRangeState;
export type TimeControlStore = Readable<TimeControlState>;

export const timeControlStateSelector = ([
  metricsView,
  timeRangeResponse,
  metricsExplorer,
]: [
  QueryObserverResult<V1MetricsViewSpec, RpcStatus>,
  QueryObserverResult<V1MetricsViewTimeRangeResponse, unknown>,
  MetricsExplorerEntity,
]): TimeControlState => {
  const hasTimeSeries = Boolean(metricsView.data?.timeDimension);
  if (
    !metricsView.data ||
    !metricsExplorer ||
    !metricsView.data ||
    !timeRangeResponse ||
    !timeRangeResponse.isSuccess
  ) {
    return {
      isFetching: metricsView.isFetching || timeRangeResponse.isRefetching,
      ready: !metricsExplorer || !hasTimeSeries,
    } as TimeControlState;
  }

  const allTimeRange = {
    name: TimeRangePreset.ALL_TIME,
    start: new Date(timeRangeResponse.data.timeRangeSummary.min),
    end: new Date(timeRangeResponse.data.timeRangeSummary.max),
  };
  const minTimeGrain =
    (metricsView.data.smallestTimeGrain as V1TimeGrain) ||
    V1TimeGrain.TIME_GRAIN_UNSPECIFIED;
  const defaultTimeRange = isoDurationToFullTimeRange(
    metricsView.data.defaultTimeRange,
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
    metricsView,
    metricsExplorer,
    allTimeRange,
    timeRangeState,
  );

  return {
    isFetching: false,
    minTimeGrain,
    allTimeRange,
    defaultTimeRange,
    ready: true,

    ...timeRangeState,

    ...comparisonTimeRangeState,
  } as TimeControlState;
};

export function createTimeControlStore(ctx: StateManagers) {
  return derived(
    [useMetricsView(ctx), createTimeRangeSummary(ctx), ctx.dashboardStore],
    timeControlStateSelector,
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
  metricsView: V1MetricsViewSpec,
  metricsExplorer: MetricsExplorerEntity,
  allTimeRange: DashboardTimeControls,
  timeRangeState: TimeRangeState,
): ComparisonTimeRangeState {
  const selectedComparisonTimeRange = getComparisonTimeRange(
    metricsView,
    metricsExplorer,
    allTimeRange,
    timeRangeState.selectedTimeRange,
    metricsExplorer.selectedComparisonTimeRange,
  );
  const showComparison = Boolean(
    metricsExplorer.showTimeComparison && selectedComparisonTimeRange?.start,
  );
  let comparisonAdjustedStart: string | undefined = undefined;
  let comparisonAdjustedEnd: string | undefined = undefined;
  if (showComparison && selectedComparisonTimeRange) {
    const adjustedComparisonTime = getAdjustedFetchTime(
      selectedComparisonTimeRange.start,
      selectedComparisonTimeRange.end,
      metricsExplorer.selectedTimezone,
      timeRangeState.selectedTimeRange.interval,
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

    const comparisonRange = getComparionRangeForScrub(
      timeRangeState.selectedTimeRange.start,
      timeRangeState.selectedTimeRange.end,
      selectedComparisonTimeRange.start,
      selectedComparisonTimeRange.end,
      start,
      end,
    );
    comparisonTimeStart = comparisonRange.start;
    comparisonTimeEnd = comparisonRange.end;
  }

  return {
    showComparison,
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
    metricsExplorer.selectedTimeRange.interval,
    timeGrainOptions,
    minTimeGrain,
  );

  let timeGrain: V1TimeGrain;
  if (isValidTimeGrain) {
    timeGrain = metricsExplorer.selectedTimeRange.interval;
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
  metricsView: V1MetricsViewSpec,
  metricsExplorer: MetricsExplorerEntity,
  allTimeRange: DashboardTimeControls,
  timeRange: DashboardTimeControls,
  comparisonTimeRange: DashboardTimeControls,
) {
  if (!comparisonTimeRange || !timeRange.name) return undefined;

  let selectedComparisonTimeRange: DashboardTimeControls | undefined =
    undefined;
  if (!comparisonTimeRange?.name) {
    const comparisonOption = DEFAULT_TIME_RANGES[
      timeRange.name as TimeComparisonOption
    ]?.defaultComparison as TimeComparisonOption;
    const range = getTimeComparisonParametersForComponent(
      comparisonOption ??
        metricsView.availableTimeRanges?.find(
          (tr) => tr.range === timeRange.name,
        )?.comparisonOffsets?.[0].offset,
      allTimeRange.start,
      allTimeRange.end,
      timeRange.start,
      timeRange.end,
    );

    if (range.isComparisonRangeAvailable) {
      selectedComparisonTimeRange = {
        start: range.start,
        end: range.end,
        name: comparisonOption,
      };
    }
  } else if (comparisonTimeRange.name === TimeComparisonOption.CUSTOM) {
    selectedComparisonTimeRange = comparisonTimeRange;
  } else if (!metricsExplorer.showTimeComparison) {
    return undefined;
  } else {
    // variable time range of some kind.
    const comparisonOption = comparisonTimeRange.name as TimeComparisonOption;
    const range = getComparisonRange(
      timeRange.start,
      timeRange.end,
      comparisonOption,
    );

    selectedComparisonTimeRange = {
      ...range,
      name: comparisonOption,
    };
  }

  return selectedComparisonTimeRange;
}

/**
 * Fills in start and end dates based on selected time range and all time range.
 */
export function selectedTimeRangeSelector([
  metricsView,
  timeRangeResponse,
  explorer,
]: [
  QueryObserverResult<V1MetricsViewSpec, RpcStatus>,
  QueryObserverResult<V1MetricsViewTimeRangeResponse, unknown>,
  MetricsExplorerEntity,
]) {
  if (!metricsView.data || !timeRangeResponse.data?.timeRangeSummary) {
    return undefined;
  }

  const allTimeRange = {
    name: TimeRangePreset.ALL_TIME,
    start: new Date(timeRangeResponse.data.timeRangeSummary.min),
    end: new Date(timeRangeResponse.data.timeRangeSummary.max),
  };
  const defaultTimeRange = isoDurationToFullTimeRange(
    metricsView.data.defaultTimeRange,
    allTimeRange.start,
    allTimeRange.end,
    explorer.selectedTimezone,
  );

  return getTimeRange(explorer, allTimeRange, defaultTimeRange);
}
