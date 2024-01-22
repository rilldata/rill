import type { TimeControlStore } from "@rilldata/web-common/features/dashboards/time-controls/time-control-store";
import { get } from "svelte/store";
import { runtime } from "../../../runtime-client/runtime-store";
import { useDashboardStore } from "web-common/src/features/dashboards/stores/dashboard-stores";
import type {
  V1ExportFormat,
  createQueryServiceExport,
} from "@rilldata/web-common/runtime-client";

export default async function exportMetrics({
  query,
  metricViewName,
  format,
  timeControlStore,
}: {
  query: ReturnType<typeof createQueryServiceExport>;
  metricViewName: string;
  format: V1ExportFormat;
  // we need this from argument since getContext is called to get the state managers
  // which cannot run outside of component initialisation
  timeControlStore: TimeControlStore;
}) {
  const dashboardStore = useDashboardStore(metricViewName);
  const timeControlState = get(timeControlStore);

  const dashboard = get(dashboardStore);
  const result = await get(query).mutateAsync({
    instanceId: get(runtime).instanceId,
    data: {
      format,
      query: {
        metricsViewRowsRequest: {
          instanceId: get(runtime).instanceId,
          metricsViewName: metricViewName,
          where: dashboard.whereFilter,
          timeStart: timeControlState.timeStart,
          timeEnd: timeControlState.timeEnd,
        },
      },
    },
  });
  const downloadUrl = `${get(runtime).host}${result.downloadUrlPath}`;

  window.open(downloadUrl, "_self");
}
