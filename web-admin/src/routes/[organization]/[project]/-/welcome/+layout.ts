import { fileArtifacts } from "@rilldata/web-common/features/entity-management/file-artifacts.ts";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
import {
  getRuntimeServiceListFilesQueryKey,
  runtimeServiceListFiles,
  type V1ListFilesResponse,
} from "@rilldata/web-common/runtime-client";
import { getCloudRuntimeClient } from "@rilldata/web-admin/lib/runtime-client.ts";

export async function load({ parent }) {
  const { runtime } = await parent();
  const client = getCloudRuntimeClient(runtime);

  // Set the client on fileArtifacts early so child page load functions
  // (e.g., files/[...file]/+page.ts) can access it before components render.
  fileArtifacts.setClient(client);

  const files = await queryClient.fetchQuery<V1ListFilesResponse>({
    queryKey: getRuntimeServiceListFilesQueryKey(client.instanceId, {}),
    queryFn: ({ signal }) => {
      return runtimeServiceListFiles(client, {}, { signal });
    },
  });

  return {
    files: files.files ?? [],
  };
}
