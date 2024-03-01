import type { V1Bookmark } from "@rilldata/web-admin/client";
import { useMetricsView } from "@rilldata/web-common/features/dashboards/selectors";
import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
import { createQueryServiceMetricsViewSchema } from "@rilldata/web-common/runtime-client";
import { get } from "svelte/store";

export function createBookmarkApplier(
  instanceId: string,
  metricsViewName: string,
) {
  const metricsViewSpec = useMetricsView(instanceId, metricsViewName);
  const metricsSchema = createQueryServiceMetricsViewSchema(
    instanceId,
    metricsViewName,
  );

  return (bookmark: V1Bookmark) => {
    const metricsViewSpecResp = get(metricsViewSpec);
    const metricsSchemaResp = get(metricsSchema);
    if (
      !bookmark.data ||
      !metricsViewSpecResp.data ||
      !metricsSchemaResp.data?.schema
    )
      return;
    metricsExplorerStore.syncFromUrl(
      metricsViewName,
      decodeURIComponent(bookmark.data),
      metricsViewSpecResp.data,
      metricsSchemaResp.data.schema,
    );
  };
}
