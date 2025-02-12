import BarChart from "@rilldata/web-common/components/icons/BarChart.svelte";
import LineChart from "@rilldata/web-common/components/icons/LineChart.svelte";
import StackedArea from "@rilldata/web-common/components/icons/StackedArea.svelte";
import StackedBar from "@rilldata/web-common/components/icons/StackedBar.svelte";
import { getRillTheme } from "@rilldata/web-common/components/vega/vega-config";
import { V1TimeGrain } from "@rilldata/web-common/runtime-client";
import merge from "deepmerge";
import type { Config } from "vega-lite";
import { generateVLAreaChartSpec } from "./area/spec";
import { generateVLBarChartSpec } from "./bar-chart/spec";
import { generateVLLineChartSpec } from "./line-chart/spec";
import type { ChartDataResult } from "./selector";
import { generateVLStackedBarChartSpec } from "./stacked-bar/spec";
import type { ChartConfig, ChartMetadata, ChartType } from "./types";

export function generateSpec(
  chartType: ChartType,
  chartConfig: ChartConfig,
  data: ChartDataResult,
) {
  if (data.isFetching || data.error) return {};
  switch (chartType) {
    case "bar_chart":
      return generateVLBarChartSpec(chartConfig, data);
    case "stacked_bar":
      return generateVLStackedBarChartSpec(chartConfig, data);
    case "line_chart":
      return generateVLLineChartSpec(chartConfig, data);
    case "area_chart":
      return generateVLAreaChartSpec(chartConfig, data);
  }
}

export const chartMetadata: ChartMetadata[] = [
  { type: "line_chart", title: "Line", icon: LineChart },
  { type: "bar_chart", title: "Bar", icon: BarChart },
  { type: "stacked_bar", title: "Stacked Bar", icon: StackedBar },
  { type: "area_chart", title: "Stacked Area", icon: StackedArea },
];

export function mergedVlConfig(config: string): Config {
  const defaultConfig = getRillTheme(true);
  let parsedConfig: Config;

  try {
    parsedConfig = JSON.parse(config) as Config;
  } catch {
    console.warn("Invalid JSON config");
    return defaultConfig;
  }

  const reverseArrayMerge = (
    destinationArray: unknown[],
    sourceArray: unknown[],
  ) => [...sourceArray, ...destinationArray];

  return merge(defaultConfig, parsedConfig, { arrayMerge: reverseArrayMerge });
}

export function getChartTitle(config: ChartConfig, data: ChartDataResult) {
  const xLabel = config.x?.field
    ? data.fields[config.x.field]?.displayName || config.x.field
    : "";

  const yLabel = config.y?.field
    ? data.fields[config.y.field]?.displayName || config.y.field
    : "";

  const colorLabel =
    typeof config.color === "object" && config.color?.field
      ? data.fields[config.color.field]?.displayName || config.color.field
      : "";

  return colorLabel
    ? `${yLabel} per ${xLabel} split by ${colorLabel}`
    : `${yLabel} per ${xLabel}`;
}

export const timeGrainToVegaTimeUnitMap: Record<V1TimeGrain, string> = {
  [V1TimeGrain.TIME_GRAIN_MILLISECOND]: "yearmonthdatehoursminutesseconds",
  [V1TimeGrain.TIME_GRAIN_SECOND]: "yearmonthdatehoursminutesseconds",
  [V1TimeGrain.TIME_GRAIN_MINUTE]: "yearmonthdatehoursminutes",
  [V1TimeGrain.TIME_GRAIN_HOUR]: "yearmonthdatehours",
  [V1TimeGrain.TIME_GRAIN_DAY]: "yearmonthdate",
  [V1TimeGrain.TIME_GRAIN_WEEK]: "yearweek",
  [V1TimeGrain.TIME_GRAIN_MONTH]: "yearmonth",
  [V1TimeGrain.TIME_GRAIN_QUARTER]: "yearquarter",
  [V1TimeGrain.TIME_GRAIN_YEAR]: "year",
  [V1TimeGrain.TIME_GRAIN_UNSPECIFIED]: "yearmonthdate",
};
