import { browser } from "$app/environment";
import { writable } from "svelte/store";

/** Creates a store whose value is stored in localStorage as a string.
 * Only JSON-serializable values can be used.
 */
export function localStorageStore<T>(itemKey: string, defaultValue: T) {
  let value;
  if (browser) {
    value = JSON.parse(localStorage.getItem(itemKey));
  } else {
    value = defaultValue;
  }

  const {
    subscribe,
    set: setStore,
    update: updateStore,
  } = writable<T>(value ?? defaultValue);
  return {
    subscribe,
    set(value: T) {
      setStore(value);
      localStorage.setItem(itemKey, JSON.stringify(value));
    },
    update(f) {
      updateStore((state) => {
        f(state);
        localStorage.setItem(itemKey, JSON.stringify(state));
        return state;
      });
    },
    reset() {
      localStorage.setItem(itemKey, JSON.stringify(defaultValue));
    },
  };
}
