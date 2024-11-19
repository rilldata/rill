import { ScrubBoxColor } from "@rilldata/web-common/features/dashboards/time-series/chart-colors";
import type { ChartField } from "./build-template";
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
      as: ["key", "measure"],
    },
    {
      calculate: `(datum['key'] === '${quantitativeFields[0].name}' ? datum['ts'] : datum['comparison\\.ts'])`,
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
    y: {
      field: "measure",
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
          test: `datum.key === '${quantitativeFields[0].name}'`,
          value: 0.8,
        },
        {
          test: `datum.key === '${quantitativeFields[1].name}'`,
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
      field: "key",
      type: "nominal",
      title: "Measure",
      sort: null,
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
