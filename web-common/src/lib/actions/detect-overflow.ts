/**
 * Svelte action that reports whether a node's content is horizontally
 * truncated (`scrollWidth > clientWidth`). Useful for showing a tooltip only
 * when text is actually clipped. Re-checks on element resize and whenever the
 * action parameter changes (e.g. the rendered text updates).
 */
export type OverflowCallback = (isOverflowing: boolean) => void;

export function detectOverflow(node: HTMLElement, callback: OverflowCallback) {
  let current = callback;

  function check() {
    current(node.scrollWidth > node.clientWidth);
  }

  let observer: ResizeObserver | undefined;
  if (typeof ResizeObserver !== "undefined") {
    observer = new ResizeObserver(check);
    observer.observe(node);
  }
  check();

  return {
    update(next: OverflowCallback) {
      current = next;
      check();
    },
    destroy() {
      observer?.disconnect();
    },
  };
}
