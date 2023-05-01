import { invalidationForMetricsViewData } from "@rilldata/web-local/lib/svelte-query/invalidation";
import type { QueryClient } from "@tanstack/svelte-query";

export async function invalidateDashboardsQueries(
  queryClient: QueryClient,
  dashboardNames: Array<string>
) {
  // TODO: do a greater refactor of invalidations and make this O(N) instead of O(NM)
  queryClient.removeQueries({
    predicate: (query) =>
      dashboardNames.some((dashboardName) =>
        invalidationForMetricsViewData(query, dashboardName)
      ),
    type: "inactive",
  });
  return queryClient.invalidateQueries({
    predicate: (query) =>
      dashboardNames.some((dashboardName) =>
        invalidationForMetricsViewData(query, dashboardName)
      ),
    type: "active",
  });
}
