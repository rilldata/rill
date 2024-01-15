import type { TimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import type {
  V1ExportFormat,
  createQueryServiceExport,
} from "@rilldata/web-common/runtime-client";
import { get } from "svelte/store";
import { useDashboardStore } from "web-common/src/features/dashboards/stores/dashboard-stores";
import { runtime } from "../../../runtime-client/runtime-store";

export default async function exportTDD({
  query,
  metricViewName,
  format,
  timeControlStore,
  timeDimension,
}: {
  query: ReturnType<typeof createQueryServiceExport>;
  metricViewName: string;
  format: V1ExportFormat;
  // we need this from argument since getContext is called to get the state managers
  // which cannot run outside of component initialisation
  timeControlStore: TimeControlStore;
  timeDimension: string;
}) {
  const dashboardStore = useDashboardStore(metricViewName);
  const timeControlState = get(timeControlStore);

  const dashboard = get(dashboardStore);

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
          metricsView: metricViewName,
          offset: "0",
          pivotOn: [timeDimension], // spreads the time dimension across columns
          priority: undefined,
          sort: undefined, // future work
          timeEnd: undefined,
          timeRange: {
            start: timeControlState.timeStart,
            end: timeControlState.timeEnd,
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
