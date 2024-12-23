import type { ChartConfig } from "@rilldata/web-common/features/canvas/components/charts/types";
import type { VisualizationSpec } from "svelte-vega";
import type { ChartDataResult } from "../selector";

export function generateVLStackedBarChartSpec(
  config: ChartConfig,
  data: ChartDataResult,
): VisualizationSpec {
  return {
    $schema: "https://vega.github.io/schema/vega-lite/v5.json",
    mark: "bar",
    width: "container",
    data: { name: "metrics-view" },
    autosize: { type: "fit" },
    encoding: {
      ...(config.x && {
        x: {
          field: config.x.field,
          title: data.dimension?.displayName || config.x.field,
          type: config.x.type,
          ...(config.x.timeUnit && { timeUnit: config.x.timeUnit }),
        },
      }),
      ...(config.y && {
        y: {
          field: config.y.field,
          title: data.measure?.displayName || config.y.field,
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
    },
  };
}
