import { writable } from "svelte/store";

export const perfTestStore = writable<{
  batch: boolean;
  cache: boolean;
}>({
  batch: false,
  cache: false,
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
    return state;
  });
}
