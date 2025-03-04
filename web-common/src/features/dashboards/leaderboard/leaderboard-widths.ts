import { clamp } from "@rilldata/web-common/lib/clamp";
import { get, writable, type Writable } from "svelte/store";

export const DEFAULT_COL_WIDTH = 80;
export const DEFAULT_CONTEXT_COLUMN_WIDTH = 60;
export const MEASURE_SPACING_WIDTH = 16;

const MIN_COL_WIDTH = 56;
const MAX_COL_WIDTH = 164;
const PADDING = 16;

class ColumnStore {
  private value: Writable<number>;
  private defaultWidth: number;

  subscribe: (this: void, run: (value: number) => void) => () => void;

  constructor(defaultWidth: number = DEFAULT_COL_WIDTH) {
    this.defaultWidth = defaultWidth;
    this.value = writable(defaultWidth);
    this.subscribe = this.value.subscribe;
  }

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
    this.value.set(this.defaultWidth);
  }
}

export const valueColumn = new ColumnStore(DEFAULT_COL_WIDTH);
export const deltaColumn = new ColumnStore(DEFAULT_CONTEXT_COLUMN_WIDTH);
