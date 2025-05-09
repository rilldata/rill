import type { Query } from "@tanstack/svelte-query";
import type {
  V1ListResourcesResponse,
  V1Resource,
} from "@rilldata/web-common/runtime-client";
import type { HTTPError } from "@rilldata/web-common/runtime-client/fetchWrapper";
import { writable, type Writable } from "svelte/store";

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

export function pollUntilResourcesReconciled(
  currentInterval: number,
  data: V1ListResourcesResponse | undefined,
  query: Query<V1ListResourcesResponse, HTTPError>,
): number | false {
  if (query.state.error) return false;
  if (!data?.resources) return false; // Stop polling if no resources

  const hasErrors = data.resources.some(isResourceErrored);
  const hasReconcilingResources = data.resources.some(isResourceReconciling);

  if (hasErrors || !hasReconcilingResources) {
    return false;
  }

  return Math.min(currentInterval * BACKOFF_FACTOR, MAX_REFETCH_INTERVAL);
}

function createRefetchIntervalStore() {
  const store: Writable<number> = writable(INITIAL_REFETCH_INTERVAL);
  let currentValue = INITIAL_REFETCH_INTERVAL;

  store.subscribe((value) => {
    currentValue = value;
  });

  return {
    subscribe: store.subscribe,
    reset: () => {
      store.set(INITIAL_REFETCH_INTERVAL);
      return INITIAL_REFETCH_INTERVAL;
    },
    calculatePollingInterval: (
      query: Query<V1ListResourcesResponse, HTTPError>,
    ): number | false => {
      const newInterval = pollUntilResourcesReconciled(
        currentValue,
        query.state.data,
        query,
      );

      if (newInterval === false) {
        if (currentValue !== INITIAL_REFETCH_INTERVAL) {
          // Only reset if we were actually polling
          store.set(INITIAL_REFETCH_INTERVAL);
        }
        return false;
      }

      store.set(newInterval);
      return newInterval;
    },
  };
}

export const refetchIntervalStore = createRefetchIntervalStore();
