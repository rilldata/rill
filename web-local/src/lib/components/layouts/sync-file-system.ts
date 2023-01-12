/* Poll the filesystem under 3 scenarios:
 * - Scenario 1. The user navigates to a new page
 * - Scenario 2. Every X seconds
 * - Scenario 3. The user returns focus to the browser tab
 *
 * It's slightly complicated because we sync a different file depending on the page we're on.
 */

import { afterNavigate, beforeNavigate } from "$app/navigation";
import {
  getRuntimeServiceListFilesQueryKey,
  runtimeServiceReconcile,
} from "@rilldata/web-common/runtime-client";
import type { Page } from "@sveltejs/kit";
import type { QueryClient } from "@sveltestack/svelte-query";
import { get, Readable, Writable } from "svelte/store";
import type { RuntimeState } from "../../application-state-stores/application-store";
import type { FileArtifactsStore } from "../../application-state-stores/file-artifacts-store";
import { invalidateAfterReconcile } from "../../svelte-query/invalidation";
import { getFilePathFromPagePath } from "../../util/entity-mappers";

const SYNC_FILE_SYSTEM_INTERVAL_MILLISECONDS = 5000;

export function syncFileSystemPeriodically(
  queryClient: QueryClient,
  runtimeStore: Writable<RuntimeState>,
  page: Readable<Page<Record<string, string>, string>>,
  fileArtifactsStore: FileArtifactsStore
) {
  let syncFileSystemInterval: NodeJS.Timer;
  let syncFileSystemOnVisibleDocument: () => void;
  let afterNavigateRanOnce: boolean;

  afterNavigate(async () => {
    // on first page load, afterNavigate races against the async onMount, which sets the runtimeInstanceId
    const runtimeInstanceId = await waitForRuntimeInstanceId(runtimeStore);

    // on first page load, afterNavigate runs twice, but we only want to run the below code once
    if (afterNavigateRanOnce) return;

    // Scenario 1: sync when the user navigates to a new page
    syncFileSystem(queryClient, runtimeInstanceId, page, 1, fileArtifactsStore);

    // setup Scenario 2: sync every X seconds
    syncFileSystemInterval = setInterval(
      async () =>
        await syncFileSystem(
          queryClient,
          runtimeInstanceId,
          page,
          2,
          fileArtifactsStore
        ),
      SYNC_FILE_SYSTEM_INTERVAL_MILLISECONDS
    );

    // setup Scenario 3: sync when the user returns focus to the browser tab
    syncFileSystemOnVisibleDocument = async () => {
      if (document.visibilityState === "visible") {
        await syncFileSystem(
          queryClient,
          runtimeInstanceId,
          page,
          3,
          fileArtifactsStore
        );
      }
    };
    window.addEventListener("focus", syncFileSystemOnVisibleDocument);

    afterNavigateRanOnce = true;
  });

  beforeNavigate(() => {
    // teardown Scenario 2
    clearInterval(syncFileSystemInterval);

    // teardown Scenario 3
    window.removeEventListener("focus", syncFileSystemOnVisibleDocument);

    afterNavigateRanOnce = false;
  });
}

async function syncFileSystem(
  queryClient: QueryClient,
  instanceId: string,
  page: Readable<Page<Record<string, string>, string>>,
  id: number,
  fileArtifactsStore: FileArtifactsStore
) {
  await queryClient.invalidateQueries(
    getRuntimeServiceListFilesQueryKey(instanceId)
  );

  const pagePath = get(page).url.pathname;
  console.log("syncFileSystem", instanceId, pagePath, id);
  if (!isPathToAsset(pagePath)) return;

  const filePath = getFilePathFromPagePath(pagePath);
  fileArtifactsStore.setIsReconciling(filePath, true);
  const resp = await runtimeServiceReconcile(instanceId, {
    changedPaths: [filePath],
  });
  fileArtifactsStore.setErrors(resp.affectedPaths, resp.errors);
  fileArtifactsStore.setIsReconciling(filePath, false);
  invalidateAfterReconcile(queryClient, instanceId, resp);
}

function isPathToAsset(path: string) {
  return (
    path.startsWith("/source") ||
    path.startsWith("/model") ||
    path.startsWith("/dashboard")
  );
}

async function waitForRuntimeInstanceId(runtimeStore: Writable<RuntimeState>) {
  let runtimeInstanceId;
  while (!runtimeInstanceId) {
    await new Promise((resolve) => setTimeout(resolve, 100));
    runtimeInstanceId = get(runtimeStore).instanceId;
  }
  return runtimeInstanceId;
}
