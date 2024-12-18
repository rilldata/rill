import { ChartType } from "@rilldata/web-common/components/vega/types";
import type { VisualizationSpec } from "svelte-vega";

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
