import { writable } from "svelte/store";

/**
 * Preview mode is URL-driven: the root layout syncs this store based on the
 * current URL (preview routes → true, developer routes → false, shared routes
 * keep the current value).  Using a plain writable (no localStorage) means
 * closing the browser and returning always defaults to developer mode.
 */
export const previewModeStore = writable<boolean>(false);
