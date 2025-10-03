import BarChart from "@rilldata/web-common/components/icons/BarChart.svelte";
import Donut from "@rilldata/web-common/components/icons/Donut.svelte";
import Funnel from "@rilldata/web-common/components/icons/Funnel.svelte";
import Heatmap from "@rilldata/web-common/components/icons/Heatmap.svelte";
import LineChart from "@rilldata/web-common/components/icons/LineChart.svelte";
import MultiChart from "@rilldata/web-common/components/icons/MultiChart.svelte";
import StackedArea from "@rilldata/web-common/components/icons/StackedArea.svelte";
import StackedBar from "@rilldata/web-common/components/icons/StackedBar.svelte";
import StackedBarFull from "@rilldata/web-common/components/icons/StackedBarFull.svelte";
import type { ComponentType, SvelteComponent } from "svelte";
import type { Readable } from "svelte/store";
import type { VisualizationSpec } from "svelte-vega";
import { generateVLAreaChartSpec } from "./cartesian/area/spec";
import { generateVLBarChartSpec } from "./cartesian/bar-chart/spec";
import {
  CartesianChartProvider,
  type CartesianChartSpec,
} from "./cartesian/CartesianChartProvider";
import { generateVLLineChartSpec } from "./cartesian/line-chart/spec";
import { generateVLMultiMetricChartSpec } from "./cartesian/multi-metric-chart";
import { generateVLStackedBarChartSpec } from "./cartesian/stacked-bar/default";
import { generateVLStackedBarNormalizedSpec } from "./cartesian/stacked-bar/normalized";
import { CircularChartProvider } from "./circular/CircularChartProvider";
import { generateVLPieChartSpec } from "./circular/pie";
import { ComboChartProvider } from "./combo/ComboChartProvider";
import { generateVLComboChartSpec } from "./combo/spec";
import { FunnelChartProvider } from "./funnel/FunnelChartProvider";
import { generateVLFunnelChartSpec } from "./funnel/spec";
import { HeatmapChartProvider } from "./heatmap/HeatmapChartProvider";
import { generateVLHeatmapSpec } from "./heatmap/spec";
import type {
  ChartDataResult,
  ChartProvider,
  ChartSpec,
  ChartType,
} from "./types";
import { isMultiFieldConfig } from "./util.ts";

export interface ChartMetadataConfig {
  title: string;
  icon: ComponentType<SvelteComponent>;
  provider: new (
    spec: Readable<ChartSpec>,
    defaultOptions?: Record<string, unknown>,
  ) => ChartProvider;
  generateSpec: (config: ChartSpec, data: ChartDataResult) => VisualizationSpec;
  hideFromSelector?: boolean;
}

export const CHART_CONFIG: Record<ChartType, ChartMetadataConfig> = {
  bar_chart: {
    title: "Bar",
    icon: BarChart,
    provider: CartesianChartProvider,
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
    provider: CartesianChartProvider,
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
    provider: CartesianChartProvider,
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
    provider: CartesianChartProvider,
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
    provider: CartesianChartProvider,
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
    provider: CircularChartProvider,
    generateSpec: generateVLPieChartSpec,
  },
  pie_chart: {
    title: "Pie",
    icon: Donut,
    provider: CircularChartProvider,
    generateSpec: generateVLPieChartSpec,
    hideFromSelector: true,
  },
  funnel_chart: {
    title: "Funnel",
    icon: Funnel,
    provider: FunnelChartProvider,
    generateSpec: generateVLFunnelChartSpec,
  },
  heatmap: {
    title: "Heatmap",
    icon: Heatmap,
    provider: HeatmapChartProvider,
    generateSpec: generateVLHeatmapSpec,
  },
  combo_chart: {
    title: "Combo",
    icon: MultiChart,
    provider: ComboChartProvider,
    generateSpec: generateVLComboChartSpec,
  },
};

export const CHART_TYPES = Object.keys(CHART_CONFIG) as ChartType[];

export const VISIBLE_CHART_TYPES = CHART_TYPES.filter(
  (type) => !CHART_CONFIG[type].hideFromSelector,
);
