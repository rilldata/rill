import {
  invalidateAllResources,
  invalidateResourceResponse,
} from "@rilldata/web-common/features/entity-management/resource-invalidations";
import { WatchRequestClient } from "@rilldata/web-common/runtime-client/watch-request-client";
import type { V1WatchResourcesResponse } from "@rilldata/web-common/runtime-client";
import type { QueryClient } from "@tanstack/svelte-query";

export function createWatchResourceClient(queryClient: QueryClient) {
  const watchResourcesClient =
    new WatchRequestClient<V1WatchResourcesResponse>();

  watchResourcesClient.on("response", (res) =>
    handleWatchResourceResponse(queryClient, res),
  );
  watchResourcesClient.on("reconnect", () =>
    invalidateAllResources(queryClient),
  );

  return watchResourcesClient;
}

function handleWatchResourceResponse(
  queryClient: QueryClient,
  res: V1WatchResourcesResponse,
) {
  if (!res.resource) return;

  invalidateResourceResponse(queryClient, res);
}
