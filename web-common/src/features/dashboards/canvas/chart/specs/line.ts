import type { ChartTypeConfig } from "@rilldata/web-common/features/dashboards/canvas/types";
import type { VisualizationSpec } from "svelte-vega";

export function generateVLLineChartSpec(
  config: ChartTypeConfig,
): VisualizationSpec {
  return {
    $schema: "https://vega.github.io/schema/vega-lite/v5.json",
    mark: {
      type: "line",
      point: true,
    },
    width: "container",
    autosize: { type: "fit" },
    data: { name: "metrics-view" },
    encoding: {
      ...(config.data?.x?.field && {
        x: {
          field: config.data.x.field,
          type: config.data.x.type,
          ...(config.data.x.timeUnit && { timeUnit: config.data.x.timeUnit }),
        },
      }),
      ...(config.data.y?.field && {
        y: {
          field: config.data.y.field,
          type: config.data.y.type,
          ...(config.data.y.timeUnit && { timeUnit: config.data.y.timeUnit }),
        },
      }),
      ...(config.data.color?.field && {
        color: {
          field: config.data.color.field,
          type: config.data.color.type,
        },
      }),
    },
  };
}
