import { EntityType } from "@rilldata/web-common/lib/entity";
import type { V1ReconcileResponse } from "@rilldata/web-common/runtime-client";
import {
  getRuntimeServiceGetCatalogEntryQueryKey,
  getRuntimeServiceGetFileQueryKey,
  getRuntimeServiceListCatalogEntriesQueryKey,
  getRuntimeServiceListFilesQueryKey,
} from "@rilldata/web-common/runtime-client";
import {
  getFilePathFromNameAndType,
  getNameFromFile,
} from "@rilldata/web-local/lib/util/entity-mappers";
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
    reconcileResponse.affectedPaths.map((path) =>
      getInvalidationsForPath(queryClient, path)
    )
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
  const r = new RegExp(
    `/v1/instances/[a-zA-Z0-9-]+/metrics-views/${metricsViewName}/`
  );
  return queryClient.refetchQueries({
    predicate: (query) =>
      typeof query.queryKey[0] === "string" && r.test(query.queryKey[0]),
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

export const removeModelQueries = async (
  queryClient: QueryClient,
  instanceId: string,
  name: string
) => {
  const path = getFilePathFromNameAndType(name, EntityType.Model);

  // remove affected catalog entries and files
  await Promise.all([
    queryClient.removeQueries(
      getRuntimeServiceGetFileQueryKey(instanceId, path)
    ),
    queryClient.removeQueries(
      getRuntimeServiceGetCatalogEntryQueryKey(instanceId, name)
    ),
  ]);

  // remove profiling queries
  return queryClient.removeQueries({
    predicate: (query) => {
      return invalidationForProfileQueries(query.queryHash, name);
    },
  });
};
