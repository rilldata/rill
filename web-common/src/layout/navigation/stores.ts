import { page } from "$app/stores";
import { get, writable } from "svelte/store";

/** maintains the global state across all navigation entries.
 * We use this store to track immediate clicks, rather than use the page store
 * from sveltekit. This is because the page store is updated after the navigation
 * while we need to capture immediate clicks for styling information.
 */
export const currentHref = writable<string>(undefined);

/**
 * For now, we will reactively update currentHref when the page store changes
 * and the pathname doesn't match. This enables us to reduce the latency in
 * the perceived click event of a menu item. Once we are able to reduce route
 * change loads to < 50ms, we can re-address whether we need this store.
 */
page.subscribe((pageState) => {
  const href = get(currentHref);
  const pathname = pageState?.url?.pathname;
  // prevent the currentHref from being set to undefined.
  if (!pathname) return;
  // only update the currentHref if needed
  if (href !== pathname) currentHref.set(pathname);
});
