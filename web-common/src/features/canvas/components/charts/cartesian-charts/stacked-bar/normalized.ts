import type { TooltipValue } from "@rilldata/web-common/features/canvas/components/charts/types";
import type { VisualizationSpec } from "svelte-vega";
import {
  createDefaultTooltipEncoding,
  createEncoding,
  createSingleLayerBaseSpec,
} from "../../builder";
import type { ChartDataResult } from "../../types";
import type { CartesianChartSpec } from "../CartesianChart";

export function generateVLStackedBarNormalizedSpec(
  config: CartesianChartSpec,
  data: ChartDataResult,
): VisualizationSpec {
  const spec = createSingleLayerBaseSpec("bar");
  const baseEncoding = createEncoding(config, data);

  if (baseEncoding.y && config.y?.field) {
    const yField = config.y.field;

    baseEncoding.y = {
      ...baseEncoding.y,
      stack: "normalize",
      ...(baseEncoding.y && {
        scale: {
          zero: false,
        },
      }),
      axis: {
        ...(!config.y.showAxisTitle && { title: null }),
        format: ".0%",
      },
    };

    // Add a transform to calculate the percentage
    spec.transform = [
      {
        joinaggregate: [
          {
            op: "sum",
            field: yField,
            as: "total",
          },
        ],
        groupby: config.x?.field ? [config.x.field] : [],
      },
      {
        calculate: `datum['${yField}'] / datum.total`,
        as: "percentage",
      },
    ];

    // Add percentage to tooltip
    const tooltipValues = createDefaultTooltipEncoding(
      [config.x, config.y, config.color],
      data,
    );
    baseEncoding.tooltip = tooltipValues
      .map((t: TooltipValue) => {
        if (t.field === yField) {
          return [
            {
              ...t,
            },
            {
              ...t,
              title: `${t.title} (%)`,
              field: "percentage",
              formatType: undefined,
              format: ".1%",
            },
          ];
        }
        return t;
      })
      .flat();
  }

  spec.encoding = baseEncoding;
  return spec;
}
