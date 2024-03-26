import { fileArtifactsStore } from "@rilldata/web-common/features/entity-management/file-artifacts-store";
import {
  getRuntimeServiceGetFileQueryKey,
  getRuntimeServiceListFilesQueryKey,
  V1WatchFilesResponse,
} from "@rilldata/web-common/runtime-client";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { WatchRequestClient } from "@rilldata/web-common/runtime-client/watch-request-client";
import type { QueryClient } from "@tanstack/svelte-query";
import { get } from "svelte/store";
import { removeLeadingSlash } from "./entity-mappers";

export function startWatchFilesClient(queryClient: QueryClient) {
  return new WatchRequestClient<V1WatchFilesResponse>(
    (runtime) =>
      `${runtime.host}/v1/instances/${runtime.instanceId}/files/watch`,
    (res) => handleWatchFileResponse(queryClient, res),
    () => invalidateAllFiles(queryClient),
  ).start();
}

function handleWatchFileResponse(
  queryClient: QueryClient,
  res: V1WatchFilesResponse,
) {
  if (!res?.path || res.path.includes(".db")) return;

  // Watch file returns events for all files under the project. Ignore everything except .sql, .yaml & .yml
  if (
    !res.path.endsWith(".sql") &&
    !res.path.endsWith(".yaml") &&
    !res.path.endsWith(".yml")
  )
    return;

  const instanceId = get(runtime).instanceId;
  const cleanedPath = removeLeadingSlash(res.path);
  // invalidations will wait until the re-fetched query is completed
  // so, we should not `await` here on `refetchQueries`
  switch (res.event) {
    case "FILE_EVENT_WRITE":
      void queryClient.refetchQueries(
        getRuntimeServiceGetFileQueryKey(instanceId, cleanedPath),
      );
      fileArtifactsStore.fileUpdated(cleanedPath);
      break;

    case "FILE_EVENT_DELETE":
      queryClient.removeQueries(
        getRuntimeServiceGetFileQueryKey(instanceId, cleanedPath),
      );
      fileArtifactsStore.fileDeleted(cleanedPath);
      break;
  }
  // TODO: should this be throttled?
  void queryClient.refetchQueries(
    getRuntimeServiceListFilesQueryKey(instanceId),
  );
}

async function invalidateAllFiles(queryClient: QueryClient) {
  // TODO: reset project parser errors

  const instanceId = get(runtime).instanceId;
  return queryClient.resetQueries({
    predicate: (query) =>
      query.queryHash.includes(`v1/instances/${instanceId}/files`),
  });
}
