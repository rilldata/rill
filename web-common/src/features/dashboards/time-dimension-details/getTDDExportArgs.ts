import { getQuerySortType } from "@rilldata/web-common/features/dashboards/leaderboard/leaderboard-utils";
import { SortDirection } from "@rilldata/web-common/features/dashboards/proto-state/derived-types";
import { useMetricsView } from "@rilldata/web-common/features/dashboards/selectors/index";
import type { StateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
import {
  createAndExpression,
  createSubQueryExpression,
  filterExpressions,
  sanitiseExpression,
} from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { MetricsExplorerEntity } from "@rilldata/web-common/features/dashboards/stores/metrics-explorer-entity";
import {
  TimeControlState,
  useTimeControlStore,
} from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import {
  mapComparisonTimeRange,
  mapTimeRange,
} from "@rilldata/web-common/features/dashboards/time-controls/time-range-mappers";
import {
  V1MetricsViewAggregationRequest,
  V1MetricsViewSpec,
} from "@rilldata/web-common/runtime-client";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { derived, get, Readable } from "svelte/store";

export function tddExportArgsSelector(
  metricsViewName: string,
  dashboardState: MetricsExplorerEntity,
  timeControlState: TimeControlState,
  metricsView: V1MetricsViewSpec | undefined,
) {
  if (
    !metricsView ||
    !timeControlState.ready ||
    !dashboardState.tdd.expandedMeasureName
  )
    return undefined;

  const timeRange = mapTimeRange(timeControlState, metricsView);
  if (!timeRange) return undefined;

  const comparisonTimeRange = mapComparisonTimeRange(
    dashboardState,
    timeControlState,
    timeRange,
  );

  // api now expects measure names for which comparison are calculated
  let comparisonMeasures: string[] = [];
  if (comparisonTimeRange) {
    comparisonMeasures = [dashboardState.tdd.expandedMeasureName];
  }

  // CAST SAFETY: exports are only available in TDD when a comparison dimension is selected
  const dimensionName = dashboardState.selectedComparisonDimension as string;

  const where =
    filterExpressions(dashboardState.whereFilter, () => true) ??
    createAndExpression([]);
  where.cond?.exprs?.push(
    ...dashboardState.dimensionThresholdFilters.map((dt) =>
      createSubQueryExpression(dt.name, undefined, dt.filter),
    ),
  );

  return {
    instanceId: get(runtime).instanceId,
    metricsViewName,
    dimensions: [
      { name: dimensionName },
      {
        name: metricsView.timeDimension ?? "",
        timeGrain: dashboardState.selectedTimeRange?.interval,
        timeZone: dashboardState.selectedTimezone,
      },
    ],
    measures: [{ name: dashboardState.tdd.expandedMeasureName }],
    comparisonMeasures: comparisonMeasures,
    timeRange,
    ...(comparisonTimeRange ? { comparisonTimeRange } : {}),
    sort: [
      {
        name: dashboardState.tdd.expandedMeasureName,
        desc: dashboardState.sortDirection === SortDirection.DESCENDING,
        sortType: getQuerySortType(dashboardState.dashboardSortType),
      },
    ],
    where: sanitiseExpression(where, undefined),
    offset: "0",
  };
}

export function getTDDExportArgs(
  ctx: StateManagers,
): Readable<V1MetricsViewAggregationRequest | undefined> {
  return derived(
    [
      ctx.metricsViewName,
      ctx.dashboardStore,
      useTimeControlStore(ctx),
      useMetricsView(ctx),
    ],
    ([metricsViewName, dashboardState, timeControlState, metricsView]) =>
      tddExportArgsSelector(
        metricsViewName,
        dashboardState,
        timeControlState,
        metricsView.data,
      ),
  );
}
