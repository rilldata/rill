import BarChart from "@rilldata/web-common/components/icons/BarChart.svelte";
import Donut from "@rilldata/web-common/components/icons/Donut.svelte";
import Funnel from "@rilldata/web-common/components/icons/Funnel.svelte";
import Heatmap from "@rilldata/web-common/components/icons/Heatmap.svelte";
import LineChart from "@rilldata/web-common/components/icons/LineChart.svelte";
import MultiChart from "@rilldata/web-common/components/icons/MultiChart.svelte";
import StackedArea from "@rilldata/web-common/components/icons/StackedArea.svelte";
import StackedBar from "@rilldata/web-common/components/icons/StackedBar.svelte";
import StackedBarFull from "@rilldata/web-common/components/icons/StackedBarFull.svelte";
import { isMultiFieldConfig } from "@rilldata/web-common/features/canvas/components/charts/util.ts";
import type { BaseCanvasComponentConstructor } from "@rilldata/web-common/features/canvas/components/util";
import type { ComponentType, SvelteComponent } from "svelte";
import type { VisualizationSpec } from "svelte-vega";
import type {
  ChartDataResult,
  ChartType,
} from "../../../components/charts/types.ts";
import { generateVLAreaChartSpec } from "./cartesian-charts/area/spec";
import { generateVLBarChartSpec } from "./cartesian-charts/bar-chart/spec";
import type { CartesianChartSpec } from "./cartesian-charts/CartesianChart";
import { CartesianChartComponent } from "./cartesian-charts/CartesianChart";
import { generateVLLineChartSpec } from "./cartesian-charts/line-chart/spec";
import { generateVLMultiMetricChartSpec } from "./cartesian-charts/multi-metric-chart.ts";
import { generateVLStackedBarChartSpec } from "./cartesian-charts/stacked-bar/default";
import { generateVLStackedBarNormalizedSpec } from "./cartesian-charts/stacked-bar/normalized";
import {
  CircularChartComponent,
  type CircularChartSpec,
} from "./circular-charts/CircularChart";
import { generateVLPieChartSpec } from "./circular-charts/pie";
import {
  ComboChartComponent,
  type ComboChartSpec,
} from "./combo-charts/ComboChart";
import { generateVLComboChartSpec } from "./combo-charts/spec";
import {
  FunnelChartComponent,
  type FunnelChartSpec,
} from "./funnel-charts/FunnelChart";
import { generateVLFunnelChartSpec } from "./funnel-charts/spec";
import {
  HeatmapChartComponent,
  type HeatmapChartSpec,
} from "./heatmap-charts/HeatmapChart";
import { generateVLHeatmapSpec } from "./heatmap-charts/spec";

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

export interface ChartMetadataConfig {
  title: string;
  icon: ComponentType<SvelteComponent>;
  component: BaseCanvasComponentConstructor<ChartSpec>;
  generateSpec: (config: ChartSpec, data: ChartDataResult) => VisualizationSpec;
  hideFromSelector?: boolean;
}

export const CHART_CONFIG: Record<ChartType, ChartMetadataConfig> = {
  bar_chart: {
    title: "Bar",
    icon: BarChart,
    component: CartesianChartComponent,
    generateSpec: (config: ChartSpec, data: ChartDataResult) => {
      const cartesianConfig = config as CartesianChartSpec;
      const isMultiMeasure = isMultiFieldConfig(cartesianConfig.y);
      return isMultiMeasure
        ? generateVLMultiMetricChartSpec(cartesianConfig, data, "grouped_bar")
        : generateVLBarChartSpec(cartesianConfig, data);
    },
  },
  line_chart: {
    title: "Line",
    icon: LineChart,
    component: CartesianChartComponent,
    generateSpec: (config: ChartSpec, data: ChartDataResult) => {
      const cartesianConfig = config as CartesianChartSpec;
      const isMultiMeasure = isMultiFieldConfig(cartesianConfig.y);
      return isMultiMeasure
        ? generateVLMultiMetricChartSpec(cartesianConfig, data, "line")
        : generateVLLineChartSpec(cartesianConfig, data);
    },
  },
  area_chart: {
    title: "Stacked Area",
    icon: StackedArea,
    component: CartesianChartComponent,
    generateSpec: (config: ChartSpec, data: ChartDataResult) => {
      const cartesianConfig = config as CartesianChartSpec;
      const isMultiMeasure = isMultiFieldConfig(cartesianConfig.y);
      return isMultiMeasure
        ? generateVLMultiMetricChartSpec(cartesianConfig, data, "stacked_area")
        : generateVLAreaChartSpec(cartesianConfig, data);
    },
  },
  stacked_bar: {
    title: "Stacked Bar",
    icon: StackedBar,
    component: CartesianChartComponent,
    generateSpec: (config: ChartSpec, data: ChartDataResult) => {
      const cartesianConfig = config as CartesianChartSpec;
      const isMultiMeasure = isMultiFieldConfig(cartesianConfig.y);
      return isMultiMeasure
        ? generateVLMultiMetricChartSpec(cartesianConfig, data, "stacked_bar")
        : generateVLStackedBarChartSpec(cartesianConfig, data);
    },
  },
  stacked_bar_normalized: {
    title: "Stacked Bar Normalized",
    icon: StackedBarFull,
    component: CartesianChartComponent,
    generateSpec: (config: ChartSpec, data: ChartDataResult) => {
      const cartesianConfig = config as CartesianChartSpec;
      const isMultiMeasure = isMultiFieldConfig(cartesianConfig.y);
      return isMultiMeasure
        ? generateVLMultiMetricChartSpec(
            cartesianConfig,
            data,
            "stacked_bar_normalized",
          )
        : generateVLStackedBarNormalizedSpec(cartesianConfig, data);
    },
  },
  donut_chart: {
    title: "Donut",
    icon: Donut,
    component: CircularChartComponent,
    generateSpec: generateVLPieChartSpec,
  },
  pie_chart: {
    title: "Pie",
    icon: Donut,
    component: CircularChartComponent,
    generateSpec: generateVLPieChartSpec,
    hideFromSelector: true,
  },
  funnel_chart: {
    title: "Funnel",
    icon: Funnel,
    component: FunnelChartComponent,
    generateSpec: generateVLFunnelChartSpec,
  },
  heatmap: {
    title: "Heatmap",
    icon: Heatmap,
    component: HeatmapChartComponent,
    generateSpec: generateVLHeatmapSpec,
  },
  combo_chart: {
    title: "Combo",
    icon: MultiChart,
    component: ComboChartComponent,
    generateSpec: generateVLComboChartSpec,
  },
};

export const CHART_TYPES = Object.keys(CHART_CONFIG) as ChartType[];

export const VISIBLE_CHART_TYPES = CHART_TYPES.filter(
  (type) => !CHART_CONFIG[type].hideFromSelector,
);
