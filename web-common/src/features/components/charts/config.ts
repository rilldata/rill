import BarChart from "@rilldata/web-common/components/icons/BarChart.svelte";
import Donut from "@rilldata/web-common/components/icons/Donut.svelte";
import Funnel from "@rilldata/web-common/components/icons/Funnel.svelte";
import Heatmap from "@rilldata/web-common/components/icons/Heatmap.svelte";
import LineChart from "@rilldata/web-common/components/icons/LineChart.svelte";
import MultiChart from "@rilldata/web-common/components/icons/MultiChart.svelte";
import StackedArea from "@rilldata/web-common/components/icons/StackedArea.svelte";
import StackedBar from "@rilldata/web-common/components/icons/StackedBar.svelte";
import StackedBarFull from "@rilldata/web-common/components/icons/StackedBarFull.svelte";
import { generateVLAreaChartSpec } from "@rilldata/web-common/features/canvas/components/charts/cartesian-charts/area/spec";
import { generateVLBarChartSpec } from "@rilldata/web-common/features/canvas/components/charts/cartesian-charts/bar-chart/spec";
import type { CartesianChartSpec } from "@rilldata/web-common/features/canvas/components/charts/cartesian-charts/CartesianChart";
import { generateVLLineChartSpec } from "@rilldata/web-common/features/canvas/components/charts/cartesian-charts/line-chart/spec";
import { generateVLMultiMetricChartSpec } from "@rilldata/web-common/features/canvas/components/charts/cartesian-charts/multi-metric-chart.ts";
import { generateVLStackedBarChartSpec } from "@rilldata/web-common/features/canvas/components/charts/cartesian-charts/stacked-bar/default";
import { generateVLStackedBarNormalizedSpec } from "@rilldata/web-common/features/canvas/components/charts/cartesian-charts/stacked-bar/normalized";
import { generateVLPieChartSpec } from "@rilldata/web-common/features/canvas/components/charts/circular-charts/pie";
import { generateVLComboChartSpec } from "@rilldata/web-common/features/canvas/components/charts/combo-charts/spec";
import { generateVLFunnelChartSpec } from "@rilldata/web-common/features/canvas/components/charts/funnel-charts/spec";
import { generateVLHeatmapSpec } from "@rilldata/web-common/features/canvas/components/charts/heatmap-charts/spec";
import {
  type ChartDataResult,
  type ChartType,
} from "@rilldata/web-common/features/components/charts";
import type { ChartSpec } from "@rilldata/web-common/features/components/charts/types";
import { isMultiFieldConfig } from "@rilldata/web-common/features/components/charts/util.ts";
import type { ComponentType, SvelteComponent } from "svelte";
import type { VisualizationSpec } from "svelte-vega";

export interface ChartMetadataConfig {
  title: string;
  icon: ComponentType<SvelteComponent>;
  generateSpec: (config: ChartSpec, data: ChartDataResult) => VisualizationSpec;
  hideFromSelector?: boolean;
}

export const CHART_CONFIG: Record<ChartType, ChartMetadataConfig> = {
  bar_chart: {
    title: "Bar",
    icon: BarChart,
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
    generateSpec: generateVLPieChartSpec,
  },
  pie_chart: {
    title: "Pie",
    icon: Donut,
    generateSpec: generateVLPieChartSpec,
    hideFromSelector: true,
  },
  funnel_chart: {
    title: "Funnel",
    icon: Funnel,
    generateSpec: generateVLFunnelChartSpec,
  },
  heatmap: {
    title: "Heatmap",
    icon: Heatmap,
    generateSpec: generateVLHeatmapSpec,
  },
  combo_chart: {
    title: "Combo",
    icon: MultiChart,
    generateSpec: generateVLComboChartSpec,
  },
};

export const CHART_TYPES = Object.keys(CHART_CONFIG) as ChartType[];

export const VISIBLE_CHART_TYPES = CHART_TYPES.filter(
  (type) => !CHART_CONFIG[type].hideFromSelector,
);
