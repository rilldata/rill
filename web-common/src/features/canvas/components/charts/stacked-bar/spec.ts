import type { ChartConfig } from "@rilldata/web-common/features/canvas/components/charts/types";
import type { VisualizationSpec } from "svelte-vega";
import {
  createEncoding,
  createLegendParam,
  createOpacityEncoding,
  createSingleLayerBaseSpec,
} from "../builder";
import type { ChartDataResult } from "../selector";

export function generateVLStackedBarChartSpec(
  config: ChartConfig,
  data: ChartDataResult,
): VisualizationSpec {
  const spec = createSingleLayerBaseSpec("bar");
  const baseEncoding = createEncoding(config, data);

  if (config.color && typeof config.color === "object") {
    baseEncoding.opacity = createOpacityEncoding("legend");
    spec.params = [createLegendParam("legend", config.color.field)];
  }
  spec.encoding = baseEncoding;
  return spec;
}
