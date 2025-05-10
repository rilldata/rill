import BarChart from "@rilldata/web-common/components/icons/BarChart.svelte";
import Donut from "@rilldata/web-common/components/icons/Donut.svelte";
import Heatmap from "@rilldata/web-common/components/icons/Heatmap.svelte";
import LineChart from "@rilldata/web-common/components/icons/LineChart.svelte";
import StackedArea from "@rilldata/web-common/components/icons/StackedArea.svelte";
import StackedBar from "@rilldata/web-common/components/icons/StackedBar.svelte";
import StackedBarFull from "@rilldata/web-common/components/icons/StackedBarFull.svelte";
import type { BaseCanvasComponentConstructor } from "@rilldata/web-common/features/canvas/components/util";
import type { ComponentType, SvelteComponent } from "svelte";
import type { VisualizationSpec } from "svelte-vega";
import { generateVLAreaChartSpec } from "./cartesian-charts/area/spec";
import { generateVLBarChartSpec } from "./cartesian-charts/bar-chart/spec";
import type { CartesianChartSpec } from "./cartesian-charts/CartesianChart";
import { CartesianChartComponent } from "./cartesian-charts/CartesianChart";
import { generateVLLineChartSpec } from "./cartesian-charts/line-chart/spec";
import { generateVLStackedBarChartSpec } from "./cartesian-charts/stacked-bar/default";
import { generateVLStackedBarNormalizedSpec } from "./cartesian-charts/stacked-bar/normalized";
import {
  CircularChartComponent,
  type CircularChartSpec,
} from "./circular-charts/CircularChart";
import { generateVLPieChartSpec } from "./circular-charts/pie";
import {
  HeatmapChartComponent,
  type HeatmapChartSpec,
} from "./heatmap-charts/HeatmapChart";
import { generateVLHeatmapSpec } from "./heatmap-charts/spec";
import type { ChartDataResult, ChartType } from "./types";

export { default as Chart } from "./Chart.svelte";

export type ChartComponent =
  | typeof CartesianChartComponent
  | typeof CircularChartComponent
  | typeof HeatmapChartComponent;

export type ChartSpec =
  | CartesianChartSpec
  | CircularChartSpec
  | HeatmapChartSpec;

export function getChartComponent(
  type: ChartType,
): BaseCanvasComponentConstructor {
  switch (type) {
    case "bar_chart":
    case "line_chart":
    case "area_chart":
    case "stacked_bar":
    case "stacked_bar_normalized":
      return CartesianChartComponent;
    case "donut_chart":
      return CircularChartComponent;
    case "heatmap":
      return HeatmapChartComponent;
    default:
      throw new Error(`Unsupported chart type: ${type}`);
  }
}

export interface ChartMetadataConfig {
  title: string;
  icon: ComponentType<SvelteComponent>;
  component: BaseCanvasComponentConstructor;
  generateSpec: (config: ChartSpec, data: ChartDataResult) => VisualizationSpec;
}

export const CHART_CONFIG: Record<ChartType, ChartMetadataConfig> = {
  bar_chart: {
    title: "Bar",
    icon: BarChart,
    component: CartesianChartComponent,
    generateSpec: generateVLBarChartSpec,
  },
  line_chart: {
    title: "Line",
    icon: LineChart,
    component: CartesianChartComponent,
    generateSpec: generateVLLineChartSpec,
  },
  area_chart: {
    title: "Stacked Area",
    icon: StackedArea,
    component: CartesianChartComponent,
    generateSpec: generateVLAreaChartSpec,
  },
  stacked_bar: {
    title: "Stacked Bar",
    icon: StackedBar,
    component: CartesianChartComponent,
    generateSpec: generateVLStackedBarChartSpec,
  },
  stacked_bar_normalized: {
    title: "Stacked Bar Normalized",
    icon: StackedBarFull,
    component: CartesianChartComponent,
    generateSpec: generateVLStackedBarNormalizedSpec,
  },
  donut_chart: {
    title: "Donut",
    icon: Donut,
    component: CircularChartComponent,
    generateSpec: generateVLPieChartSpec,
  },
  heatmap: {
    title: "Heatmap",
    icon: Heatmap,
    component: HeatmapChartComponent,
    generateSpec: generateVLHeatmapSpec,
  },
};

export const CHART_TYPES = Object.keys(CHART_CONFIG) as ChartType[];
