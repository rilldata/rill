import { getNameFromFile } from "@rilldata/web-common/features/entity-management/entity-mappers";
import { useRuntimeServiceListFiles } from "@rilldata/web-common/runtime-client";

export function useAllNames(instanceId: string) {
  return useRuntimeServiceListFiles(
    instanceId,
    {
      glob: "{sources,models,dashboards}/*.{yaml,sql}",
    },
    {
      query: {
        select: (data) =>
          data.paths?.map((path) => getNameFromFile(path)) ?? [],
      },
    }
  );
}
