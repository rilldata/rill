import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import {
  createRuntimeServiceListResources,
  type V1Resource,
} from "@rilldata/web-common/runtime-client";

// Use the ListResources API to get the target dashboard
// The JWT generated via a "magic" token will only have access to one dashboard, so we can assume the first one is the correct one
export function useShareableURLMetricsView(
  instanceId: string,
  enabled: boolean,
) {
  return createRuntimeServiceListResources<V1Resource | undefined>(
    instanceId,
    {
      kind: ResourceKind.MetricsView,
    },
    {
      query: {
        select: (data) => data?.resources?.[0],
        enabled: !!instanceId && enabled,
      },
    },
  );
}
