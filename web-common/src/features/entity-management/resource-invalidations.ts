import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import {
  getRuntimeServiceGetResourceQueryKey,
  getRuntimeServiceListResourcesQueryKey,
  V1GetResourceResponse,
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
import { get } from "svelte/types/runtime/store";

const MainEntities: {
  [kind in ResourceKind]?: boolean;
} = {
  [ResourceKind.ProjectParser]: false,
  [ResourceKind.Source]: true,
  [ResourceKind.Model]: true,
  [ResourceKind.MetricsView]: true,
};

export function invalidateResourceResponse(
  queryClient: QueryClient,
  res: V1WatchResourcesResponse
) {
  // only process for the `ResourceKind` present in `MainEntities`
  if (!(res.name.kind in MainEntities)) return;

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

  // only re-fetch list queries for kinds in `MainEntities` and is ture
  if (MainEntities[res.name.kind]) {
    queryClient.refetchQueries(
      // we only use individual kind's queries
      getRuntimeServiceListResourcesQueryKey(instanceId, {
        kind: res.name.kind,
      })
    );
  }
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
