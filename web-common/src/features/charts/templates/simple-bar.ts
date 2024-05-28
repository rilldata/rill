import { ChartField } from "@rilldata/web-common/features/charts/templates/build-template";
import { singleLayerBaseSpec } from "./utils";

export function buildSimpleBar(
  timeField: ChartField,
  quantitativeField: ChartField,
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
        type: "quantitative",
        formatType: "measureFormatter",
      },
    ],
  };

  baseSpec.params = [
    {
      name: "hover",
      select: {
        type: "point",
        on: "pointerover",
        clear: "pointerout",
        encodings: ["x"],
      },
    },
  ];

  return baseSpec;
}
