import type { V1Bookmark } from "@rilldata/web-admin/client";
import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
import { useValidExplore } from "@rilldata/web-common/features/explores/selectors";
import { createQueryServiceMetricsViewSchema } from "@rilldata/web-common/runtime-client";
import { get } from "svelte/store";

export function createBookmarkApplier(instanceId: string, exploreName: string) {
  const validExploreSpec = useValidExplore(instanceId, exploreName);
  const metricsSchema = createQueryServiceMetricsViewSchema(
    instanceId,
    exploreName,
  );

  return (bookmark: V1Bookmark) => {
    const validExploreSpecResp = get(validExploreSpec);
    const metricsSchemaResp = get(metricsSchema);
    if (
      !bookmark.data ||
      !validExploreSpecResp.data?.metricsView ||
      !validExploreSpecResp.data?.explore ||
      !metricsSchemaResp.data?.schema
    ) {
      return;
    }
    metricsExplorerStore.syncFromUrl(
      exploreName,
      decodeURIComponent(bookmark.data),
      validExploreSpecResp.data.metricsView,
      validExploreSpecResp.data.explore,
      metricsSchemaResp.data.schema,
    );
  };
}
