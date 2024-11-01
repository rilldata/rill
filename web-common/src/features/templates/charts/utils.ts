import { ChartType } from "@rilldata/web-common/features/canvas-components/types";
import type { ChartProperties } from "@rilldata/web-common/features/templates/types";
import type { VisualizationSpec } from "svelte-vega";
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
  bar_chart: ChartType.BAR,
  grouped_bar_chart: ChartType.GROUPED_BAR,
  stacked_bar_chart: ChartType.STACKED_BAR,
  line_chart: ChartType.LINE,
  area_chart: ChartType.AREA,
  stacked_area_chart: ChartType.STACKED_AREA,
};

export function getSpecFromTemplateProperties(
  renderer: string,
  properties: ChartProperties,
) {
  if (!properties.x || !properties.y) {
    return undefined;
  }
  const chartType = templateNameToChartEnumMap[renderer];

  if (!chartType) return undefined;

  const timeFields = [{ name: properties.x, label: properties.x }];
  const quantitativeFields = [{ name: properties.y, label: properties.y }];

  const spec = buildVegaLiteSpec(chartType, timeFields, quantitativeFields, []);
  return spec;
}
