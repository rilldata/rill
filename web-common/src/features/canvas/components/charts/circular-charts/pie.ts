import type { VisualizationSpec } from "svelte-vega";
import type { Config } from "vega-lite";
import type { ExprRef, SignalRef } from "vega-typings";
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
  /**
   * The layout property is not typed in the current version of Vega-Lite.
   * This will be fixed when we upgrade to Svelte 5 and subseqent Vega-Lite versions.
   */
  const vegaConfig = createConfigWithLegend(
    config,
    config.color,
    {
      legend: {
        layout: {
          right: { anchor: "middle" },
          left: { anchor: "middle" },
          top: { anchor: "middle" },
          bottom: { anchor: "middle" },
        },
      },
    } as unknown as Config<ExprRef | SignalRef>,
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
