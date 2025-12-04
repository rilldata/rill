import {
  sanitizeFieldName,
  sanitizeValueForVega,
} from "@rilldata/web-common/components/vega/util";
import type { TooltipValue } from "@rilldata/web-common/features/components/charts/types";
import type { VisualizationSpec } from "svelte-vega";
import type { Field } from "vega-lite/build/src/channeldef";
import type { LayerSpec } from "vega-lite/build/src/spec/layer";
import type { UnitSpec } from "vega-lite/build/src/spec/unit";
import type { Transform } from "vega-lite/build/src/transform";
import {
  buildHoverPointOverlay,
  buildHoverRuleLayer,
  createColorEncoding,
  createConfigWithLegend,
  createMultiLayerBaseSpec,
  createPositionEncoding,
} from "../builder";
import type { ChartDataResult } from "../types";
import type { CartesianChartSpec } from "./CartesianChartProvider";

export function generateVLMultiMetricChartSpec(
  config: CartesianChartSpec,
  data: ChartDataResult,
  markType:
    | "grouped_bar"
    | "stacked_bar"
    | "stacked_bar_normalized"
    | "stacked_area"
    | "line" = "grouped_bar",
): VisualizationSpec {
  const measureField = "Measure";
  const valueField = "value";

  const spec = createMultiLayerBaseSpec();
  const vegaConfig = createConfigWithLegend(config, config.color);

  const measures = config.y?.fields || [];

  const measureDisplayNames: Record<string, string> = {};
  measures.forEach((measure) => {
    measureDisplayNames[measure] = data.fields[measure]?.displayName || measure;
  });

  const transforms: Transform[] = [
    {
      fold: measures,
      as: [measureField, valueField],
    },
  ];
  spec.transform = transforms;

  spec.encoding = { x: createPositionEncoding(config.x, data) };

  const xField = sanitizeValueForVega(config.x?.field);

  const legend = {
    labelExpr:
      Object.entries(measureDisplayNames)
        .map(([key, value]) => `datum.value === '${key}' ? '${value}' : `)
        .join("") + "datum.value",
  };

  const baseColorEncoding = {
    ...createColorEncoding(config.color, data),
    field: measureField,
    title: measureField,
    legend,
  };

  if (typeof baseColorEncoding === "object" && "scale" in baseColorEncoding) {
    baseColorEncoding.scale!.domain = measures;
  }

  const baseYEncoding = {
    field: valueField,
    type: "quantitative" as const,
    title: "Value",
    axis: {
      ...(!config.y?.showAxisTitle && { title: null }),
    },
    scale: {
      ...(config.y?.zeroBasedOrigin !== true && { zero: false }),
      ...(config.y?.min !== undefined && { domainMin: config.y.min }),
      ...(config.y?.max !== undefined && { domainMax: config.y.max }),
    },
  };

  const sumYEncoding = {
    aggregate: "sum" as const,
    ...baseYEncoding,
  };

  const stackedYEncoding = {
    ...sumYEncoding,
    stack: "zero" as const,
  };

  const normalizedYEncoding = {
    ...baseYEncoding,
    stack: "normalize" as const,
    scale: {
      zero: false,
    },
    axis: {
      title: null,
      format: ".0%",
    },
  };

  // Build multi-value tooltip for hover rule
  let multiValueTooltipChannel: TooltipValue[] | undefined;
  if (config.x && measures.length > 0) {
    multiValueTooltipChannel = [
      {
        field: xField,
        title: data.fields[config.x.field]?.displayName || config.x.field,
        type: config.x?.type === "value" ? "nominal" : config.x.type,
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
    (markType === "stacked_bar" ||
      markType === "stacked_bar_normalized" ||
      markType === "grouped_bar")
  ) {
    xBand = 0.5;
  }

  const hoverRuleLayer = buildHoverRuleLayer({
    xField,
    defaultTooltip: [],
    multiValueTooltipChannel,
    primaryColor: data.theme.primary,
    xBand: xBand,
    isDarkMode: data.isDarkMode,
    pivot:
      xField && measures.length && multiValueTooltipChannel?.length
        ? { field: measureField, value: valueField, groupby: [xField] }
        : undefined,
    isBarMark:
      markType === "stacked_bar" ||
      markType === "stacked_bar_normalized" ||
      markType === "grouped_bar",
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
            y: stackedYEncoding,
            color: baseColorEncoding,
          },
        },
      ];
      break;
    }
    case "stacked_bar_normalized": {
      spec.layer = [
        hoverRuleLayer,
        {
          mark: { type: "bar", clip: true, width: { band: 0.9 } },
          encoding: {
            y: normalizedYEncoding,
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
