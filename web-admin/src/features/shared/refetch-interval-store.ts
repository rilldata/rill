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
 * @param prevMeta Previous meta object (should contain refetchInterval)
 * @returns Updated meta object with new refetchInterval
 */
export function updateSmartRefetchMeta(
  resources: any[] | undefined,
  prevMeta: { refetchInterval?: number | false } = {},
): { refetchInterval: number | false } {
  if (!resources) {
    return { ...prevMeta, refetchInterval: false };
  }
  const hasReconciling = resources.some(isResourceReconciling);
  if (!hasReconciling) {
    return { ...prevMeta, refetchInterval: false };
  }
  // Backoff logic: update the interval, but reset if a new cycle starts
  const current =
    typeof prevMeta.refetchInterval === "number"
      ? prevMeta.refetchInterval
      : INITIAL_REFETCH_INTERVAL;
  const next = Math.min(current * BACKOFF_FACTOR, MAX_REFETCH_INTERVAL);
  return { ...prevMeta, refetchInterval: next };
}

/**
 * Creates a smart refetch interval function that uses query.meta to store state.
 * This approach keeps refetch state per query and encapsulates all logic in the refetchInterval parameter.
 *
 * @param query The TanStack query object
 * @returns The refetch interval (number in ms or false to disable)
 */
export function createSmartRefetchInterval(query: any): number | false {
  if (!query.state.data?.resources) {
    return false;
  }

  // Get or initialize meta with refetch state
  const currentMeta = query.meta || {};
  const updatedMeta = updateSmartRefetchMeta(
    query.state.data.resources,
    currentMeta,
  );

  // Update query meta with new refetch state
  if (query.meta !== updatedMeta) {
    query.meta = updatedMeta;
  }

  return updatedMeta.refetchInterval;
}
