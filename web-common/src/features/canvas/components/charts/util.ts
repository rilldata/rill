import { generateVLBarChartSpec } from "./bar-chart/spec";
import { generateVLLineChartSpec } from "./line-chart/spec";
import { generateVLStackedBarChartSpec } from "./stacked-bar/spec";
import type { ChartConfig, ChartType } from "./types";

export function generateSpec(chartType: ChartType, chartConfig: ChartConfig) {
  switch (chartType) {
    case "bar_chart":
      return generateVLBarChartSpec(chartConfig);
    case "stacked_bar":
      return generateVLStackedBarChartSpec(chartConfig);
    case "line_chart":
      return generateVLLineChartSpec(chartConfig);
    default:
      return generateVLBarChartSpec(chartConfig);
  }
}
