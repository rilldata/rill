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
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { get } from "svelte/store";
import { buildWhereParamForDimensionTableAndTDDExports } from "../../exports/export-filters";
import { dimensionSearchText as dimensionSearchTextStore } from "../stores/dashboard-stores";

export function getDimensionTableExportQuery(
  ctx: StateManagers,
  isScheduled: boolean,
): V1Query | undefined {
  const metricsViewName = get(ctx.metricsViewName);
  const dashboardState = get(ctx.dashboardStore);
  const timeControlState = get(useTimeControlStore(ctx));
  const validSpecStore = get(ctx.validSpecStore);
  const dimensionSearchText = get(dimensionSearchTextStore);

  if (!validSpecStore.data?.explore || !timeControlState.ready)
    return undefined;

  let timeRange: V1TimeRange | undefined;
  let comparisonTimeRange: V1TimeRange | undefined;
  if (isScheduled) {
    timeRange = mapSelectedTimeRangeToV1TimeRange(
      timeControlState,
      dashboardState.selectedTimezone,
      validSpecStore.data.explore,
    );
    comparisonTimeRange = mapSelectedComparisonTimeRangeToV1TimeRange(
      timeControlState,
      timeRange,
    );
  } else {
    // NOTE: This is currently needed to ensure the on-demand exports have the same time range as seen on-screen. Currently,
    // the client-side interpretation of time ranges is not the same as the server-side interpretation.
    timeRange = {
      start: timeControlState.timeStart,
      end: timeControlState.timeEnd,
    };
    comparisonTimeRange = {
      start: timeControlState.comparisonTimeStart,
      end: timeControlState.comparisonTimeEnd,
    };
  }
  if (!timeRange) return undefined;

  const query: V1Query = {
    metricsViewAggregationRequest: getDimensionTableAggregationRequestForTime(
      metricsViewName,
      dashboardState,
      timeRange,
      comparisonTimeRange,
      dimensionSearchText,
    ),
  };

  return query;
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

  const where = buildWhereParamForDimensionTableAndTDDExports(
    dashboardState.whereFilter,
    dashboardState.dimensionThresholdFilters,
    dashboardState.selectedDimensionName!, // must exist when viewing a dimension table
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
