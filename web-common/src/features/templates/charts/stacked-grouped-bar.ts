import { ChartField } from "./build-template";
import { sanitizeValueForVega, singleLayerBaseSpec } from "./utils";

export function buildStackedGroupedBar(
  timeFields: ChartField[],
  quantitativeFields: ChartField[],
  nominalField: ChartField,
) {
  const baseSpec = singleLayerBaseSpec();

  const primaryTimeField = timeFields[0];
  baseSpec.transform = [
    {
      fold: quantitativeFields.map((field) => sanitizeValueForVega(field.name)),
      as: ["Measure", "Value"],
    },
  ];

  baseSpec.mark = {
    type: "bar",
    width: { band: 1 },
    clip: true,
  };

  baseSpec.encoding = {
    x: { field: primaryTimeField.name, type: "temporal", bandPosition: 0 },
    y: {
      field: "Value",
      type: "quantitative",
      title: quantitativeFields[0].label,
    },
    opacity: {
      condition: [
        {
          param: "hover",
          empty: false,
          value: 1,
        },
        {
          test: `datum.Measure === '${quantitativeFields[0].name}'`,
          value: 0.8,
        },
        {
          test: `datum.Measure === '${quantitativeFields[1].name}'`,
          value: 0.4,
        },
      ],
      value: 0.8,
    },
    color: {
      field: nominalField.name,
      type: "nominal",
      legend: null,
    },
    xOffset: {
      field: "Measure",
      type: "nominal",
      title: "Measure",
      sort: null,
    },
    tooltip: [
      {
        field: primaryTimeField.tooltipName
          ? primaryTimeField.tooltipName
          : primaryTimeField.name,
        type: "temporal",
        title: "Time",
        format: "%b %d, %Y %H:%M",
      },
      {
        title: quantitativeFields[0].label,
        field: "Value",
        type: "quantitative",
        formatType: quantitativeFields[0].formatterFunction || "number",
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
        clear: "pointerout",
        encodings: ["x", "xOffset", "color"],
      },
    },
  ];

  return baseSpec;
}
