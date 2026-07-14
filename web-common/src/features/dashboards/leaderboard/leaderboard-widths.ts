import { clamp } from "@rilldata/web-common/lib/clamp";
import { localStorageStore } from "@rilldata/web-common/lib/store-utils";
import { get, writable, type Writable } from "svelte/store";

export const DEFAULT_COLUMN_WIDTH = 110;
export const COMPARISON_COLUMN_WIDTH = 64;
export const MEASURES_PADDING = 16;

const MIN_COL_WIDTH = 56;
const MAX_COL_WIDTH = 164;

export const DEFAULT_DIMENSION_COLUMN_WIDTH = 164;
export const MIN_DIMENSION_COLUMN_WIDTH = 120;
export const MAX_DIMENSION_COLUMN_WIDTH = 480;

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

// Width of the dimension column in explore leaderboards. Shared by all
// leaderboards so resizing one keeps them symmetric; persisted so the
// width is remembered across sessions.
export const dimensionColumn = localStorageStore<number>(
  "leaderboard-dimension-column-width",
  DEFAULT_DIMENSION_COLUMN_WIDTH,
);
