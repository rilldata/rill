import { get } from "svelte/store";

/** a workaround for getting the client-side store in a page.ts file */
export async function entityExists(store, id) {
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
  const modelExists =
    get(store)?.entities?.some((entity) => entity.id === id) || false;
  unsubscribe();
  return modelExists;
}
