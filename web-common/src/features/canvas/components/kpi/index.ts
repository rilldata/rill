import type { EphemeralMeasureSpec } from "@rilldata/web-common/features/dashboards/ephemeral-measures/canvas";
import type {
  ComponentCommonProperties,
  ComponentComparisonOptions,
  ComponentFilterProperties,
} from "../types";

export { default as KPI } from "./KPI.svelte";

export const KPI_SIZES = ["small", "medium", "large", "xlarge"] as const;
export type KPISize = (typeof KPI_SIZES)[number];
export const DEFAULT_KPI_SIZE: KPISize = "medium";

/**
 * The size preset a KPI renders at. An absent or unknown value (the runtime
 * rejects unknown ones, but an unvalidated draft can still carry one) renders
 * at the default size.
 */
export function resolveKPISize(size: unknown): KPISize {
  return KPI_SIZES.includes(size as KPISize)
    ? (size as KPISize)
    : DEFAULT_KPI_SIZE;
}

/**
 * Font size of the value in pixels per size preset.
 * KPI.svelte sets the same sizes in rem (`--kpi-value-size`).
 */
export const KPI_VALUE_FONT_SIZES: Record<KPISize, number> = {
  small: 20,
  medium: 30,
  large: 44,
  xlarge: 64,
};

/**
 * Minimum width in pixels of the column holding a medium value
 * and all three comparison values;
 * each comparison value left out takes 10px off.
 */
const MEDIUM_BIG_NUMBER_MIN_WIDTH = 160;

/**
 * Sparkline minimum width, gap between the value column and a right-hand sparkline,
 * and card padding per size preset.
 * They match the spacing KPI.svelte and KPIGrid.svelte apply for the same preset.
 */
const KPI_SIZE_SPACING: Record<
  KPISize,
  { sparkline: number; gap: number; padding: number }
> = {
  small: { sparkline: 96, gap: 4, padding: 24 },
  medium: { sparkline: 128, gap: 8, padding: 32 },
  large: { sparkline: 160, gap: 12, padding: 40 },
  xlarge: { sparkline: 160, gap: 16, padding: 48 },
};

/**
 * Minimum width in pixels of the column holding the value and its comparison values.
 * It grows in proportion to the value's font size,
 * so a value that fits the medium column also fits the column of a larger preset.
 */
export function getBigNumberMinWidth(
  size?: KPISize,
  comparisonCount = 3,
): number {
  const mediumWidth = MEDIUM_BIG_NUMBER_MIN_WIDTH - 10 * (3 - comparisonCount);
  return Math.ceil(
    (mediumWidth * KPI_VALUE_FONT_SIZES[size ?? DEFAULT_KPI_SIZE]) /
      KPI_VALUE_FONT_SIZES.medium,
  );
}

/**
 * Minimum width in pixels of a KPI card, used to pick the grid's column count.
 */
export function getMinWidth(
  sparkline: "none" | "bottom" | "right" | undefined,
  size?: KPISize,
): number {
  const bigNumber = getBigNumberMinWidth(size);
  const spacing = KPI_SIZE_SPACING[size ?? DEFAULT_KPI_SIZE];
  switch (sparkline) {
    case "right":
      return spacing.sparkline + spacing.gap + bigNumber + spacing.padding;
    case "none":
    case "bottom":
    default:
      return bigNumber + spacing.padding;
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
  // by finding the largest factor of itemCount >= 2 that fits
  let bestColumns = -1;

  for (let cols = Math.min(maxColumns, itemCount); cols >= 2; cols--) {
    if (itemCount % cols === 0) {
      // Perfect fit with multiple columns - no empty cells
      bestColumns = cols;
      break;
    }
  }

  // If a perfect factor >= 2 was found, use it
  if (bestColumns !== -1) {
    return bestColumns;
  }

  // No perfect factor >= 2 found (prime numbers or only factor is 1)
  // Find the column count that minimizes empty cells while preferring more columns
  // Use maxColumns as starting point since having more columns is generally better
  bestColumns = maxColumns;
  let minEmptyCells = Infinity;

  for (let cols = Math.min(maxColumns, itemCount); cols >= 2; cols--) {
    const rows = Math.ceil(itemCount / cols);
    const totalCells = rows * cols;
    const emptyCells = totalCells - itemCount;

    if (emptyCells < minEmptyCells) {
      minEmptyCells = emptyCells;
      bestColumns = cols;
    }
  }

  return bestColumns;
}

export interface KPISpec
  extends ComponentCommonProperties,
    ComponentFilterProperties {
  metrics_view: string;
  measure: string;
  // Ad-hoc measures derived from existing measures via an arithmetic
  // expression; `measure` may name one of them.
  adhoc_measures?: EphemeralMeasureSpec[];
  // Defaults to "bottom"
  sparkline?: "none" | "bottom" | "right";
  // Scales the card's text, spacing and minimum footprint. Defaults to "medium"
  size?: KPISize;
  // Defaults to "delta" and "percent_change"
  comparison?: ComponentComparisonOptions[];
  hide_time_range?: boolean;
}
