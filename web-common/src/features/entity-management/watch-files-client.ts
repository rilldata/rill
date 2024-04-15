import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
import {
  getRuntimeServiceGetFileQueryKey,
  getRuntimeServiceListFilesQueryKey,
  V1WatchFilesResponse,
} from "@rilldata/web-common/runtime-client";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { WatchRequestClient } from "@rilldata/web-common/runtime-client/watch-request-client";
import { get } from "svelte/store";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";

export function createWatchFilesClient() {
  const client = new WatchRequestClient<V1WatchFilesResponse>();
  client.on("response", handleWatchFileResponse);
  client.on("reconnect", invalidateAllFiles);

  return client;
}

function handleWatchFileResponse(res: V1WatchFilesResponse) {
  if (!res?.path || res.path.includes(".db")) return;

  const instanceId = get(runtime).instanceId;
  // invalidations will wait until the re-fetched query is completed
  // so, we should not `await` here on `refetchQueries`
  switch (res.event) {
    case "FILE_EVENT_WRITE":
      void queryClient.resetQueries(
        getRuntimeServiceGetFileQueryKey(instanceId, res.path),
      );
      fileArtifacts.fileUpdated(res.path);
      break;

    case "FILE_EVENT_DELETE":
      void queryClient.resetQueries(
        getRuntimeServiceGetFileQueryKey(instanceId, res.path),
      );
      fileArtifacts.fileDeleted(res.path);
      break;
  }
  // TODO: should this be throttled?
  void queryClient.refetchQueries(
    getRuntimeServiceListFilesQueryKey(instanceId),
  );
}

async function invalidateAllFiles() {
  // TODO: reset project parser errors
  const instanceId = get(runtime).instanceId;
  return queryClient.resetQueries({
    predicate: (query) =>
      query.queryHash.includes(`v1/instances/${instanceId}/files`),
  });
}
