import {
  createRuntimeServiceListFiles,
  runtimeServiceListFiles,
} from "@rilldata/web-common/runtime-client";

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

export async function isProjectInitialized(
  instanceId: string
): Promise<boolean> {
  const data = await runtimeServiceListFiles(instanceId, {
    glob: "rill.yaml",
  });

  // Return true if `rill.yaml` exists, else false
  return data.paths.length === 1;
}
