import {
  isColumnProfilingQuery,
  isProfilingQuery,
  isTableProfilingQuery,
} from "@rilldata/web-common/runtime-client/query-matcher";
import type { Query, QueryClient } from "@tanstack/svelte-query";

// invalidation helpers

export function invalidateRuntimeQueries(queryClient: QueryClient) {
  return queryClient.resetQueries({
    predicate: (query) =>
      typeof query.queryKey[0] === "string" &&
      query.queryKey[0].startsWith("/v1/instances"),
  });
}

export function isMetricsViewQuery(queryHash: string, metricsViewName: string) {
  const r = new RegExp(
    `/v1/instances/[a-zA-Z0-9-]+/queries/metrics-views/${metricsViewName}/`,
  );
  return r.test(queryHash);
}
export function invalidationForMetricsViewData(
  query: Query,
  metricsViewName: string,
) {
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
  void queryClient.resetQueries({
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
  return queryClient.resetQueries({
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
