import { invalidate } from "$app/navigation";
import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
import { eventBus } from "@rilldata/web-common/lib/event-bus/event-bus";
import { Throttler } from "@rilldata/web-common/lib/throttler";
import type { QueryClient } from "@tanstack/svelte-query";
import {
  getRuntimeServiceGetFileQueryKey,
  getRuntimeServiceGitStatusQueryKey,
  getRuntimeServiceIssueDevJWTQueryKey,
  getRuntimeServiceListFilesQueryKey,
  V1FileEvent,
  type V1WatchFilesResponse,
} from "@rilldata/web-common/runtime-client";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";

const REFETCH_LIST_FILES_THROTTLE_MS = 100;

export interface FileInvalidatorState {
  /** Paths the watcher has seen. Used to decide whether a refetchListFiles is needed. */
  seenFiles: Set<string>;
  refetchListFilesThrottle: Throttler;
}

export function createFileInvalidatorState(): FileInvalidatorState {
  return {
    seenFiles: new Set<string>(),
    refetchListFilesThrottle: new Throttler(
      REFETCH_LIST_FILES_THROTTLE_MS,
      REFETCH_LIST_FILES_THROTTLE_MS,
    ),
  };
}

/**
 * File-event handler. Refetches file content on write, clears cached content
 * on delete, maintains `seenFiles` bookkeeping, and runs a throttled
 * listFiles refetch when the working set changes. `/rill.yaml` is the one
 * path with extra work: it drives the dev JWT cache and app:init loader.
 */
export async function handleFileEvent(
  event: V1WatchFilesResponse,
  queryClient: QueryClient,
  runtimeClient: RuntimeClient,
  state: FileInvalidatorState,
): Promise<void> {
  if (!event?.path || event.path.includes(".db")) return;

  const { instanceId } = runtimeClient;

  const isNew = !state.seenFiles.has(event.path);

  if (!event.isDir) {
    switch (event.event) {
      case V1FileEvent.FILE_EVENT_WRITE:
        await fileArtifacts.getFileArtifact(event.path).fetchContent(true);
        if (event.path === "/rill.yaml") {
          void queryClient.invalidateQueries({
            queryKey: getRuntimeServiceIssueDevJWTQueryKey(instanceId),
          });
          await invalidate("app:init");
          eventBus.emit("rill-yaml-updated");
        }
        state.seenFiles.add(event.path);
        break;

      case V1FileEvent.FILE_EVENT_DELETE:
        void queryClient.resetQueries({
          queryKey: getRuntimeServiceGetFileQueryKey(instanceId, {
            path: event.path,
          }),
        });
        fileArtifacts.removeFile(event.path);
        // The dev JWT is intentionally NOT invalidated on delete: the key is
        // project-bound and the next load of rill.yaml handles re-issuance.
        if (event.path === "/rill.yaml") {
          await invalidate("app:init");
        }
        state.seenFiles.delete(event.path);
        break;
    }

    // Keep the cloud editor's commit button in sync with the working tree.
    void queryClient.invalidateQueries({
      queryKey: getRuntimeServiceGitStatusQueryKey(instanceId, {}),
    });
  }

  // Throttle: when many files arrive at once (e.g. initial sync), one refetch
  // covers the batch.
  if (isNew || event.event === V1FileEvent.FILE_EVENT_DELETE) {
    state.refetchListFilesThrottle.throttle(() =>
      queryClient.refetchQueries({
        queryKey: getRuntimeServiceListFilesQueryKey(instanceId),
      }),
    );
  }
}
