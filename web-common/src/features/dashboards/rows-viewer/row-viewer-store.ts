import { writable } from "svelte/store";

/**
 * This store keeps track the HTML ref of the row viewer element.
 * This is needed to make sure that clicking on it isn't registed
 * as an outside click and does not remove the highlighted cell
 */
export const rowViewerStore = writable<HTMLElement | null>(null);
