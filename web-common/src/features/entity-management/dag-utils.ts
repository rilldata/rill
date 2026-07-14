import {
  getRuntimeServiceListResourcesQueryKey,
  runtimeServiceListResources,
  type V1Resource,
} from "@rilldata/web-common/runtime-client";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import { queryClient } from "../../lib/svelte-query/globalQueryClient";

export async function isLeafResource(
  resource: V1Resource,
  client: RuntimeClient,
) {
  const allResources = await queryClient.fetchQuery({
    queryKey: getRuntimeServiceListResourcesQueryKey(client.instanceId, {}),
    queryFn: () => runtimeServiceListResources(client, {}),
  });

  if (!allResources || !allResources.resources) return false;

  const hasDownstreamResource = allResources.resources.some((r: V1Resource) =>
    r.meta?.refs?.some((ref) => ref.name === resource.meta?.name?.name),
  );

  return !hasDownstreamResource;
}
