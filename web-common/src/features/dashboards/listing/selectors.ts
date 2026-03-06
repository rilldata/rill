import type { V1Resource } from "@rilldata/web-common/runtime-client";
import { createRuntimeServiceListResources } from "@rilldata/web-common/runtime-client";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import type { CreateQueryResult } from "@tanstack/svelte-query";

/**
 * Generic dashboard listing query that filters for Explore and Canvas resources.
 * Callers can pass additional query options (e.g., refetchInterval) to customize behavior.
 */
export function useDashboards(
  client: RuntimeClient,
  queryOptions?: Record<string, unknown>,
): CreateQueryResult<V1Resource[]> {
  return createRuntimeServiceListResources(
    client,
    {},
    {
      query: {
        select: (data) => {
          return (data.resources ?? []).filter(
            (res) => res.canvas || res.explore,
          );
        },
        enabled: !!client.instanceId,
        ...queryOptions,
      },
    },
  );
}
