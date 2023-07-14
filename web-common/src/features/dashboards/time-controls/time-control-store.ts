import type { RpcStatus } from "@rilldata/web-admin/client";
import {
  MetricsExplorerEntity,
  metricsExplorerStore,
} from "@rilldata/web-common/features/dashboards/dashboard-stores";
import type { StateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
import {
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
  ISODurationToTimePreset,
} from "@rilldata/web-common/lib/time/ranges";
import {
  TimeComparisonOption,
  TimeRange,
  TimeRangePreset,
} from "@rilldata/web-common/lib/time/types";
import type { TimeRangeType } from "@rilldata/web-common/lib/time/types";
import type { DashboardTimeControls } from "@rilldata/web-common/lib/time/types";
import {
  createQueryServiceColumnTimeRange,
  createRuntimeServiceGetCatalogEntry,
  V1ColumnTimeRangeResponse,
  V1MetricsView,
  V1TimeGrain,
} from "@rilldata/web-common/runtime-client";
import type { QueryObserverResult } from "@tanstack/query-core";
import type { CreateQueryResult } from "@tanstack/svelte-query";
import { getContext } from "svelte";
import { derived } from "svelte/store";
import type { Readable } from "svelte/store";

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
  defaultTimeRange?: TimeRangeType;
  minTimeGrain?: V1TimeGrain;
  allTimeRange?: TimeRange;

  ready?: boolean;
} & TimeRangeState &
  ComparisonTimeRangeState;
export type TimeControlStore = Readable<TimeControlState>;

export const DEFAULT_TIME_STORE_KEY = "time-store";
export function getTimeControlStore(): TimeControlStore {
  return getContext(DEFAULT_TIME_STORE_KEY);
}

// TODO: remove once #2749 is merged
export function useMetaQuery<T = V1MetricsView>(
  ctx: StateManagers,
  selector?: (meta: V1MetricsView) => T
): Readable<QueryObserverResult<T | V1MetricsView, RpcStatus>> {
  return derived(
    [ctx.runtime, ctx.metricsViewName],
    ([runtime, metricViewName], set) => {
      const store = createRuntimeServiceGetCatalogEntry(
        runtime.instanceId,
        metricViewName,
        {
          query: {
            select: (data) =>
              selector
                ? selector(data?.entry?.metricsView)
                : data?.entry?.metricsView,
            queryClient: ctx.queryClient,
          },
        }
      );
      return store.subscribe(set);
    }
  );
}

function createTimeRangeSummary(
  ctx: StateManagers
): CreateQueryResult<V1ColumnTimeRangeResponse> {
  return derived(
    [ctx.runtime, useMetaQuery(ctx)],
    ([runtime, metricsView], set) =>
      createQueryServiceColumnTimeRange(
        runtime.instanceId,
        metricsView.data?.model,
        {
          columnName: metricsView.data?.timeDimension,
        },
        {
          query: {
            enabled: !!metricsView.data?.timeDimension,
            queryClient: ctx.queryClient,
          },
        }
      ).subscribe(set)
  );
}

export function createTimeControlStore(ctx: StateManagers) {
  return derived(
    [
      ctx.metricsViewName,
      useMetaQuery(ctx),
      createTimeRangeSummary(ctx),
      ctx.dashboardStore,
    ],
    ([metricsViewName, metricsView, timeRangeResponse, metricsExplorer]) => {
      const hasTimeSeries = Boolean(metricsView.data?.timeDimension);
      if (!timeRangeResponse || !timeRangeResponse.isSuccess) {
        if (!hasTimeSeries && !metricsExplorer.defaultsSelected) {
          // TODO: refactor this when everything is moved to new architecture
          metricsExplorerStore.allDefaultsSelected(metricsViewName);
        }

        return {
          isFetching: metricsView.isFetching || timeRangeResponse.isRefetching,
          ready: !hasTimeSeries,
        } as TimeControlState;
      }

      if (!metricsExplorer.defaultsSelected) {
        // TODO: refactor this when everything is moved to new architecture
        metricsExplorerStore.allDefaultsSelected(metricsViewName);
      }
      const allTimeRange = {
        name: TimeRangePreset.ALL_TIME,
        start: new Date(timeRangeResponse.data.timeRangeSummary.min),
        end: new Date(timeRangeResponse.data.timeRangeSummary.max),
      };
      const defaultTimeRange = ISODurationToTimePreset(
        metricsView.data.defaultTimeRange
      );
      const minTimeGrain =
        (metricsView.data.defaultTimeRange as V1TimeGrain) ||
        V1TimeGrain.TIME_GRAIN_UNSPECIFIED;

      const timeRangeState = calculateTimeRangePartial(
        metricsExplorer,
        allTimeRange,
        defaultTimeRange,
        minTimeGrain
      );

      const comparisonTimeRangeState = calculateComparisonTimeRangePartial(
        metricsExplorer,
        allTimeRange,
        timeRangeState
      );

      return {
        isFetching: false,
        defaultTimeRange,
        minTimeGrain,
        allTimeRange,
        ready: true,

        ...timeRangeState,

        ...comparisonTimeRangeState,
      } as TimeControlState;
    }
  ) as TimeControlStore;
}

/**
 * Calculates time range and grain from all time range and selected time range name.
 * Also adds start, end and their adjusted counterparts as strings ready to use in requests.
 */
