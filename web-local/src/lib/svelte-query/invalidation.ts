import {
  getRuntimeServiceGetCatalogEntryQueryKey,
  getRuntimeServiceGetFileQueryKey,
  getRuntimeServiceListCatalogEntriesQueryKey,
  getRuntimeServiceListFilesQueryKey,
} from "@rilldata/web-common/runtime-client";
import type { V1ReconcileResponse } from "@rilldata/web-common/runtime-client";
import { getNameFromFile } from "@rilldata/web-local/lib/util/entity-mappers";
import type { QueryClient } from "@sveltestack/svelte-query";

// invalidation helpers

export const invalidateAfterReconcile = async (
  queryClient: QueryClient,
  instanceId: string,
  reconcileResponse: V1ReconcileResponse
) => {
  await Promise.all([
    queryClient.invalidateQueries(
      getRuntimeServiceListFilesQueryKey(instanceId)
    ),
    queryClient.invalidateQueries(
      getRuntimeServiceListCatalogEntriesQueryKey(instanceId)
    ),
  ]);
  await Promise.all(
    reconcileResponse.affectedPaths
      .map((affectedPath) => [
        queryClient.invalidateQueries(
          getRuntimeServiceGetFileQueryKey(instanceId, affectedPath)
        ),
        queryClient.invalidateQueries(
          getRuntimeServiceGetCatalogEntryQueryKey(
            instanceId,
            getNameFromFile(affectedPath)
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
  return queryClient.invalidateQueries({
    predicate: (query) =>
      typeof query.queryKey[0] === "string" &&
      query.queryKey[0].startsWith(
        `/v1/instances/${instanceId}/metrics-views/${metricsViewName}/`
      ),
  });
};
