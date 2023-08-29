import { WatchRequestClient } from "@rilldata/web-common/features/reconcile/WatchRequestClient";
import {
  getRuntimeServiceGetResourceQueryKey,
  getRuntimeServiceListResourcesQueryKey,
  V1Resource,
} from "@rilldata/web-common/runtime-client";
import type {
  RuntimeServiceWatchResourcesParams,
  V1WatchResourcesResponse,
} from "@rilldata/web-common/runtime-client";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import type { QueryClient } from "@tanstack/svelte-query";

export function resourcesRequests(queryClient: QueryClient) {
  let client: WatchRequestClient<
    RuntimeServiceWatchResourcesParams,
    V1WatchResourcesResponse
  >;

  return runtime.subscribe((runtime) => {
    client?.cancel();
    if (!runtime?.instanceId || !runtime.host) return;

    client = new WatchRequestClient<
      RuntimeServiceWatchResourcesParams,
      V1WatchResourcesResponse
    >(
      `${runtime.host}/v1/instances/${runtime.instanceId}/resources/-/watch`,
      // TODO: filters
      {}
    );

    handleResourceResponses(client, queryClient, runtime.instanceId);
  });
}

async function handleResourceResponses(
  client: WatchRequestClient<
    RuntimeServiceWatchResourcesParams,
    V1WatchResourcesResponse
  >,
  queryClient: QueryClient,
  instanceId: string
) {
  for await (const res of client.send(() =>
    invalidateAllResources(queryClient, instanceId)
  )) {
    if (!res.resource) continue;
    console.log(`File: ${res.event}: ${res.resource.meta.name.name}`);

    // invalidations will wait until the re-fetched query is completed
    // so, we should not `await` here
    switch (res.event) {
      case "RESOURCE_EVENT_ADDED":
        queryClient.refetchQueries(
          getRuntimeServiceListResourcesQueryKey(instanceId)
        );
      // eslint-disable-next-line no-fallthrough
      case "RESOURCE_EVENT_UPDATED_SPEC":
      case "RESOURCE_EVENT_UPDATED_STATE":
        invalidateResource(queryClient, instanceId, res.resource);
        break;

      case "RESOURCE_EVENT_DELETED":
        invalidateRemovedResource(queryClient, instanceId, res.resource);
        queryClient.refetchQueries(
          getRuntimeServiceListResourcesQueryKey(instanceId)
        );
        break;
    }
  }
}

async function invalidateResource(
  queryClient: QueryClient,
  instanceId: string,
  resource: V1Resource
) {
  return queryClient.refetchQueries(
    getRuntimeServiceGetResourceQueryKey(instanceId, {
      "name.name": resource.meta.name.name,
      "name.kind": resource.meta.name.kind,
    })
  );
  // TODO: invalidate individual queries when we swap over
}

async function invalidateRemovedResource(
  queryClient: QueryClient,
  instanceId: string,
  resource: V1Resource
) {
  queryClient.removeQueries(
    getRuntimeServiceGetResourceQueryKey(instanceId, {
      "name.name": resource.meta.name.name,
      "name.kind": resource.meta.name.kind,
    })
  );
  // TODO: remove individual queries when we swap over
}

async function invalidateAllResources(
  queryClient: QueryClient,
  instanceId: string
) {
  queryClient.removeQueries({
    type: "inactive",
    predicate: (query) =>
      query.queryHash.includes(`v1/instances/${instanceId}/resources`),
  });

  return queryClient.refetchQueries({
    type: "active",
    predicate: (query) =>
      query.queryHash.includes(`v1/instances/${instanceId}/resources`),
  });
  // TODO: invalidate individual queries when we swap over
}
