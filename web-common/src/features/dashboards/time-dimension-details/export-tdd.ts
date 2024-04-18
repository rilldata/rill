import { getDimensionFilterWithSearch } from "@rilldata/web-common/features/dashboards/dimension-table/dimension-table-utils";
import { getResolvedMeasureFilters } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-utils";
import { sanitiseExpression } from "@rilldata/web-common/features/dashboards/stores/filter-utils";
import type {
  V1ExportFormat,
  createQueryServiceExport,
} from "@rilldata/web-common/runtime-client";
import { get } from "svelte/store";
import { runtime } from "../../../runtime-client/runtime-store";
import type { StateManagers } from "../state-managers/state-managers";

export default async function exportTDD({
  ctx,
  query,
  format,
  timeDimension,
}: {
  ctx: StateManagers;
  query: ReturnType<typeof createQueryServiceExport>;
  format: V1ExportFormat;
  timeDimension: string;
}) {
  const metricsView = get(ctx.metricsViewName);
  const dashboard = get(ctx.dashboardStore);
  const selectedTimeRange = get(
    ctx.selectors.timeRangeSelectors.selectedTimeRangeState,
  );
  const measureFilters = await getResolvedMeasureFilters(ctx);
  // CAST SAFETY: exports are only available in TDD when a comparison dimension is selected
  const dimensionName = dashboard.selectedComparisonDimension as string;

  const where = sanitiseExpression(
    getDimensionFilterWithSearch(
      dashboard?.whereFilter,
      dashboard?.dimensionSearchText ?? "",
      dimensionName,
    ),
    measureFilters,
  );

  const result = await get(query).mutateAsync({
    instanceId: get(runtime).instanceId,
    data: {
      format,
      query: {
        metricsViewAggregationRequest: {
          dimensions: [
            { name: dimensionName },
            {
              name: timeDimension,
              timeGrain: dashboard.selectedTimeRange?.interval,
              timeZone: dashboard.selectedTimezone,
            },
          ],
          where,
          instanceId: get(runtime).instanceId,
          limit: undefined, // the backend handles export limits
          measures: [{ name: dashboard.tdd.expandedMeasureName }],
          metricsView,
          offset: "0",
          pivotOn: [timeDimension], // spreads the time dimension across columns
          sort: undefined, // future work
          timeRange: {
            start: selectedTimeRange?.start.toISOString(),
            end: selectedTimeRange?.end.toISOString(),
          },
        },
      },
    },
  });

  const downloadUrl = `${get(runtime).host}${result.downloadUrlPath}`;

  window.open(downloadUrl, "_self");
}
