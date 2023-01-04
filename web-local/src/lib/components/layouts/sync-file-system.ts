/* Poll the filesystem when:
 * 1. The document becomes visible
 * 2. Every X seconds
 *
 * It's slightly complicated because we sync a different file depending on the page we're on.
 */

import { beforeNavigate } from "$app/navigation";
import {
  getRuntimeServiceGetFileQueryKey,
  getRuntimeServiceListFilesQueryKey,
} from "@rilldata/web-common/runtime-client";
import type { Page } from "@sveltejs/kit";
import type { QueryClient } from "@sveltestack/svelte-query";
import { getFilePathFromPagePath } from "../../util/entity-mappers";

export async function syncFileSystem(
  queryClient: QueryClient,
  instanceId: string,
  current_page: Page
) {
  const pagePath = current_page.url.pathname;
  console.log("syncFileSystem", instanceId, pagePath);
  if (!instanceId) return;

  await Promise.all([
    queryClient.invalidateQueries(
      getRuntimeServiceGetFileQueryKey(
        instanceId,
        getFilePathFromPagePath(pagePath)
      )
    ),
    // TODO: should we also invalidate ListCatalogObjects?
    queryClient.invalidateQueries(
      getRuntimeServiceListFilesQueryKey(instanceId)
    ),
  ]);

  // TODO: call reconcile
}

const POLL_FILE_SYSTEM_MILLISECONDS = 3000;

export async function syncFileSystemOnInterval(
  queryClient: QueryClient,
  instanceId: string,
  page: Page
) {
  const interval = setInterval(
    async () => await syncFileSystem(queryClient, instanceId, page),
    POLL_FILE_SYSTEM_MILLISECONDS
  );

  beforeNavigate(() => {
    clearInterval(interval);
  });
}

export async function syncFileSystemOnVisibleDocument(
  queryClient: QueryClient,
  instanceId: string,
  page: Page
) {
  const _syncFileSystemOnVisibleDocument = async () => {
    if (document.visibilityState === "visible") {
      await syncFileSystem(queryClient, instanceId, page);
    }
  };

  document.addEventListener(
    "visibilitychange",
    _syncFileSystemOnVisibleDocument
  );

  beforeNavigate(() => {
    // TODO: this doesn't seem to remove the listener.
    document.removeEventListener(
      "visibilitychange",
      _syncFileSystemOnVisibleDocument
    );
  });
}
