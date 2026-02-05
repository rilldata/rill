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
import {
  ScatterPlotChartComponent,
  type ScatterPlotCanvasChartSpec,
} from "./variants/ScatterPlotChart";

export { default as Chart } from "./CanvasChart.svelte";

export type ChartComponent =
  | typeof CartesianChartComponent
  | typeof CircularChartComponent
  | typeof FunnelChartComponent
  | typeof HeatmapChartComponent
  | typeof ComboChartComponent
  | typeof ScatterPlotChartComponent;

export type CanvasChartSpec =
  | CartesianCanvasChartSpec
  | CircularCanvasChartSpec
  | FunnelCanvasChartSpec
  | HeatmapCanvasChartSpec
  | ComboCanvasChartSpec
  | ScatterPlotCanvasChartSpec;

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
    case "scatter_plot":
      return ScatterPlotChartComponent;
    default:
      throw new Error("Unsupported chart type: " + type);
  }
}

<<<<<<< HEAD
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
    title: "Area",
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
=======
export type CanvasChartConfig = ChartMetadataConfig & {
  component: BaseCanvasComponentConstructor<CanvasChartSpec>;
>>>>>>> main
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
