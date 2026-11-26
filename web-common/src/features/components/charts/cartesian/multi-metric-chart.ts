import {
  sanitizeFieldName,
  sanitizeValueForVega,
} from "@rilldata/web-common/components/vega/util";
import { SortOrderField } from "@rilldata/web-common/features/components/charts/comparison-builder";
import type { TooltipValue } from "@rilldata/web-common/features/components/charts/types";
import { ComparisonDeltaPreviousSuffix } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
import { FormatPreset } from "@rilldata/web-common/lib/number-formatting/humanizer-types";
import type { MetricsViewSpecMeasure } from "@rilldata/web-common/runtime-client";
import { formatSpecifier } from "d3-format";
import type { VisualizationSpec } from "svelte-vega";
import type { ExprRef, SignalRef } from "vega";
import type { Config } from "vega-lite";
import type {
  Field,
  NumericMarkPropDef,
  OffsetDef,
} from "vega-lite/types_unstable/channeldef.js";
import type { Encoding } from "vega-lite/types_unstable/encoding.js";
import type { LayerSpec } from "vega-lite/types_unstable/spec/layer.js";
import type { UnitSpec } from "vega-lite/types_unstable/spec/unit.js";
import type { Transform } from "vega-lite/types_unstable/transform.js";
import {
  buildHoverPointOverlay,
  buildHoverRuleLayer,
  centerGroupedBars,
  createColorEncoding,
  createConfigWithLegend,
  createMultiLayerBaseSpec,
  createPositionEncoding,
} from "../builder";
import type { ChartDataResult, ChartFieldsMap } from "../types";
import type { CartesianChartSpec } from "./CartesianChartProvider";
import {
  isHorizontal,
  toVerticalSpec,
  transposeCartesianSpec,
} from "./orientation";

export function generateVLMultiMetricChartSpec(
  chartConfig: CartesianChartSpec,
  data: ChartDataResult,
  markType:
    | "grouped_bar"
    | "stacked_bar"
    | "stacked_bar_normalized"
    | "stacked_area"
    | "line" = "grouped_bar",
): VisualizationSpec {
  // The spec is built for the vertical layout and, for the bar variants,
  // transposed at the end when the measures sit on x.
  const horizontal =
    isHorizontal(chartConfig) &&
    (markType === "grouped_bar" ||
      markType === "stacked_bar" ||
      markType === "stacked_bar_normalized");
  const config = toVerticalSpec(chartConfig);

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
      fold: fieldsToFold.map((field) => sanitizeValueForVega(field)),
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

  // Axis label layout depends on the channel the axis ends up on.
  spec.encoding = {
    x: {
      ...createPositionEncoding(
        config.x,
        data,
        horizontal ? "y" : "x",
        vegaConfig,
      ),
      bandPosition: 0,
    },
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
      // An axis format in the encoding would override the chart's own `vl_config`.
      ...(!formatsAxisLabels(vegaConfig, horizontal ? "x" : "y") &&
        createSharedAxisFormat(measures, data.fields)),
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
    xField: config.x?.field,
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
    case "stacked_bar":
    case "stacked_bar_normalized": {
      const stackedBarEncoding: Partial<Encoding<Field>> = {
        y: markType === "stacked_bar" ? stackedYEncoding : normalizedYEncoding,
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
        centerGroupedBars(
          {
            mark: { type: "bar", clip: true, width: { band: 0.9 } },
            encoding: stackedBarEncoding,
          },
          vegaConfig,
        ),
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
        centerGroupedBars(
          {
            mark: { type: "bar", clip: true, width: { band: 0.9 } },
            encoding: barEncoding,
          },
          vegaConfig,
        ),
      ];
      break;
    }
  }

  const result: VisualizationSpec = {
    ...spec,
    ...(vegaConfig && { config: vegaConfig }),
  };

  return horizontal ? transposeCartesianSpec(result) : result;
}

/**
 * Whether a chart's Vega-Lite config sets the tick labels of the quantitative axis
 * on the given channel: a number format, or a format or label expression in any of
 * the axis config blocks that Vega-Lite applies to that axis.
 */
function formatsAxisLabels(
  vegaConfig: Config<ExprRef | SignalRef> | undefined,
  channel: "x" | "y",
): boolean {
  if (!vegaConfig) return false;
  if (
    vegaConfig.numberFormat !== undefined ||
    vegaConfig.numberFormatType !== undefined
  ) {
    return true;
  }
  const axisConfigs =
    channel === "x"
      ? [
          vegaConfig.axis,
          vegaConfig.axisQuantitative,
          vegaConfig.axisX,
          vegaConfig.axisXQuantitative,
          vegaConfig.axisBottom,
          vegaConfig.axisTop,
        ]
      : [
          vegaConfig.axis,
          vegaConfig.axisQuantitative,
          vegaConfig.axisY,
          vegaConfig.axisYQuantitative,
          vegaConfig.axisLeft,
          vegaConfig.axisRight,
        ];
  return axisConfigs.some(
    (axis) =>
      axis?.format !== undefined ||
      axis?.formatType !== undefined ||
      axis?.labelExpr !== undefined,
  );
}

// SI prefixes below one, from yocto to milli, as d3 writes them.
const SI_FRACTION_PREFIXES = "yzafpn\u00b5m";

/**
 * Tick label format for the value axis that the measures share.
 *
 * The axis has no single measure to take a formatter from, so its labels come from
 * the theme's SI number format, which writes fractions with milli and micro prefixes:
 * a 90% share reads "900m", and the labels change style once a value reaches 1.
 * Percentages are the one format that scales the number itself, so when every measure
 * is one the axis is formatted in percent; leaving the precision out lets Vega pick it
 * from the tick step, which keeps neighboring ticks distinct at any magnitude.
 * Other axes keep the theme format, except that a label with a fractional SI prefix
 * is written as a plain decimal instead.
 */
function createSharedAxisFormat(
  measures: string[],
  fields: ChartFieldsMap,
): { format: string } | { labelExpr: string } {
  if (
    measures.length > 0 &&
    measures.every((measure) => isPercentage(fields[measure]))
  ) {
    return { format: ",%" };
  }
  // The tick's value as JavaScript writes it, which keeps every digit that sets
  // it apart from its neighbors, with the minus sign d3 gives the other labels.
  const plainDecimal = `(datum.value < 0 ? '\u2212' : '') + toString(abs(datum.value))`;
  return {
    labelExpr: `indexof('${SI_FRACTION_PREFIXES}', slice(datum.label, -1)) >= 0 ? ${plainDecimal} : datum.label`,
  };
}

/**
 * Whether a measure's values are shown multiplied by 100, as percentages.
 * A d3 format wins over the preset, as it does in `createMeasureValueFormatter`.
 */
function isPercentage(field: ChartFieldsMap[string]): boolean {
  const measure = field as MetricsViewSpecMeasure | undefined;
  if (!measure?.formatD3) {
    return measure?.formatPreset === FormatPreset.PERCENTAGE;
  }
  try {
    return ["%", "p"].includes(formatSpecifier(measure.formatD3).type);
  } catch {
    // The value formatter ignores a format that d3 cannot parse, preset included.
    return false;
  }
}
