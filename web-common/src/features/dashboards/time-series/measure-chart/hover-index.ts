import { writable } from "svelte/store";
import type { ScaleLinear } from "d3-scale";

function createHoverIndex() {
  const { subscribe, set: _set } = writable<number | undefined>(undefined);
  let currentOwner: string | null = null;
  let _xScale: ScaleLinear<number, number> | null = null;

  return {
    subscribe,
    set(index: number, owner: string) {
      currentOwner = owner;
      _set(index);
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
