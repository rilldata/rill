import type { V1Resource } from "@rilldata/web-common/runtime-client";
import { createRuntimeServiceListResources } from "@rilldata/web-common/runtime-client";
import type { CreateQueryResult } from "@tanstack/svelte-query";

/**
 * Generic dashboard listing query that filters for Explore and Canvas resources.
 * Callers can pass additional query options (e.g., refetchInterval) to customize behavior.
 */
export function useDashboards(
  instanceId: string,
  queryOptions?: Record<string, unknown>,
): CreateQueryResult<V1Resource[]> {
  return createRuntimeServiceListResources(instanceId, undefined, {
    query: {
      select: (data) => {
        return (data.resources ?? []).filter((res) => res.canvas || res.explore);
      },
      enabled: !!instanceId,
      ...queryOptions,
    },
  });
}
