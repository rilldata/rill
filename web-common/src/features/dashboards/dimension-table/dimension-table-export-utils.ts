import { getComparisonRequestMeasures } from "@rilldata/web-common/features/dashboards/dashboard-utils";
import {
  ComparisonDeltaAbsoluteSuffix,
  ComparisonDeltaRelativeSuffix,
} from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
import { SortDirection } from "@rilldata/web-common/features/dashboards/proto-state/derived-types";
import { useMetricsView } from "@rilldata/web-common/features/dashboards/selectors/index";
import type { StateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
import { sanitiseExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import {
  mapComparisonTimeRange,
  mapTimeRange,
} from "@rilldata/web-common/features/dashboards/time-controls/time-range-mappers";
import { DashboardState_LeaderboardSortType } from "@rilldata/web-common/proto/gen/rill/ui/v1/dashboard_pb";
import type {
  V1MetricsViewAggregationMeasure,
  V1MetricsViewAggregationRequest,
} from "@rilldata/web-common/runtime-client";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { derived, get, Readable } from "svelte/store";

export function getDimensionTableExportArgs(
  ctx: StateManagers,
): Readable<V1MetricsViewAggregationRequest | undefined> {
  return derived(
    [
      ctx.metricsViewName,
      ctx.dashboardStore,
      useTimeControlStore(ctx),
      useMetricsView(ctx),
    ],
    ([metricViewName, dashboardState, timeControlState, metricsView]) => {
      if (!metricsView.data || !timeControlState.ready) return undefined;

      const timeRange = mapTimeRange(timeControlState, metricsView.data);
      if (!timeRange) return undefined;

      const comparisonTimeRange = mapComparisonTimeRange(
        dashboardState,
        timeControlState,
        timeRange,
      );

      // api now expects measure names for which comparison are calculated
      let comparisonMeasures: string[] = [];
      if (comparisonTimeRange) {
        comparisonMeasures = [dashboardState.leaderboardMeasureName];
      }

      const measures: V1MetricsViewAggregationMeasure[] = [
        ...dashboardState.visibleMeasureKeys,
      ].map((name) => ({
        name: name,
      }));

      let apiSortName = dashboardState.leaderboardMeasureName;
      if (comparisonTimeRange && timeControlState.selectedComparisonTimeRange) {
        // insert beside the correct measure
        measures.splice(
          measures.findIndex((m) => m.name === apiSortName),
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

      return {
        instanceId: get(runtime).instanceId,
        metricsView: metricViewName,
        dimension: {
          name: dashboardState.selectedDimensionName,
        },
        measures,
        comparisonMeasures: comparisonMeasures,
        timeRange,
        ...(comparisonTimeRange ? { comparisonTimeRange } : {}),
        sort: [
          {
            name: apiSortName,
            desc: dashboardState.sortDirection === SortDirection.DESCENDING,
          },
        ],
        where: sanitiseExpression(dashboardState.whereFilter, undefined),
        offset: "0",
      };
    },
  );
}
