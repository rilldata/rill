import type { VisualizationSpec } from "svelte-vega";
import {
  createConfigWithLegend,
  createEncoding,
  createSingleLayerBaseSpec,
} from "../../builder";
import type { ChartDataResult } from "../../types";
import type { CartesianChartSpec } from "../CartesianChart";

export function generateVLStackedBarChartSpec(
  config: CartesianChartSpec,
  data: ChartDataResult,
): VisualizationSpec {
  const spec = createSingleLayerBaseSpec("bar");
  spec.encoding = createEncoding(config, data);

  const vegaConfig = createConfigWithLegend(config, config.color);

  return {
    ...spec,
    ...(vegaConfig && { config: vegaConfig }),
  };
}
