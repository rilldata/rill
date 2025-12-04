import type {
  V1ListResourcesResponse,
  V1Resource,
} from "@rilldata/web-common/runtime-client";
import type { HTTPError } from "@rilldata/web-common/runtime-client/fetchWrapper";
import type { Query } from "@tanstack/svelte-query";

export const INITIAL_REFETCH_INTERVAL = 200; // Start at 200ms for immediate feedback
export const MAX_REFETCH_INTERVAL = 2_000; // Cap at 2s
export const BACKOFF_FACTOR = 1.5;

export function isResourceReconciling(resource: V1Resource) {
  return (
    resource?.meta?.reconcileStatus === "RECONCILE_STATUS_PENDING" ||
    resource?.meta?.reconcileStatus === "RECONCILE_STATUS_RUNNING"
  );
}

/**
 * Updates the meta object with a smart refetch interval based on resource states.
 * @param resources Array of resources to check
 * @param prevMeta Previous meta object (should contain refetchInterval and wasReconciling)
 * @returns Updated meta object with new refetchInterval and wasReconciling state
 */
function updateSmartRefetchMeta(
  resources: any[] | undefined,
  prevMeta: { refetchInterval?: number | false; wasReconciling?: boolean } = {},
): { refetchInterval: number | false; wasReconciling: boolean } {
  if (!resources) {
    return { refetchInterval: false, wasReconciling: false };
  }

  const hasReconciling = resources.some(isResourceReconciling);
  const wasReconciling = prevMeta.wasReconciling || false;

  if (!hasReconciling) {
    // No reconciling resources - stop polling
    return { refetchInterval: false, wasReconciling: false };
  }

  // Resources are reconciling
  if (!wasReconciling) {
    // NEW reconciliation cycle - reset to initial interval
    return {
      refetchInterval: INITIAL_REFETCH_INTERVAL,
      wasReconciling: true,
    };
  }

  // CONTINUING reconciliation cycle - apply backoff
  const current =
    typeof prevMeta.refetchInterval === "number"
      ? prevMeta.refetchInterval
      : INITIAL_REFETCH_INTERVAL;
  const next = Math.min(current * BACKOFF_FACTOR, MAX_REFETCH_INTERVAL);

  return { refetchInterval: next, wasReconciling: true };
}

// WeakMap to store refetch state associated with each query
const queryRefetchStateMap = new WeakMap<
  any,
  { refetchInterval?: number | false; wasReconciling?: boolean }
>();

/**
 * Creates a smart refetch interval function that uses a WeakMap to store state.
 * This approach keeps refetch state per query without mutating the query object.
 *
 * @param query The TanStack query object
 * @returns The refetch interval (number in ms or false to disable)
 */
export function createSmartRefetchInterval(
  query: Query<
    V1ListResourcesResponse,
    HTTPError,
    V1ListResourcesResponse,
    readonly unknown[]
  >,
): number | false {
  if (!query.state.data?.resources) {
    return false;
  }

  const resources = query.state.data.resources;

  // If there are no resources at all, use a fixed refetch interval
  // This handles the case during initial deployment creation when parser hasn't run yet
  if (resources.length === 0) {
    return INITIAL_REFETCH_INTERVAL;
  }

  // Get or initialize state from WeakMap
  const currentState = queryRefetchStateMap.get(query) || {};
  const updatedState = updateSmartRefetchMeta(resources, currentState);

  // Store updated state in WeakMap
  queryRefetchStateMap.set(query, updatedState);

  return updatedState.refetchInterval;
}
