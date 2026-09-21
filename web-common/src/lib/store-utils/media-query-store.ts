import { readable, type Readable } from "svelte/store";

/**
 * Tracks whether a CSS media query matches.
 * Unlike `hidden`/`sm:block` classes, rendering from it with `{#if}` keeps the side of a breakpoint
 * that isn't shown from mounting at all, along with its queries and connections.
 */
export function mediaQueryStore(query: string): Readable<boolean> {
  return readable(false, (set) => {
    if (typeof window === "undefined") return;
    const mediaQueryList = window.matchMedia(query);
    const update = () => set(mediaQueryList.matches);
    update();
    mediaQueryList.addEventListener("change", update);
    return () => mediaQueryList.removeEventListener("change", update);
  });
}

/** Below Tailwind's `sm` breakpoint, matching its `max-sm:` variant. */
export const belowSm = mediaQueryStore("not all and (min-width: 640px)");
