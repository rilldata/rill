import { page } from "$app/stores";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import { createRuntimeServiceListResources } from "@rilldata/web-common/runtime-client";
import { get } from "svelte/store";
import { isMagicLinkPage } from "../navigation/nav-utils";

// Use the ListResources API to get the target dashboard
// The JWT generated via a "magic" token will only have access to one dashboard, so we can assume the first one is the correct one
export function useShareableURLMetricsView(instanceId: string) {
  return createRuntimeServiceListResources(
    instanceId,
    {
      kind: ResourceKind.MetricsView,
    },
    {
      query: {
        select: (data) => ({
          resource: data.resources[0],
        }),
        enabled: isMagicLinkPage(get(page)),
      },
    },
  );
}
