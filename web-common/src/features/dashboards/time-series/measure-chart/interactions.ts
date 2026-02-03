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

const EMPTY_HOVER: HoverState = {
  index: null,
  screenX: null,
  screenY: null,
  isHovered: false,
};

/**
 * Create a simple hover state store.
 */
export function createHoverState(): Writable<HoverState> {
  return writable<HoverState>(EMPTY_HOVER);
}

export { EMPTY_HOVER };

/**
 * Helper to get ordered start/end dates.
 */
export function getOrderedDates(
  start: Date | null,
  end: Date | null,
): { start: Date | null; end: Date | null } {
  if (!start || !end) return { start, end };
  return start.getTime() > end.getTime()
    ? { start: end, end: start }
    : { start, end };
}
