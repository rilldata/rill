import { writable } from "svelte/store";

/**
 * True when preview mode was forced on by the CLI `--preview` flag.
 * Distinct from `previewModeStore`, which tracks the current URL-derived
 * mode. When locked, the UI must not offer a "return to editor" affordance.
 */
export const previewLocked = writable<boolean>(false);
