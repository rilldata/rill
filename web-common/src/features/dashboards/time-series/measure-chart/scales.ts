import { scaleLinear } from "d3-scale";
import { max, min } from "d3-array";
import type {
  TimeSeriesPoint,
  DimensionSeriesData,
  ChartConfig,
  ChartScales,
  PlotBounds,
} from "./types";

interface ExtentConfig {
  includeZero: boolean;
  paddingFactor: number;
  minRange?: number;
}

/**
 * Default extent configuration.
 */
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
 * Compute X extent as index range [0, N-1].
 */
export function computeXExtent(data: TimeSeriesPoint[]): [number, number] {
  return [0, Math.max(0, data.length - 1)];
}

/**
 * Create index-based X scale.
 */
export function createXScale(
  data: TimeSeriesPoint[],
  plotBounds: PlotBounds,
): ChartScales["x"] {
  return scaleLinear<number>()
    .domain(computeXExtent(data))
    .range([plotBounds.left, plotBounds.left + plotBounds.width]);
}

/**
 * Create Y scale from value extent.
 */
export function createYScale(
  yExtent: [number, number],
  plotBounds: PlotBounds,
): ChartScales["y"] {
  return scaleLinear<number>()
    .domain(yExtent)
    .range([plotBounds.top + plotBounds.height, plotBounds.top]); // Inverted for SVG
}

/**
 * Create both X and Y scales.
 */
export function createScales(
  data: TimeSeriesPoint[],
  dimensionData: DimensionSeriesData[],
  showComparison: boolean,
  plotBounds: PlotBounds,
): ChartScales {
  const yRawExtent = computeYExtent(data, dimensionData, showComparison);
  const yExtent = computeNiceYExtent(yRawExtent[0], yRawExtent[1]);

  return {
    x: createXScale(data, plotBounds),
    y: createYScale(yExtent, plotBounds),
  };
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
    right: 40,
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

/**
 * Update scales with new data while preserving animation continuity.
 */
export function updateScalesWithData(
  existingScales: ChartScales | null,
  data: TimeSeriesPoint[],
  dimensionData: DimensionSeriesData[],
  showComparison: boolean,
  plotBounds: PlotBounds,
): ChartScales {
  // Always recompute scales - animation is handled by tweening the domain values
  return createScales(data, dimensionData, showComparison, plotBounds);
}
