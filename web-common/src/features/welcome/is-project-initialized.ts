import {
  V1ListFilesResponse,
  getRuntimeServiceListFilesQueryKey,
  runtimeServiceListFiles,
} from "@rilldata/web-common/runtime-client";
import type { QueryClient } from "@tanstack/query-core";

export async function isProjectInitialized(
  queryClient: QueryClient,
  instanceId: string,
) {
  const files = await queryClient.fetchQuery<V1ListFilesResponse>({
    queryKey: getRuntimeServiceListFilesQueryKey(instanceId, undefined),
    queryFn: ({ signal }) => {
      return runtimeServiceListFiles(instanceId, undefined, signal);
    },
  });

  // Return true if `rill.yaml` exists, else false
  return files.files?.some((file) => file.path === "/rill.yaml");
}
