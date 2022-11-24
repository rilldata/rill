import {
  useRuntimeServiceGetFile,
  useRuntimeServiceListFiles,
} from "@rilldata/web-common/runtime-client";
import { parse } from "yaml";

/**
 * Calls {@link useRuntimeServiceListFiles} using glob to select only sources.
 * Returns just the source names from the files.
 */
export function useSourceNames(instanceId: string) {
  return useRuntimeServiceListFiles(
    instanceId,
    {
      glob: "sources/*.{sql,yaml}",
    },
    {
      query: {
        refetchInterval: 1000,
        select: (data) =>
          data.paths
            ?.filter((path) => path.includes("sources/"))
            .map((path) => path.replace("/sources/", "").replace(".yaml", "")),
      },
    }
  );
}

export function useSourceFromYaml(instanceId: string, filePath: string) {
  return useRuntimeServiceGetFile(instanceId, filePath, {
    query: {
      select: (data) => (data.blob ? parse(data.blob) : {}),
    },
  });
}
