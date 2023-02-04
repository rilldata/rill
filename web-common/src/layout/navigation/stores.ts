import { page } from "$app/stores";
import { get, writable } from "svelte/store";

/** maintains the global state across all navigation entries.
 * We use this store to track immediate clicks, rather than use the page store
 * from sveltekit. This is because the page store is updated after the navigation
 * while we need to capture immediate clicks.
 */
export const currentHref = writable<string>(undefined);

page.subscribe((pageState) => {
  const href = get(currentHref);
  const pathname = pageState?.url?.pathname;
  // This is a hack to prevent the currentHref from being set to undefined.
  if (!pathname) return;
  // look at only the first two segments of the path. We map all file type views to these two for now.
  // later, we will have better decision rules for reactively setting this variable.
  const segments = pathname.split("/").slice(0, 3);
  const path = segments.join("/");
  if (href !== path) currentHref.set(path);
});
