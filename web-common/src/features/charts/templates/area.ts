import { ChartField } from "./build-template";
import { multiLayerBaseSpec } from "./utils";

export function buildArea(
  timeField: ChartField,
  quantitativeField: ChartField,
) {
  const baseSpec = multiLayerBaseSpec();

  baseSpec.encoding = {
    x: { field: timeField.name, type: "temporal" },
    y: {
      field: quantitativeField.name,
      type: "quantitative",
    },
  };
  baseSpec.layer = [
    {
      mark: { type: "area", clip: true },
    },
    {
      transform: [{ filter: { param: "hover", empty: false } }],
      mark: { type: "rule", color: "royalblue", strokeWidth: 3 },
    },
    {
      mark: { type: "point", filled: true, size: 50 },
      encoding: {
        opacity: {
          condition: { param: "hover", value: 1, empty: false },
          value: 0,
        },
        tooltip: [
          {
            field: timeField.name,
            type: "temporal",
            title: "Time",
            format: "%b %d, %Y %H:%M",
          },
          {
            title: quantitativeField.label,
            field: quantitativeField.name,
            type: "quantitative",
          },
        ],
      },

      params: [
        {
          name: "hover",
          select: {
            type: "point",
            encodings: ["x"],
            nearest: true,
            on: "mouseover",
          },
        },
      ],
    },
  ];

  return baseSpec;
}
