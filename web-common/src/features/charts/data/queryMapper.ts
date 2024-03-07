import { createQueryServiceMetricsViewAggregation } from "@rilldata/web-common/runtime-client";
import type { QueryClient } from "@tanstack/svelte-query";

export const queryNameToFunctionMap: Record<
  string,
  (...args: unknown[]) => unknown
> = {
  MetricsViewAggregation: createQueryServiceMetricsViewAggregation,
};

export function getQueryFromDataSpec(
  instanceId: string,
  queryClient: QueryClient,
  queryName: string,
  queryArgsJson: string,
) {
  // FIXME: Expect JSON directly from runtime API
  // Add better type support for JSON args
  const queryArgs = JSON.parse(queryArgsJson);
  if (!queryArgs) return null;

  if (!queryArgs?.["metrics_view"]) return null;
  const { metrics_view: metricViewName, ...body } = queryArgs;

  const queryFunction = queryNameToFunctionMap[queryName];

  if (!queryFunction) {
    console.warn(`No query function found for query name: ${queryName}`);
    return null;
  }

  return queryFunction(instanceId, metricViewName, body, {
    query: { queryClient },
  });
}
