import { ChartField } from "./build-template";
import { repeatedLayerBaseSpec } from "./utils";

export function buildMultiMeasureGroupedBar(
  timeField: ChartField,
  quantitativeField: ChartField[],
) {
  const measures = quantitativeField.map((field) => field.name);
  const baseSpec = repeatedLayerBaseSpec(measures);

  baseSpec.spec = {
    mark: {
      type: "bar",
      width: { band: 1 },
      clip: true,
    },
    encoding: {
      x: { field: timeField.name, type: "temporal", bandPosition: 0 },
      y: { field: { repeat: "layer" }, type: "quantitative" },
      opacity: {
        condition: { param: "hover", empty: false, value: 1 },
        value: 0.8,
      },
      color: { datum: { repeat: "layer" }, legend: null },
      xOffset: { datum: { repeat: "layer" } },
      tooltip: [
        {
          field: timeField.name,
          type: "temporal",
          title: "Time",
          format: "%b %d, %Y %H:%M",
        },
      ],
    },
  };

  baseSpec.params = [
    {
      name: "hover",
      select: {
        type: "point",
        on: "pointerover",
        encodings: ["x", "color"],
        nearest: true,
      },
    },
  ];

  return baseSpec;
}
