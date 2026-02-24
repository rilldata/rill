import { getComparisonRequestMeasures } from "@rilldata/web-common/features/dashboards/dashboard-utils";
import {
  ComparisonDeltaAbsoluteSuffix,
  ComparisonDeltaRelativeSuffix,
} from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
import { SortDirection } from "@rilldata/web-common/features/dashboards/proto-state/derived-types";
import type { StateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
import type { ExploreState } from "@rilldata/web-common/features/dashboards/stores/explore-state";
import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import {
  mapSelectedComparisonTimeRangeToV1TimeRange,
  mapSelectedTimeRangeToV1TimeRange,
} from "@rilldata/web-common/features/dashboards/time-controls/time-range-mappers";
import { DashboardState_LeaderboardSortType } from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import type {
  V1MetricsViewAggregationMeasure,
  V1MetricsViewAggregationRequest,
  V1Query,
  V1TimeRange,
} from "@rilldata/web-common/runtime-client";
import { get } from "svelte/store";
import { buildWhereParamForDimensionTableAndTDDExports } from "../../exports/export-filters";
import { dimensionSearchText as dimensionSearchTextStore } from "../stores/dashboard-stores";

export function getDimensionTableExportQuery(
  ctx: StateManagers,
  isScheduled: boolean,
): V1Query | undefined {
  const metricsViewName = get(ctx.metricsViewName);
  const exploreState = get(ctx.dashboardStore);
  const timeControlState = get(useTimeControlStore(ctx));
  const validSpecStore = get(ctx.validSpecStore);
  const dimensionSearchText = get(dimensionSearchTextStore);

  if (!validSpecStore.data?.explore || !timeControlState.ready)
    return undefined;

  let timeRange: V1TimeRange | undefined;
  if (isScheduled) {
    timeRange = mapSelectedTimeRangeToV1TimeRange(
      timeControlState.selectedTimeRange,
      exploreState.selectedTimezone,
      validSpecStore.data.explore,
    );
  } else {
    timeRange = {
      start: timeControlState.timeStart,
      end: timeControlState.timeEnd,
    };
  }
  if (!timeRange) return undefined;

  let comparisonTimeRange: V1TimeRange | undefined;
  if (timeControlState.showTimeComparison) {
    if (isScheduled) {
      comparisonTimeRange = mapSelectedComparisonTimeRangeToV1TimeRange(
        timeControlState.selectedComparisonTimeRange,
        timeControlState.showTimeComparison,
        timeRange,
      );
    } else {
      comparisonTimeRange = {
        start: timeControlState.comparisonTimeStart,
        end: timeControlState.comparisonTimeEnd,
      };
    }
  }

  const query: V1Query = {
    metricsViewAggregationRequest: getDimensionTableAggregationRequestForTime({
      instanceId: ctx.runtimeClient.instanceId,
      metricsViewName,
      exploreState,
      timeRange,
      comparisonTimeRange,
      dimensionSearchText,
    }),
  };

  return query;
}

export function getDimensionTableAggregationRequestForTime({
  instanceId,
  metricsViewName,
  exploreState,
  timeRange,
  comparisonTimeRange,
  dimensionSearchText,
}: {
  instanceId: string;
  metricsViewName: string;
  exploreState: ExploreState;
  timeRange: V1TimeRange;
  comparisonTimeRange: V1TimeRange | undefined;
  dimensionSearchText: string;
}): V1MetricsViewAggregationRequest {
  const measures: V1MetricsViewAggregationMeasure[] =
    exploreState.visibleMeasures.map((name) => ({
      name: name,
    }));

  let apiSortName = exploreState.leaderboardSortByMeasureName;
  if (!exploreState.visibleMeasures.includes(apiSortName)) {
    // if selected sort measure is not visible add it to list
    measures.push({ name: apiSortName });
  }
  if (comparisonTimeRange) {
    // insert beside the correct measure
    measures.splice(
      measures.findIndex((m) => m.name === apiSortName) + 1,
      0,
      ...getComparisonRequestMeasures(apiSortName),
    );
    switch (exploreState.dashboardSortType) {
      case DashboardState_LeaderboardSortType.DELTA_ABSOLUTE:
        apiSortName += ComparisonDeltaAbsoluteSuffix;
        break;
      case DashboardState_LeaderboardSortType.DELTA_PERCENT:
        apiSortName += ComparisonDeltaRelativeSuffix;
        break;
    }
  }

  const where = buildWhereParamForDimensionTableAndTDDExports(
    exploreState.whereFilter,
    exploreState.dimensionThresholdFilters,
    exploreState.selectedDimensionName!, // must exist when viewing a dimension table
    dimensionSearchText,
  );

  return {
    instanceId,
    metricsView: metricsViewName,
    dimensions: [
      {
        name: exploreState.selectedDimensionName,
      },
    ],
    measures,
    timeRange,
    ...(comparisonTimeRange ? { comparisonTimeRange } : {}),
    sort: [
      {
        name: apiSortName,
        desc: exploreState.sortDirection === SortDirection.DESCENDING,
      },
    ],
    where,
    offset: "0",
  };
}
