import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import {
  createRuntimeServiceGetResource,
  createRuntimeServiceListResources,
} from "@rilldata/web-common/runtime-client";
import { smartRefetchIntervalFunc } from "@rilldata/web-admin/lib/refetch-interval-store";

export function useAPIs(instanceId: string, enabled = true) {
  return createRuntimeServiceListResources(
    instanceId,
    {
      kind: ResourceKind.API,
    },
    {
      query: {
        enabled: enabled && !!instanceId,
        refetchOnMount: true,
        refetchInterval: smartRefetchIntervalFunc,
      },
    },
  );
}

export function useAPI(instanceId: string, name: string) {
  return createRuntimeServiceGetResource(instanceId, {
    "name.name": name,
    "name.kind": ResourceKind.API,
  });
}
