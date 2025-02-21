import { getComparisonRequestMeasures } from "@rilldata/web-common/features/dashboards/dashboard-utils";
import {
  ComparisonDeltaAbsoluteSuffix,
  ComparisonDeltaRelativeSuffix,
} from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
import { SortDirection } from "@rilldata/web-common/features/dashboards/proto-state/derived-types";
import type { StateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import {
  mapComparisonTimeRange,
  mapTimeRange,
} from "@rilldata/web-common/features/dashboards/time-controls/time-range-mappers";
import { DashboardState_LeaderboardSortType } from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import type {
  V1Expression,
  V1MetricsViewAggregationMeasure,
  V1MetricsViewAggregationRequest,
  V1TimeRange,
} from "@rilldata/web-common/runtime-client";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { derived, get, type Readable } from "svelte/store";
import { mergeMeasureFilters } from "../filters/measure-filters/measure-filter-utils";
import { dimensionSearchText } from "../stores/dashboard-stores";
import { sanitiseExpression } from "../stores/filter-utils";
import { getDimensionFilterWithSearch } from "./dimension-table-utils";

export function getDimensionTableExportArgs(
  ctx: StateManagers,
): Readable<V1MetricsViewAggregationRequest | undefined> {
  return derived(
    [
      ctx.metricsViewName,
      ctx.dashboardStore,
      useTimeControlStore(ctx),
      ctx.validSpecStore,
      dimensionSearchText,
    ],
    ([
      metricsViewName,
      dashboardState,
      timeControlState,
      validSpecStore,
      dimensionSearchText,
    ]) => {
      if (!validSpecStore.data?.explore || !timeControlState.ready)
        return undefined;

      const timeRange = mapTimeRange(
        timeControlState,
        dashboardState.selectedTimezone,
        validSpecStore.data.explore,
      );
      if (!timeRange) return undefined;

      const comparisonTimeRange = mapComparisonTimeRange(
        timeControlState,
        timeRange,
      );

      return getDimensionTableAggregationRequestForTime(
        metricsViewName,
        dashboardState,
        timeRange,
        comparisonTimeRange,
        dimensionSearchText,
      );
    },
  );
}

export function getDimensionTableAggregationRequestForTime(
  metricsView: string,
  dashboardState: MetricsExplorerEntity,
  timeRange: V1TimeRange,
  comparisonTimeRange: V1TimeRange | undefined,
  dimensionSearchText: string,
): V1MetricsViewAggregationRequest {
  const measures: V1MetricsViewAggregationMeasure[] = [
    ...dashboardState.visibleMeasureKeys,
  ].map((name) => ({
    name: name,
  }));

  let apiSortName = dashboardState.leaderboardMeasureName;
  if (!dashboardState.visibleMeasureKeys.has(apiSortName)) {
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
    switch (dashboardState.dashboardSortType) {
      case DashboardState_LeaderboardSortType.DELTA_ABSOLUTE:
        apiSortName += ComparisonDeltaAbsoluteSuffix;
        break;
      case DashboardState_LeaderboardSortType.DELTA_PERCENT:
        apiSortName += ComparisonDeltaRelativeSuffix;
        break;
    }
  }

  const where = buildWhereParam(
    dashboardState,
    dashboardState.selectedDimensionName!,
    dimensionSearchText,
  );

  return {
    instanceId: get(runtime).instanceId,
    metricsView,
    dimensions: [
      {
        name: dashboardState.selectedDimensionName,
      },
    ],
    measures,
    timeRange,
    ...(comparisonTimeRange ? { comparisonTimeRange } : {}),
    sort: [
      {
        name: apiSortName,
        desc: dashboardState.sortDirection === SortDirection.DESCENDING,
      },
    ],
    where,
    offset: "0",
  };
}

export function buildWhereParam(
  dashboard: MetricsExplorerEntity,
  dimensionName: string,
  searchText: string,
) {
  let dimensionFilter: V1Expression | undefined;
  if (searchText) {
    dimensionFilter = getDimensionFilterWithSearch(
      dashboard?.whereFilter,
      searchText,
      dimensionName,
    );
  } else {
    dimensionFilter = dashboard?.whereFilter;
  }
  const where = mergeMeasureFilters(dashboard, dimensionFilter);
  const sanitisedWhere = sanitiseExpression(where, undefined);
  return sanitisedWhere;
}
