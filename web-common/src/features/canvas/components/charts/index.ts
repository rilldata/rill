import type { BaseCanvasComponentConstructor } from "@rilldata/web-common/features/canvas/components/util";
import {
  CHART_CONFIG,
  type ChartMetadataConfig,
} from "@rilldata/web-common/features/components/charts/config";
import type { ChartType } from "@rilldata/web-common/features/components/charts/types";
import type { CartesianChartSpec } from "./cartesian-charts/CartesianChart";
import { CartesianChartComponent } from "./cartesian-charts/CartesianChart";
import {
  CircularChartComponent,
  type CircularChartSpec,
} from "./circular-charts/CircularChart";
import {
  ComboChartComponent,
  type ComboChartSpec,
} from "./combo-charts/ComboChart";
import {
  FunnelChartComponent,
  type FunnelChartSpec,
} from "./funnel-charts/FunnelChart";
import {
  HeatmapChartComponent,
  type HeatmapChartSpec,
} from "./heatmap-charts/HeatmapChart";

export { default as Chart } from "./CanvasChart.svelte";

export type ChartComponent =
  | typeof CartesianChartComponent
  | typeof CircularChartComponent
  | typeof FunnelChartComponent
  | typeof HeatmapChartComponent
  | typeof ComboChartComponent;

export type ChartSpec =
  | CartesianChartSpec
  | CircularChartSpec
  | FunnelChartSpec
  | HeatmapChartSpec
  | ComboChartSpec;

export function getChartComponent(
  type: ChartType,
): BaseCanvasComponentConstructor<ChartSpec> {
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
  component: BaseCanvasComponentConstructor<ChartSpec>;
};

export const CANVAS_CHART_CONFIG: Record<ChartType, CanvasChartConfig> =
  Object.fromEntries(
    Object.entries(CHART_CONFIG).map(([type, config]) => [
      type,
      {
        ...config,
        component: getChartComponent(type as ChartType),
      },
    ]),
  ) as Record<ChartType, CanvasChartConfig>;
