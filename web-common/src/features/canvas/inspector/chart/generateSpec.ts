import type { ChartTypeConfig } from "@rilldata/web-common/features/dashboards/canvas/types";
import { generateVLBarChartSpec } from "./specs/bar";
import { generateVLLineChartSpec } from "./specs/line";
import { generateVLStackedBarChartSpec } from "./specs/stacked-bar";

export function generateSpec(chartConfig: ChartTypeConfig) {
  const chartType = chartConfig.chartType;

  switch (chartType) {
    case "bar":
      return generateVLBarChartSpec(chartConfig);
    case "stacked-bar":
      return generateVLStackedBarChartSpec(chartConfig);
    case "line":
      return generateVLLineChartSpec(chartConfig);
    default:
      return generateVLBarChartSpec(chartConfig);
  }
}
