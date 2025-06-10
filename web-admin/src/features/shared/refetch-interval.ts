import type { Query } from "@tanstack/svelte-query";
import type {
  V1ListResourcesResponse,
  V1Resource,
} from "@rilldata/web-common/runtime-client";
import type { HTTPError } from "@rilldata/web-common/runtime-client/fetchWrapper";

export const INITIAL_REFETCH_INTERVAL = 200; // Start at 200ms for immediate feedback
export const MAX_REFETCH_INTERVAL = 2_000; // Cap at 2s
export const BACKOFF_FACTOR = 1.5;

export function isResourceErrored(resource: V1Resource) {
  return !!resource?.meta?.reconcileError;
}

export function isResourceReconciling(resource: V1Resource) {
  return (
    resource?.meta?.reconcileStatus === "RECONCILE_STATUS_PENDING" ||
    resource?.meta?.reconcileStatus === "RECONCILE_STATUS_RUNNING"
  );
}

export function calculateRefetchInterval(
  query: Query<V1ListResourcesResponse, HTTPError>,
): number | false {
  if (query.state.error) return false;
  if (!query.state.data?.resources) return false; // Stop polling if no resources

  const hasReconcilingResources = query.state.data.resources.some(
    isResourceReconciling,
  );

  // Only stop polling if there are no reconciling resources
  if (!hasReconcilingResources) {
    return false;
  }

  // Get the current interval from the query's state
  const currentInterval =
    query.state.dataUpdateCount === 0
      ? INITIAL_REFETCH_INTERVAL
      : INITIAL_REFETCH_INTERVAL *
        Math.pow(BACKOFF_FACTOR, Math.min(query.state.dataUpdateCount, 5));

  return Math.min(currentInterval, MAX_REFETCH_INTERVAL);
}
