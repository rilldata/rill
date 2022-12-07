import type { V1ReconcileResponse } from "@rilldata/web-common/runtime-client";
import {
  getRuntimeServiceGetCatalogEntryQueryKey,
  getRuntimeServiceGetFileQueryKey,
  getRuntimeServiceGetTableCardinalityQueryKey,
  getRuntimeServiceGetTableRowsQueryKey,
  getRuntimeServiceListCatalogEntriesQueryKey,
  getRuntimeServiceListFilesQueryKey,
  getRuntimeServiceProfileColumnsQueryKey,
} from "@rilldata/web-common/runtime-client";
import { getNameFromFile } from "@rilldata/web-local/lib/util/entity-mappers";
import type { QueryClient } from "@sveltestack/svelte-query";

// invalidation helpers

export const invalidateAfterReconcile = async (
  queryClient: QueryClient,
  instanceId: string,
  reconcileResponse: V1ReconcileResponse
) => {
  // invalidate lists of catalog entries and files
  await Promise.all([
    queryClient.refetchQueries(getRuntimeServiceListFilesQueryKey(instanceId)),
    queryClient.refetchQueries(
      getRuntimeServiceListCatalogEntriesQueryKey(instanceId)
    ),
  ]);

  // invalidate affected catalog entries and files
  await Promise.all(
    reconcileResponse.affectedPaths
      .map((path) => [
        queryClient.refetchQueries(
          getRuntimeServiceGetFileQueryKey(instanceId, path)
        ),
        queryClient.refetchQueries(
          getRuntimeServiceGetCatalogEntryQueryKey(
            instanceId,
            getNameFromFile(path)
          )
        ),
      ])
      .flat()
  );
  // invalidate tablewide profiling queries
  // (applies to sources and models, but not dashboards)
  await Promise.all(
    reconcileResponse.affectedPaths
      .map((path) => [
        queryClient.invalidateQueries(
          getRuntimeServiceGetTableCardinalityQueryKey(
            instanceId,
            getNameFromFile(path)
          )
        ),
        queryClient.invalidateQueries(
          getRuntimeServiceGetTableRowsQueryKey(
            instanceId,
            getNameFromFile(path)
          )
        ),
        queryClient.invalidateQueries(
          getRuntimeServiceProfileColumnsQueryKey(
            instanceId,
            getNameFromFile(path)
          )
        ),
        getInvalidationsForPath(queryClient, path),
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
