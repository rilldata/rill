import { writable } from "svelte/store";

/**
 * Store for tracking whether the graph is in expanded view mode.
 * This is needed because window.history.replaceState doesn't trigger $page updates.
 */
export const isGraphExpanded = writable<boolean>(false);
