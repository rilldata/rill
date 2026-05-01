import { writable } from "svelte/store";

/**
 * Preview mode is URL-driven: the root layout syncs this store based on the
 * current URL (preview routes → true, developer routes → false, shared routes
 * keep the current value).  Using a plain writable (no localStorage) means
 * closing the browser and returning always defaults to developer mode.
 */
export const previewModeStore = writable<boolean>(false);

/**
 * True only when the runtime was started with `--preview`. The Preview/Edit
 * toggle in the navbar is hidden when this is true: the user cannot exit
 * preview mode without restarting the CLI.
 */
export const previewModeLocked = writable<boolean>(false);
