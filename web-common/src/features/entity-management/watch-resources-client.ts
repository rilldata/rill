import {
  invalidateAllResources,
  invalidateResourceResponse,
} from "@rilldata/web-common/features/entity-management/resource-invalidations";
import { WatchRequestClient } from "@rilldata/web-common/runtime-client/watch-request-client";
import type { V1WatchResourcesResponse } from "@rilldata/web-common/runtime-client";
import type { QueryClient } from "@tanstack/svelte-query";

export function startWatchResourcesClient(queryClient: QueryClient) {
  return new WatchRequestClient<V1WatchResourcesResponse>(
    (runtime) =>
      `${runtime.host}/v1/instances/${runtime.instanceId}/resources/-/watch`,
    (res) => handleWatchResourceResponse(queryClient, res),
    () => invalidateAllResources(queryClient),
  ).start();
}

function handleWatchResourceResponse(
  queryClient: QueryClient,
  res: V1WatchResourcesResponse,
) {
  if (!res.resource) return;

  invalidateResourceResponse(queryClient, res);
}
