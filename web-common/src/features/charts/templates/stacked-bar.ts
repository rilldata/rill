import { singleLayerBaseSpec } from "./utils";

export function buildStackedBar(
  timeField: string,
  quantitativeField: string,
  nominalField: string,
) {
  const baseSpec = singleLayerBaseSpec();

  baseSpec.mark = {
    type: "bar",
    width: { band: 0.75 },
    clip: true,
  };
  baseSpec.encoding = {
    x: { field: timeField, type: "temporal", bandPosition: 0 },
    y: { field: quantitativeField, type: "quantitative" },
    opacity: {
      condition: { param: "hover", empty: false, value: 1 },
      value: 0.8,
    },
    color: {
      field: nominalField,
      type: "nominal",
      legend: null,
    },
    tooltip: [
      {
        field: timeField,
        type: "temporal",
        title: "Time",
        format: "%b %d, %Y %H:%M",
      },
      { field: quantitativeField, type: "quantitative" },
      { field: nominalField, type: "nominal" },
    ],
  };

  baseSpec.params = [
    {
      name: "hover",
      select: {
        type: "point",
        on: "pointerover",
        encodings: ["color"],
      },
    },
  ];

  return baseSpec;
}
