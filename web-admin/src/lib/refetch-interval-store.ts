import type {
  V1ListResourcesResponse,
  V1Resource,
} from "@rilldata/web-common/runtime-client";
import type { ConnectError } from "@connectrpc/connect";
import type { Query } from "@tanstack/svelte-query";

export const INITIAL_REFETCH_INTERVAL = 500; // Start at 500ms for quick feedback
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
 * Creates a smart refetch interval function that only considers resources
 * matching a predicate. Use this when the query fetches all resources but
 * only a subset is relevant (e.g., useDashboards fetches everything but
 * only cares about canvas/explore).
 */
export function createSmartRefetchInterval(
  isRelevant: (resource: V1Resource) => boolean,
) {
  return function refetchInterval(
    query: Query<
      V1ListResourcesResponse,
      ConnectError,
      V1ListResourcesResponse,
      readonly unknown[]
    >,
  ): number | false {
    const resources = query.state.data?.resources;

    // No data (query errored or hasn't resolved) or empty resource list
    // (runtime just started, parser hasn't created resources yet): keep
    // polling so we pick up resources once the runtime is ready.
    if (!resources || resources.length === 0) {
      return MAX_REFETCH_INTERVAL;
    }

    const relevantResources = resources.filter(isRelevant);

    // When no relevant resources exist, fall back to non-ProjectParser
    // resources. If any are reconciling (models, sources being built),
    // relevant resources may still appear as reconciliation progresses.
    //
    // Additionally, if non-parser resources are all idle but the parser
    // is still reconciling, include it: during wake-up the parser creates
    // resources incrementally, so explores/canvases may not exist yet
    // even though earlier resources (sources, models) have finished.
    let toCheck: V1Resource[];
    if (relevantResources.length > 0) {
      toCheck = relevantResources;
    } else {
      const nonParser = resources.filter((r) => !r.projectParser);
      const parserReconciling = resources.some(
        (r) => !!r.projectParser && isResourceReconciling(r),
      );
      if (nonParser.length > 0 && parserReconciling) {
        toCheck = resources;
      } else {
        toCheck = nonParser;
      }
    }

    const currentState = queryRefetchStateMap.get(query) || {};
    const updatedState = updateSmartRefetchMeta(toCheck, currentState);
    queryRefetchStateMap.set(query, updatedState);

    return updatedState.refetchInterval;
  };
}

/**
 * A smart refetch interval function that uses a WeakMap to store state.
 * This approach keeps refetch state per query without mutating the query object.
 * Checks ALL resources in the response; use createSmartRefetchInterval
 * when you need to scope to a subset.
 */
export function smartRefetchIntervalFunc(
  query: Query<
    V1ListResourcesResponse,
    ConnectError,
    V1ListResourcesResponse,
    readonly unknown[]
  >,
): number | false {
  if (!query.state.data?.resources) {
    return false;
  }

  const resources = query.state.data.resources;

  // Get or initialize state from WeakMap
  const currentState = queryRefetchStateMap.get(query) || {};
  const updatedState = updateSmartRefetchMeta(resources, currentState);

  // Store updated state in WeakMap
  queryRefetchStateMap.set(query, updatedState);

  return updatedState.refetchInterval;
}
