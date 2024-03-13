import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
import type {
  V1MetricsViewSpec,
  V1MetricsViewTimeRangeResponse,
  V1StructType,
} from "@rilldata/web-common/runtime-client";
import { get } from "svelte/store";

export function syncDashboardState(
  metricViewName: string,
  metricsViewSpec: V1MetricsViewSpec | undefined,
  metricsViewSchema: V1StructType | undefined,
  timeRangeQuery: V1MetricsViewTimeRangeResponse | undefined,
  preloadUrlState: string | null,
) {
  if (!metricsViewSpec || !metricsViewSchema) return;
  if (metricViewName in get(metricsExplorerStore).entities) {
    metricsExplorerStore.sync(metricViewName, metricsViewSpec);
  } else {
    metricsExplorerStore.init(metricViewName, metricsViewSpec, timeRangeQuery);
    if (preloadUrlState) {
      metricsExplorerStore.syncFromUrl(
        metricViewName,
        preloadUrlState,
        metricsViewSpec,
        metricsViewSchema,
      );
      // Call sync to make sure changes in dashboard are honoured
      metricsExplorerStore.sync(metricViewName, metricsViewSpec);
    }
  }
}
