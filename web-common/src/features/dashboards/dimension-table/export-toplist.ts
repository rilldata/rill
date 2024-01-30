import { getMeasureFilters } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils";
import type { StateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
import { sanitiseExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type {
  V1ExportFormat,
  V1MetricsViewAggregationMeasure,
  createQueryServiceExport,
} from "@rilldata/web-common/runtime-client";
import { get } from "svelte/store";
import { runtime } from "../../../runtime-client/runtime-store";
import { getQuerySortType } from "../leaderboard/leaderboard-utils";
import { SortDirection } from "../proto-state/derived-types";

export default async function exportToplist({
  ctx,
  query,
  format,
}: {
  ctx: StateManagers;
  query: ReturnType<typeof createQueryServiceExport>;
  format: V1ExportFormat;
}) {
  const metricsViewName = get(ctx.metricsViewName);
  const dashboard = get(ctx.dashboardStore);
  const timeControlState = get(
    ctx.selectors.timeRangeSelectors.timeControlsState,
  );
  const measureFilters = await getMeasureFilters(ctx);

  const result = await get(query).mutateAsync({
    instanceId: get(runtime).instanceId,
    data: {
      format,
      query: {
        metricsViewComparisonRequest: {
          instanceId: get(runtime).instanceId,
          metricsViewName,
          dimension: {
            name: dashboard.selectedDimensionName,
          },
          measures: [...dashboard.visibleMeasureKeys].map(
            (name) =>
              <V1MetricsViewAggregationMeasure>{
                name: name,
              },
          ),
          timeRange: {
            start: timeControlState.timeStart,
            end: timeControlState.timeEnd,
          },
          comparisonTimeRange: {
            start: timeControlState.comparisonTimeStart,
            end: timeControlState.comparisonTimeEnd,
          },
          sort: [
            {
              name: dashboard.leaderboardMeasureName,
              desc: dashboard.sortDirection === SortDirection.DESCENDING,
              sortType: getQuerySortType(dashboard.dashboardSortType),
            },
          ],
          where: sanitiseExpression(dashboard.whereFilter, measureFilters),
          limit: undefined, // the backend handles export limits
          offset: "0",
        },
      },
    },
  });

  const downloadUrl = `${get(runtime).host}${result.downloadUrlPath}`;

  window.open(downloadUrl, "_self");
}
