import {
  useRuntimeServiceGetFile,
  useRuntimeServiceListFiles,
} from "@rilldata/web-common/runtime-client";

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
