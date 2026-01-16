import {
  sanitizeFieldName,
  sanitizeValueForVega,
} from "@rilldata/web-common/components/vega/util";
import { SortOrderField } from "@rilldata/web-common/features/components/charts/comparison-builder";
import type { TooltipValue } from "@rilldata/web-common/features/components/charts/types";
import { ComparisonDeltaPreviousSuffix } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
import type { VisualizationSpec } from "svelte-vega";
import type {
  Field,
  NumericMarkPropDef,
  OffsetDef,
} from "vega-lite/build/src/channeldef";
import type { Encoding } from "vega-lite/build/src/encoding";
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
  const measureNormalizedField = "Measure_normalized";

  const spec = createMultiLayerBaseSpec();
  const vegaConfig = createConfigWithLegend(config, config.color);

  const measures = config.y?.fields || [];

  const measureDisplayNames: Record<string, string> = {};
  measures.forEach((measure) => {
    measureDisplayNames[measure] = data.fields[measure]?.displayName || measure;
  });

  // Check if comparison mode is enabled
  const hasComparison = data.hasComparison;

  // Build the list of fields to fold
  // In comparison mode, include both current and previous measures
  const fieldsToFold: string[] = [];
  if (hasComparison) {
    measures.forEach((measure) => {
      fieldsToFold.push(measure);
      fieldsToFold.push(measure + ComparisonDeltaPreviousSuffix);
    });
  } else {
    fieldsToFold.push(...measures);
  }

  const transforms: Transform[] = [
    {
      fold: fieldsToFold,
      as: [measureField, valueField],
    },
  ];

  // Add comparison-specific transforms
  if (hasComparison) {
    // Create a normalized measure name field (without _prev suffix)
    transforms.push({
      calculate: `replace(datum['${measureField}'], '${ComparisonDeltaPreviousSuffix}', '')`,
      as: measureNormalizedField,
    });

    // Create a period field for stacked bar xOffset
    transforms.push({
      calculate: `indexof(datum['${measureField}'], '${ComparisonDeltaPreviousSuffix}') >= 0 ? 'comparison' : 'current'`,
      as: "period",
    });

    // Create a sort order field to ensure current appears before comparison using period
    transforms.push({
      calculate: `datum['period'] === 'comparison' ? 1 : 0`,
      as: SortOrderField,
    });
  }

  spec.transform = transforms;

  spec.encoding = {
    x: { ...createPositionEncoding(config.x, data), bandPosition: 0 },
  };

  const xField = sanitizeValueForVega(config.x?.field);

  const legend = {
    labelExpr:
      Object.entries(measureDisplayNames)
        .map(([key, value]) => `datum.value === '${key}' ? '${value}' : `)
        .join("") + "datum.value",
  };

  // In comparison mode, use the normalized measure name for coloring
  // so both current and comparison bars have the same color per measure
  const colorField = hasComparison ? measureNormalizedField : measureField;

  const baseColorEncoding = {
    ...createColorEncoding(config.color, data),
    field: colorField,
    title: measureField,
    legend,
  };

  const opacityComparisonEncodingMeasure: NumericMarkPropDef<Field> = {
    condition: [
      {
        test: `indexof(datum['${measureField}'], '${ComparisonDeltaPreviousSuffix}') >= 0`,
        value: 0.4,
      },
    ],
    value: 1,
  };

  const opacityComparisonEncodingPeriod: NumericMarkPropDef<Field> = {
    condition: [
      {
        test: "datum.period === 'comparison'",
        value: 0.4,
      },
    ],
    value: 1,
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

    if (hasComparison) {
      measures.forEach((measure) => {
        // Current period value
        multiValueTooltipChannel!.push({
          field: sanitizeValueForVega(measure),
          title: measureDisplayNames[measure],
          type: "quantitative",
          formatType: sanitizeFieldName(measure),
        });

        // Comparison period value
        multiValueTooltipChannel!.push({
          field: sanitizeValueForVega(measure) + ComparisonDeltaPreviousSuffix,
          title: measureDisplayNames[measure] + ComparisonDeltaPreviousSuffix,
          type: "quantitative",
          formatType: sanitizeFieldName(measure),
        });
      });
    } else {
      measures.forEach((measure) => {
        multiValueTooltipChannel!.push({
          field: sanitizeValueForVega(measure),
          title: measureDisplayNames[measure],
          type: "quantitative",
          formatType: sanitizeFieldName(measure),
        });
      });
    }

    multiValueTooltipChannel = multiValueTooltipChannel.slice(0, 50);
  }

  const hoverRuleLayer = buildHoverRuleLayer({
    xField,
    defaultTooltip: [],
    multiValueTooltipChannel,
    primaryColor: data.theme.primary,
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
      const lineLayerEncoding: Partial<Encoding<Field>> = {
        y: baseYEncoding,
        color: baseColorEncoding,
      };

      if (hasComparison) {
        lineLayerEncoding.detail = {
          field: measureField,
          type: "nominal",
        };
        lineLayerEncoding.opacity = opacityComparisonEncodingPeriod;
      }

      const layers: Array<LayerSpec<Field> | UnitSpec<Field>> = [
        {
          encoding: lineLayerEncoding,
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
      const stackedBarEncoding: Partial<Encoding<Field>> = {
        y: stackedYEncoding,
        color: baseColorEncoding,
      };

      if (hasComparison) {
        const xOffset: OffsetDef<Field> = {
          field: "period",
          sort: { field: SortOrderField },
        };
        stackedBarEncoding.xOffset = xOffset;
        stackedBarEncoding.opacity = opacityComparisonEncodingPeriod;
      }

      spec.layer = [
        hoverRuleLayer,
        {
          mark: { type: "bar", clip: true, width: { band: 0.9 } },
          encoding: stackedBarEncoding,
        },
      ];
      break;
    }
    case "stacked_bar_normalized": {
      const normalizedBarEncoding: Partial<Encoding<Field>> = {
        y: normalizedYEncoding,
        color: baseColorEncoding,
      };

      if (hasComparison) {
        const xOffset: OffsetDef<Field> = {
          field: "period",
          sort: { field: SortOrderField },
        };
        normalizedBarEncoding.xOffset = xOffset;
        normalizedBarEncoding.opacity = opacityComparisonEncodingPeriod;
      }

      spec.layer = [
        hoverRuleLayer,
        {
          mark: { type: "bar", clip: true, width: { band: 0.9 } },
          encoding: normalizedBarEncoding,
        },
      ];
      break;
    }
    case "grouped_bar": {
      const barEncoding: Partial<Encoding<Field>> = {
        y: sumYEncoding,
        color: baseColorEncoding,
      };

      if (hasComparison) {
        const xOffset: OffsetDef<Field> = {
          field: measureField,
          sort: { field: SortOrderField },
        };
        barEncoding.xOffset = xOffset;
        barEncoding.opacity = opacityComparisonEncodingMeasure;
      } else {
        // Normal mode: group by measure only
        const xOffset: OffsetDef<Field> = { field: measureField };
        barEncoding.xOffset = xOffset;
      }

      spec.layer = [
        hoverRuleLayer,
        {
          mark: { type: "bar", clip: true, width: { band: 0.9 } },
          encoding: barEncoding,
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
