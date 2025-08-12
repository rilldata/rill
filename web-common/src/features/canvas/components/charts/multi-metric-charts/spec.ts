import {
  sanitizeFieldName,
  sanitizeValueForVega,
} from "@rilldata/web-common/components/vega/util";
import type { TooltipValue } from "@rilldata/web-common/features/canvas/components/charts/types";
import type { VisualizationSpec } from "svelte-vega";
import type { Field } from "vega-lite/build/src/channeldef";
import type { LayerSpec } from "vega-lite/build/src/spec/layer";
import type { UnitSpec } from "vega-lite/build/src/spec/unit";
import {
  buildHoverPointOverlay,
  buildHoverRuleLayer,
  createConfigWithLegend,
  createMultiLayerBaseSpec,
  createPositionEncoding,
} from "../builder";
import type { ChartDataResult } from "../types";
import type { MultiMetricChartSpec } from "./MultiMetricChart";

export function generateVLMultiMetricChartSpec(
  config: MultiMetricChartSpec,
  data: ChartDataResult,
): VisualizationSpec {
  const measureField = "Measure";
  const valueField = "value";

  const spec = createMultiLayerBaseSpec();
  const vegaConfig = createConfigWithLegend(config, {
    field: measureField,
    type: "nominal",
  });

  // Fold measures into a long format for easier encoding
  const measures = config.measures || [];

  const measureDisplayNames: Record<string, string> = {};
  measures.forEach((measure) => {
    measureDisplayNames[measure] = data.fields[measure]?.displayName || measure;
  });

  spec.transform = [
    {
      fold: measures,
      as: [measureField, valueField],
    },
  ];

  spec.encoding = { x: createPositionEncoding(config.x, data) };

  const markType = config.mark_type || "stacked_bar";
  const xField = sanitizeValueForVega(config.x?.field);

  const legend = {
    labelExpr:
      Object.entries(measureDisplayNames)
        .map(([key, value]) => `datum.value === '${key}' ? '${value}' : `)
        .join("") + "datum.value",
  };

  const baseColorEncoding = {
    field: measureField,
    type: "nominal" as const,
    legend,
  };

  const baseYEncoding = {
    field: valueField,
    type: "quantitative" as const,
    title: null,
  };

  const sumYEncoding = {
    aggregate: "sum" as const,
    ...baseYEncoding,
  };

  const stackedYEncoding = {
    ...sumYEncoding,
    stack: "zero" as const,
  };

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

  let xBand: number | undefined = undefined;

  if (
    config.x?.type === "temporal" &&
    (markType === "stacked_bar" || markType === "grouped_bar")
  ) {
    xBand = 0.5;
  }

  const hoverRuleLayer = buildHoverRuleLayer({
    xField,
    defaultTooltip: [],
    multiValueTooltipChannel,
    primaryColor: data.theme.primary,
    xBand: xBand,
    pivot:
      xField && measures.length && multiValueTooltipChannel?.length
        ? { field: measureField, value: valueField, groupby: [xField] }
        : undefined,
  });

  const hoverPointLayer = buildHoverPointOverlay();

  switch (markType) {
    case "line": {
      const layers: Array<LayerSpec<Field> | UnitSpec<Field>> = [
        {
          encoding: {
            y: baseYEncoding,
            color: baseColorEncoding,
          },
          layer: [{ mark: { type: "line", clip: true } }, hoverPointLayer],
        },
        hoverRuleLayer,
      ];
      spec.layer = layers;
      break;
    }
    case "stacked_area": {
      const layers: Array<LayerSpec<Field> | UnitSpec<Field>> = [
        {
          encoding: {
            y: stackedYEncoding,
            color: baseColorEncoding,
          },
          layer: [
            { mark: { type: "area", clip: true } },
            { mark: { type: "line", opacity: 0.5 } },
            hoverPointLayer,
          ],
        },
        hoverRuleLayer,
      ];
      spec.layer = layers;
      break;
    }
    case "stacked_bar": {
      spec.layer = [
        hoverRuleLayer,
        {
          mark: { type: "bar", clip: true, width: { band: 0.9 } },
          encoding: {
            y: sumYEncoding,
            color: baseColorEncoding,
          },
        },
      ];
      break;
    }
    case "grouped_bar": {
      spec.layer = [
        hoverRuleLayer,
        {
          mark: { type: "bar", clip: true, width: { band: 0.9 } },
          encoding: {
            y: sumYEncoding,
            xOffset: { field: measureField },
            color: baseColorEncoding,
          },
        },
      ];
      break;
    }
  }

  return {
    ...spec,
    ...(vegaConfig && { config: vegaConfig }),
  };
}