function calculateTimeRangePartial(
  metricsExplorer: MetricsExplorerEntity,
  allTimeRange: DashboardTimeControls,
  defaultTimeRange: string,
  minTimeGrain: V1TimeGrain
): TimeRangeState {
  const selectedTimeRange = getTimeRange(
    metricsExplorer,
    allTimeRange,
    defaultTimeRange
  );
  selectedTimeRange.interval = getTimeGrain(
    metricsExplorer,
    selectedTimeRange,
    minTimeGrain
  );
  const { start: adjustedStart, end: adjustedEnd } = getAdjustedFetchTime(
    selectedTimeRange.start,
    selectedTimeRange.end,
    selectedTimeRange.interval
  );

  return {
    selectedTimeRange,
    timeStart: selectedTimeRange.start.toISOString(),
    adjustedStart,
    timeEnd: selectedTimeRange.end.toISOString(),
    adjustedEnd,
  };
}

/**
 * Calculates time range and grain for comparison based on time range and comparison selection.
 * Also adds start, end and their adjusted counterparts as strings ready to use in requests.
 */
function calculateComparisonTimeRangePartial(
  metricsExplorer: MetricsExplorerEntity,
  allTimeRange: DashboardTimeControls,
  timeRangeState: TimeRangeState
): ComparisonTimeRangeState {
  const selectedComparisonTimeRange = getComparisonTimeRange(
    allTimeRange,
    timeRangeState.selectedTimeRange,
    metricsExplorer.selectedComparisonTimeRange
  );
  const showComparison = Boolean(
    metricsExplorer.showComparison && selectedComparisonTimeRange?.start
  );
  let comparisonAdjustedStart: string;
  let comparisonAdjustedEnd: string;
  if (showComparison && selectedComparisonTimeRange) {
    const adjustedComparisonTime = getAdjustedFetchTime(
      selectedComparisonTimeRange.start,
      selectedComparisonTimeRange.end,
      timeRangeState.selectedTimeRange.interval
    );
    comparisonAdjustedStart = adjustedComparisonTime.start;
    comparisonAdjustedEnd = adjustedComparisonTime.end;
  }

  return {
    showComparison,
    selectedComparisonTimeRange,
    comparisonTimeStart: selectedComparisonTimeRange?.start.toISOString(),
    comparisonAdjustedStart,
    comparisonTimeEnd: selectedComparisonTimeRange?.end.toISOString(),
    comparisonAdjustedEnd,
  };
}

function getTimeRange(
  metricsExplorer: MetricsExplorerEntity,
  allTimeRange: DashboardTimeControls,
  defaultTimeRange: string
) {
  let timeRange: DashboardTimeControls;
  if (!metricsExplorer?.selectedTimeRange) {
    timeRange = convertTimeRangePreset(
      defaultTimeRange,
      allTimeRange.start,
      allTimeRange.end
    );
  } else {
    if (metricsExplorer.selectedTimeRange.name === TimeRangePreset.CUSTOM) {
      /** set the time range to the fixed custom time range */
      timeRange = {
        name: TimeRangePreset.CUSTOM,
        start: new Date(metricsExplorer.selectedTimeRange.start),
        end: new Date(metricsExplorer.selectedTimeRange.end),
      };
    } else {
      /** rebuild off of relative time range */
      timeRange = convertTimeRangePreset(
        metricsExplorer.selectedTimeRange?.name ?? TimeRangePreset.ALL_TIME,
        allTimeRange.start,
        allTimeRange.end
      );
    }
  }
  return timeRange;
}

function getTimeGrain(
  metricsExplorer: MetricsExplorerEntity,
  timeRange: DashboardTimeControls,
  minTimeGrain: V1TimeGrain
) {
  let timeGrain: V1TimeGrain;

  if (!metricsExplorer?.selectedTimeRange) {
    timeGrain = getDefaultTimeGrain(timeRange.start, timeRange.end).grain;
  } else {
    const timeGrainOptions = getAllowedTimeGrains(
      timeRange.start,
      timeRange.end
    );
    const isValidTimeGrain = checkValidTimeGrain(
      metricsExplorer.selectedTimeRange.interval,
      timeGrainOptions,
      minTimeGrain
    );

    if (isValidTimeGrain) {
      timeGrain = metricsExplorer.selectedTimeRange.interval;
    } else {
      const defaultTimeGrain = getDefaultTimeGrain(
        timeRange.start,
        timeRange.end
      ).grain;
      timeGrain = findValidTimeGrain(
        defaultTimeGrain,
        timeGrainOptions,
        minTimeGrain
      );
    }
  }

  return timeGrain;
}

function getComparisonTimeRange(
  allTimeRange: DashboardTimeControls,
  timeRange: DashboardTimeControls,
  comparisonTimeRange: DashboardTimeControls
) {
  if (!comparisonTimeRange) return undefined;

  let selectedComparisonTimeRange: DashboardTimeControls;
  if (!comparisonTimeRange?.name) {
    const comparisonOption = DEFAULT_TIME_RANGES[timeRange.name]
      ?.defaultComparison as TimeComparisonOption;
    const range = getTimeComparisonParametersForComponent(
      comparisonOption,
      allTimeRange.start,
      allTimeRange.end,
      timeRange.start,
      timeRange.end
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
  } else {
    // variable time range of some kind.
    const comparisonOption = comparisonTimeRange.name as TimeComparisonOption;
    const range = getComparisonRange(
      timeRange.start,
      timeRange.end,
      comparisonOption
    );

    selectedComparisonTimeRange = {
      ...range,
      name: comparisonOption,
    };
  }

  return selectedComparisonTimeRange;
}
