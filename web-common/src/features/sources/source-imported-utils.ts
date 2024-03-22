import { fileArtifactsStore } from "@rilldata/web-common/features/entity-management/file-artifacts-store";
import { waitForResourceUpdate } from "@rilldata/web-common/features/entity-management/resource-status-utils";
import { sourceImportedName } from "@rilldata/web-common/features/sources/sources-store";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import type { QueryClient } from "@tanstack/svelte-query";
import { get } from "svelte/store";

export function checkSourceImported(
  queryClient: QueryClient,
  sourceName: string,
  filePath: string,
) {
  const lastUpdatedOn = fileArtifactsStore.getLastStateUpdatedOn(filePath);
  if (lastUpdatedOn) return; // For now only show for fresh sources
  waitForResourceUpdate(queryClient, get(runtime).instanceId, filePath).then(
    (success) => {
      if (!success) return;
      sourceImportedName.set(sourceName);
      // TODO: telemetry
    },
  );
}
