import type { ChartTypeConfig } from "@rilldata/web-common/features/dashboards/canvas/types";
import type { VisualizationSpec } from "svelte-vega";

export function generateVLBarChartSpec(
  config: ChartTypeConfig,
): VisualizationSpec {
  return {
    $schema: "https://vega.github.io/schema/vega-lite/v5.json",
    mark: "bar",
    width: "container",
    data: { name: "metrics-view" },
    encoding: {
      ...(config.data.x && {
        x: {
          field: config.data.x.field,
          type: config.data.x.type,
          ...(config.data.x.timeUnit && { timeUnit: config.data.x.timeUnit }),
        },
      }),
      ...(config.data.y && {
        y: {
          field: config.data.y.field,
          type: config.data.y.type,
          ...(config.data.y.timeUnit && { timeUnit: config.data.y.timeUnit }),
        },
      }),
      ...(config.data.color && {
        color: {
          field: config.data.color.field,
          type: config.data.color.type,
          ...(config.data.color.timeUnit && {
            timeUnit: config.data.color.timeUnit,
          }),
        },
      }),
    },
  };
}
