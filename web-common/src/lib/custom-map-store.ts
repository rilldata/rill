import { get, writable } from "svelte/store";

export function createCustomMapStore<T>() {
  const { subscribe, set, update } = writable(new Map<string, T>());

  return {
    subscribe,
    read: () => get({ subscribe }),
    getNonReactive: (name: string) => {
      return get({ subscribe }).get(name);
    },
    set: (name: string, component: T) => {
      update((map) => {
        map.set(name, component);
        return map;
      });
    },
    delete: (name: string) => {
      update((map) => {
        map.delete(name);
        return map;
      });
    },
    reset: () => set(new Map()),
  };
}
