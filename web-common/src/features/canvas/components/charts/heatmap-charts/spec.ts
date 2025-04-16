import type { Field } from "vega-lite/build/src/channeldef";
import type { TopLevelUnitSpec } from "vega-lite/build/src/spec/unit";
import {
  createColorEncoding,
  createDefaultTooltipEncoding,
  createSingleLayerBaseSpec,
  createXEncoding,
  createYEncoding,
} from "../builder";
import type { ChartDataResult } from "../types";
import type { HeatmapChartSpec } from "./HeatmapChart";

export function generateVLHeatmapSpec(
  config: HeatmapChartSpec,
  data: ChartDataResult,
): TopLevelUnitSpec<Field> {
  const spec = createSingleLayerBaseSpec("rect");

  return {
    ...spec,
    encoding: {
      x: createXEncoding(config.x, data),
      y: createYEncoding(config.y, data),
      color: createColorEncoding(config.color, data),
      tooltip: createDefaultTooltipEncoding(
        [config.x, config.y, config.color],
        data,
      ),
    },
    config: {
      axis: { grid: true, tickBand: "extent" },
      axisX: {
        grid: true,
        gridDash: [],
        tickBand: "extent",
      },
    },
  };
}
