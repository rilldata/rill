import { getRuntimeServiceGetInstanceQueryKey } from "@rilldata/web-common/runtime-client/v2/gen/runtime-service";
import {
  isColumnProfilingQuery,
  isProfilingQuery,
  isTableProfilingQuery,
} from "@rilldata/web-common/runtime-client/query-matcher";
import type { Query, QueryClient } from "@tanstack/svelte-query";

/** Matches the new key format for a given instanceId. */
function isRuntimeQueryForInstance(
  queryKey: readonly unknown[],
  instanceId: string,
): boolean {
  // Format: [ServiceName, methodName, instanceId, request]
  if (queryKey.length >= 3 && queryKey[2] === instanceId) {
    const svc = queryKey[0];
    return (
      svc === "QueryService" ||
      svc === "RuntimeService" ||
      svc === "ConnectorService"
    );
  }
  return false;
}

/** Checks if a query key matches a metrics view query (by name). */
function isMetricsViewQueryKey(
  queryKey: readonly unknown[],
  metricsViewName: string,
): boolean {
  // Format: ["QueryService", "metricsView*", instanceId, { metricsViewName, ... }]
  if (
    queryKey[0] === "QueryService" &&
    typeof queryKey[1] === "string" &&
    queryKey[1].startsWith("metricsView")
  ) {
    const request = queryKey[3];
    return (
      typeof request === "object" &&
      request !== null &&
      (request as Record<string, unknown>).metricsViewName === metricsViewName
    );
  }
  return false;
}

/** Checks if a query key matches a component resolve query (by name). */
function isComponentResolveKey(
  queryKey: readonly unknown[],
  componentName: string,
): boolean {
  // Format: ["QueryService", "resolveComponent", instanceId, { component: name }]
  if (queryKey[0] === "QueryService" && queryKey[1] === "resolveComponent") {
    const request = queryKey[3];
    return (
      typeof request === "object" &&
      request !== null &&
      (request as Record<string, unknown>).component === componentName
    );
  }
  return false;
}

// --- invalidation helpers ---

export function invalidateRuntimeQueries(
  queryClient: QueryClient,
  instanceId: string,
) {
  return queryClient.resetQueries({
    predicate: (query) => isRuntimeQueryForInstance(query.queryKey, instanceId),
  });
}

export function invalidationForMetricsViewData(
  query: Query,
  metricsViewName: string,
) {
  return isMetricsViewQueryKey(query.queryKey, metricsViewName);
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
  // First, refetch the instance query. Instance has feature flag that depends on the user.
  await queryClient.refetchQueries({
    type: "active",
    queryKey: getRuntimeServiceGetInstanceQueryKey(instanceId),
  });

  // Second, refetch the resource entries (which returns the available dimensions and measures)
  await queryClient.refetchQueries({
    type: "active",
    predicate: (query) => {
      const key = query.queryKey;
      // Format: ["RuntimeService", "getResource" or "listResources", instanceId, ...]
      return (
        key[0] === "RuntimeService" &&
        (key[1] === "getResource" || key[1] === "listResources") &&
        key[2] === instanceId
      );
    },
  });

  // Third, reset queries for all metrics views. This will cause the active queries to refetch.
  // Note: This is a confusing hack. At time of writing, neither `queryClient.refetchQueries`
  // nor `queryClient.invalidateQueries` were working as expected. Perhaps there's a race condition somewhere.
  void queryClient.resetQueries({
    predicate: (query: Query) => {
      const key = query.queryKey;
      // Format: ["QueryService", "metricsView*", instanceId, ...]
      return (
        key[0] === "QueryService" &&
        typeof key[1] === "string" &&
        key[1].startsWith("metricsView") &&
        key[2] === instanceId
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

export async function invalidateComponentData(
  queryClient: QueryClient,
  name: string,
  failed: boolean,
) {
  const matchesComponent = (query: Query) =>
    isComponentResolveKey(query.queryKey, name);

  queryClient.removeQueries({
    predicate: matchesComponent,
    type: "inactive",
  });
  if (failed) return;

  return queryClient.resetQueries({
    predicate: matchesComponent,
    type: "active",
  });
}
