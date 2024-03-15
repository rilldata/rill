import {
  getNameFromFile,
  removeLeadingSlash,
} from "@rilldata/web-common/features/entity-management/entity-mappers";
import { fileArtifactsStore } from "@rilldata/web-common/features/entity-management/file-artifacts-store";
import { getMapFromArray } from "@rilldata/web-common/lib/arrayUtils";
import type { V1ReconcileResponse } from "@rilldata/web-common/runtime-client";
import {
  getRuntimeServiceGetCatalogEntryQueryKey,
  getRuntimeServiceGetFileQueryKey,
  getRuntimeServiceListCatalogEntriesQueryKey,
  getRuntimeServiceListFilesQueryKey,
} from "@rilldata/web-common/runtime-client";
import {
  isColumnProfilingQuery,
  isProfilingQuery,
  isTableProfilingQuery,
} from "@rilldata/web-common/runtime-client/query-matcher";
import type { Query, QueryClient } from "@tanstack/svelte-query";
import { get } from "svelte/store";

// invalidation helpers

export function invalidateRuntimeQueries(queryClient: QueryClient) {
  return queryClient.resetQueries({
    predicate: (query) =>
      typeof query.queryKey[0] === "string" &&
      query.queryKey[0].startsWith("/v1/instances"),
  });
}

export const invalidateAfterReconcile = async (
  queryClient: QueryClient,
  instanceId: string,
  reconcileResponse: V1ReconcileResponse,
) => {
  const erroredMapByPath = getMapFromArray(
    reconcileResponse.errors,
    (reconcileError) => reconcileError.filePath,
  );

  // invalidate lists of catalog entries and files
  await Promise.all([
    queryClient.refetchQueries(getRuntimeServiceListFilesQueryKey(instanceId)),
    queryClient.refetchQueries(
      getRuntimeServiceListCatalogEntriesQueryKey(instanceId),
    ),
    // TODO: There are other list calls with filters for model and metrics view.
    //       We should perhaps have a single call and filter required items in a selector
    queryClient.refetchQueries(
      getRuntimeServiceListCatalogEntriesQueryKey(instanceId, {
        type: "OBJECT_TYPE_SOURCE",
      }),
    ),
  ]);

  // invalidate affected catalog entries and files
  await Promise.all(
    reconcileResponse.affectedPaths
      .map((path) => [
        queryClient.refetchQueries(
          getRuntimeServiceGetFileQueryKey(
            instanceId,
            removeLeadingSlash(path),
          ),
        ),
        queryClient.refetchQueries(
          getRuntimeServiceGetCatalogEntryQueryKey(
            instanceId,
            get(fileArtifactsStore).entities[path]?.name ??
              getNameFromFile(path),
          ),
        ),
      ])
      .flat(),
  );
  // invalidate tablewide profiling queries
  // (applies to sources and models, but not dashboards)
  await Promise.all(
    reconcileResponse.affectedPaths.map((path) =>
      getInvalidationsForPath(queryClient, path, erroredMapByPath.has(path)),
    ),
  );
};

const getInvalidationsForPath = (
  queryClient: QueryClient,
  filePath: string,
  failed: boolean,
) => {
  const name = getNameFromFile(filePath);
  if (filePath.startsWith("/dashboards")) {
    return invalidateMetricsViewData(queryClient, name, failed);
  } else {
    return invalidateProfilingQueries(queryClient, name, failed);
  }
};

export function isMetricsViewQuery(queryHash, metricsViewName: string) {
  const r = new RegExp(
    `/v1/instances/[a-zA-Z0-9-]+/queries/metrics-views/${metricsViewName}/`,
  );
  return r.test(queryHash);
}
export function invalidationForMetricsViewData(query, metricsViewName: string) {
  return (
    typeof query.queryKey[0] === "string" &&
    isMetricsViewQuery(query.queryKey[0], metricsViewName)
  );
}

export const invalidateMetricsViewData = (
  queryClient: QueryClient,
  metricsViewName: string,
  failed: boolean,
) => {
  // remove inactive queries, this is needed since these would be re-fetched with incorrect filter
  // invalidateQueries by itself doesnt work as of now.
  // reference: https://github.com/rilldata/rill/pull/2027#discussion_r1161672656
  queryClient.removeQueries({
    predicate: (query) =>
      invalidationForMetricsViewData(query, metricsViewName),
    type: "inactive",
  });
  // do not re-fetch for failed entities.
  if (failed) return Promise.resolve();

  return queryClient.resetQueries({
    predicate: (query) =>
      invalidationForMetricsViewData(query, metricsViewName),
    type: "active",
  });
};

export async function invalidateAllMetricsViews(
  queryClient: QueryClient,
  instanceId: string,
) {
  // First, refetch the resource entries (which returns the available dimensions and measures)
  await queryClient.refetchQueries({
    predicate: (query) =>
      typeof query.queryKey[0] === "string" &&
      query.queryKey[0].startsWith(`/v1/instances/${instanceId}/resource`),
  });

  // Second, reset queries for all metrics views. This will cause the active queries to refetch.
  // Note: This is a confusing hack. At time of writing, neither `queryClient.refetchQueries`
  // nor `queryClient.invalidateQueries` were working as expected. Perhaps there's a race condition somewhere.
  queryClient.resetQueries({
    predicate: (query: Query) => {
      return (
        typeof query.queryKey[0] === "string" &&
        query.queryKey[0].startsWith(
          `/v1/instances/${instanceId}/queries/metrics-views`,
        )
      );
    },
  });

  // Additionally, reset the queries for the rows viewer, which have custom query keys
  queryClient.resetQueries({
    predicate: (query: Query) => {
      return (
        typeof query.queryKey[0] === "string" &&
        (query.queryKey[0].startsWith(`dashboardFilteredRowsCt`) ||
          query.queryKey[0].startsWith(`dashboardAllRowsCt`))
      );
    },
  });
}

export async function invalidateProfilingQueries(
  queryClient: QueryClient,
  name: string,
  failed: boolean,
) {
  queryClient.removeQueries({
    predicate: (query) => isProfilingQuery(query, name),
    type: "inactive",
  });
  // do not re-fetch for failed entities.
  if (failed) return Promise.resolve();

  queryClient.removeQueries({
    predicate: (query) => isColumnProfilingQuery(query, name),
    type: "active",
  });

  return queryClient.resetQueries({
    predicate: (query) => isTableProfilingQuery(query, name),
    type: "active",
  });
}

export const removeEntityQueries = async (
  queryClient: QueryClient,
  instanceId: string,
  path: string,
) => {
  const name = getNameFromFile(path);
  // remove affected catalog entries and files
  await Promise.all([
    queryClient.removeQueries(
      getRuntimeServiceGetFileQueryKey(instanceId, removeLeadingSlash(path)),
    ),
    queryClient.removeQueries(
      getRuntimeServiceGetCatalogEntryQueryKey(instanceId, name),
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
      predicate: (query) => isProfilingQuery(query, name),
    });
  }
};
