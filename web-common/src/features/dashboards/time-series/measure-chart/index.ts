// Main component
export { default as MeasureChart } from "./MeasureChart.svelte";

// Sub-components
export { default as MeasureChartTooltip } from "./MeasureChartTooltip.svelte";
export { default as MeasureChartScrub } from "./MeasureChartScrub.svelte";

// Types
export type {
  TimeSeriesPoint,
  DimensionSeriesData,
  ChartConfig,
  ChartScales,
  ChartSeries,
  ChartMode,
  ScrubState,
  HoverState,
  BisectedPoint,
  InteractionState,
  MeasureChartProps,
} from "./types";

// Utilities
export {
  computeChartConfig,
  computeNiceYExtent,
  computeYExtent,
  computeXExtent,
  createScales,
  createXScale,
  createYScale,
} from "./scales";

export {
  createChartInteractions,
  createVisibilityObserver,
  getOrderedDates,
} from "./interactions";

// Data fetching hooks
export {
  useMeasureTimeSeries,
  useMeasureTimeSeriesData,
  transformTimeSeriesData,
} from "./use-measure-time-series";

export {
  useMeasureTotals,
  useMeasureTotalsData,
  computeComparisonMetrics,
} from "./use-measure-totals";
