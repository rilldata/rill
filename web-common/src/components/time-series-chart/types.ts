/**
 * A data point for time series chart rendering.
 * Uses index-based positioning for the x-axis, with originalDate preserved for display.
 */
export type ChartDataPoint = {
  /** Index for x-axis positioning (0, 1, 2, ...) */
  index: number;
  /** Original date for display purposes (tooltips, range labels) */
  originalDate: Date;
  value: number | null | undefined;
};
