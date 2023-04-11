import { getNameFromFile } from "@rilldata/web-common/features/entity-management/entity-mappers";
import { createRuntimeServiceListFiles } from "../../runtime-client";

export function useAllNames(instanceId: string) {
  return createRuntimeServiceListFiles(
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
