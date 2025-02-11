import type {
  ChartConfig,
  TooltipValue,
} from "@rilldata/web-common/features/canvas/components/charts/types";
import { sanitizeValueForVega } from "@rilldata/web-common/features/templates/charts/utils";
import type { VisualizationSpec } from "svelte-vega";
import {
  createColorEncoding,
  createDefaultTooltipEncoding,
  createMultiLayerBaseSpec,
  createXEncoding,
  createYEncoding,
} from "../builder";
import type { ChartDataResult } from "../selector";

export function generateVLAreaChartSpec(
  config: ChartConfig,
  data: ChartDataResult,
): VisualizationSpec {
  const spec = createMultiLayerBaseSpec();

  const colorField =
    typeof config.color === "object" ? config.color.field : undefined;
  const xField = config.x?.field;
  const yField = config.y?.field;

  const defaultTooltipChannel = createDefaultTooltipEncoding(config, data);
  let multiValueTooltipChannel: TooltipValue[] | undefined;

  if (colorField && config.x && yField) {
    multiValueTooltipChannel = data.data?.map((value) => ({
      field: sanitizeValueForVega(value?.[colorField]),
      type: "quantitative",
      formatType: yField,
    }));

    multiValueTooltipChannel.unshift({
      field: sanitizeValueForVega(config.x.field),
      title: data.fields[config.x.field]?.displayName || config.x.field,
      type: config.x?.type,
      ...(config.x.type === "temporal" && { format: "%b %d, %Y %H:%M" }),
    });
  }

  spec.encoding = { x: createXEncoding(config, data) };

  spec.layer = [
    {
      encoding: {
        y: { ...createYEncoding(config, data), stack: "zero" },
        color: createColorEncoding(config, data),
      },
      layer: [
        { mark: "area" },
        {
          transform: [{ filter: { param: "hover", empty: false } }],
          mark: {
            type: "point",
            filled: true,
            opacity: 1,
            size: 40,
            clip: true,
            stroke: "white",
            strokeWidth: 1,
          },
        },
        { mark: { type: "line", opacity: 0.5 } },
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
        y: { value: -400 },
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

  return spec;
}
