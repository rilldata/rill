import type { ChartConfig } from "@rilldata/web-common/features/canvas/components/charts/types";
import type { VisualizationSpec } from "svelte-vega";
import type { ChartDataResult } from "../selector";

export function generateVLBarChartSpec(
  config: ChartConfig,
  data: ChartDataResult,
): VisualizationSpec {
  return {
    $schema: "https://vega.github.io/schema/vega-lite/v5.json",
    mark: "bar",
    width: "container",
    autosize: { type: "fit" },
    data: { name: "metrics-view" },
    encoding: {
      ...(config.x && {
        x: {
          field: config.x.field,
          title: data.fields[config.x.field]?.displayName || config.x.field,
          type: config.x.type,
          ...(config.x.timeUnit && { timeUnit: config.x.timeUnit }),
        },
      }),
      ...(config.y && {
        y: {
          field: config.y.field,
          title: data.fields[config.y.field]?.displayName || config.y.field,
          type: config.y.type,
          ...(config.y.timeUnit && { timeUnit: config.y.timeUnit }),
        },
      }),
      ...(config.color &&
        typeof config.color === "object" && {
          color: {
            field: config.color.field,
            title:
              data.fields[config.color.field]?.displayName ||
              config.color.field,
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
        typeof config.color === "object" &&
        config.x && {
          xOffset: {
            field: config.color.field,
            title:
              data.fields[config.color.field]?.displayName ||
              config.color.field,
          },
        }),
    },
  };
}
