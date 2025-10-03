import type { BaseCanvasComponentConstructor } from "@rilldata/web-common/features/canvas/components/util";
import {
  CHART_CONFIG,
  type ChartMetadataConfig,
} from "@rilldata/web-common/features/components/charts/config";
import type { ChartType } from "@rilldata/web-common/features/components/charts/types";
import {
  CartesianChartComponent,
  type CartesianCanvasChartSpec,
} from "./variants/CartesianChart";
import {
  CircularChartComponent,
  type CircularCanvasChartSpec,
} from "./variants/CircularChart";
import {
  ComboChartComponent,
  type ComboCanvasChartSpec,
} from "./variants/ComboChart";
import {
  FunnelChartComponent,
  type FunnelCanvasChartSpec,
} from "./variants/FunnelChart";
import {
  HeatmapChartComponent,
  type HeatmapCanvasChartSpec,
} from "./variants/HeatmapChart";

export { default as Chart } from "./CanvasChart.svelte";

export type ChartComponent =
  | typeof CartesianChartComponent
  | typeof CircularChartComponent
  | typeof FunnelChartComponent
  | typeof HeatmapChartComponent
  | typeof ComboChartComponent;

export type CanvasChartSpec =
  | CartesianCanvasChartSpec
  | CircularCanvasChartSpec
  | FunnelCanvasChartSpec
  | HeatmapCanvasChartSpec
  | ComboCanvasChartSpec;

export function getCanvasChartComponent(
  type: ChartType,
): BaseCanvasComponentConstructor<CanvasChartSpec> {
  switch (type) {
    case "bar_chart":
    case "line_chart":
    case "area_chart":
    case "stacked_bar":
    case "stacked_bar_normalized":
      return CartesianChartComponent;
    case "donut_chart":
    case "pie_chart":
      return CircularChartComponent;
    case "funnel_chart":
      return FunnelChartComponent;
    case "heatmap":
      return HeatmapChartComponent;
    case "combo_chart":
      return ComboChartComponent;
    default:
      throw new Error("Unsupported chart type: " + type);
  }
}

export type CanvasChartConfig = ChartMetadataConfig & {
  component: BaseCanvasComponentConstructor<CanvasChartSpec>;
};

export const CANVAS_CHART_CONFIG: Record<ChartType, CanvasChartConfig> =
  Object.fromEntries(
    Object.entries(CHART_CONFIG).map(([type, config]) => [
      type,
      {
        ...config,
        component: getCanvasChartComponent(type as ChartType),
      },
    ]),
  ) as Record<ChartType, CanvasChartConfig>;
