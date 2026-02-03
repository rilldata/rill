import type { ScaleLinear } from "d3-scale";
import type { DateTime } from "luxon";

/**
 * Strongly-typed time series data point.
 * Replaces the string-based accessor pattern from TimeSeriesDatum.
 */
export interface TimeSeriesPoint {
  /** Primary timestamp for the data point */
  ts: DateTime;
  /** The measure value (nullable for gaps) */
  value: number | null;
  /** Comparison value when time comparison is active */
  comparisonValue?: number | null;
  /** Comparison timestamp */
  comparisonTs?: DateTime;
}

/**
 * Dimension comparison data item.
 */
export interface DimensionSeriesData {
  /** Dimension value (e.g., "USA", "Canada") */
  dimensionValue: string | null;
  /** Color for this dimension series */
  color: string;
  /** Time series data for this dimension */
  data: TimeSeriesPoint[];
  /** Loading state */
  isFetching: boolean;
  /** Total value for percent calculations */
  total?: number;
}

/**
 * Chart margin configuration.
 */
export interface ChartMargin {
  top: number;
  right: number;
  bottom: number;
  left: number;
}

/**
 * Computed plot bounds within the SVG.
 */
export interface PlotBounds {
  left: number;
  right: number;
  top: number;
  bottom: number;
  width: number;
  height: number;
}

/**
 * Chart configuration (replaces GraphicContext props).
 */
export interface ChartConfig {
  width: number;
  height: number;
  margin: ChartMargin;
  plotBounds: PlotBounds;
}

/**
 * Scale types for the chart.
 */
export interface ChartScales {
  x: ScaleLinear<number, number>;
  y: ScaleLinear<number, number>;
}

/**
 * Scrub/selection state.
 */
export interface ScrubState {
  /** Start index (fractional, from xScale.invert) */
  startIndex: number | null;
  /** End index (fractional, from xScale.invert) */
  endIndex: number | null;
  isScrubbing: boolean;
}

/**
 * Mouseover/hover state.
 */
export interface HoverState {
  /** Hovered index (fractional) */
  index: number | null;
  /** Screen x coordinate */
  screenX: number | null;
  /** Screen y coordinate */
  screenY: number | null;
  /** Is mouse currently over the chart */
  isHovered: boolean;
}

/**
 * A generic series descriptor for the pure TimeSeriesChart renderer.
 * Decoupled from measure/dimension semantics.
 */
export interface ChartSeries {
  /** Unique identifier for this series */
  id: string;
  /** Values array — one per bucket, null for gaps */
  values: (number | null)[];
  /** Stroke/fill color */
  color: string;
  /** Dash pattern for the stroke */
  strokeDasharray?: string;
  /** Opacity override (default 1) */
  opacity?: number;
  /** Area gradient colors — only the first/primary series typically gets this */
  areaGradient?: { dark: string; light: string };
  /** Stroke width */
  strokeWidth?: number;
}

/**
 * Rendering mode for TimeSeriesChart.
 */
export type ChartMode = "line" | "bar";
