import { min, max } from "d3-array";
import type {
  TimeSeriesPoint,
  DimensionSeriesData,
  ChartConfig,
} from "./types";

interface ExtentConfig {
  includeZero: boolean;
  paddingFactor: number;
  minRange?: number;
}

/**
 * Default extent configuration.
 */
export const LINE_MODE_MIN_POINTS = 6;
export const X_PAD = 8;
export const MARGIN_RIGHT = 40;

const DEFAULT_EXTENT_CONFIG: ExtentConfig = {
  includeZero: true,
  paddingFactor: 1.3,
};

/**
 * Compute nice Y extent with padding.
 * Sets extents to 0 if it makes sense; otherwise, inflates each extent component.
 * Ported from utils.ts niceMeasureExtents.
 */
export function computeNiceYExtent(
  smallest: number,
  largest: number,
  config: ExtentConfig = DEFAULT_EXTENT_CONFIG,
): [number, number] {
  const { includeZero, paddingFactor, minRange } = config;

  // Handle edge case where both are 0
  if (smallest === 0 && largest === 0) {
    return [0, 1];
  }

  // Handle NaN or invalid values
  if (!Number.isFinite(smallest) || !Number.isFinite(largest)) {
    return [0, 1];
  }

  let yMin: number;
  let yMax: number;

  if (includeZero) {
    // Include zero in the extent when appropriate
    yMin = smallest < 0 ? smallest * paddingFactor : 0;
    yMax = largest > 0 ? largest * paddingFactor : 0;
  } else {
    yMin = smallest * paddingFactor;
    yMax = largest * paddingFactor;
  }

  // Ensure minimum range if specified
  if (minRange !== undefined && yMax - yMin < minRange) {
    const mid = (yMin + yMax) / 2;
    yMin = mid - minRange / 2;
    yMax = mid + minRange / 2;
  }

  return [yMin, yMax];
}

/**
 * Compute combined Y extent from all data sources.
 * Includes main data, comparison data, and dimension data.
 */
export function computeYExtent(
  data: TimeSeriesPoint[],
  dimensionData: DimensionSeriesData[],
  showComparison: boolean,
): [number, number] {
  const values: number[] = [];
  const hasDimensionData = dimensionData.length > 0;

  // Main data values — skip when dimension comparison is active,
  // since only individual dimension series are rendered (not the aggregate).
  if (!hasDimensionData) {
    for (const d of data) {
      if (d.value !== null && Number.isFinite(d.value)) {
        values.push(d.value);
      }
      if (
        showComparison &&
        d.comparisonValue !== null &&
        d.comparisonValue !== undefined &&
        Number.isFinite(d.comparisonValue)
      ) {
        values.push(d.comparisonValue);
      }
    }
  }

  // Dimension data values
  for (const dim of dimensionData) {
    for (const d of dim.data) {
      if (d.value !== null && Number.isFinite(d.value)) {
        values.push(d.value);
      }
    }
  }

  if (values.length === 0) {
    return [0, 1];
  }

  return [min(values) ?? 0, max(values) ?? 1];
}

/**
 * Compute chart configuration from dimensions.
 */
export function computeChartConfig(
  width: number,
  height: number,
  isExpanded: boolean,
): ChartConfig {
  const margin = {
    top: 4, // Space for data readout labels
    right: MARGIN_RIGHT,
    bottom: isExpanded ? 25 : 10,
    left: 0,
  };

  const plotWidth = Math.max(0, width - margin.left - margin.right);
  const plotHeight = Math.max(0, height - margin.top - margin.bottom);

  return {
    width,
    height,
    margin,
    plotBounds: {
      left: margin.left,
      right: margin.left + plotWidth,
      top: margin.top,
      bottom: margin.top + plotHeight,
      width: plotWidth,
      height: plotHeight,
    },
  };
}
