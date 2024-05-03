import { getQuerySortType } from "@rilldata/web-common/features/dashboards/leaderboard/leaderboard-utils";
import { SortDirection } from "@rilldata/web-common/features/dashboards/proto-state/derived-types";
import { useMetricsView } from "@rilldata/web-common/features/dashboards/selectors/index";
import type { StateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
import { sanitiseExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import { useTimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import {
  mapComparisonTimeRange,
  mapTimeRange,
} from "@rilldata/web-common/features/dashboards/time-controls/time-range-mappers";
import { type V1MetricsViewComparisonRequest } from "@rilldata/web-common/runtime-client";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { derived, get, Readable } from "svelte/store";

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
