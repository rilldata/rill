import { get, writable } from "svelte/store";

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
  // TODO: any way to ignore dev tools focusing triggering these?
  window.addEventListener("blur", () => {
    if (get(focusNotifier)) return;
    pageInFocus.set(false);
  });
  window.addEventListener("focus", () => {
    pageInFocus.set(true);
  });
}
