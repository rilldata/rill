import { ChartField } from "@rilldata/web-common/features/templates/charts/build-template";
import { singleLayerBaseSpec, multiLayerBaseSpec } from "./utils";
import { ScrubBoxColor, ScrubMutedColor, VEGA_BAR_WIDTH_SIZE } from "@rilldata/web-common/features/dashboards/time-series/chart-colors";

export function buildSimpleBarSingleLayer(
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
    x: {
      field: timeField.name,
      type: "temporal",
      bandPosition: 0,
    },
    y: { field: quantitativeField.name, type: "quantitative" },
    opacity: {
      condition: [
        {
          param: "hover",
          empty: false,
          value: 1,
        },
        {
          param: "brush",
          empty: false,
          value: 1,
        },
      ],
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
        formatType: quantitativeField.formatterFunction || "number",
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

export function buildSimpleBarMultiLayer(
  timeField: ChartField,
  quantitativeField: ChartField,
) {
  const baseSpec = multiLayerBaseSpec();

  baseSpec.layer = [
    // Full-height hover layer
    {
      params: [
        {
          name: "hover",
          select: { type: "point", on: "pointerover", clear: "pointerout" },
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
      ],
      mark: {
        type: "bar",
        color: ScrubMutedColor,
        tooltip: false,
        width: VEGA_BAR_WIDTH_SIZE,
        // width: { signal: "barWidth" },
      },
      encoding: {
        x: {
          field: timeField.name,
          type: "temporal",
          bandPosition: 0,
          axis: {
            orient: "top",
            title: null,
          },
        },
        opacity: {
          condition: { param: "hover", empty: false, value: 0.5 },
          value: 0,
        },
        detail: {
          field: quantitativeField.name,
          type: "quantitative",
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
            formatType: quantitativeField.formatterFunction || "number",
          },
        ],
      },
    },
    // Main bar layer
    {
      mark: {
        type: "bar",
        width: VEGA_BAR_WIDTH_SIZE,
        // width: { signal: "barWidth" },
      },
      encoding: {
        x: {
          field: timeField.name,
          type: "temporal",
          bandPosition: 0,
        },
        y: {
          field: quantitativeField.name,
          type: "quantitative",
          axis: {
            orient: "right",
            title: null,
            format: "~s",
          },
        },
      },
    },
  ];

  return baseSpec;
}
