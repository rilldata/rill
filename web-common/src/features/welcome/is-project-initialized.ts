import {
  V1ListFilesResponse,
  getRuntimeServiceListFilesQueryKey,
  runtimeServiceListFiles,
} from "@rilldata/web-common/runtime-client";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";

export async function isProjectInitialized(instanceId: string) {
  const files = await queryClient.fetchQuery<V1ListFilesResponse>({
    queryKey: getRuntimeServiceListFilesQueryKey(instanceId, undefined),
    queryFn: ({ signal }) => {
      return runtimeServiceListFiles(instanceId, undefined, signal);
    },
  });

  // Return true if `rill.yaml` exists, else false
  return files.files?.some((file) => file.path === "/rill.yaml");
}
