import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import { resourcesStore } from "@rilldata/web-common/features/entity-management/resources-store";
import { ReconcileStatus } from "@rilldata/web-common/proto/gen/rill/runtime/v1/resources_pb";
import {
  getRuntimeServiceGetResourceQueryKey,
  getRuntimeServiceListResourcesQueryKey,
  V1GetResourceResponse,
  V1ReconcileStatus,
  V1Resource,
} from "@rilldata/web-common/runtime-client";
import type { V1WatchResourcesResponse } from "@rilldata/web-common/runtime-client";
import {
  invalidateMetricsViewData,
  invalidateProfilingQueries,
  invalidationForMetricsViewData,
} from "@rilldata/web-common/runtime-client/invalidation";
import { isProfilingQuery } from "@rilldata/web-common/runtime-client/query-matcher";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import type { QueryClient } from "@tanstack/svelte-query";
import { get } from "svelte/store";

const MainResources: {
  [kind in ResourceKind]?: true;
} = {
  [ResourceKind.Source]: true,
  [ResourceKind.Model]: true,
  [ResourceKind.MetricsView]: true,
};
const UsedResources: {
  [kind in ResourceKind]?: true;
} = {
  [ResourceKind.ProjectParser]: true,
  ...MainResources,
};

export function invalidateResourceResponse(
  queryClient: QueryClient,
  res: V1WatchResourcesResponse
) {
  // only process for the `ResourceKind` present in `UsedResources`
  if (
    !UsedResources[res.name.kind] ||
    res.resource.meta.reconcileStatus !==
      V1ReconcileStatus.RECONCILE_STATUS_IDLE
  )
    return;

  console.log(
    `[${res.resource.meta.reconcileStatus}] ${res.name.kind}/${res.name.name}`
  );
  const instanceId = get(runtime).instanceId;
  // invalidations will wait until the re-fetched query is completed
  // so, we should not `await` here
  switch (res.event) {
    case "RESOURCE_EVENT_WRITE":
      invalidateResource(queryClient, instanceId, res.resource);
      break;

    case "RESOURCE_EVENT_DELETE":
      invalidateRemovedResource(queryClient, instanceId, res.resource);
      break;
  }

  // only re-fetch list queries for kinds in `MainResources` and is ture
  if (!MainResources[res.name.kind]) return;
  resourcesStore.setResource(res.resource);
  return queryClient.refetchQueries(
    // we only use individual kind's queries
    getRuntimeServiceListResourcesQueryKey(instanceId, {
      kind: res.name.kind,
    })
  );
}

async function invalidateResource(
  queryClient: QueryClient,
  instanceId: string,
  resource: V1Resource
) {
  const failed = !!resource.meta.reconcileError;

  // set the data directly since we have the full resource already
  // TODO: test this thoroughly to make sure this doesnt break anything
  queryClient.setQueryData(
    getRuntimeServiceGetResourceQueryKey(instanceId, {
      "name.name": resource.meta.name.name,
      "name.kind": resource.meta.name.kind,
    }),
    { resource } as V1GetResourceResponse
  );
  switch (resource.meta.name.kind) {
    case ResourceKind.Source:
    case ResourceKind.Model:
      return invalidateProfilingQueries(
        queryClient,
        resource.meta.name.name,
        failed
      );

    case ResourceKind.MetricsView:
      return invalidateMetricsViewData(
        queryClient,
        resource.meta.name.name,
        failed
      );
  }
}

async function invalidateRemovedResource(
  queryClient: QueryClient,
  instanceId: string,
  resource: V1Resource
) {
  queryClient.removeQueries(
    getRuntimeServiceGetResourceQueryKey(instanceId, {
      "name.name": resource.meta.name.name,
      "name.kind": resource.meta.name.kind,
    })
  );
  switch (resource.meta.name.kind) {
    case ResourceKind.Source:
    case ResourceKind.Model:
      queryClient.removeQueries({
        predicate: (query) => isProfilingQuery(query, resource.meta.name.name),
      });
      break;

    case ResourceKind.MetricsView:
      queryClient.removeQueries({
        predicate: (query) =>
          invalidationForMetricsViewData(query, resource.meta.name.name),
      });
      break;
  }
}

export async function invalidateAllResources(queryClient: QueryClient) {
  return queryClient.resetQueries({
    type: "inactive",
    predicate: (query) =>
      query.queryHash.includes(`v1/instances/${get(runtime).instanceId}`),
  });
}
