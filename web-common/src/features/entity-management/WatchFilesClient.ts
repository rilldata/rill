import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
import {
  getRuntimeServiceGetFileQueryKey,
  getRuntimeServiceIssueDevJWTQueryKey,
  getRuntimeServiceListFilesQueryKey,
  V1FileEvent,
  V1WatchFilesResponse,
} from "@rilldata/web-common/runtime-client";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { WatchRequestClient } from "@rilldata/web-common/runtime-client/watch-request-client";
import { get } from "svelte/store";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";

export class WatchFilesClient {
  private readonly client: WatchRequestClient<V1WatchFilesResponse>;
  private seenFiles = new Set<string>();

  public constructor() {
    this.client = new WatchRequestClient<V1WatchFilesResponse>();
    this.client.on("response", (res) => this.handleWatchFileResponse(res));
    this.client.on("reconnect", () => this.invalidateAllFiles());
  }

  public static New() {
    const watcher = new WatchFilesClient();
    return watcher.client;
  }

  private invalidateAllFiles() {
    // TODO: reset project parser errors
    const instanceId = get(runtime).instanceId;
    return queryClient.resetQueries({
      predicate: (query) =>
        query.queryHash.includes(`v1/instances/${instanceId}/files`),
    });
  }

  private handleWatchFileResponse(res: V1WatchFilesResponse) {
    if (!res?.path || res.path.includes(".db")) return;

    const instanceId = get(runtime).instanceId;
    const isNew = this.seenFiles.has(res.path);

    // invalidations will wait until the re-fetched query is completed
    // so, we should not `await` here on `refetchQueries`
    if (!res.isDir) {
      switch (res.event) {
        case V1FileEvent.FILE_EVENT_WRITE:
          void queryClient.refetchQueries(
            getRuntimeServiceGetFileQueryKey(instanceId, { path: res.path }),
          );
          void fileArtifacts.fileUpdated(res.path);
          if (res.path === "/rill.yaml") {
            // If it's a rill.yaml file, invalidate the dev JWT queries
            void queryClient.invalidateQueries(
              getRuntimeServiceIssueDevJWTQueryKey(),
            );
          }
          this.seenFiles.add(res.path);
          break;

        case V1FileEvent.FILE_EVENT_DELETE:
          void queryClient.resetQueries(
            getRuntimeServiceGetFileQueryKey(instanceId, { path: res.path }),
          );
          fileArtifacts.fileDeleted(res.path);
          this.seenFiles.delete(res.path);
          break;
      }
    }
    if (isNew || res.event === V1FileEvent.FILE_EVENT_DELETE) {
      // TODO: should this be throttled?
      void queryClient.refetchQueries(
        getRuntimeServiceListFilesQueryKey(instanceId),
      );
    }
  }
}
