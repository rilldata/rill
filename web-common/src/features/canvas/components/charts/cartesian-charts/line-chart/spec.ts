import type { TooltipValue } from "@rilldata/web-common/features/canvas/components/charts/types";
import { sanitizeFieldName } from "@rilldata/web-common/features/canvas/components/charts/util";
import { sanitizeValueForVega } from "@rilldata/web-common/features/templates/charts/utils";
import type { VisualizationSpec } from "svelte-vega";
import {
  createColorEncoding,
  createConfigWithLegend,
  createDefaultTooltipEncoding,
  createMultiLayerBaseSpec,
  createPositionEncoding,
} from "../../builder";
import type { ChartDataResult } from "../../types";
import type { CartesianChartSpec } from "../CartesianChart";

export function generateVLLineChartSpec(
  config: CartesianChartSpec,
  data: ChartDataResult,
): VisualizationSpec {
  const spec = createMultiLayerBaseSpec();
  const vegaConfig = createConfigWithLegend(config, config.color);

  const colorField =
    typeof config.color === "object" ? config.color.field : undefined;
  const xField = sanitizeValueForVega(config.x?.field);
  const yField = sanitizeValueForVega(config.y?.field);

  const defaultTooltipChannel = createDefaultTooltipEncoding(
    [config.x, config.y, config.color],
    data,
  );
  let multiValueTooltipChannel: TooltipValue[] | undefined;

  if (colorField && config.x && yField) {
    multiValueTooltipChannel = data.data?.map((value) => ({
      field: sanitizeValueForVega(value?.[colorField]),
      type: "quantitative",
      formatType: sanitizeFieldName(yField),
    }));

    multiValueTooltipChannel.unshift({
      field: xField,
      title: data.fields[config.x.field]?.displayName || config.x.field,
      type: config.x?.type,
      ...(config.x.type === "temporal" && { format: "%b %d, %Y %H:%M" }),
    });

    multiValueTooltipChannel = multiValueTooltipChannel.slice(0, 50);
  }

  spec.encoding = { x: createPositionEncoding(config.x, data) };

  spec.layer = [
    {
      encoding: {
        y: createPositionEncoding(config.y, data),
        color: createColorEncoding(config.color, data),
      },
      layer: [
        { mark: "line" },
        {
          transform: [{ filter: { param: "hover", empty: false } }],
          mark: {
            type: "point",
            filled: true,
            opacity: 1,
            size: 50,
            clip: true,
            stroke: "white",
            strokeWidth: 1,
          },
        },
      ],
    },
    {
      transform:
        xField && yField && colorField && multiValueTooltipChannel?.length
          ? [
              {
                pivot: colorField,
                value: yField,
                groupby: [xField],
              },
            ]
          : [],
      mark: {
        type: "rule",
        clip: true,
      },
      encoding: {
        x: {
          field: xField,
          ...(yField && config.x?.sort === "y"
            ? {
                sort: {
                  field: yField,
                  order: "ascending",
                },
              }
            : yField && config.x?.sort === "-y"
              ? {
                  sort: {
                    field: yField,
                    order: "descending",
                  },
                }
              : {}),
        },
        color: {
          condition: [
            {
              param: "hover",
              empty: false,
              value: "var(--color-primary-300)",
            },
          ],
          value: "transparent",
        },
        tooltip: multiValueTooltipChannel?.length
          ? multiValueTooltipChannel
          : defaultTooltipChannel,
      },
      params: [
        {
          name: "hover",
          select: {
            type: "point",
            encodings: ["x"],
            nearest: true,
            on: "pointerover",
            clear: "pointerout",
          },
        },
      ],
    },
  ];

  return {
    ...spec,
    ...(vegaConfig && { config: vegaConfig }),
  };
}
