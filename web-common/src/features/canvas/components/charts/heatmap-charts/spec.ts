import type { Field } from "vega-lite/build/src/channeldef";
import type { TopLevelUnitSpec } from "vega-lite/build/src/spec/unit";
import {
  createColorEncoding,
  createConfigWithLegend,
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

  const vegaConfig = createConfigWithLegend(
    config,
    config.color,
    {
      axis: { grid: true, tickBand: "extent" },
      axisX: {
        grid: true,
        gridDash: [],
        tickBand: "extent",
      },
    },
    "right",
  );

  const xEncoding = createPositionEncoding(config.x, data);
  const yEncoding = createPositionEncoding(config.y, data);

  if (config.x?.type === "nominal" && config.color?.field) {
    xEncoding.sort = {
      op: "sum",
      field: config.color.field,
      order: "descending",
    };
  } else if (config.y?.type === "nominal" && config.color?.field) {
    yEncoding.sort = {
      op: "sum",
      field: config.color.field,
      order: "descending",
    };
  }

  return {
    ...spec,
    encoding: {
      x: xEncoding,
      y: yEncoding,
      color: createColorEncoding(config.color, data),
      tooltip: createDefaultTooltipEncoding(
        [config.x, config.y, config.color],
        data,
      ),
    },
    ...(vegaConfig && { config: vegaConfig }),
  };
}
