import { writable } from "svelte/store";

/**
 * Used to notify the viewport listener that focus will be taken off.
 * Right now this is used for file upload.
 * True = focus has been taken intentionally, False otherwise
 */
export const focusNotifier = writable<boolean>(false);
export function takeFocus() {
  focusNotifier.set(true);
}
export function giveBackFocus() {
  focusNotifier.set(false);
}

export const pageInFocus = writable<boolean>(true);

export function addViewportListener() {
  document.addEventListener("visibilitychange", () => {
    pageInFocus.set(!document.hidden);
  });
}
