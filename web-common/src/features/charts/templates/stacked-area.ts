import { ChartField } from "./build-template";
import { multiLayerBaseSpec } from "./utils";

/** Temporary solution for the lack of vega lite type exports */
interface TooltipValue {
  title?: string;
  field: string;
  format?: string;
  type: "quantitative" | "temporal" | "nominal" | "ordinal";
}
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
    nominalField?.values?.map((value) => {
      return {
        field: value === null ? "null" : value,
        type: "quantitative",
      };
    });

  if (multiValueTooltipChannel?.length) {
    multiValueTooltipChannel.unshift({
      field: timeField.name,
      type: "temporal",
      title: "Time",
      format: "%b %d, %Y %H:%M",
    });
  }

  baseSpec.encoding = {
    x: { field: timeField.name, type: "temporal" },
    y: {
      field: quantitativeField.name,
      type: "quantitative",
      stack: "zero",
    },
    color: { field: nominalField?.name, type: "nominal", legend: null },
  };

  if (nominalField?.values?.length) {
    baseSpec.transform = [
      {
        calculate: `indexof([${nominalField.values
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
      mark: { type: "area", clip: true, opacity: 0.7 },
    },
    {
      mark: { type: "line", strokeWidth: 1, clip: true },
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
          condition: {
            param: "hover",
            empty: false,
            value: "var(--color-primary-300)",
          },
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
