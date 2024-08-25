import { ChartField } from "./build-template";
import { sanitizeValueForVega, singleLayerBaseSpec } from "./utils";

export function buildGroupedComparisonBar(
  timeFields: ChartField[],
  quantitativeFields: ChartField[],
  nominalField: ChartField,
) {
  const baseSpec = singleLayerBaseSpec();

  const primaryTimeField = timeFields[0];
  const measureName = sanitizeValueForVega(quantitativeFields[0].name);
  const nominalFieldName = sanitizeValueForVega(nominalField.name);

  baseSpec.transform = [
    // Sanitize and transform comparison data in the right time format
    {
      timeUnit: "yearmonthdate",
      field: "comparison\\.ts",
      as: "comparison_ts",
    },
    // Expand datum to have a key field to differentiate between current and comparison data
    { fold: ["ts", "comparison_ts"], as: ["key", "value"] },
    // Add a measure field to hold the right measure value
    {
      calculate: `(datum['key'] === 'comparison_ts' ? datum['comparison.${measureName}'] : datum['${measureName}'])`,
      as: "measure",
    },
    {
      calculate: `(datum['key'] === 'comparison_ts' ? datum['${nominalFieldName}'] + 'Comparison' : datum['${nominalFieldName}'])`,
      as: "nominalField",
    },
    // Add a time field to hold the right time value
    {
      calculate:
        "(datum['key'] === 'comparison_ts' ? datum['comparison_ts'] : datum['ts'])",
      as: "time",
    },
  ];

  baseSpec.mark = {
    type: "bar",
    width: { band: 1 },
    clip: true,
  };

  baseSpec.encoding = {
    x: { field: primaryTimeField.name, type: "temporal", bandPosition: 0 },
    y: { field: "measure", type: "quantitative" },
    opacity: {
      condition: [
        {
          param: "hover",
          empty: false,
          value: 1,
        },
        {
          test: `datum.key === 'ts'`,
          value: 0.8,
        },
        {
          test: `datum.key === 'comparison_ts'`,
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
      field: "nominalField",
    },
    tooltip: [
      {
        field: "time",
        type: "temporal",
        title: "Time",
        format: "%b %d, %Y %H:%M",
      },
      {
        title: quantitativeFields[0].label,
        field: "measure",
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
