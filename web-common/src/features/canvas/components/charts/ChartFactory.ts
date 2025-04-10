import { CartesianChartComponent } from "./CartesianChart";
import type { ChartType } from "./types";

type ChartComponent = typeof CartesianChartComponent;

export function createChart(type: ChartType): ChartComponent {
  switch (type) {
    case "bar_chart":
    case "line_chart":
    case "area_chart":
    case "stacked_bar":
    case "stacked_bar_normalized":
      return CartesianChartComponent;

    // Add other chart types here as they are implemented
    default:
      throw new Error(`Unsupported chart type: ${type}`);
  }
}
