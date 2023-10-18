import { writable } from "svelte/store";

export type TimeDimensionDetailsStore = {
  highlightedCol: number | null;
  highlightedRow: number | null;
  scrubbedCols: [number, number] | null;
  visibleDimensions: string[];
  // TODO: replace with dashboard store
  filterMode: "include" | "exclude";
  filteredValues: string[];
};

export function createTimeDimensionDetailsStore() {
  // Store of state to share between line chart and table
  const store = writable<TimeDimensionDetailsStore>({
    highlightedCol: null,
    highlightedRow: null,
    scrubbedCols: [8, 12],
    visibleDimensions: ["Value A1"],
    // TODO: this will be replaced with dashboard store
    filterMode: "include",
    filteredValues: ["Value A1"],
  });
  return store;
}

const VISIBLE_LIMIT = 8;
export function toggleVisibleDimensions(
  state: TimeDimensionDetailsStore,
  text: string
) {
  if (state.visibleDimensions.includes(text)) {
    state.visibleDimensions = state.visibleDimensions.filter((d) => d !== text);
  } else {
    // FIFO up to visible limit
    state.visibleDimensions.push(text);
    while (state.visibleDimensions.length > VISIBLE_LIMIT) {
      state.visibleDimensions.shift();
    }
  }
}

const VISIBLE_COLORS = [
  "pink-600",
  "cyan-600",
  "red-600",
  "blue-600",
  "green-600",
  "orange-600",
  "purple-600",
];
export function getVisibleDimensionColor(
  state: TimeDimensionDetailsStore,
  text: string
) {
  return VISIBLE_COLORS[state.visibleDimensions.indexOf(text)] ?? "gray-900";
}

// TODO: this will be replaced with dashboard store
export function toggleFilter(state: TimeDimensionDetailsStore, value: string) {
  if (state.filteredValues.includes(value)) {
    state.filteredValues = state.filteredValues.filter((d) => d !== value);
  } else {
    state.filteredValues.push(value);
  }
}
