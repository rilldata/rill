import { clamp } from "@rilldata/web-common/lib/clamp";
import { get, writable, type Writable } from "svelte/store";

export const DEFAULT_COLUMN_WIDTH = 110;
export const COMPARISON_COLUMN_WIDTH = 64;
export const MEASURES_PADDING = 16;

const MIN_COL_WIDTH = 56;
const MAX_COL_WIDTH = 164;

class ColumnStore {
  private value: Writable<number>;
  private defaultWidth: number;

  subscribe: (this: void, run: (value: number) => void) => () => void;

  constructor(defaultWidth: number = DEFAULT_COLUMN_WIDTH) {
    this.defaultWidth = defaultWidth;
    this.value = writable(defaultWidth);
    this.subscribe = this.value.subscribe;
  }

  update = (newValue: number) => {
    newValue = clamp(
      MIN_COL_WIDTH,
      Math.ceil(newValue) + MEASURES_PADDING,
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

export const valueColumn = new ColumnStore(DEFAULT_COLUMN_WIDTH);
export const deltaColumn = new ColumnStore(COMPARISON_COLUMN_WIDTH);
