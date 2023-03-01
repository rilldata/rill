import {
  StructTypeField,
  useRuntimeServiceGetCatalogEntry,
  useRuntimeServiceGetFile,
  useRuntimeServiceListFiles,
} from "@rilldata/web-common/runtime-client";
import { TIMESTAMPS } from "../../lib/duckdb-data-types";

export function useModelNames() {
  return useRuntimeServiceListFiles(
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

export function useModelFileIsEmpty(modelName) {
  return useRuntimeServiceGetFile(`/models/${modelName}.sql`, {
    query: {
      select(data) {
        return data?.blob?.length === 0;
      },
    },
  });
}

export function useModelTimestampColumns(modelName: string) {
  return useRuntimeServiceGetCatalogEntry(modelName, {
    query: {
      select: (data) =>
        data?.entry?.model?.schema?.fields?.filter((field: StructTypeField) =>
          TIMESTAMPS.has(field.type.code as string)
        ) ?? [].map((field) => field.name),
    },
  });
}
