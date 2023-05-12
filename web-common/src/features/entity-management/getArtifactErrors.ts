import { getMapFromArray } from "@rilldata/web-common/lib/arrayUtils";
import type { V1ReconcileResponse } from "@rilldata/web-common/runtime-client";
import {
  runtimeServiceListCatalogEntries,
  runtimeServiceListFiles,
  runtimeServiceReconcile,
} from "@rilldata/web-common/runtime-client";

export async function getArtifactErrors(
  instanceId: string
): Promise<V1ReconcileResponse> {
  const files = await runtimeServiceListFiles(instanceId, {
    glob: "{sources,models,dashboards}/*.{yaml,sql}",
  });
  const catalogs = await runtimeServiceListCatalogEntries(instanceId);
  const catalogsMap = getMapFromArray(
    catalogs.entries,
    (catalog) => catalog.path
  );
  const missingFiles = files.paths.filter(
    (filePath) => !catalogsMap.has(filePath)
  );
  return runtimeServiceReconcile(instanceId, {
    changedPaths: missingFiles,
    forcedPaths: missingFiles,
    dry: true,
  });
}
