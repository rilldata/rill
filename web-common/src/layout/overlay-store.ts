import type { SvelteComponent } from "svelte";
import { writable } from "svelte/store";

interface Overlay {
  title: string;
  detail?: {
    component: typeof SvelteComponent<any>;
    props: Record<string, unknown>;
  };
}

const { subscribe, set } = writable<Overlay | null>(null);
let timeout: NodeJS.Timeout;
let isCleared: boolean = false;

export const overlay = {
  subscribe,
  set: (overlay: Overlay | null) => {
    isCleared = false;
    set(overlay);
  },
  /**
   * `setDebounced` is a debounced version of the set method.
   *
   * The overlay will be displayed only if it hasn't been cleared before the specified delay elapses.
   */
  setDebounced: (overlay: Overlay | null, delay: number = 300) => {
    isCleared = false;
    clearTimeout(timeout);
    timeout = setTimeout(() => {
      if (!isCleared) {
        set(overlay);
      }
    }, delay);
  },
  clear: () => {
    isCleared = true;
    set(null);
  },
};
