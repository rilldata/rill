import { Writable, get, writable } from "svelte/store";

export function previousValueStore<T extends number | string>(
  anotherStore: Writable<T>,
): Writable<T> {
  let previousValue = get(anotherStore);
  const store = writable(previousValue);
  anotherStore.subscribe(($currentValue) => {
    store.set(previousValue);
    previousValue = $currentValue;
  });
  return store;
}
