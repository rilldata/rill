import {
  sanitizeValueForVega,
  sanitizeValuesForSpec,
} from "@rilldata/web-common/components/vega/util";
import type { TooltipValue } from "@rilldata/web-common/features/components/charts/types";
import { ScrubBoxColor } from "@rilldata/web-common/features/dashboards/time-series/chart-colors";
import type { ChartField } from "./build-template";
import { multiLayerBaseSpec } from "./utils";

export function buildStackedArea(
  timeField: ChartField,
  quantitativeField: ChartField,
  nominalField: ChartField,
) {
  const baseSpec = multiLayerBaseSpec();

  const defaultTooltipChannel: TooltipValue[] = [
    {
      title: quantitativeField.label,
      field: quantitativeField.name,
      formatType: quantitativeField.formatterFunction || "number",
      type: "quantitative",
    },
    {
      field: timeField.name,
      type: "temporal",
      title: "Time",
      format: "%b %d, %Y %H:%M",
    },
    {
      title: nominalField.label,
      field: nominalField.name,
      type: "nominal",
    },
  ];

  const multiValueTooltipChannel: TooltipValue[] | undefined =
    nominalField?.values?.map((value) => ({
      field: sanitizeValueForVega(value),
      type: "quantitative",
      formatType: quantitativeField.formatterFunction || "number",
    }));

  if (multiValueTooltipChannel?.length) {
    multiValueTooltipChannel.unshift({
      field: timeField.name,
      type: "temporal",
      title: "Time",
      format: "%b %d, %Y %H:%M",
    });
  }

  baseSpec.encoding = {
    x: {
      field: timeField.name,
      type: "temporal",
    },
    y: {
      field: quantitativeField.name,
      type: "quantitative",
      stack: "zero",
    },
    color: {
      field: nominalField?.name,
      type: "nominal",
      legend: null,
    },
  };

  if (nominalField?.values?.length) {
    const values = sanitizeValuesForSpec(nominalField.values);
    baseSpec.transform = [
      {
        calculate: `indexof([${values
          ?.map((v) => `'${v}'`)
          .reverse()
          .join(",")}], datum.${nominalField?.name})`,
        as: "order",
      },
    ];
    baseSpec.encoding.order = { field: "order", type: "ordinal" };
  }

  baseSpec.layer = [
    {
      mark: {
        type: "area",
        clip: true,
      },
      encoding: {
        opacity: {
          condition: {
            param: "brush",
            empty: false,
            value: 1,
          },
          value: 0.7,
        },
      },
    },
    {
      mark: {
        type: "line",
        strokeWidth: 1,
        clip: true,
      },
    },
    {
      transform: multiValueTooltipChannel?.length
        ? [
            {
              pivot: nominalField.name,
              value: quantitativeField.name,
              groupby: [timeField.name],
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
        {
          name: "brush",
          select: {
            type: "interval",
            encodings: ["x"],
            mark: {
              fill: ScrubBoxColor,
              fillOpacity: 0.2,
              stroke: ScrubBoxColor,
              strokeWidth: 1,
              strokeOpacity: 0.8,
            },
          },
        },
      ],
    },
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
  ];

  return baseSpec;
}
