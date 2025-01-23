import type { ChartConfig } from "@rilldata/web-common/features/canvas/components/charts/types";
import type { VisualizationSpec } from "svelte-vega";
import { createEncoding, createSingleLayerBaseSpec } from "../builder";
import type { ChartDataResult } from "../selector";

export function generateVLBarChartSpec(
  config: ChartConfig,
  data: ChartDataResult,
): VisualizationSpec {
  const spec = createSingleLayerBaseSpec("bar");
  const baseEncoding = createEncoding(config, data);

  if (config.color && typeof config.color === "object" && config.x) {
    baseEncoding.xOffset = {
      field: config.color.field,
      title: data.fields[config.color.field]?.displayName || config.color.field,
    };
  }

  spec.encoding = baseEncoding;
  return spec;
}
