import { get } from "svelte/store";
import { runtime } from "../../../runtime-client/runtime-store";
import { useDashboardStore, useFetchTimeRange } from "../dashboard-stores";
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
  const fetchTimeStore = useFetchTimeRange(metricViewName);

  const dashboard = get(dashboardStore);
  const time = get(fetchTimeStore);
  const result = await get(query).mutateAsync({
    instanceId: get(runtime).instanceId,
    data: {
      format,
      metricsViewRowsRequest: {
        instanceId: get(runtime).instanceId,
        metricsViewName: metricViewName,
        filter: dashboard.filters,
        timeStart: time?.start?.toISOString(),
        timeEnd: time?.end?.toISOString(),
      },
    },
  });
  const downloadUrl = `${get(runtime).host}${result.downloadUrlPath}`;

  window.open(downloadUrl, "_self");
}
