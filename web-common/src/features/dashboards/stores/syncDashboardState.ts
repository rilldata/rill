import type { CompoundQueryResult } from "@rilldata/web-common/features/compound-query-result";
import {
  createMetricsViewSchema,
  createTimeRangeSummary,
  useMetricsView,
} from "@rilldata/web-common/features/dashboards/selectors/index";
import type { StateManagers } from "@rilldata/web-common/features/dashboards/state-managers/state-managers";
import { metricsExplorerStore } from "@rilldata/web-common/features/dashboards/stores/dashboard-stores";
import { derived, get } from "svelte/store";

export function createDashboardStateSync(
  ctx: StateManagers,
  preloadUrlStore: CompoundQueryResult<string>,
) {
  return derived(
    [
      useMetricsView(ctx),
      createTimeRangeSummary(ctx),
      createMetricsViewSchema(ctx),
      preloadUrlStore,
    ],
    ([
      metricsViewSpecRes,
      timeRangeRes,
      metricsViewSchemaRes,
      preloadUrlRes,
    ]) => {
      if (
        // still fetching
        metricsViewSpecRes.isFetching ||
        timeRangeRes.isFetching ||
        metricsViewSchemaRes.isFetching ||
        preloadUrlRes.isFetching ||
        // requests errored out
        !metricsViewSpecRes.data ||
        !timeRangeRes.data ||
        !metricsViewSchemaRes.data?.schema
      ) {
        return false;
      }

      const metricViewName = get(ctx.metricsViewName);
      if (metricViewName in get(metricsExplorerStore).entities) {
        // Successive syncs with metrics view spec
        metricsExplorerStore.sync(metricViewName, metricsViewSpecRes.data);
      } else {
        // Running for the 1st time. Initialise the dashboard store.
        metricsExplorerStore.init(
          metricViewName,
          metricsViewSpecRes.data,
          timeRangeRes.data,
        );
        if (preloadUrlRes.data) {
          // If there is data to be loaded, load it during the init
          metricsExplorerStore.syncFromUrl(
            metricViewName,
            preloadUrlRes.data,
            metricsViewSpecRes.data,
            metricsViewSchemaRes.data.schema,
          );
          // Call sync to make sure changes in dashboard are honoured
          metricsExplorerStore.sync(metricViewName, metricsViewSpecRes.data);
        }
      }
      return true;
    },
  );
}
