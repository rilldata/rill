import {
  StructTypeField,
  useRuntimeServiceGetCatalogEntry,
  useRuntimeServiceGetFile,
  useRuntimeServiceListFiles,
} from "@rilldata/web-common/runtime-client";
import { TIMESTAMPS } from "../../lib/duckdb-data-types";

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

export function useModelFileIsEmpty(instanceId, modelName) {
  return useRuntimeServiceGetFile(instanceId, `/models/${modelName}.sql`, {
    query: {
      select(data) {
        return data?.blob?.length === 0;
      },
    },
  });
}

export function useModelTimestampColumns(
  instanceId: string,
  modelName: string
) {
  return useRuntimeServiceGetCatalogEntry(instanceId, modelName, {
    query: {
      select: (data) =>
        data?.entry?.model?.schema?.fields?.filter((field: StructTypeField) =>
          TIMESTAMPS.has(field.type.code as string)
        ) ?? [].map((field) => field.name),
    },
  });
}
