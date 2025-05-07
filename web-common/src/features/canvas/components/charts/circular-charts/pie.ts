import type { VisualizationSpec } from "svelte-vega";
import type { Config } from "vega-lite";
import {
  createColorEncoding,
  createConfig,
  createDefaultTooltipEncoding,
  createPositionEncoding,
  createSingleLayerBaseSpec,
} from "../builder";
import type { ChartDataResult } from "../types";
import type { CircularChartSpec } from "./CircularChart";

function getInnerRadius(innerRadiusPercentage: number | undefined) {
  if (!innerRadiusPercentage) return 0;

  if (innerRadiusPercentage >= 100 || innerRadiusPercentage < 0) {
    console.warn("Inner radius percentage must be between 0 and 100");
    return { expr: `0.5*min(width,height)/2` };
  }

  const decimal = innerRadiusPercentage / 100;
  return { expr: `${decimal}*min(width,height)/2` };
}

/**
 * The layout property is not typed in the current version of Vega-Lite.
 * This will be fixed when we upgrade to Svelte 5 and subseqent Vega-Lite versions.
 */
export function generateVLPieChartSpec(
  config: CircularChartSpec,
  data: ChartDataResult,
): VisualizationSpec {
  const spec = createSingleLayerBaseSpec("arc");
  const vegaConfig = createConfig(config, {
    legend: {
      orient: "right",
      layout: {
        right: { anchor: "middle" },
      },
    },
  } as unknown as Config);

  spec.mark = {
    type: "arc",
    innerRadius: getInnerRadius(config.innerRadius),
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
