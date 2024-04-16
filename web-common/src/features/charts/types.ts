import { TDDChart } from "@rilldata/web-common/features/dashboards/time-dimension-details/types";

/**
 * Type definitions for common chart types supported by Vega
 */
export enum ChartType {
  BAR = "bar",
  STACKED_BAR = "stacked_bar",
  GROUPED_BAR = "grouped_bar",
  AREA = "area",
  STACKED_AREA = "stacked_area",
  SCATTER = "scatter",
  LINE = "line",
  PIE = "pie",
  DONUT = "donut",
  HEATMAP = "heatmap",
}

export const TDDChartMap = {
  [TDDChart.GROUPED_BAR]: ChartType.GROUPED_BAR,
  [TDDChart.STACKED_BAR]: ChartType.STACKED_BAR,
  [TDDChart.STACKED_AREA]: ChartType.STACKED_AREA,
};
