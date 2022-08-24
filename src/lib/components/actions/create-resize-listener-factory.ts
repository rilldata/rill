import { writable } from "svelte/store";

/**
 * creates a store whose value is an HTML element, and an action that when
 * put on a DOM element, will listen to it.
 */
export function createResizeListenerActionFactory() {
  const { set, subscribe } = writable(undefined);
  return {
    observedNode: { subscribe },
    listenToNodeResize(node) {
      const observer = new ResizeObserver(() => {
        set(node);
      });
      observer.observe(node);
      return {
        destroy() {
          observer.disconnect();
        },
      };
    },
  };
}
