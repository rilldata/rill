import { getQuerySortType } from "@rilldata/web-common/features/dashboards/leaderboard/leaderboard-utils";
import { SortDirection } from "@rilldata/web-common/features/dashboards/proto-state/derived-types";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import type { TimeControlState } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import {
  TimeComparisonOption,
  TimeRangePreset,
  TimeRangeType,
} from "@rilldata/web-common/lib/time/types";
import type {
  V1MetricsViewComparisonRequest,
  V1MetricsViewSpec,
  V1TimeRange,
} from "@rilldata/web-common/runtime-client";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { get } from "svelte/store";

// Temporary. A future PR will add iso to the selection itself.
const TIME_RANGES_TO_ISO: Record<TimeRangeType, string> = {
  [TimeRangePreset.ALL_TIME]: "inf",
  [TimeRangePreset.LAST_SIX_HOURS]: "PT6H",
  [TimeRangePreset.LAST_24_HOURS]: "PT24H",
  [TimeRangePreset.LAST_7_DAYS]: "P7D",
  [TimeRangePreset.LAST_14_DAYS]: "P14D",
  [TimeRangePreset.LAST_4_WEEKS]: "P4W",
  [TimeRangePreset.LAST_12_MONTHS]: "P12M",
  [TimeRangePreset.TODAY]: "rill-TD",
  [TimeRangePreset.WEEK_TO_DATE]: "rill-WTD",
  [TimeRangePreset.MONTH_TO_DATE]: "rill-MTD",
  [TimeRangePreset.QUARTER_TO_DATE]: "rill-QTD",
  [TimeRangePreset.YEAR_TO_DATE]: "rill-YTD",
};

export function getDimensionTableExportArgs(
  metricViewName: string,
  dashboardStore: MetricsExplorerEntity,
  timeControlState: TimeControlState,
  metricsView: V1MetricsViewSpec | undefined
): V1MetricsViewComparisonRequest | undefined {
  if (!metricsView) return undefined;

  const timeRange = getTimeRange(timeControlState, metricsView);

  const comparisonTimeRange = getComparisonTimeRange(
    dashboardStore,
    timeControlState,
    timeRange
  );

  return {
    instanceId: get(runtime).instanceId,
    metricsViewName: metricViewName,
    dimension: {
      name: dashboardStore.selectedDimensionName,
    },
    measures: dashboardStore.selectedMeasureNames.map((name) => ({
      name: name,
    })),
    ...(timeRange ? { timeRange } : {}),
    ...(comparisonTimeRange ? { comparisonTimeRange } : {}),
    sort: [
      {
        name: dashboardStore.leaderboardMeasureName,
        desc: dashboardStore.sortDirection === SortDirection.DESCENDING,
        type: getQuerySortType(dashboardStore.dashboardSortType),
      },
    ],
    filter: dashboardStore.filters,
    offset: "0",
  };
}

/**
 * Fills in isoDuration based on selection.
 * This is used by scheduled report by using report run time as end time.
 */
function getTimeRange(
  timeControlState: TimeControlState,
  metricsView: V1MetricsViewSpec
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
      timeRange.isoDuration =
        TIME_RANGES_TO_ISO[timeControlState.selectedTimeRange.name];
      break;
  }

  return timeRange;
}

/**
 * Fills in isoDuration and isoOffset based on selection.
 * This is used by scheduled report by using report run time as end time.
 */
function getComparisonTimeRange(
  dashboardStore: MetricsExplorerEntity,
  timeControlState: TimeControlState,
  timeRange: V1TimeRange | undefined
) {
  if (
    !timeRange ||
    dashboardStore.selectedComparisonDimension ||
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
