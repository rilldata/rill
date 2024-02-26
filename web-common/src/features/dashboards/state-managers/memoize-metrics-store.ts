import { Readable, derived } from "svelte/store";
import type { StateManagers } from "./state-managers";

/**
 * Higher order function to create a memoized store based on metrics view name
 */
export function memoizeMetricsStore<Store extends Readable<any>>(
  storeGetter: (ctx: StateManagers) => Store,
) {
  const cache = new Map<string, Store>();
  return (ctx: StateManagers): Store => {
    return derived([ctx.metricsViewName], ([name], set) => {
      let store = cache.get(name);
      if (!store) {
        store = storeGetter(ctx);
        cache.set(name, store);
      }
      return store.subscribe(set);
    });
  };
}
