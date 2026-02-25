import type { V1Resource } from "@rilldata/web-common/runtime-client";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import {
  getRuntimeServiceListResourcesQueryKey,
  runtimeServiceListResources,
} from "@rilldata/web-common/runtime-client/v2/gen/runtime-service";
import { queryClient } from "../../lib/svelte-query/globalQueryClient";

export async function isLeafResource(
  client: RuntimeClient,
  resource: V1Resource,
) {
  const allResources = await queryClient.fetchQuery({
    queryKey: getRuntimeServiceListResourcesQueryKey(
      client.instanceId,
      undefined,
    ),
    queryFn: () => runtimeServiceListResources(client, {}),
  });

  if (!allResources || !allResources.resources) return false;

  const hasDownstreamResource = allResources.resources.some((r: V1Resource) =>
    r.meta?.refs?.some((ref) => ref.name === resource.meta?.name?.name),
  );

  return !hasDownstreamResource;
}
