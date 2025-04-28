import type { VisualizationSpec } from "svelte-vega";
import {
  createColorEncoding,
  createConfigWithLegend,
  createDefaultTooltipEncoding,
  createPositionEncoding,
  createSingleLayerBaseSpec,
} from "../builder";
import type { ChartDataResult } from "../types";
import type { CircularChartSpec } from "./CircularChart";

export function generateVLPieChartSpec(
  config: CircularChartSpec,
  data: ChartDataResult,
): VisualizationSpec {
  const spec = createSingleLayerBaseSpec("arc");
  const vegaConfig = createConfigWithLegend(
    config,
    config.color,
    undefined,
    "right",
  );

  spec.mark = {
    type: "arc",
    innerRadius: config.innerRadius || 0,
  };
  const theta = createPositionEncoding(config.measure, data);
  const color = createColorEncoding(config.color, data);
  const tooltip = createDefaultTooltipEncoding(
    [config.measure, config.color],
    data,
  );

  return {
    ...spec,
    encoding: {
      theta,
      color,
      tooltip,
    },
    ...(vegaConfig && { config: vegaConfig }),
  };
}
