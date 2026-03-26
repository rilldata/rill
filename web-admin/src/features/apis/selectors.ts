import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import {
  createRuntimeServiceGetResource,
  createRuntimeServiceListResources,
} from "@rilldata/web-common/runtime-client";
import { smartRefetchIntervalFunc } from "@rilldata/web-admin/lib/refetch-interval-store";

export function useAPIs(client: RuntimeClient, enabled = true) {
  return createRuntimeServiceListResources(
    client,
    {
      kind: ResourceKind.API,
    },
    {
      query: {
        enabled,
        refetchOnMount: true,
        refetchInterval: smartRefetchIntervalFunc,
      },
    },
  );
}

export function useAPI(client: RuntimeClient, name: string) {
  return createRuntimeServiceGetResource(client, {
    "name.name": name,
    "name.kind": ResourceKind.API,
  });
}
