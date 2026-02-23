import { writable } from "svelte/store";
import type { ScaleLinear } from "d3-scale";

export interface HoverRange {
  start: number;
  end: number;
}

function createHoverIndex() {
  const { subscribe, set: _set } = writable<HoverRange | undefined>(undefined);
  let currentOwner: string | null = null;
  let _xScale: ScaleLinear<number, number> | null = null;

  return {
    subscribe,
    /** Set a single hovered index (start === end). */
    set(index: number, owner: string) {
      currentOwner = owner;
      _set({ start: index, end: index });
    },
    /** Set a range of highlighted indices. */
    setRange(start: number, end: number, owner: string) {
      currentOwner = owner;
      _set({ start: Math.min(start, end), end: Math.max(start, end) });
    },
    clear(owner: string) {
      if (owner === currentOwner) {
        currentOwner = null;
        _set(undefined);
      }
    },
    registerScale(scale: ScaleLinear<number, number>) {
      _xScale = scale;
    },
    get xScale() {
      return _xScale;
    },
  };
}

export const hoverIndex = createHoverIndex();
