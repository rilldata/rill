import type { Field } from "vega-lite/build/src/channeldef";
import type { TopLevelUnitSpec } from "vega-lite/build/src/spec/unit";
import {
  createColorEncoding,
  createConfig,
  createDefaultTooltipEncoding,
  createPositionEncoding,
  createSingleLayerBaseSpec,
} from "../builder";
import type { ChartDataResult } from "../types";
import type { HeatmapChartSpec } from "./HeatmapChart";

export function generateVLHeatmapSpec(
  config: HeatmapChartSpec,
  data: ChartDataResult,
): TopLevelUnitSpec<Field> {
  const spec = createSingleLayerBaseSpec("rect");

  const vegaConfig = createConfig(config, {
    legend: {
      orient: "bottom",
    },
    axis: { grid: true, tickBand: "extent" },
    axisX: {
      grid: true,
      gridDash: [],
      tickBand: "extent",
    },
  });

  return {
    ...spec,
    encoding: {
      x: createPositionEncoding(config.x, data),
      y: createPositionEncoding(config.y, data),
      color: createColorEncoding(config.color, data),
      tooltip: createDefaultTooltipEncoding(
        [config.x, config.y, config.color],
        data,
      ),
    },
    ...(vegaConfig && { config: vegaConfig }),
  };
}
