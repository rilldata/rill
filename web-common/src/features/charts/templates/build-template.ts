import { ChartTypes } from "@rilldata/web-common/features/charts/types";
import { VisualizationSpec } from "svelte-vega";

const BAR_LIKE_CHARTS = [
  ChartTypes.BAR,
  ChartTypes.GROUPED_BAR,
  ChartTypes.STACKED_BAR,
];
const LINE_LIKE_CHARTS = [
  ChartTypes.LINE,
  ChartTypes.AREA,
  ChartTypes.STACKED_AREA,
];

function singleLayerBaseSpec() {
  const baseSpec: VisualizationSpec = {
    $schema: "https://vega.github.io/schema/vega-lite/v5.json",
    description: "A simple single layered chart with embedded data.",
    width: "container",
    data: { name: "table" },
    mark: "point",
  };

  return baseSpec;
}

function multiLayerBaseSpec() {
  const baseSpec: VisualizationSpec = {
    $schema: "https://vega.github.io/schema/vega-lite/v5.json",
    width: "container",
    data: { name: "table" },
    layer: [],
  };

  return baseSpec;
}

export function buildVegaLiteSpec(
  chartType: ChartTypes,
  timeFields: string[],
  quantitativeFields: string[],
  nominalFields: string[] = [],
): VisualizationSpec {
  if (!timeFields.length) throw "No time fields found";
  const hasNominalFields = nominalFields.length > 0;

  if (BAR_LIKE_CHARTS.includes(chartType)) {
    const baseSpec = singleLayerBaseSpec();
    const expandBars = hasNominalFields && chartType === ChartTypes.GROUPED_BAR;

    baseSpec.mark = {
      type: "bar",
      width: { band: expandBars ? 1 : 0.75 },
      clip: true,
    };
    baseSpec.encoding = {
      x: { field: timeFields[0], type: "temporal", bandPosition: 0 },
      y: { field: quantitativeFields[0], type: "quantitative" },
      opacity: {
        condition: { param: "hover", empty: false, value: 1 },
        value: 0.8,
      },
      tooltip: [
        {
          field: timeFields[0],
          type: "temporal",
          title: "Time",
          format: "%b %d, %Y %H:%M",
        },
        { field: quantitativeFields[0], type: "quantitative" },
        { field: nominalFields[0], type: "nominal" },
      ],
    };

    baseSpec.params = [
      {
        name: "hover",
        select: {
          type: "point",
          on: "pointerover",
          fields: ["x"],
          nearest: true,
        },
      },
    ];

    if (chartType === ChartTypes.GROUPED_BAR) {
      baseSpec.encoding.color = {
        field: nominalFields[0],
        type: "nominal",
        legend: null,
      };

      baseSpec.encoding.xOffset = {
        field: nominalFields[0],
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
    } else if (chartType === ChartTypes.STACKED_BAR) {
      baseSpec.encoding.color = {
        field: nominalFields[0],
        type: "nominal",
        legend: null,
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
    }

    return baseSpec;
  } else if (LINE_LIKE_CHARTS.includes(chartType)) {
    const baseSpec = singleLayerBaseSpec();

    if (chartType == ChartTypes.AREA) {
      baseSpec.mark = { type: "area", clip: true };
      baseSpec.encoding = {
        x: { field: timeFields[0], type: "temporal" },
        y: { field: quantitativeFields[0], type: "quantitative" },
      };
    } else if (chartType == ChartTypes.LINE) {
      baseSpec.mark = { type: "line", clip: true };
      baseSpec.encoding = {
        x: { field: timeFields[0], type: "temporal" },
        y: { field: quantitativeFields[0], type: "quantitative" },
        ...(hasNominalFields && {
          color: { field: nominalFields[0], type: "nominal" },
        }),
      };
    } else if (chartType == ChartTypes.STACKED_AREA) {
      const layeredBaseSpec = multiLayerBaseSpec();

      layeredBaseSpec.layer = [
        {
          mark: { type: "area", clip: true },
          encoding: {
            x: { field: timeFields[0], type: "temporal" },
            y: {
              field: quantitativeFields[0],
              type: "quantitative",
              stack: "zero",
            },
            color: { field: nominalFields[0], type: "nominal", legend: null },
            opacity: {
              condition: { param: "hover", empty: false, value: 1 },
              value: 0.7,
            },
          },
          params: [
            {
              name: "hover",
              select: { type: "point", on: "pointerover" },
            },
          ],
        },
        {
          mark: { type: "line", strokeWidth: 1, clip: true },
          encoding: {
            x: { field: timeFields[0], type: "temporal" },
            y: {
              field: quantitativeFields[0],
              type: "quantitative",
              stack: "zero",
            },
            stroke: { field: nominalFields[0], type: "nominal", legend: null },
          },
        },
      ];

      return layeredBaseSpec;
    }

    return baseSpec;
  } else {
    throw new Error(`Chart type "${chartType}" not supported.`);
  }
}
