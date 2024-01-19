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

  const result = await get(query).mutateAsync({
    instanceId: get(runtime).instanceId,
    data: {
      format,
      query: {
        metricsViewAggregationRequest: {
          dimensions: [
            { name: dashboard.selectedComparisonDimension },
            {
              name: timeDimension,
              timeGrain: dashboard.selectedTimeRange?.interval,
              timeZone: dashboard.selectedTimezone,
            },
          ],
          filter: dashboard.filters,
          having: undefined,
          instanceId: get(runtime).instanceId,
          limit: undefined, // the backend handles export limits
          measures: [{ name: dashboard.expandedMeasureName }],
          metricsView,
          offset: "0",
          pivotOn: [timeDimension], // spreads the time dimension across columns
          priority: undefined,
          sort: undefined, // future work
          timeEnd: undefined,
          timeRange: {
            start: selectedTimeRange?.start.toISOString(),
            end: selectedTimeRange?.end.toISOString(),
          },
          timeStart: undefined,
          where: undefined,
        },
      },
    },
  });

  const downloadUrl = `${get(runtime).host}${result.downloadUrlPath}`;

  window.open(downloadUrl, "_self");
}
