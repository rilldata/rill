import type { Reference } from "@rilldata/web-common/features/models/utils/get-table-references";
import { writable } from "svelte/store";

export type QueryHighlightState = Array<Reference>;

/**
 * This store is currently used to enable highlighting within the query interface.
 * Any part of the application can set this.
 * If we enable this kind of interaction, we should support having this store have an explicit set
 * operation based on types.
 */
export function createQueryHighlightStore() {
  const { subscribe, set } = writable<QueryHighlightState>(undefined);
  return {
    subscribe,
    set,
  };
}
