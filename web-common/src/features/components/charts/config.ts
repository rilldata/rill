import BarChart from "@rilldata/web-common/components/icons/BarChart.svelte";
import Donut from "@rilldata/web-common/components/icons/Donut.svelte";
import Funnel from "@rilldata/web-common/components/icons/Funnel.svelte";
import Heatmap from "@rilldata/web-common/components/icons/Heatmap.svelte";
import LineChart from "@rilldata/web-common/components/icons/LineChart.svelte";
import MultiChart from "@rilldata/web-common/components/icons/MultiChart.svelte";
import ScatterChart from "@rilldata/web-common/components/icons/ScatterChart.svelte";
import StackedArea from "@rilldata/web-common/components/icons/StackedArea.svelte";
import StackedBar from "@rilldata/web-common/components/icons/StackedBar.svelte";
import StackedBarFull from "@rilldata/web-common/components/icons/StackedBarFull.svelte";
import type { ComponentType, SvelteComponent } from "svelte";
import type { Readable } from "svelte/store";
import { CartesianChartProvider } from "./cartesian/CartesianChartProvider";
import { CircularChartProvider } from "./circular/CircularChartProvider";
import { ComboChartProvider } from "./combo/ComboChartProvider";
import { FunnelChartProvider } from "./funnel/FunnelChartProvider";
import { HeatmapChartProvider } from "./heatmap/HeatmapChartProvider";
import { ScatterPlotChartProvider } from "./scatter/ScatterPlotChartProvider";
import { SPEC_GENERATORS, type SpecGenerator } from "./spec-generators";
import type { ChartProvider, ChartSpec, ChartType } from "./types";

export interface ChartMetadataConfig {
  title: string;
  icon: ComponentType<SvelteComponent>;
  provider: new (
    spec: Readable<ChartSpec>,
    defaultOptions?: Record<string, unknown>,
  ) => ChartProvider;
  // The spec builders live in `spec-generators.ts` so that renderers can import them
  // without this module, which also pulls in the icons and the query providers.
  generateSpec: SpecGenerator;
  hideFromSelector?: boolean;
}

export const CHART_CONFIG: Record<ChartType, ChartMetadataConfig> = {
  bar_chart: {
    title: "Bar",
    icon: BarChart,
    provider: CartesianChartProvider,
    generateSpec: SPEC_GENERATORS.bar_chart,
  },
  line_chart: {
    title: "Line",
    icon: LineChart,
    provider: CartesianChartProvider,
    generateSpec: SPEC_GENERATORS.line_chart,
  },
  area_chart: {
    title: "Stacked Area",
    icon: StackedArea,
    provider: CartesianChartProvider,
    generateSpec: SPEC_GENERATORS.area_chart,
  },
  stacked_bar: {
    title: "Stacked Bar",
    icon: StackedBar,
    provider: CartesianChartProvider,
    generateSpec: SPEC_GENERATORS.stacked_bar,
  },
  stacked_bar_normalized: {
    title: "Stacked Bar Normalized",
    icon: StackedBarFull,
    provider: CartesianChartProvider,
    generateSpec: SPEC_GENERATORS.stacked_bar_normalized,
  },
  donut_chart: {
    title: "Donut",
    icon: Donut,
    provider: CircularChartProvider,
    generateSpec: SPEC_GENERATORS.donut_chart,
  },
  pie_chart: {
    title: "Pie",
    icon: Donut,
    provider: CircularChartProvider,
    generateSpec: SPEC_GENERATORS.pie_chart,
    hideFromSelector: true,
  },
  funnel_chart: {
    title: "Funnel",
    icon: Funnel,
    provider: FunnelChartProvider,
    generateSpec: SPEC_GENERATORS.funnel_chart,
  },
  heatmap: {
    title: "Heatmap",
    icon: Heatmap,
    provider: HeatmapChartProvider,
    generateSpec: SPEC_GENERATORS.heatmap,
  },
  combo_chart: {
    title: "Combo",
    icon: MultiChart,
    provider: ComboChartProvider,
    generateSpec: SPEC_GENERATORS.combo_chart,
  },
  scatter_plot: {
    title: "Scatter",
    icon: ScatterChart,
    provider: ScatterPlotChartProvider,
    generateSpec: SPEC_GENERATORS.scatter_plot,
  },
};

export const CHART_TYPES = Object.keys(CHART_CONFIG) as ChartType[];

export const VISIBLE_CHART_TYPES = CHART_TYPES.filter(
  (type) => !CHART_CONFIG[type].hideFromSelector,
);
