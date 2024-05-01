import { ChartField } from "./build-template";
import { singleLayerBaseSpec } from "./utils";

export function buildStackedBar(
  timeField: ChartField,
  quantitativeField: ChartField,
  nominalField: ChartField,
) {
  const baseSpec = singleLayerBaseSpec();

  baseSpec.mark = {
    type: "bar",
    width: { band: 0.75 },
    clip: true,
  };
  baseSpec.encoding = {
    x: { field: timeField.name, type: "temporal", bandPosition: 0 },
    y: { field: quantitativeField.name, type: "quantitative" },
    opacity: {
      condition: { param: "hover", empty: false, value: 1 },
      value: 0.8,
    },
    color: {
      field: nominalField.name,
      type: "nominal",
      legend: null,
    },
    tooltip: [
      {
        field: timeField.tooltipName ? timeField.tooltipName : timeField.name,
        type: "temporal",
        title: "Time",
        format: "%b %d, %Y %H:%M",
      },
      {
        title: quantitativeField.label,
        field: quantitativeField.name,
        formatType: "measureFormatter",
        type: "quantitative",
      },
      { title: nominalField.label, field: nominalField.name, type: "nominal" },
    ],
  };

  baseSpec.params = [
    {
      name: "hover",
      select: {
        type: "point",
        on: "pointerover",
        encodings: ["x", "color"],
      },
    },
  ];

  return baseSpec;
}
