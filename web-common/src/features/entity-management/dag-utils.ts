import {
  runtimeServiceListResources,
  type V1Resource,
} from "@rilldata/web-common/runtime-client";
import { getRuntimeServiceListResourcesQueryKey } from "@rilldata/web-common/runtime-client/v2/gen/runtime-service";
import { queryClient } from "../../lib/svelte-query/globalQueryClient";

export async function isLeafResource(resource: V1Resource, instanceId: string) {
  const allResources = await queryClient.fetchQuery({
    queryKey: getRuntimeServiceListResourcesQueryKey(instanceId, undefined),
    queryFn: () => runtimeServiceListResources(instanceId, undefined),
  });

  if (!allResources || !allResources.resources) return false;

  const hasDownstreamResource = allResources.resources.some((r: V1Resource) =>
    r.meta?.refs?.some((ref) => ref.name === resource.meta?.name?.name),
  );

  return !hasDownstreamResource;
}
