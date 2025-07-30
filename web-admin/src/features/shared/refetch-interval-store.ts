import type { V1Resource } from "@rilldata/web-common/runtime-client";

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

/**
 * Updates the meta object with a smart refetch interval based on resource states.
 * @param resources Array of resources to check
 * @param prevMeta Previous meta object (should contain refetchInterval and wasReconciling)
 * @returns Updated meta object with new refetchInterval and wasReconciling state
 */
export function updateSmartRefetchMeta(
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
      wasReconciling: hasReconciling,
    };
  }

  // CONTINUING reconciliation cycle - apply backoff
  const current =
    typeof prevMeta.refetchInterval === "number"
      ? prevMeta.refetchInterval
      : INITIAL_REFETCH_INTERVAL;
  const next = Math.min(current * BACKOFF_FACTOR, MAX_REFETCH_INTERVAL);

  return { refetchInterval: next, wasReconciling: hasReconciling };
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
export function createSmartRefetchInterval(query: any): number | false {
  if (!query.state.data?.resources) {
    return false;
  }

  // Get or initialize state from WeakMap
  const currentState = queryRefetchStateMap.get(query) || {};
  const updatedState = updateSmartRefetchMeta(
    query.state.data.resources,
    currentState,
  );

  // Store updated state in WeakMap
  queryRefetchStateMap.set(query, updatedState);

  return updatedState.refetchInterval;
}
