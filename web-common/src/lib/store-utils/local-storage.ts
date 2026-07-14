import { browser } from "$app/environment";
import { writable } from "svelte/store";
import { debounce } from "../create-debouncer";

/** Creates a store whose value is stored in localStorage as a string.
 * Only JSON-serializable values can be used.
 */
export function localStorageStore<T>(itemKey: string, defaultValue: T) {
  const store = writable<T>(defaultValue);

  if (browser) {
    try {
      // Accessing localStorage can throw, not just return null: a sandboxed or
      // storage-partitioned iframe (embed mode) raises a SecurityError. Keep the
      // read inside the try so initialization degrades to in-memory only.
      const stored = localStorage.getItem(itemKey);
      if (stored !== null) {
        const parsed = JSON.parse(stored);
        if (parsed !== undefined) {
          store.set(parsed);
        }
      }
    } catch {
      // ignore: localStorage unavailable or unreadable
    }
  }
  const debouncer = debounce((v: T) => {
    try {
      localStorage.setItem(itemKey, JSON.stringify(v));
    } catch {
      // ignore: localStorage unavailable or over quota (e.g. embed iframe)
    }
  }, 300);
  store.subscribe(debouncer);

  return {
    ...store,
    reset() {
      store.set(defaultValue);
      try {
        localStorage.setItem(itemKey, JSON.stringify(defaultValue));
      } catch {
        // ignore: localStorage unavailable
      }
    },
  };
}

/**
 * Simplified version of localStorageStore that only stores value on an explicit set call.
 * Where as localStorageStore will store the default value as well.
 */
export function explicitLocalStorageStore<T>(itemKey: string, defaultValue: T) {
  let initValue: T = defaultValue;
  if (browser) {
    const stored = localStorage.getItem(itemKey);
    if (stored !== null) {
      try {
        const parsed = JSON.parse(stored);
        if (parsed !== undefined) {
          initValue = parsed;
        }
      } catch {
        // ignore
      }
    }
  }

  const store = writable(initValue);

  return {
    ...store,
    set: (value: T) => {
      store.set(value);
      try {
        localStorage.setItem(itemKey, JSON.stringify(value));
      } catch {
        // no-op
      }
    },
  };
}
