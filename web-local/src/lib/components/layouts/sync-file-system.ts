/* Poll the filesystem when:
 * 1. The document becomes visible
 * 2. Every X seconds
 *
 * It's slightly complicated because we sync a different file depending on the page we're on.
 */

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
  current_page: Page,
  id: number
) {
  if (!instanceId) return;

  const pagePath = current_page.url.pathname;
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
