import { SortDirection } from "@rilldata/web-common/features/dashboards/proto-state/derived-types";
import type { StateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
import type { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import {
  type TimeControlState,
  useTimeControlStore,
} from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import { mapSelectedTimeRangeToV1TimeRange } from "@rilldata/web-common/features/dashboards/time-controls/time-range-mappers";
import type {
  V1ExploreSpec,
  V1MetricsViewAggregationMeasure,
  V1MetricsViewAggregationRequest,
  V1MetricsViewSpec,
  V1Query,
} from "@rilldata/web-common/runtime-client";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { get } from "svelte/store";
import { buildWhereParamForDimensionTableAndTDDExports } from "../../exports/export-filters";
import { dimensionSearchText as dimensionSearchTextStore } from "../stores/dashboard-stores";

export function getTDDExportQuery(
  ctx: StateManagers,
  isScheduled: boolean,
): V1Query {
  const metricsViewName = get(ctx.metricsViewName);
  const dashboardState = get(ctx.dashboardStore);
  const timeControlState = get(useTimeControlStore(ctx));
  const validSpec = get(ctx.validSpecStore);
  const dimensionSearchText = get(dimensionSearchTextStore);

  const query: V1Query = {
    metricsViewAggregationRequest: getTDDAggregationRequest(
      metricsViewName,
      dashboardState,
      timeControlState,
      validSpec.data?.metricsView,
      validSpec.data?.explore,
      dimensionSearchText,
      isScheduled,
    ),
  };

  return query;
}

function getTDDAggregationRequest(
  metricsViewName: string,
  dashboardState: MetricsExplorerEntity,
  timeControlState: TimeControlState,
  metricsView: V1MetricsViewSpec | undefined,
  explore: V1ExploreSpec | undefined,
  dimensionSearchText: string,
  isScheduled: boolean,
): undefined | V1MetricsViewAggregationRequest {
  if (
    !metricsView ||
    !explore ||
    !timeControlState.ready ||
    !dashboardState.tdd.expandedMeasureName
  )
    return undefined;

  const timeRange = mapSelectedTimeRangeToV1TimeRange(
    timeControlState,
    dashboardState.selectedTimezone,
    explore,
  );
  if (!timeRange) return undefined;
  if (!isScheduled) {
    // To match the UI's time range, we must explicitly specify `timeEnd` for on-demand exports
    timeRange.end = timeControlState.timeEnd;
  }

  const measures: V1MetricsViewAggregationMeasure[] = [
    { name: dashboardState.tdd.expandedMeasureName },
  ];

  // CAST SAFETY: exports are only available in TDD when a comparison dimension is selected
  const dimensionName = dashboardState.selectedComparisonDimension as string;
  const timeDimension = metricsView.timeDimension ?? "";

  return {
    instanceId: get(runtime).instanceId,
    metricsView: metricsViewName,
    dimensions: [
      { name: dimensionName },
      {
        name: metricsView.timeDimension ?? "",
        timeGrain: dashboardState.selectedTimeRange?.interval,
        timeZone: dashboardState.selectedTimezone,
      },
    ],
    measures,
    timeRange,
    pivotOn: [timeDimension],
    sort: [
      {
        name: dimensionName,
        desc: dashboardState.sortDirection === SortDirection.DESCENDING,
      },
    ],
    where: buildWhereParamForDimensionTableAndTDDExports(
      dashboardState.whereFilter,
      dashboardState.dimensionThresholdFilters,
      dimensionName,
      dimensionSearchText,
    ),
    offset: "0",
  };
}
