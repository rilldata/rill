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
