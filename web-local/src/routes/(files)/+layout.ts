import {
  getRuntimeServiceListFilesQueryKey,
  runtimeServiceListFiles,
} from "@rilldata/web-common/runtime-client/index.js";
import type { QueryFunction } from "@tanstack/svelte-query";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.js";
import { buildMapStructure } from "@rilldata/web-common/layout/file-explorer/createMap.js";

export async function load({ depends, parent }) {
  depends("files");

  const { instanceId } = await parent();

  const queryFn: QueryFunction<
    Awaited<ReturnType<typeof runtimeServiceListFiles>>
  > = ({ signal }) => runtimeServiceListFiles(instanceId, {}, signal);

  const filesQuery = queryClient.fetchQuery({
    queryKey: getRuntimeServiceListFilesQueryKey(instanceId, {}),
    queryFn,
    cacheTime: 0,
  });

  const files = await filesQuery;

  const filtered =
    files.paths?.filter((p) => {
      return !p.startsWith("/tmp/");
    }) || [];

  const fileStructure = buildMapStructure(filtered);

  return {
    files: fileStructure,
  };
}
