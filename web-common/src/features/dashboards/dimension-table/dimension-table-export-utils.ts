import { getQuerySortType } from "@rilldata/web-common/features/dashboards/leaderboard/leaderboard-utils";
import { SortDirection } from "@rilldata/web-common/features/dashboards/proto-state/derived-types";
import { useMetricsView } from "@rilldata/web-common/features/dashboards/selectors/index";
import type { StateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
import { sanitiseExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import type { TimeControlState } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import {
  TimeComparisonOption,
  TimeRangePreset,
} from "@rilldata/web-common/lib/time/types";
import {
  type V1MetricsViewComparisonRequest,
  type V1MetricsViewSpec,
  V1TimeGrain,
  type V1TimeRange,
} from "@rilldata/web-common/runtime-client";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { Readable, derived, get } from "svelte/store";

export function getDimensionTableExportArgs(
  ctx: StateManagers,
): Readable<V1MetricsViewComparisonRequest | undefined> {
  return derived(
    [
      ctx.metricsViewName,
      ctx.dashboardStore,
      useTimeControlStore(ctx),
      useMetricsView(ctx),
    ],
    ([metricViewName, dashboardState, timeControlState, metricsView]) => {
      if (!metricsView.data || !timeControlState.ready) return undefined;

      const timeRange = getTimeRange(timeControlState, metricsView.data);
      if (!timeRange) return undefined;

      const comparisonTimeRange = getComparisonTimeRange(
        dashboardState,
        timeControlState,
        timeRange,
      );

      // api now expects measure names for which comparison are calculated
      let comparisonMeasures: string[] = [];
      if (comparisonTimeRange) {
        comparisonMeasures = [dashboardState.leaderboardMeasureName];
      }

      return {
        instanceId: get(runtime).instanceId,
        metricsViewName: metricViewName,
        dimension: {
          name: dashboardState.selectedDimensionName,
        },
        measures: [...dashboardState.visibleMeasureKeys].map((name) => ({
          name: name,
        })),
        comparisonMeasures: comparisonMeasures,
        timeRange,
        ...(comparisonTimeRange ? { comparisonTimeRange } : {}),
        sort: [
          {
            name: dashboardState.leaderboardMeasureName,
            desc: dashboardState.sortDirection === SortDirection.DESCENDING,
            sortType: getQuerySortType(dashboardState.dashboardSortType),
          },
        ],
        where: sanitiseExpression(dashboardState.whereFilter, undefined),
        offset: "0",
      };
    },
  );
}

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

/**
 * Fills in isoDuration based on selection.
 * This is used by scheduled report by using report run time as end time.
 */
function getTimeRange(
  timeControlState: TimeControlState,
  metricsView: V1MetricsViewSpec,
) {
  if (!timeControlState.selectedTimeRange?.name) return undefined;

  const timeRange: V1TimeRange = {};
  switch (timeControlState.selectedTimeRange.name) {
    case TimeRangePreset.DEFAULT:
      timeRange.isoDuration = metricsView.defaultTimeRange;
      break;

    case TimeRangePreset.CUSTOM:
      timeRange.start = timeControlState.timeStart;
      timeRange.end = timeControlState.timeEnd;
      break;

    default:
      if (timeControlState.selectedTimeRange.name in PreviousCompleteRangeMap) {
        // Backend doesn't support previous complete ranges since it has offset built in.
        // We add the offset manually as a workaround for now
        timeRange.isoDuration =
          PreviousCompleteRangeMap[
            timeControlState.selectedTimeRange.name
          ]?.isoDuration;
        timeRange.isoOffset =
          PreviousCompleteRangeMap[
            timeControlState.selectedTimeRange.name
          ]?.isoOffset;
        timeRange.roundToGrain =
          PreviousCompleteRangeMap[
            timeControlState.selectedTimeRange.name
          ]?.roundToGrain;
      } else {
        timeRange.isoDuration = timeControlState.selectedTimeRange.name;
      }
      break;
  }

  return timeRange;
}

/**
 * Fills in isoDuration and isoOffset based on selection.
 * This is used by scheduled report by using report run time as end time.
 */
function getComparisonTimeRange(
  dashboardState: MetricsExplorerEntity,
  timeControlState: TimeControlState,
  timeRange: V1TimeRange | undefined,
) {
  if (
    !timeRange ||
    dashboardState.selectedComparisonDimension ||
    !timeControlState.showComparison ||
    !timeControlState.selectedComparisonTimeRange?.name
  ) {
    return undefined;
  }

  const comparisonTimeRange: V1TimeRange = {};
  switch (timeControlState.selectedComparisonTimeRange.name) {
    default:
      comparisonTimeRange.isoOffset =
        timeControlState.selectedComparisonTimeRange.name;
    // eslint-disable-next-line no-fallthrough
    case TimeComparisonOption.CONTIGUOUS:
      comparisonTimeRange.isoDuration = timeRange.isoDuration;
      break;

    case TimeComparisonOption.CUSTOM:
      comparisonTimeRange.start = timeControlState.comparisonTimeStart;
      comparisonTimeRange.end = timeControlState.comparisonTimeEnd;
      break;
  }
  return comparisonTimeRange;
}
