import { writable } from "svelte/store";

export const pageInFocus = writable<boolean>(true);

export function addViewportListener() {
  document.addEventListener("visibilitychange", () => {
    pageInFocus.set(!document.hidden);
  });
}
