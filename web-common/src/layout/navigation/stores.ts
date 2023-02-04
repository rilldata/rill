import { writable } from "svelte/store";
/** maintains the global state across all navigation entries.
 * We use this store to track immediate clicks, rather than use the page store
 * from sveltekit. This is because the page store is updated after the navigation
 * while we need to capture immediate clicks.
 */
export const currentHref = writable<string>(undefined);
