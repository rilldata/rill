import { VisualizationSpec } from "svelte-vega";
import { ChartType } from "../types";
import { buildVegaLiteSpec } from "./build-template";

export function singleLayerBaseSpec() {
  const baseSpec: VisualizationSpec = {
    $schema: "https://vega.github.io/schema/vega-lite/v5.json",
    description: "A simple single layered chart with embedded data.",
    width: "container",
    data: { name: "table" },
    mark: "point",
  };

  return baseSpec;
}

export function multiLayerBaseSpec() {
  const baseSpec: VisualizationSpec = {
    $schema: "https://vega.github.io/schema/vega-lite/v5.json",
    width: "container",
    data: { name: "table" },
    layer: [],
  };
  return baseSpec;
}

export function sanitizeValueForVega(value: unknown) {
  if (typeof value === "string") {
    return value.replace(/[\.\-\{\}\[\]]/g, (match) => `\\${match}`); //eslint-disable-line
  } else {
    return String(value);
  }
}

export function sanitizeValuesForSpec(values: unknown[]) {
  return values.map((value) => sanitizeValueForVega(value));
}

export const templateNameToChartEnumMap = {
  bar: ChartType.BAR,
  grouped_bar: ChartType.GROUPED_BAR,
  stacked_bar: ChartType.STACKED_BAR,
  line: ChartType.LINE,
  area: ChartType.AREA,
  stacked_area: ChartType.STACKED_AREA,
};

export function getSpecFromTemplateProperties(properties) {
  if (!properties.name || !properties.x || !properties.y) {
    return undefined;
  }
  const chartType = templateNameToChartEnumMap[properties?.name];

  const timeFields = [{ name: properties.x, label: properties.x }];
  const quantitativeFields = [{ name: properties.y, label: properties.y }];

  const spec = buildVegaLiteSpec(chartType, timeFields, quantitativeFields, []);
  return spec;
}
