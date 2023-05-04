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
import type { QueryClient } from "@tanstack/svelte-query";
import { Readable, Writable, get } from "svelte/store";
import { overlay } from "../../layout/overlay-store";
import { invalidateAfterReconcile } from "../../runtime-client/invalidation";
import type { Runtime } from "../../runtime-client/runtime-store";
import type { FeatureFlags } from "../feature-flags";
import { getFilePathFromPagePath } from "./entity-mappers";
import {
  FileArtifactsStore,
  getIsFileReconcilingStore,
} from "./file-artifacts-store";

const SYNC_FILE_SYSTEM_INTERVAL_MILLISECONDS = 60000;
const RECONCILE_OVERLAY_DELAY_MILLISECONDS = 1000;

export async function syncFileSystemPeriodically(
  queryClient: QueryClient,
  runtimeStore: Writable<Runtime>,
  featureFlags: Writable<FeatureFlags>,
  page: Readable<Page<Record<string, string>, string>>,
  fileArtifactsStore: FileArtifactsStore
) {
  let syncFileSystemInterval: NodeJS.Timer;
  // let syncFileSystemOnVisibleDocument: () => void;
  let afterNavigateRanOnce: boolean;

  afterNavigate(async () => {
    const runtimeInstanceId = get(runtimeStore).instanceId;
    if (get(featureFlags).readOnly) return;

    // on first page load, afterNavigate runs twice, but we only want to run the below code once
    if (afterNavigateRanOnce) return;

    // Scenario 1: sync when the user navigates to a new page
    // syncFileSystem(queryClient, runtimeInstanceId, page, fileArtifactsStore);

    // setup Scenario 2: sync every X seconds
    syncFileSystemInterval = setInterval(
      async () =>
        await syncFileSystem(
          queryClient,
          runtimeInstanceId,
          page,
          fileArtifactsStore
        ),
      SYNC_FILE_SYSTEM_INTERVAL_MILLISECONDS
    );

    // setup Scenario 3: sync when the user returns focus to the browser tab
    // syncFileSystemOnVisibleDocument = async () => {
    //   if (document.visibilityState === "visible") {
    //     await syncFileSystem(
    //       queryClient,
    //       runtimeInstanceId,
    //       page,
    //       fileArtifactsStore
    //     );
    //   }
    // };
    // window.addEventListener("focus", syncFileSystemOnVisibleDocument);

    afterNavigateRanOnce = true;
  });

  beforeNavigate(() => {
    // teardown Scenario 2
    clearInterval(syncFileSystemInterval);

    // teardown Scenario 3
    // window.removeEventListener("focus", syncFileSystemOnVisibleDocument);

    afterNavigateRanOnce = false;
  });
}

async function syncFileSystem(
  queryClient: QueryClient,
  instanceId: string,
  page: Readable<Page<Record<string, string>, string>>,
  fileArtifactsStore: FileArtifactsStore
) {
  await queryClient.invalidateQueries(
    getRuntimeServiceListFilesQueryKey(instanceId)
  );

  const pagePath = get(page).url.pathname;
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

export function addReconcilingOverlay(pagePath: string) {
  if (pagePath === "/") return;

  const filePath = getFilePathFromPagePath(pagePath);
  const isFileReconcilingStore = getIsFileReconcilingStore(filePath);

  // we debounce the overlay so that it doesn't flash on the screen for a split second
  let delayedOverlayTimeout: NodeJS.Timeout;

  isFileReconcilingStore.subscribe((isFileReconciling) => {
    if (isFileReconciling) {
      delayedOverlayTimeout = setTimeout(() => {
        overlay.set({
          title: `Updating project — this could take a moment`,
        });
      }, RECONCILE_OVERLAY_DELAY_MILLISECONDS);
    } else {
      clearTimeout(delayedOverlayTimeout);
      overlay.set(null);
    }
  });
}
