import type { ChartConfig } from "@rilldata/web-common/features/canvas/components/charts/types";
import type { VisualizationSpec } from "svelte-vega";

export function generateVLBarChartSpec(config: ChartConfig): VisualizationSpec {
  return {
    $schema: "https://vega.github.io/schema/vega-lite/v5.json",
    mark: "line",
    width: "container",
    autosize: { type: "fit" },
    data: { name: "metrics-view" },
    encoding: {
      ...(config.x && {
        x: {
          field: config.x.field,
          type: config.x.type,
          ...(config.x.timeUnit && { timeUnit: config.x.timeUnit }),
        },
      }),
      ...(config.y && {
        y: {
          field: config.y.field,
          type: config.y.type,
          ...(config.y.timeUnit && { timeUnit: config.y.timeUnit }),
        },
      }),
      ...(config.color &&
        typeof config.color === "object" && {
          color: {
            field: config.color.field,
            type: config.color.type,
            ...(config.color.timeUnit && {
              timeUnit: config.color.timeUnit,
            }),
          },
        }),
      ...(config.color &&
        typeof config.color === "string" && {
          color: { value: config.color },
        }),
      ...(config.color &&
        config.x && {
          xOffset: {
            field:
              typeof config.color === "object"
                ? config.color.field
                : config.color,
          },
        }),
    },
  };
}
