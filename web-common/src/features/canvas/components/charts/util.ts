import BarChart from "@rilldata/web-common/components/icons/BarChart.svelte";
import LineChart from "@rilldata/web-common/components/icons/LineChart.svelte";
import StackedBar from "@rilldata/web-common/components/icons/StackedBar.svelte";
import { generateVLBarChartSpec } from "./bar-chart/spec";
import { generateVLLineChartSpec } from "./line-chart/spec";
import { generateVLStackedBarChartSpec } from "./stacked-bar/spec";

import type { ChartDataResult } from "./selector";
import type { ChartConfig, ChartMetadata, ChartType } from "./types";

export function generateSpec(
  chartType: ChartType,
  chartConfig: ChartConfig,
  data: ChartDataResult,
) {
  if (data.isFetching || data.error) return {};
  switch (chartType) {
    case "bar_chart":
      return generateVLBarChartSpec(chartConfig, data);
    case "stacked_bar":
      return generateVLStackedBarChartSpec(chartConfig, data);
    case "line_chart":
      return generateVLLineChartSpec(chartConfig, data);
    default:
      return generateVLBarChartSpec(chartConfig, data);
  }
}

export const chartMetadata: ChartMetadata[] = [
  { id: "line_chart", title: "Line", icon: LineChart },
  { id: "bar_chart", title: "Bar", icon: BarChart },
  { id: "stacked_bar", title: "Stacked Bar", icon: StackedBar },
];
