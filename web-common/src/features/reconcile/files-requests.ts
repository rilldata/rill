import { WatchRequestClient } from "@rilldata/web-common/features/reconcile/WatchRequestClient";
import {
  getRuntimeServiceGetFileQueryKey,
  getRuntimeServiceListFilesQueryKey,
} from "@rilldata/web-common/runtime-client";
import type {
  RuntimeServiceWatchFilesParams,
  V1WatchFilesResponse,
} from "@rilldata/web-common/runtime-client";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import type { QueryClient } from "@tanstack/svelte-query";

export function filesRequests(queryClient: QueryClient) {
  let client: WatchRequestClient<
    RuntimeServiceWatchFilesParams,
    V1WatchFilesResponse
  >;

  return runtime.subscribe((runtime) => {
    client?.cancel();
    if (!runtime?.instanceId || !runtime.host) return;

    client = new WatchRequestClient<
      RuntimeServiceWatchFilesParams,
      V1WatchFilesResponse
    >(
      `${runtime.host}/v1/instances/${runtime.instanceId}/resources/-/watch`,
      // TODO: filters
      {}
    );

    handleFileResponses(client, queryClient, runtime.instanceId);
  });
}

async function handleFileResponses(
  client: WatchRequestClient<
    RuntimeServiceWatchFilesParams,
    V1WatchFilesResponse
  >,
  queryClient: QueryClient,
  instanceId: string
) {
  for await (const res of client.send(() =>
    // When there is a reconnection we need to invalidate all files.
    // This is to make sure invalidations for files changed when disconnected goes through
    invalidateAllFiles(queryClient, instanceId)
  )) {
    console.log(`File: ${res.event}: ${res.path}`);

    // invalidations will wait until the re-fetched query is completed
    // so, we should not `await` here
    switch (res.event) {
      case "FILE_EVENT_WRITE":
        queryClient.refetchQueries(
          getRuntimeServiceGetFileQueryKey(instanceId, res.path)
        );
        break;

      case "FILE_EVENT_DELETE":
        queryClient.removeQueries(
          getRuntimeServiceGetFileQueryKey(instanceId, res.path)
        );
        break;
    }
    // TODO: should this be throttled?
    queryClient.refetchQueries(getRuntimeServiceListFilesQueryKey(instanceId));
  }
}

async function invalidateAllFiles(
  queryClient: QueryClient,
  instanceId: string
) {
  queryClient.removeQueries({
    type: "inactive",
    predicate: (query) =>
      query.queryHash.includes(`v1/instances/${instanceId}/files`),
  });

  return queryClient.refetchQueries({
    type: "active",
    predicate: (query) =>
      query.queryHash.includes(`v1/instances/${instanceId}/files`),
  });
}
