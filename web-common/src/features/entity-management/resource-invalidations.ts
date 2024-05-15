import { ThrottlerMap } from "@rilldata/web-common/lib/throttler";
import {
  getRuntimeServiceGetResourceQueryKey,
  V1Resource,
} from "@rilldata/web-common/runtime-client";
import type { QueryClient } from "@tanstack/svelte-query";

const throttler = new ThrottlerMap(100);

export function throttledRefreshResource(
  queryClient: QueryClient,
  instanceId: string,
  res: V1Resource,
) {
  const key = `${res.meta?.name?.kind}/${res.meta?.name?.name}`;
  throttler.throttle(key, () => {
    queryClient.setQueryData(
      getRuntimeServiceGetResourceQueryKey(instanceId, {
        "name.name": res.meta?.name?.name,
        "name.kind": res.meta?.name?.kind,
      }),
      {
        resource: res,
      },
      {
        updatedAt: Date.now(),
      },
    );
  });
}

export function refreshResource(
  queryClient: QueryClient,
  instanceId: string,
  res: V1Resource,
) {
  return queryClient.resetQueries(
    getRuntimeServiceGetResourceQueryKey(instanceId, {
      "name.name": res.meta?.name?.name,
      "name.kind": res.meta?.name?.kind,
    }),
  );
}
