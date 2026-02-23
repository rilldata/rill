import { writable, type Writable } from "svelte/store";
import type { HoverState } from "./types";

/**
 * Create an IntersectionObserver-based visibility store.
 * Used for lazy-loading chart data when the chart scrolls into view.
 */
export function createVisibilityObserver(rootMargin = "120px"): {
  visible: Writable<boolean>;
  observe: (element: HTMLElement, root?: HTMLElement | null) => () => void;
} {
  const visible = writable(false);

  function observe(
    element: HTMLElement,
    root: HTMLElement | null = null,
  ): () => void {
    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          visible.set(true);
          observer.unobserve(element);
        }
      },
      { root, rootMargin, threshold: 0 },
    );
    observer.observe(element);
    return () => observer.disconnect();
  }

  return { visible, observe };
}

export const EMPTY_HOVER: HoverState = {
  index: null,
  screenX: null,
  screenY: null,
  isHovered: false,
};
