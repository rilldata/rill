import { clamp } from "@rilldata/web-common/lib/clamp";
import { localStorageStore } from "@rilldata/web-common/lib/store-utils";
import { get, writable, type Writable } from "svelte/store";

export const DEFAULT_COLUMN_WIDTH = 110;
export const COMPARISON_COLUMN_WIDTH = 64;
export const MEASURES_PADDING = 16;

const MIN_COL_WIDTH = 56;
const MAX_COL_WIDTH = 164;

// A comparison cell loses 8px to its own horizontal padding. The rest is
// breathing room, so the value doesn't butt up against the previous column.
const COMPARISON_PADDING = 16;

export const DEFAULT_DIMENSION_COLUMN_WIDTH = 164;
export const MIN_DIMENSION_COLUMN_WIDTH = 120;
export const MAX_DIMENSION_COLUMN_WIDTH = 480;

class ColumnStore {
  private value: Writable<number>;
  private defaultWidth: number;
  private minWidth: number;
  private padding: number;

  subscribe: (this: void, run: (value: number) => void) => () => void;

  constructor(
    defaultWidth: number = DEFAULT_COLUMN_WIDTH,
    minWidth: number = MIN_COL_WIDTH,
    padding: number = MEASURES_PADDING,
  ) {
    this.defaultWidth = defaultWidth;
    this.minWidth = minWidth;
    this.padding = padding;
    this.value = writable(defaultWidth);
    this.subscribe = this.value.subscribe;
  }

  update = (newValue: number) => {
    newValue = clamp(
      this.minWidth,
      Math.ceil(newValue) + this.padding,
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

// Delta columns hold formatted measure values, which can be far wider than the
// default comparison width, e.g. a currency delta like "-$1,234,567.89".
export const deltaColumn = new ColumnStore(
  COMPARISON_COLUMN_WIDTH,
  COMPARISON_COLUMN_WIDTH,
  COMPARISON_PADDING,
);

// Width of the dimension column in explore leaderboards. Shared by all
// leaderboards so resizing one keeps them symmetric; persisted so the
// width is remembered across sessions.
export const dimensionColumn = localStorageStore<number>(
  "leaderboard-dimension-column-width",
  DEFAULT_DIMENSION_COLUMN_WIDTH,
);
