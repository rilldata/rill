import { CartesianChartComponent } from "./CartesianChart";
import type { ChartType } from "./types";

export { default as Chart } from "./Chart.svelte";

export type ChartComponent = typeof CartesianChartComponent;

export function getChartComponent(type: ChartType): ChartComponent {
  switch (type) {
    case "bar_chart":
    case "line_chart":
    case "area_chart":
    case "stacked_bar":
    case "stacked_bar_normalized":
      return CartesianChartComponent;

    default:
      throw new Error(`Unsupported chart type: ${type}`);
  }
}
