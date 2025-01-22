import type { ChartConfig } from "@rilldata/web-common/features/canvas/components/charts/types";
import type { VisualizationSpec } from "svelte-vega";
import { createEncoding, createSingleLayerBaseSpec } from "../builder";
import type { ChartDataResult } from "../selector";

export function generateVLLineChartSpec(
  config: ChartConfig,
  data: ChartDataResult,
): VisualizationSpec {
  const spec = createSingleLayerBaseSpec("line");
  spec.encoding = createEncoding(config, data);
  return spec;
}
