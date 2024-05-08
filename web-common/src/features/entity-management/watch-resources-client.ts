import {
  invalidateAllResources,
  invalidateResourceResponse,
} from "@rilldata/web-common/features/entity-management/resource-invalidations";
import { WatchRequestClient } from "@rilldata/web-common/runtime-client/watch-request-client";
import type { V1WatchResourcesResponse } from "@rilldata/web-common/runtime-client";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";

export function createWatchResourceClient() {
  const client = new WatchRequestClient<V1WatchResourcesResponse>();
  client.on("response", handleWatchResourceResponse);
  client.on("reconnect", invalidateAllResources);

  return client;
}

function handleWatchResourceResponse(res: V1WatchResourcesResponse) {
  void invalidateResourceResponse(queryClient, res);
}
