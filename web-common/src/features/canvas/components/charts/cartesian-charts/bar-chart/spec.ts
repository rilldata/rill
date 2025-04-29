import type { VisualizationSpec } from "svelte-vega";
import {
  createConfig,
  createEncoding,
  createSingleLayerBaseSpec,
} from "../../builder";
import type { ChartDataResult } from "../../types";
import type { CartesianChartSpec } from "../CartesianChart";

export function generateVLBarChartSpec(
  config: CartesianChartSpec,
  data: ChartDataResult,
): VisualizationSpec {
  const spec = createSingleLayerBaseSpec("bar");
  const baseEncoding = createEncoding(config, data);
  const vegaConfig = createConfig(config);

  if (config.color && typeof config.color === "object" && config.x) {
    baseEncoding.xOffset = {
      field: config.color.field,
      title: data.fields[config.color.field]?.displayName || config.color.field,
    };
  }
  spec.encoding = baseEncoding;

  return {
    ...spec,
    ...(vegaConfig && { config: vegaConfig }),
  };
}
