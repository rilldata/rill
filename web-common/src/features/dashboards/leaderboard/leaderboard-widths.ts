import { clamp } from "@rilldata/web-common/lib/clamp";
import { get, writable } from "svelte/store";

export const DEFAULT_COL_WIDTH = 60;
const MIN_COL_WIDTH = 56;
const MAX_COL_WIDTH = 164;
const PADDING = 16;

class ColumnStore {
  value = writable(DEFAULT_COL_WIDTH);

  subscribe = this.value.subscribe;

  update = (newValue: number) => {
    newValue = clamp(
      MIN_COL_WIDTH,
      Math.ceil(newValue) + PADDING,
      MAX_COL_WIDTH,
    );
    const value = get(this.value);

    if (newValue > value) {
      this.value.set(newValue);
    }
  };

  reset() {
    this.value.set(DEFAULT_COL_WIDTH);
  }
}

export const valueColumn = new ColumnStore();
export const deltaColumn = new ColumnStore();
