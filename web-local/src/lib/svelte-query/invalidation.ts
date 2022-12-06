import {
  getRuntimeServiceGetCatalogEntryQueryKey,
  getRuntimeServiceGetFileQueryKey,
  getRuntimeServiceGetTableCardinalityQueryKey,
  getRuntimeServiceGetTableRowsQueryKey,
  getRuntimeServiceListCatalogEntriesQueryKey,
  getRuntimeServiceListFilesQueryKey,
  getRuntimeServiceProfileColumnsQueryKey,
  V1ReconcileResponse,
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
      ])
      .flat()
  );
};

export const invalidateMetricsViewData = (
  queryClient: QueryClient,
  instanceId: string,
  metricsViewName: string
) => {
  return queryClient.refetchQueries({
    predicate: (query) =>
      typeof query.queryKey[0] === "string" &&
      query.queryKey[0].startsWith(
        `/v1/instances/${instanceId}/metrics-views/${metricsViewName}/`
      ),
  });
};
