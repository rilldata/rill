import {
  V1ListFilesResponse,
  createRuntimeServiceListFiles,
  getRuntimeServiceListFilesQueryKey,
  runtimeServiceListFiles,
} from "@rilldata/web-common/runtime-client";
import type { QueryClient } from "@tanstack/query-core";

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
    },
  );
}

export async function isProjectInitialized(
  instanceId: string,
): Promise<boolean> {
  const data = await runtimeServiceListFiles(instanceId, {
    glob: "rill.yaml",
  });

  // Return true if `rill.yaml` exists, else false
  return data.paths.length === 1;
}

// V2 is an improvement because it uses the queryClient to cache the result
export async function isProjectInitializedV2(
  queryClient: QueryClient,
  instanceId: string,
) {
  const rillYAMLFiles = await queryClient.fetchQuery<V1ListFilesResponse>({
    queryKey: getRuntimeServiceListFilesQueryKey(instanceId, {
      glob: "rill.yaml",
    }),
    queryFn: () => {
      return runtimeServiceListFiles(instanceId, {
        glob: "rill.yaml",
      });
    },
  });

  // Return true if `rill.yaml` exists, else false
  return rillYAMLFiles.paths.length === 1;
}
