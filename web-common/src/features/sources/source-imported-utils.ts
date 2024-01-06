import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import { waitForResourceUpdate } from "@rilldata/web-common/features/entity-management/resource-status-utils";
import { getLastStateUpdatedOnByKindAndName } from "@rilldata/web-common/features/entity-management/resources-store";
import { sourceImportedName } from "@rilldata/web-common/features/sources/sources-store";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import type { QueryClient } from "@tanstack/svelte-query";
import { get } from "svelte/store";

export function checkSourceImported(
  queryClient: QueryClient,
  sourceName: string,
  filePath: string,
) {
  const lastUpdatedOn = getLastStateUpdatedOnByKindAndName(
    ResourceKind.Source,
    sourceName,
  );
  if (lastUpdatedOn) return; // For now only show for fresh sources
  waitForResourceUpdate(
    queryClient,
    get(runtime).instanceId,
    filePath,
    ResourceKind.Source,
    sourceName,
  ).then((success) => {
    if (!success) return;
    sourceImportedName.set(sourceName);
    // TODO: telemetry
  });
}
