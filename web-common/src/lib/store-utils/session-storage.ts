import { browser } from "$app/environment";
import { writable } from "svelte/store";
import { debounce } from "../create-debouncer";

/** Creates a store whose value is stored in sessionStorage as a string.
 * Only JSON-serializable values can be used.
 */
export function sessionStorageStore<T>(
  itemKey: string,
  defaultValue: T | undefined = undefined,
) {
  const store = writable<T>(defaultValue);

  const loadData = () => {
    if (!browser) return;

    const stored = sessionStorage.getItem(itemKey);
    if (stored !== null) {
      try {
        const parsed = JSON.parse(stored);
        if (parsed !== undefined) {
          store.set(parsed);
        }
      } catch {
        // ignore
      }
    }
  };
  loadData();
  const debouncer = debounce((v: T) => {
    if (v === undefined) return;
    sessionStorage.setItem(itemKey, JSON.stringify(v));
  }, 300);
  store.subscribe(debouncer);

  return {
    ...store,
    reload() {
      loadData();
    },
  };
}
