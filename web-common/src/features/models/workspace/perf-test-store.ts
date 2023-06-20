import { runtimeServiceToggleCaching } from "@rilldata/web-common/runtime-client";
import { writable } from "svelte/store";

export const perfTestStore = writable<{
  batch: boolean;
  cache: boolean;
}>({
  batch: true,
  cache: true,
});

export function togglePerfBatch() {
  perfTestStore.update((state) => {
    state.batch = !state.batch;
    return state;
  });
}

export function togglePerfCache() {
  perfTestStore.update((state) => {
    state.cache = !state.cache;
    runtimeServiceToggleCaching({
      enable: state.cache,
    });
    return state;
  });
}
