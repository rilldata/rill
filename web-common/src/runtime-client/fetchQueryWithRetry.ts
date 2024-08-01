import { asyncWait } from "@rilldata/web-common/lib/waitUtils";
import { FetchQueryOptions } from "@tanstack/query-core";
import { QueryClient, QueryKey } from "@tanstack/svelte-query";

/**
 * A query cancellation due to a possible unmount will not retry `fetchQuery` even if `retry` is set.
 * This wraps it with an explicit retry.
 */
export async function fetchQueryWithRetry<T>(
  queryClient: QueryClient,
  args: FetchQueryOptions<T, unknown, T, QueryKey>,
  retryCount = 3,
  retryDelay = 100,
) {
  let lastErr;
  for (let c = 0; c < retryCount; c++) {
    try {
      return await queryClient.fetchQuery<T>(args);
    } catch (e) {
      if (c < retryCount - 1) {
        // wait for some time before retrying
        await asyncWait(retryDelay * (c + 1));
      }
      lastErr = e;
    }
  }
  throw lastErr;
}
