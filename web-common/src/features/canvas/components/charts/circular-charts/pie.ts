import type { VisualizationSpec } from "svelte-vega";
import {
  createColorEncoding,
  createDefaultTooltipEncoding,
  createSingleLayerBaseSpec,
  createXEncoding,
} from "../builder";
import type { ChartDataResult } from "../types";
import type { CircularChartSpec } from "./CircularChart";

export function generateVLPieChartSpec(
  config: CircularChartSpec,
  data: ChartDataResult,
): VisualizationSpec {
  const spec = createSingleLayerBaseSpec("arc");

  spec.mark = {
    type: "arc",
    innerRadius: config.innerRadius || 0,
  };
  const theta = createXEncoding(config.measure, data);
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
  };
}
