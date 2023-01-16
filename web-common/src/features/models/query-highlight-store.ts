import { writable } from "svelte/store";

/**
 * This store is currently used to enable highlighting within the query interface.
 * Any part of the application can set this.
 * If we enable this kind of interaction, we should support having this store have an explicit set
 * operation based on types.
 */
export function createQueryHighlightStore() {
  const { subscribe, set } = writable(undefined);
  return {
    subscribe,
    set,
  };
}
