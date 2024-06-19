import { getDimensionFilterWithSearch } from "@rilldata/web-common/features/dashboards/dimension-table/dimension-table-utils";
import { mergeMeasureFilters } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils";
import type { StateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
import { sanitiseExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import {
  fetchResource,
  ResourceKind,
} from "@rilldata/web-common/features/entity-management/resource-selectors";
import {
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
  const instanceId = get(runtime).instanceId;
  const metricsViewResource = await fetchResource(
    ctx.queryClient,
    instanceId,
    metricsViewName,
    ResourceKind.MetricsView,
  );
  const measuresSpec =
    metricsViewResource?.metricsView?.state?.validSpec?.measures ?? [];
  const measures = [...dashboard.visibleMeasureKeys]
    .filter((vm) => {
      const m = measuresSpec.find((m) => m.name === vm);
      return !!m && !m.window && !m.requiredDimensions?.length;
    })
    .map(
      (name) =>
        <V1MetricsViewAggregationMeasure>{
          name: name,
        },
    );

  // CAST SAFETY: by definition, a dimension is selected when in the Dimension Table
  const dimensionName = dashboard.selectedDimensionName as string;

  // api now expects measure names for which comparison are calculated
  let comparisonMeasures: string[] = [];
  if (
    timeControlState.comparisonTimeStart &&
    timeControlState.comparisonTimeStart
  ) {
    comparisonMeasures = [dashboard.leaderboardMeasureName];
  }

  const where = sanitiseExpression(
    mergeMeasureFilters(
      dashboard,
      getDimensionFilterWithSearch(
        dashboard?.whereFilter,
        dashboard?.dimensionSearchText ?? "",
        dimensionName,
      ),
    ),
    undefined,
  );

  const result = await get(query).mutateAsync({
    instanceId: get(runtime).instanceId,
    data: {
      format,
      query: {
        metricsViewComparisonRequest: {
          instanceId,
          metricsViewName,
          dimension: {
            name: dimensionName,
          },
          measures,
          comparisonMeasures: comparisonMeasures,
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
          where,
          limit: undefined, // the backend handles export limits
          offset: "0",
        },
      },
    },
  });

  const downloadUrl = `${get(runtime).host}${result.downloadUrlPath}`;

  window.open(downloadUrl, "_self");
}
