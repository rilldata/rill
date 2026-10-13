import type { VisualizationSpec } from "svelte-vega";
import { generateVLAreaChartSpec } from "./cartesian/area/spec";
import { generateVLBarChartSpec } from "./cartesian/bar-chart/spec";
import type { CartesianChartSpec } from "./cartesian/CartesianChartProvider";
import { generateVLLineChartSpec } from "./cartesian/line-chart/spec";
import { generateVLMultiMetricChartSpec } from "./cartesian/multi-metric-chart";
import { measureChannel } from "./cartesian/orientation";
import { generateVLStackedBarChartSpec } from "./cartesian/stacked-bar/default";
import { generateVLStackedBarNormalizedSpec } from "./cartesian/stacked-bar/normalized";
import { generateVLPieChartSpec } from "./circular/pie";
import { generateVLComboChartSpec } from "./combo/spec";
import { generateVLFunnelChartSpec } from "./funnel/spec";
import { generateVLHeatmapSpec } from "./heatmap/spec";
import { generateVLScatterPlotSpec } from "./scatter/spec";
import type { ChartDataResult, ChartSpec, ChartType } from "./types";
import { isMultiFieldConfig } from "./util";

/**
 * Builds a Vega-Lite spec from a Rill chart spec and its data.
 *
 * This module holds only the spec builders so that renderers can import them without
 * `config.ts`, which also carries the icons and the query providers (TanStack, Connect RPC).
 * Standalone bundles such as the MCP chart widget depend on that boundary.
 */
export type SpecGenerator = (
  config: ChartSpec,
  data: ChartDataResult,
) => VisualizationSpec;

type MultiMetricMarkType = Parameters<typeof generateVLMultiMetricChartSpec>[2];

/**
 * The cartesian types share one decision: a multi-measure field config draws the
 * multi-metric variant, anything else the single-measure builder.
 * The measure may sit on either axis (horizontal bars), so it is read by role.
 */
function cartesianSpecGenerator(
  singleMeasure: (
    config: CartesianChartSpec,
    data: ChartDataResult,
  ) => VisualizationSpec,
  multiMetricMarkType: MultiMetricMarkType,
): SpecGenerator {
  return (config: ChartSpec, data: ChartDataResult) => {
    const cartesianConfig = config as CartesianChartSpec;
    const isMultiMeasure = isMultiFieldConfig(
      cartesianConfig[measureChannel(cartesianConfig)],
    );
    return isMultiMeasure
      ? generateVLMultiMetricChartSpec(
          cartesianConfig,
          data,
          multiMetricMarkType,
        )
      : singleMeasure(cartesianConfig, data);
  };
}

export const SPEC_GENERATORS: Record<ChartType, SpecGenerator> = {
  bar_chart: cartesianSpecGenerator(generateVLBarChartSpec, "grouped_bar"),
  line_chart: cartesianSpecGenerator(generateVLLineChartSpec, "line"),
  area_chart: cartesianSpecGenerator(generateVLAreaChartSpec, "stacked_area"),
  stacked_bar: cartesianSpecGenerator(
    generateVLStackedBarChartSpec,
    "stacked_bar",
  ),
  stacked_bar_normalized: cartesianSpecGenerator(
    generateVLStackedBarNormalizedSpec,
    "stacked_bar_normalized",
  ),
  donut_chart: generateVLPieChartSpec as SpecGenerator,
  pie_chart: generateVLPieChartSpec as SpecGenerator,
  funnel_chart: generateVLFunnelChartSpec as SpecGenerator,
  heatmap: generateVLHeatmapSpec as SpecGenerator,
  combo_chart: generateVLComboChartSpec as SpecGenerator,
  scatter_plot: generateVLScatterPlotSpec as SpecGenerator,
};

export function generateSpec(
  chartType: ChartType,
  rillChartSpec: ChartSpec,
  data: ChartDataResult,
) {
  if (data.isFetching || data.error) return {};
  return SPEC_GENERATORS[chartType]?.(rillChartSpec, data);
}
