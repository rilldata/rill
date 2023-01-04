/* Poll the filesystem under 3 scenarios:
 * - Scenario 1. The user navigates to a new page
 * - Scenario 2. Every X seconds
 * - Scenario 3. The user returns focus to the browser tab
 *
 * It's slightly complicated because we sync a different file depending on the page we're on.
 */

import { afterNavigate, beforeNavigate } from "$app/navigation";
import {
  getRuntimeServiceGetFileQueryKey,
  getRuntimeServiceListFilesQueryKey,
} from "@rilldata/web-common/runtime-client";
import type { Page } from "@sveltejs/kit";
import type { QueryClient } from "@sveltestack/svelte-query";
import { get, Readable, Writable } from "svelte/store";
import type { RuntimeState } from "../../application-state-stores/application-store";
import { getFilePathFromPagePath } from "../../util/entity-mappers";

const SYNC_FILE_SYSTEM_INTERVAL_MILLISECONDS = 5000;

export async function syncFileSystem(
  queryClient: QueryClient,
  instanceId: string,
  page: Readable<Page<Record<string, string>, string>>,
  id: number
) {
  if (!instanceId) return;

  const pagePath = get(page).url.pathname;
  console.log("syncFileSystem", instanceId, pagePath, id);

  // invalidate `GetFile` only if on a /source, /model, or /dashboard page
  if (
    pagePath.startsWith("/model") ||
    pagePath.startsWith("/source") ||
    pagePath.startsWith("/dashboard")
  ) {
    await queryClient.invalidateQueries(
      getRuntimeServiceGetFileQueryKey(
        instanceId,
        getFilePathFromPagePath(pagePath)
      )
    );
  }

  // TODO: should we also invalidate ListCatalogObjects?
  await queryClient.invalidateQueries(
    getRuntimeServiceListFilesQueryKey(instanceId)
  );

  // TODO: call reconcile
}

export function syncFileSystemPeriodically(
  queryClient: QueryClient,
  runtimeStore: Writable<RuntimeState>,
  page: Readable<Page<Record<string, string>, string>>
) {
  let syncFileSystemInterval: any; // NodeJS.Timer
  let syncFileSystemOnVisibleDocument: () => void;
  let afterNavigateRanOnce: boolean;

  afterNavigate(async () => {
    // afterNavigate races against onMount, which sets the runtimeInstanceId
    // loop until we have a runtimeInstanceID
    let runtimeInstanceId: string;
    while (!runtimeInstanceId) {
      await new Promise((resolve) => setTimeout(resolve, 100));
      runtimeInstanceId = get(runtimeStore).instanceId;
    }

    // on first page load, afterNavigate runs twice
    // this guard clause ensures we only run the below code once
    if (afterNavigateRanOnce) return;

    // Scenario 1: sync when the user navigates to a new page
    syncFileSystem(queryClient, runtimeInstanceId, page, 1);

    // Setup scenario 2: sync every X seconds
    syncFileSystemInterval = setInterval(
      async () => await syncFileSystem(queryClient, runtimeInstanceId, page, 2),
      SYNC_FILE_SYSTEM_INTERVAL_MILLISECONDS
    );

    // Setup scenario 3: sync when the user returns focus to the browser tab
    syncFileSystemOnVisibleDocument = async () => {
      if (document.visibilityState === "visible") {
        await syncFileSystem(queryClient, runtimeInstanceId, page, 3);
      }
    };
    document.addEventListener(
      "visibilitychange",
      syncFileSystemOnVisibleDocument
    );

    afterNavigateRanOnce = true;
  });

  beforeNavigate(() => {
    // Teardown scenario 2
    clearInterval(syncFileSystemInterval);

    // Teardown scenario 3
    document.removeEventListener(
      "visibilitychange",
      syncFileSystemOnVisibleDocument
    );

    afterNavigateRanOnce = false;
  });
}
