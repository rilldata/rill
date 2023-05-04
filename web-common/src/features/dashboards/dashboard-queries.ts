import { isMetricsViewQuery } from "@rilldata/web-common/runtime-client/invalidation";
import type { QueryClient } from "@tanstack/svelte-query";

export function cancelDashboardQueries(
  queryClient: QueryClient,
  metricsViewName: string
) {
  return queryClient.cancelQueries({
    fetching: true,
    predicate: (query) => {
      return isMetricsViewQuery(query.queryHash, metricsViewName);
    },
  });
}
