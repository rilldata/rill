import { isMetricsViewData } from "@rilldata/web-local/lib/svelte-query/invalidation";
import type { QueryClient } from "@sveltestack/svelte-query";

export function cancelDashboardQueries(
  queryClient: QueryClient,
  metricsViewName: string
) {
  return queryClient.cancelQueries({
    fetching: true,
    predicate: (query) => {
      return isMetricsViewData(query.queryHash, metricsViewName);
    },
  });
}
