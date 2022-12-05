import type { V1ReconcileResponse } from "@rilldata/web-common/runtime-client";
import {
  getRuntimeServiceGetCatalogEntryQueryKey,
  getRuntimeServiceGetFileQueryKey,
  getRuntimeServiceListCatalogEntriesQueryKey,
  getRuntimeServiceListFilesQueryKey,
} from "@rilldata/web-common/runtime-client";
import { getNameFromFile } from "@rilldata/web-local/lib/util/entity-mappers";
import type { QueryClient } from "@sveltestack/svelte-query";

// invalidation helpers

export const invalidateAfterReconcile = async (
  queryClient: QueryClient,
  instanceId: string,
  reconcileResponse: V1ReconcileResponse
) => {
  await Promise.all([
    queryClient.refetchQueries(getRuntimeServiceListFilesQueryKey(instanceId)),
    queryClient.refetchQueries(
      getRuntimeServiceListCatalogEntriesQueryKey(instanceId)
    ),
  ]);
  await Promise.all(
    reconcileResponse.affectedPaths
      .map((affectedPath) => [
        queryClient.refetchQueries(
          getRuntimeServiceGetFileQueryKey(instanceId, affectedPath)
        ),
        queryClient.refetchQueries(
          getRuntimeServiceGetCatalogEntryQueryKey(
            instanceId,
            getNameFromFile(affectedPath)
          )
        ),
        getInvalidationsForPath(queryClient, affectedPath),
      ])
      .flat()
  );
};

const getInvalidationsForPath = (
  queryClient: QueryClient,
  filePath: string
) => {
  const name = getNameFromFile(filePath);
  if (filePath.startsWith("/dashboards")) {
    return invalidateMetricsViewData(queryClient, name);
  } else {
    return invalidateProfilingQueries(queryClient, name);
  }
};

export const invalidateMetricsViewData = (
  queryClient: QueryClient,
  metricsViewName: string
) => {
  return queryClient.refetchQueries({
    predicate: (query) =>
      typeof query.queryKey[0] === "string" &&
      query.queryKey[0].startsWith(
        `/v1/instances/[a-zA-Z0-9-]+/metrics-views/${metricsViewName}/`
      ),
  });
};

export function invalidationForProfileQueries(queryHash, name: string) {
  const r = new RegExp(
    `/v1/instances/[a-zA-Z0-9-]+/queries/[a-zA-Z0-9-]+/tables/${name}`
  );
  return r.test(queryHash);
}

export function invalidateProfilingQueries(
  queryClient: QueryClient,
  name: string
) {
  return queryClient.refetchQueries({
    predicate: (query) => {
      return invalidationForProfileQueries(query.queryHash, name);
    },
  });
}
