import { clamp } from "@rilldata/web-common/lib/clamp";
import { writable } from "svelte/store";
import type { LeaderboardItemData } from "./leaderboard-utils";

const MIN_COL_WIDTH = 56;
const DEFAULT_COL_WIDTH = 60;
const MAX_COL_WIDTH = 164;

export type ColumnWidths = {
  dimension: number;
  value: number;
  percentOfTotal: number;
  delta: number;
  deltaPercent: number;
};

export const LEADERBOARD_DEFAULT_COLUMN_WIDTHS: ColumnWidths = {
  dimension: 164,
  value: DEFAULT_COL_WIDTH,
  percentOfTotal: DEFAULT_COL_WIDTH,
  delta: DEFAULT_COL_WIDTH,
  deltaPercent: DEFAULT_COL_WIDTH,
};

// Create a store for the maximum column widths
export const columnWidths = writable<ColumnWidths>(
  LEADERBOARD_DEFAULT_COLUMN_WIDTHS,
);
export const processedDimensions = new Set<string>();

export function updateMaxColumnWidths(
  dimensionName: string,
  newWidths: ColumnWidths,
) {
  if (!processedDimensions.has(dimensionName) && processedDimensions.size < 4) {
    columnWidths.update((currentWidths) => ({
      dimension: Math.max(currentWidths.dimension, newWidths.dimension),
      value: Math.max(currentWidths.value, newWidths.value),
      percentOfTotal: Math.max(
        currentWidths.percentOfTotal,
        newWidths.percentOfTotal,
      ),
      delta: Math.max(currentWidths.delta, newWidths.delta),
      deltaPercent: Math.max(
        currentWidths.deltaPercent,
        newWidths.deltaPercent,
      ),
    }));
    processedDimensions.add(dimensionName);
  }
}

function estimateColumnWidth(values: unknown[]) {
  const samples = values.filter(
    (v): v is string | number => typeof v === "string" || typeof v === "number",
  );

  const maxValueLength = samples.reduce((max: number, value) => {
    const stringLength = String(value).length;
    return Math.max(max, stringLength);
  }, 0) as number;

  const pixelLength = maxValueLength * 7;
  return clamp(MIN_COL_WIDTH, pixelLength + 16, MAX_COL_WIDTH);
}

export function calculateLeaderboardColumnWidth(
  firstColumnWidth: number,
  aboveTheFold: LeaderboardItemData[],
  selectedBelowTheFold: LeaderboardItemData[],
  formatter: (
    value: string | number | null | undefined,
  ) => string | null | undefined,
): ColumnWidths {
  const rows = aboveTheFold.concat(selectedBelowTheFold);

  const columnWidths = {
    dimension: firstColumnWidth,
    value: estimateColumnWidth(rows.map((i) => formatter(i.value || 0))),
    delta: estimateColumnWidth(rows.map((i) => formatter(i.deltaAbs || 0))),
    // These are always formatted and can be contained in min width
    percentOfTotal: MIN_COL_WIDTH,
    deltaPercent: MIN_COL_WIDTH,
  };
  return columnWidths;
}

export function resetColumnWidths() {
  processedDimensions.clear();
  columnWidths.set(LEADERBOARD_DEFAULT_COLUMN_WIDTHS);
}
