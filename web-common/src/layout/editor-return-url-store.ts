import { writable } from "svelte/store";

/**
 * The most recent developer-route URL. Captured while the user is in editor
 * mode so the nav-bar "return to editor" action can send them back to where
 * they were when they entered UI-triggered preview mode.
 */
export const editorReturnUrl = writable<string | null>(null);
