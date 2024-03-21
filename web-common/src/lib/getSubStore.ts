import { type Readable, writable } from "svelte/store";

/**
 * Returns a sub store based on a key.
 * Note that it only adds a subscriber to entries only when there is no SubStore present for key.
 * This assumes that a deletion will unload the component.
 */
export function getSubStore<SubStore>(
  entries: Readable<Record<string, any>>,
  subStore: Record<string, SubStore>,
  key: string,
  defaultStore: SubStore,
): Readable<SubStore> {
  const store = writable(defaultStore);
  const unsub = entries.subscribe((e) => {
    if (!(key in entries) || !(key in subStore)) return e;
    setTimeout(() => {
      store.set(subStore[key]);
      unsub();
    }, 0);
    return e;
  });
  return store;
}
