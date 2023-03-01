import type { V1ReconcileResponse } from "@rilldata/web-common/runtime-client";
import {
  runtimeServiceListCatalogEntries,
  runtimeServiceListFiles,
  runtimeServiceReconcile,
} from "@rilldata/web-common/runtime-client";
import { getMapFromArray } from "@rilldata/web-local/lib/util/arrayUtils";

export async function getArtifactErrors(): Promise<V1ReconcileResponse> {
  const files = await runtimeServiceListFiles({
    glob: "{sources,models,dashboards}/*.{yaml,sql}",
  });
  const catalogs = await runtimeServiceListCatalogEntries();
  const catalogsMap = getMapFromArray(
    catalogs.entries,
    (catalog) => catalog.path
  );
  const missingFiles = files.paths.filter(
    (filePath) => !catalogsMap.has(filePath)
  );
  return runtimeServiceReconcile({
    changedPaths: missingFiles,
    forcedPaths: missingFiles,
    dry: true,
  });
}
