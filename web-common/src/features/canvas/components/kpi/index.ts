import type {
  ComponentCommonProperties,
  ComponentComparisonOptions,
  ComponentFilterProperties,
} from "../types";

export { default as KPI } from "./KPI.svelte";

export const SPARKLINE_MIN_WIDTH = 128;
export const BIG_NUMBER_MIN_WIDTH = 160;
export const padding = 32;
export const SPARK_RIGHT_MIN =
  SPARKLINE_MIN_WIDTH + 8 + BIG_NUMBER_MIN_WIDTH + padding;

export function getMinWidth(
  sparkline: "none" | "bottom" | "right" | undefined,
): number {
  switch (sparkline) {
    case "right":
      return SPARK_RIGHT_MIN;
    case "none":
    case "bottom":
    default:
      return BIG_NUMBER_MIN_WIDTH + padding;
  }
}

/**
 * Calculates the optimal number of columns for a grid layout to avoid
 * unnecessary whitespace. For example, with 6 items, returns 3 or 2
 * (for a 3x2 or 2x3 layout) rather than 4 (which would leave 2 empty cells).
 *
 * @param itemCount - Number of items to display in the grid
 * @param containerWidth - Available container width in pixels
 * @param minItemWidth - Minimum width per item in pixels
 * @returns Optimal number of columns
 */
export function getOptimalColumns(
  itemCount: number,
  containerWidth: number,
  minItemWidth: number,
): number {
  if (itemCount <= 0 || containerWidth <= 0 || minItemWidth <= 0) {
    return 1;
  }

  // Calculate maximum columns that fit in the container
  const maxColumns = Math.max(1, Math.floor(containerWidth / minItemWidth));

  // If only one column fits or we have one item, return 1
  if (maxColumns === 1 || itemCount === 1) {
    return 1;
  }

  // Find the optimal number of columns that minimizes whitespace
  // by finding the largest factor of itemCount that fits
  let bestColumns = 1;

  for (let cols = Math.min(maxColumns, itemCount); cols >= 1; cols--) {
    if (itemCount % cols === 0) {
      // Perfect fit - no empty cells
      bestColumns = cols;
      break;
    }
  }

  // If no perfect factor found (prime numbers), find the column count
  // that minimizes empty cells in the last row
  if (itemCount % bestColumns !== 0) {
    let minEmptyCells = itemCount;
    for (let cols = Math.min(maxColumns, itemCount); cols >= 1; cols--) {
      const rows = Math.ceil(itemCount / cols);
      const totalCells = rows * cols;
      const emptyCells = totalCells - itemCount;

      if (emptyCells < minEmptyCells) {
        minEmptyCells = emptyCells;
        bestColumns = cols;
      }
    }
  }

  return bestColumns;
}

export interface KPISpec
  extends ComponentCommonProperties,
    ComponentFilterProperties {
  metrics_view: string;
  measure: string;
  // Defaults to "bottom"
  sparkline?: "none" | "bottom" | "right";
  // Defaults to "delta" and "percent_change"
  comparison?: ComponentComparisonOptions[];
  hide_time_range?: boolean;
}
