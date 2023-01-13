import {
  useRuntimeServiceListCatalogEntries,
  useRuntimeServiceListFiles,
  V1StructType,
} from "@rilldata/web-common/runtime-client";
import { schemaHasTimestampColumn } from "@rilldata/web-local/lib/svelte-query/column-selectors";

export function useModelNames(instanceId: string) {
  return useRuntimeServiceListFiles(
    instanceId,
    {
      glob: "{sources,models,dashboards}/*.{yaml,sql}",
    },
    {
      query: {
        // refetchInterval: 1000,
        select: (data) =>
          data.paths
            ?.filter((path) => path.includes("models/"))
            .map((path) => path.replace("/models/", "").replace(".sql", ""))
            // sort alphabetically case-insensitive
            .sort((a, b) =>
              a.localeCompare(b, undefined, { sensitivity: "base" })
            ),
      },
    }
  );
}

function isDashboardable(schema: V1StructType) {
  return schemaHasTimestampColumn(schema);
}

export function useDashboardableModels(instanceId) {
  return useRuntimeServiceListCatalogEntries(
    instanceId,
    { type: "OBJECT_TYPE_MODEL" },
    {
      query: {
        select: (data) =>
          data?.entries?.filter((entry) =>
            isDashboardable(entry?.model?.schema)
          ),
      },
    }
  );
}
