import { get } from "svelte/store";
import { runtime } from "../../../runtime-client/runtime-store";
import { useDashboardStore } from "../dashboard-stores";
import type {
  V1ExportFormat,
  createQueryServiceExport,
} from "@rilldata/web-common/runtime-client";

export default async function exportMetrics({
  query,
  metricViewName,
  format,
}: {
  query: ReturnType<typeof createQueryServiceExport>;
  metricViewName: string;
  format: V1ExportFormat;
}) {
  const dashboardStore = useDashboardStore(metricViewName);
  const dashboard = get(dashboardStore);
  const result = await get(query).mutateAsync({
    instanceId: get(runtime).instanceId,
    data: {
      format,
      metricsViewRowsRequest: {
        instanceId: get(runtime).instanceId,
        metricsViewName: metricViewName,
        filter: dashboard.filters,
        timeStart: dashboard.selectedTimeRange?.start?.toISOString(),
        timeEnd: dashboard.selectedTimeRange?.end?.toISOString(),
      },
    },
  });
  const downloadUrl = `${get(runtime).host}${result.downloadUrlPath}`;

  window.open(downloadUrl, "_self");
}
