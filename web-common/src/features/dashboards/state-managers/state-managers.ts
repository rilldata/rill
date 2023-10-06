import { writable, Writable, Readable, derived, get } from "svelte/store";
import { getContext } from "svelte";
import type { QueryClient } from "@tanstack/svelte-query";
import type { Runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import {
  MetricsExplorerEntity,
  MetricsExplorerStoreType,
  metricsExplorerStore,
  updateMetricsExplorerByName,
  useDashboardStore,
} from "../dashboard-stores";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import {
  createStateManagerSelectors,
  type StateManagerSelectors,
} from "./selectors";
import { createStateManagerActions, type StateManagerActions } from "./actions";

/**
 * A MetricsExplorerMutatorClosure is a function that mutates
 * a MetricsExplorerEntity, i.e., the data single dashboard.
 * This will often be a closure over other parameters
 * that are relevant to the mutation.
 */
export type MetricsExplorerMutatorClosure = (
  metricsExplorer: MetricsExplorerEntity
) => void;

export type UpdateDashboard2ndOrderCallback = (
  callback: MetricsExplorerMutatorClosure
) => void;

export type StateManagers = {
  runtime: Writable<Runtime>;
  metricsViewName: Writable<string>;
  metricsStore: Readable<MetricsExplorerStoreType>;
  dashboardStore: Readable<MetricsExplorerEntity>;
  queryClient: QueryClient;
  setMetricsViewName: (s: string) => void;
  updateDashboard: UpdateDashboard2ndOrderCallback;
  /**
   * A collection of Readables that can be used to select data from the dashboard.
   */
  selectors: StateManagerSelectors;
  /**
   * A collection of functions that update the dashboard data model.
   */
  actions: StateManagerActions;
};

export const DEFAULT_STORE_KEY = Symbol("state-managers");

export function getStateManagers(): StateManagers {
  return getContext(DEFAULT_STORE_KEY);
}

export function createStateManagers({
  queryClient,
  metricsViewName,
}: {
  queryClient: QueryClient;
  metricsViewName: string;
}): StateManagers {
  const metricsViewNameStore = writable(metricsViewName);
  const dashboardStore: Readable<MetricsExplorerEntity> = derived(
    [metricsViewNameStore],
    ([name], set) => {
      const store = useDashboardStore(name);
      return store.subscribe(set);
    }
  );

  const updateDashboard = (
    callback: (metricsExplorer: MetricsExplorerEntity) => void
  ) => {
    const name = get(dashboardStore).name;
    // TODO: Remove dependency on MetricsExplorerStore singleton and its exports
    updateMetricsExplorerByName(name, callback);
  };

  return {
    runtime: runtime,
    metricsViewName: metricsViewNameStore,
    metricsStore: metricsExplorerStore,
    queryClient,
    dashboardStore,
    setMetricsViewName: (name) => {
      metricsViewNameStore.set(name);
    },
    updateDashboard,
    selectors: createStateManagerSelectors(dashboardStore),
    actions: createStateManagerActions(updateDashboard),
  };
}

/**
 * Higher order function to create a memoized store based on metrics view name
 */
export function memoizeMetricsStore<Store extends Readable<any>>(
  storeGetter: (ctx: StateManagers) => Store
) {
  const cache = new Map<string, Store>();
  return (ctx: StateManagers): Store => {
    return derived([ctx.metricsViewName], ([name], set) => {
      let store: Store;
      if (cache.has(name)) {
        store = cache.get(name);
      } else {
        store = storeGetter(ctx);
        cache.set(name, store);
      }
      return store.subscribe(set);
    }) as Store;
  };
}
