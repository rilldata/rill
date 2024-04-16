import { VisualizationSpec } from "svelte-vega";

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
