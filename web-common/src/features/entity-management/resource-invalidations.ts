import {
  getRuntimeServiceGetResourceQueryKey,
  V1Resource,
} from "@rilldata/web-common/runtime-client";
import type { QueryClient } from "@tanstack/svelte-query";

export function refreshResource(
  queryClient: QueryClient,
  instanceId: string,
  res: V1Resource,
) {
  return queryClient.setQueryData(
    getRuntimeServiceGetResourceQueryKey(instanceId, {
      "name.name": res.meta?.name?.name,
      "name.kind": res.meta?.name?.kind,
    }),
    {
      resource: res,
    },
  );
}
