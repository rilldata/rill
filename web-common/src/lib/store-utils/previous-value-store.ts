import { get, writable } from "svelte/store";

export function previousValueStore(anotherStore) {
  let previousValue = get(anotherStore);
  const store = writable(previousValue);
  anotherStore.subscribe(($currentValue) => {
    if (Array.isArray(previousValue)) {
      store.set([...previousValue]);
    } else if (typeof previousValue === "object" && previousValue !== null) {
      store.set({ ...previousValue });
    } else {
      store.set(previousValue);
    }
    previousValue = $currentValue;
  });
  return {
    subscribe: store.subscribe,
    set: store.set,
  };
}
