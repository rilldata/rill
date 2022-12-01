import { useRuntimeServiceListFiles } from "@rilldata/web-common/runtime-client";

export function useModelNames(instanceId: string) {
  return useRuntimeServiceListFiles(
    instanceId,
    {
      glob: "{models}/*.{sql}",
    },
    {
      query: {
        refetchInterval: 1000,
        select: (data) =>
          data.paths?.map((path) =>
            path.replace("/models/", "").replace(".sql", "")
          ),
      },
    }
  );
}
