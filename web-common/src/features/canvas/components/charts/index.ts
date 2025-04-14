import type { CartesianChartSpec } from "./cartesian-charts/CartesianChart";
import { CartesianChartComponent } from "./cartesian-charts/CartesianChart";
import type { CircularChartSpec } from "./circular-charts/CircularChart";

export { default as Chart } from "./Chart.svelte";

export type ChartComponent = typeof CartesianChartComponent;

export type ChartSpec = CartesianChartSpec | CircularChartSpec;

export type ChartType =
  | "bar_chart"
  | "line_chart"
  | "area_chart"
  | "stacked_bar"
  | "stacked_bar_normalized";

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
