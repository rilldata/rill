import { writable } from "svelte/store";

export const addSourceModal = (() => {
  const { subscribe, set, update } = writable(false);

  return {
    subscribe,
    open: () => set(true),
    close: () => set(false),
    toggle: () => update((state) => !state),
  };
})();
