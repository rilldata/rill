import { isMetricsViewQuery } from "@rilldata/web-common/runtime-client/invalidation";
import type { QueryClient } from "@tanstack/svelte-query";

// This is the name of the column that is added to the query when we want to
// count the number of filtered rows available. Stuck a UUID on the end to
// avoid collisions with other columns, this may be overkill. but this column should
// not be looked up by string name in anycase, this const should be used.
export const ROW_COUNT_INLINE_COL_NAME =
  "COUNT(*)_inline_55b1a12c-8b5d-47fc-be1c-97c121623424";
export const ROW_COUNT_INLINE_COL_EXPRESSION = "COUNT(*)";

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
