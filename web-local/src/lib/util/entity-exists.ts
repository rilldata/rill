import { get } from "svelte/store";
import type { AppStore } from "../application-state-stores/application-store";

/** a workaround for getting the client-side store in a page.ts file */
export async function entityExists(store: AppStore, id) {
  let r;
  const pr = new Promise((resolve) => {
    r = resolve;
  });
  const unsubscribe = store.subscribe((s) => {
    if (s.lastUpdated !== 0) {
      r(s);
    }
  });
  await pr;
  const storeValue = get(store);
  const modelExists =
    storeValue?.entities?.some((entity) => entity.id === id) || false;
  unsubscribe();
  return modelExists;
}
