import { writable } from "svelte/store";

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

// Store for the current refetch interval
export const refetchInterval = writable<number | false>(
  INITIAL_REFETCH_INTERVAL,
);

// Helper to reset the interval
export function resetRefetchInterval() {
  refetchInterval.set(INITIAL_REFETCH_INTERVAL);
}

/**
 * Call this function with the latest resources array after each query update.
 * It will update the refetchInterval store appropriately.
 */
export function updateSmartRefetchInterval(resources: any[] | undefined) {
  if (!resources) {
    refetchInterval.set(false);
    return;
  }
  const hasReconciling = resources.some(isResourceReconciling);
  if (!hasReconciling) {
    refetchInterval.set(false);
    return;
  }
  // Backoff logic: update the interval, but reset if a new cycle starts
  refetchInterval.update((current) => {
    if (typeof current !== "number") {
      return INITIAL_REFETCH_INTERVAL;
    }
    const next = Math.min(current * BACKOFF_FACTOR, MAX_REFETCH_INTERVAL);
    return next;
  });
}
