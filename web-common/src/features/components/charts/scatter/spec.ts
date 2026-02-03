import type { ChartDataResult } from "@rilldata/web-common/features/components/charts";
import {
  createColorEncoding,
  createConfigWithLegend,
  createDefaultTooltipEncoding,
  createPositionEncoding,
  createSingleLayerBaseSpec,
  createSizeEncoding,
} from "@rilldata/web-common/features/components/charts/builder";
import type { VisualizationSpec } from "svelte-vega";
import type { ScatterPlotChartSpec } from "./ScatterPlotChartProvider";

export function generateVLScatterPlotSpec(
  config: ScatterPlotChartSpec,
  data: ChartDataResult,
): VisualizationSpec {
  const spec = createSingleLayerBaseSpec("point");

  const vegaConfig = createConfigWithLegend(
    config,
    config.color,
    {
      axisX: { grid: true },
    },
    "top",
  );
  const tooltip = createDefaultTooltipEncoding(
    [config.dimension, config.x, config.y, config.size, config.color],
    data,
  );

  const xEncoding = createPositionEncoding(config.x, data);
  const yEncoding = createPositionEncoding(config.y, data);
  const sizeEncoding = createSizeEncoding(config.size, data);

  xEncoding.scale = { ...(xEncoding.scale ?? {}), nice: true };
  yEncoding.scale = { ...(yEncoding.scale ?? {}), nice: true };

  spec.encoding = {
    x: xEncoding,
    y: yEncoding,
    ...(config.size && {
      size: sizeEncoding,
    }),
    color: createColorEncoding(config.color, data),
    tooltip,
  };

  if (!config.size && spec.mark && typeof spec.mark === "object") {
    spec.mark = {
      ...spec.mark,
      size: 40,
    } as typeof spec.mark;
  }

  return {
    ...spec,
    selection: {
      grid: {
        type: "interval",
        bind: "scales",
        translate: "[mousedown[event.shiftKey], mouseup] > mousemove",
        zoom: "wheel![!event.shiftKey]",
        clear: "dblclick",
      },
    },
    ...(vegaConfig && { config: vegaConfig }),
  } as VisualizationSpec;
}
