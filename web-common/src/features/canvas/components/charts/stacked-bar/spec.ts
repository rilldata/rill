import type { ChartConfig } from "@rilldata/web-common/features/canvas/components/charts/types";
import type { VisualizationSpec } from "svelte-vega";
import { createEncoding, createSingleLayerBaseSpec } from "../builder";
import type { ChartDataResult } from "../selector";

export function generateVLStackedBarChartSpec(
  config: ChartConfig,
  data: ChartDataResult,
): VisualizationSpec {
  const spec = createSingleLayerBaseSpec("bar");
  spec.encoding = createEncoding(config, data);
  return spec;
}
