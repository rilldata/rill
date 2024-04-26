import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts";
import { waitForResourceUpdate } from "@rilldata/web-common/features/entity-management/resource-status-utils";
import { sourceImportedPath } from "@rilldata/web-common/features/sources/sources-store";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import type { QueryClient } from "@tanstack/svelte-query";
import { get } from "svelte/store";

export async function checkSourceImported(
  queryClient: QueryClient,
  filePath: string,
) {
  const lastUpdatedOn =
    fileArtifacts.getFileArtifact(filePath).lastStateUpdatedOn;
  if (lastUpdatedOn) return; // For now only show for fresh sources
  const success = await waitForResourceUpdate(
    queryClient,
    get(runtime).instanceId,
    filePath,
  );

  if (!success) return;
  sourceImportedPath.set(filePath);
}
