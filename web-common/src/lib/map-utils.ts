import { get, writable } from "svelte/store";

export function reverseMap<
  K extends string | number,
  V extends string | number,
>(map: Partial<Record<K, V>>): Partial<Record<V, K>> {
  const revMap = {} as Partial<Record<V, K>>;
  for (const k in map) {
    revMap[map[k] as string | number] = k;
  }
  return revMap;
}

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
