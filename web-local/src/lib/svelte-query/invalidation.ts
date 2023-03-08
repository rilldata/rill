import { getNameFromFile } from "@rilldata/web-common/features/entity-management/entity-mappers";
import { fileArtifactsStore } from "@rilldata/web-common/features/entity-management/file-artifacts-store";
import type { V1ReconcileResponse } from "@rilldata/web-common/runtime-client";
import {
  getRuntimeServiceGetCatalogEntryQueryKey,
  getRuntimeServiceGetFileQueryKey,
  getRuntimeServiceListCatalogEntriesQueryKey,
  getRuntimeServiceListFilesQueryKey,
} from "@rilldata/web-common/runtime-client";
import type { QueryClient } from "@sveltestack/svelte-query";
import { get } from "svelte/store";

// invalidation helpers

export function invalidateRuntimeQueries(queryClient: QueryClient) {
  return queryClient.invalidateQueries({
    predicate: (query) =>
      typeof query.queryKey[0] === "string" &&
      query.queryKey[0].startsWith("/v1/instances"),
  });
}

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
    queryClient.refetchQueries(
      getRuntimeServiceListCatalogEntriesQueryKey(instanceId, {
        type: "OBJECT_TYPE_SOURCE",
      })
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
            get(fileArtifactsStore).entities[path]?.name ??
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

export function invalidationForMetricsViewData(query, metricsViewName: string) {
  const r = new RegExp(
    `/v1/instances/[a-zA-Z0-9-]+/queries/metrics-views/${metricsViewName}/`
  );

  return typeof query.queryKey[0] === "string" && r.test(query.queryKey[0]);
}

export function invalidationForProfileQueries(queryHash, name: string) {
  const r = new RegExp(
    `/v1/instances/[a-zA-Z0-9-]+/queries/[a-zA-Z0-9-]+/tables/${name}`
  );
  return r.test(queryHash);
}

export const invalidateMetricsViewData = (
  queryClient: QueryClient,
  metricsViewName: string
) => {
  return queryClient.refetchQueries({
    predicate: (query) =>
      invalidationForMetricsViewData(query, metricsViewName),
  });
};

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

export const removeEntityQueries = async (
  queryClient: QueryClient,
  instanceId: string,
  path: string
) => {
  const name = getNameFromFile(path);
  // remove affected catalog entries and files
  await Promise.all([
    queryClient.removeQueries(
      getRuntimeServiceGetFileQueryKey(instanceId, path)
    ),
    queryClient.removeQueries(
      getRuntimeServiceGetCatalogEntryQueryKey(instanceId, name)
    ),
  ]);

  if (path.startsWith("/dashboards")) {
    return queryClient.removeQueries({
      predicate: (query) => {
        return invalidationForMetricsViewData(query, name);
      },
    });
  } else {
    // remove profiling queries
    return queryClient.removeQueries({
      predicate: (query) => {
        return invalidationForProfileQueries(query.queryHash, name);
      },
    });
  }
};
