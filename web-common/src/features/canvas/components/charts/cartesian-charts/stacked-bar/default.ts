import type { VisualizationSpec } from "svelte-vega";
import {
  createConfig,
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

  const vegaConfig = createConfig(config);

  return {
    ...spec,
    ...(vegaConfig && { config: vegaConfig }),
  };
}
