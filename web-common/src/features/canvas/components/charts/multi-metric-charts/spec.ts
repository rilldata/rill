import type { VisualizationSpec } from "svelte-vega";
import {
  createMultiLayerBaseSpec,
  createPositionEncoding,
  createConfigWithLegend,
} from "../builder";
import type { ChartDataResult } from "../types";
import type { MultiMetricChartSpec } from "./MultiMetricChart";
import {
  sanitizeFieldName,
  sanitizeValueForVega,
} from "@rilldata/web-common/components/vega/util";
import type { TooltipValue } from "@rilldata/web-common/features/canvas/components/charts/types";

export function generateVLMultiMetricChartSpec(
  config: MultiMetricChartSpec,
  data: ChartDataResult,
): VisualizationSpec {
  const spec = createMultiLayerBaseSpec();
  const vegaConfig = createConfigWithLegend(config, {
    field: "Measure",
    type: "nominal",
  });

  // Fold measures into a long format for easier encoding
  const measures = config.measures || [];
  spec.transform = [
    {
      fold: measures,
      as: ["Measure", "value"],
    },
  ];

  // X encoding
  spec.encoding = { x: createPositionEncoding(config.x, data) };

  const markType = config.mark_type || "grouped_bar";
  const xField = sanitizeValueForVega(config.x?.field);

  // Create measure display names map
  const measureDisplayNames: Record<string, string> = {};
  measures.forEach((measure) => {
    measureDisplayNames[measure] = data.fields[measure]?.displayName || measure;
  });

  // Build multi-value tooltip for hover rule
  let multiValueTooltipChannel: TooltipValue[] | undefined;
  if (config.x && measures.length > 0) {
    multiValueTooltipChannel = [
      {
        field: xField,
        title: data.fields[config.x.field]?.displayName || config.x.field,
        type: config.x?.type,
        ...(config.x.type === "temporal" && { format: "%b %d, %Y %H:%M" }),
      },
    ];

    measures.forEach((measure) => {
      multiValueTooltipChannel!.push({
        field: sanitizeValueForVega(measure),
        title: measureDisplayNames[measure],
        type: "quantitative",
        formatType: sanitizeFieldName(measure),
      });
    });

    multiValueTooltipChannel = multiValueTooltipChannel.slice(0, 50);
  }

  // Default tooltip with proper formatting
  const defaultTooltip: TooltipValue[] = [
    ...(config.x
      ? [
          {
            field: xField,
            title: data.fields[config.x.field]?.displayName || config.x.field,
            type: config.x.type,
            ...(config.x.type === "temporal" && { format: "%b %d, %Y %H:%M" }),
          },
        ]
      : []),
    {
      field: "Measure",
      type: "nominal",
      title: "Metric",
    },
    {
      field: "value",
      type: "quantitative",
      title: "Value",
    },
  ];

  if (markType === "line") {
    spec.layer = [
      {
        encoding: {
          y: {
            field: "value",
            type: "quantitative",
            title: "Value",
          },
          color: {
            field: "Measure",
            type: "nominal",
            legend: {
              labelExpr:
                Object.entries(measureDisplayNames)
                  .map(
                    ([key, value]) =>
                      `datum.value === '${key}' ? '${value}' : `,
                  )
                  .join("") + "datum.value",
            },
          },
        },
        layer: [
          { mark: { type: "line", clip: true } },
          {
            transform: [{ filter: { param: "hover", empty: false } }],
            mark: {
              type: "point",
              filled: true,
              opacity: 1,
              size: 50,
              clip: true,
              stroke: "white",
              strokeWidth: 1,
            },
          },
        ],
      },
      {
        transform:
          xField && measures.length && multiValueTooltipChannel?.length
            ? [
                {
                  pivot: "Measure",
                  value: "value",
                  groupby: [xField],
                },
              ]
            : [],
        mark: {
          type: "rule",
          clip: true,
        },
        encoding: {
          x: {
            field: xField,
          },
          color: {
            condition: [
              {
                param: "hover",
                empty: false,
                value: "var(--color-primary-300)",
              },
            ],
            value: "transparent",
          },
          tooltip: multiValueTooltipChannel?.length
            ? multiValueTooltipChannel
            : defaultTooltip,
        },
        params: [
          {
            name: "hover",
            select: {
              type: "point",
              encodings: ["x"],
              nearest: true,
              on: "pointerover",
              clear: "pointerout",
            },
          },
        ],
      },
    ];
  } else if (markType === "stacked_area") {
    spec.layer = [
      {
        encoding: {
          y: {
            aggregate: "sum",
            field: "value",
            type: "quantitative",
            stack: "zero",
            title: "Value",
          },
          color: {
            field: "Measure",
            type: "nominal",
            legend: {
              labelExpr:
                Object.entries(measureDisplayNames)
                  .map(
                    ([key, value]) =>
                      `datum.value === '${key}' ? '${value}' : `,
                  )
                  .join("") + "datum.value",
            },
          },
        },
        layer: [
          { mark: { type: "area", clip: true } },
          { mark: { type: "line", opacity: 0.5 } },
          {
            transform: [{ filter: { param: "hover", empty: false } }],
            mark: {
              type: "point",
              filled: true,
              opacity: 1,
              size: 50,
              clip: true,
              stroke: "white",
              strokeWidth: 1,
            },
          },
        ],
      },
      {
        transform:
          xField && measures.length && multiValueTooltipChannel?.length
            ? [
                {
                  pivot: "Measure",
                  value: "value",
                  groupby: [xField],
                },
              ]
            : [],
        mark: {
          type: "rule",
          clip: true,
        },
        encoding: {
          x: {
            field: xField,
          },
          color: {
            condition: [
              {
                param: "hover",
                empty: false,
                value: "var(--color-primary-300)",
              },
            ],
            value: "transparent",
          },
          tooltip: multiValueTooltipChannel?.length
            ? multiValueTooltipChannel
            : defaultTooltip,
        },
        params: [
          {
            name: "hover",
            select: {
              type: "point",
              encodings: ["x"],
              nearest: true,
              on: "pointerover",
              clear: "pointerout",
            },
          },
        ],
      },
    ];
  } else if (markType === "stacked_bar") {
    spec.layer = [
      {
        mark: { type: "bar", clip: true },
        encoding: {
          y: {
            aggregate: "sum",
            field: "value",
            type: "quantitative",
            title: "Value",
          },
          color: {
            field: "Measure",
            type: "nominal",
            legend: {
              labelExpr:
                Object.entries(measureDisplayNames)
                  .map(
                    ([key, value]) =>
                      `datum.value === '${key}' ? '${value}' : `,
                  )
                  .join("") + "datum.value",
            },
          },
          tooltip: defaultTooltip,
        },
      },
    ];
  } else if (markType === "grouped_bar") {
    spec.layer = [
      {
        mark: { type: "bar", clip: true },
        encoding: {
          y: {
            aggregate: "sum",
            field: "value",
            type: "quantitative",
            title: "Value",
          },
          xOffset: { field: "Measure" },
          color: {
            field: "Measure",
            type: "nominal",
            legend: {
              labelExpr:
                Object.entries(measureDisplayNames)
                  .map(
                    ([key, value]) =>
                      `datum.value === '${key}' ? '${value}' : `,
                  )
                  .join("") + "datum.value",
            },
          },
          tooltip: defaultTooltip,
        },
      },
    ];
  }

  return {
    ...spec,
    ...(vegaConfig && { config: vegaConfig }),
  };
}
