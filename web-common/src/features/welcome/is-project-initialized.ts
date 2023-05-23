import { createRuntimeServiceListFiles } from "@rilldata/web-common/runtime-client";

export function useIsProjectInitialized(instanceId: string) {
  return createRuntimeServiceListFiles(
    instanceId,
    {
      glob: "rill.yaml",
    },
    {
      query: {
        select: (data) => {
          // Return true if `rill.yaml` exists, else false
          return data.paths.length === 1;
        },
        refetchOnWindowFocus: true,
      },
    }
  );
}
