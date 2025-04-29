import { sanitizeValueForVega } from "@rilldata/web-common/components/vega/util";
import { ScrubBoxColor } from "@rilldata/web-common/features/dashboards/time-series/chart-colors";
import type { ChartField } from "./build-template";
import { singleLayerBaseSpec } from "./utils";

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
    {
      calculate: `datum.key === 'comparison_ts' ? 1 : 0`,
      as: "sortOrder",
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
          value: 0.3,
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
      sort: { field: "sortOrder" },
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
  ];

  return baseSpec;
}
