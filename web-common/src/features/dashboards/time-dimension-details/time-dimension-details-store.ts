import { writable } from "svelte/store";

export type TimeDimensionDetailsStore = {
  highlightedCol: number | null;
  highlightedRow: number | null;
  scrubbedCols: [number, number] | null;
};

export function createTimeDimensionDetailsStore() {
  // Store of state to share between line chart and table
  const store = writable<TimeDimensionDetailsStore>({
    highlightedCol: null,
    highlightedRow: null,
    scrubbedCols: [8, 12],
  });
  return store;
}
