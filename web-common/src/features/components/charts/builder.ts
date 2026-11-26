import {
  sanitizeFieldName,
  sanitizeTitleForVegaTooltip,
  sanitizeValueForVega,
} from "@rilldata/web-common/components/vega/util";
import { createBrushParam } from "@rilldata/web-common/features/components/charts/brush-builder";
import type { CartesianChartSpec } from "@rilldata/web-common/features/components/charts/cartesian/CartesianChartProvider";
import type {
  ChartDataResult,
  ChartDomainValues,
  ChartLegend,
  ChartSortDirection,
  ChartSpec,
  FieldConfig,
  TooltipValue,
} from "@rilldata/web-common/features/components/charts/types";
import {
  estimateAxisLabelWidth,
  getColorForValues,
  isDomainStringArray,
  isFieldConfig,
  mergedVlConfig,
  resolveColor,
  resolveCSSVariable,
  sanitizeSortFieldForVega,
} from "@rilldata/web-common/features/components/charts/util";
import { ComparisonDeltaPreviousSuffix } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
import {
  BarHighlightColorDark,
  BarHighlightColorLight,
} from "@rilldata/web-common/features/dashboards/time-series/chart-colors";
import {
  getDivergingColorsAsHex,
  getSequentialColorsAsHex,
} from "@rilldata/web-common/features/themes/palette-store";
import type { Color } from "chroma-js";
import merge from "deepmerge";
import type { VisualizationSpec } from "svelte-vega";
import type { ExprRef, SignalRef } from "vega";
import type { Config } from "vega-lite";
import type {
  ColorDef,
  Field,
  MarkPropDef,
  OrderFieldDef,
  PositionFieldDef,
} from "vega-lite/types_unstable/channeldef.js";
import type { Encoding } from "vega-lite/types_unstable/encoding.js";
import type { SelectionParameter } from "vega-lite/types_unstable/selection.js";
import type {
  TopLevelUnitSpec,
  UnitSpec,
} from "vega-lite/types_unstable/spec/index.js";
import type { TopLevelParameter } from "vega-lite/types_unstable/spec/toplevel.js";

export function createMultiLayerBaseSpec() {
  const baseSpec: VisualizationSpec = {
    $schema: "https://vega.github.io/schema/vega-lite/v5.json",
    width: "container",
    data: { name: "metrics-view" },
    autosize: { type: "fit" },
    background: "transparent",
    layer: [],
  };
  return baseSpec;
}

export function createSingleLayerBaseSpec(
  mark: "line" | "bar" | "point" | "area" | "arc" | "rect",
): TopLevelUnitSpec<Field> {
  return {
    $schema: "https://vega.github.io/schema/vega-lite/v5.json",
    description: `A ${mark} chart with embedded data.`,
    mark: { type: mark, clip: true },
    width: "container",
    data: { name: "metrics-view" },
    autosize: { type: "fit" },
  };
}

type PositionChannel = "x" | "y" | "theta";
type AxisProps = NonNullable<PositionFieldDef<Field>["axis"]>;

/**
 * The encoding of a position channel for a field. `config` is the chart's own config, with the
 * widget's `vl_config`; the layout of a categorical x-axis's labels takes it into account.
 */
export function createPositionEncoding(
  field: FieldConfig | undefined,
  data: ChartDataResult,
  channel: PositionChannel,
  config?: Config<ExprRef | SignalRef>,
): PositionFieldDef<Field> {
  if (!field || field.type === "value") return {};
  const metaData = data.fields[field.field];
  const temporalAxisLabelExpr =
    field.type === "temporal" && metaData && "format" in metaData
      ? getTemporalAxisLabelExpr(metaData.format)
      : undefined;
  return {
    field: sanitizeValueForVega(field.field),
    title: metaData?.displayName || field.field,
    type: field.type,
    ...(metaData && "timeUnit" in metaData && { timeUnit: metaData.timeUnit }),
    ...(field.type !== "temporal" && {
      sort:
        data.domainValues?.[field.field] ??
        sanitizeSortFieldForVega(field.sort),
    }),
    ...(field.type === "quantitative" && {
      scale: {
        ...(field.zeroBasedOrigin !== true && { zero: false }),
        ...(field.min !== undefined && { domainMin: field.min }),
        ...(field.max !== undefined && { domainMax: field.max }),
      },
    }),
    axis:
      field.axisOrient === "none"
        ? null
        : {
            ...createAxisLabelLayout(field, data, channel, config),
            ...(field.axisOrient && { orient: field.axisOrient }),
            ...(field.type === "quantitative" && {
              formatType: sanitizeFieldName(field.field),
            }),
            ...(metaData &&
              "format" in metaData && {
                format: metaData.format,
              }),
            ...(temporalAxisLabelExpr && {
              labelExpr: temporalAxisLabelExpr,
            }),
            ...(!field.showAxisTitle && { title: null }),
          },
  };
}

// Padding kept around an upright label inside its band before it counts as fitting.
const UPRIGHT_LABEL_PADDING_PX = 8;
// Narrowest band on which -45° labels stay apart: the 11px line height divided by sin(45°).
const ANGLED_LABEL_MIN_BAND_PX = 16;
// Rotated labels are truncated at this share of the plot height so the marks always keep room.
// Under `autosize: fit` the `height` signal is the plot height left after the axes, so the share
// of the whole chart the labels end up with is smaller than this fraction.
const ROTATED_LABEL_MAX_HEIGHT_FRACTION = 0.5;
const ROTATED_LABEL_MIN_LIMIT_PX = 60;
// Whether the axis a label belongs to runs along the top of the chart. It is read from the axis
// when the label is drawn, because the side can be set on the field (`axisOrient`) as well as in
// the chart's or the theme's config.
const AXIS_ON_TOP = `item.mark.group.orient === "top"`;
// The sections of a chart's config that Vega-Lite reads the properties of a categorical x-axis
// from. Which of them apply to a chart depends on its scale and on the side its axis is on.
const CATEGORICAL_X_AXIS_CONFIGS = [
  "axis",
  "axisX",
  "axisTop",
  "axisBottom",
  "axisDiscrete",
  "axisBand",
  "axisPoint",
  "axisXDiscrete",
  "axisXBand",
  "axisXPoint",
] as const;

/**
 * Axis label angle, truncation and overlap handling for a position field.
 *
 * Vega-Lite rotates nominal x-axis labels to -90° by default regardless of the available width,
 * which lets the axis consume most of a short chart. For a categorical x-axis without a configured
 * `labelAngle`, the labels are instead laid out with Vega signal expressions so they react to
 * resizes without a re-embed: upright when the widest label fits inside its band, -45° when it
 * does not, and -90° only when the bands are too narrow even for angled text. Rotated labels are
 * truncated with an ellipsis at a fraction of the plot height so the marks always keep room.
 * An explicit angle is kept as is with the same truncation, so a vertical label can never take over
 * the chart. Categorical labels are never thinned: a reader cannot infer a hidden category.
 * Every label is anchored on the center of its band: an upright label is centered on it, and a
 * rotated label meets it with the end of its text nearest the axis, through the middle of the line.
 * A `labelAlign` or `labelBaseline` that `config` (the chart's own config, with the widget's
 * `vl_config`) sets for the axis is left to it, because a property on the encoding overrides it.
 * Other channels and field types only pass through an explicit `labelAngle`.
 */
export function createAxisLabelLayout(
  field: FieldConfig,
  data: ChartDataResult,
  channel: PositionChannel,
  config?: Config<ExprRef | SignalRef>,
): Partial<AxisProps> {
  const isCategoricalX =
    channel === "x" && (field.type === "nominal" || field.type === "ordinal");
  if (!isCategoricalX) {
    return field.labelAngle !== undefined
      ? { labelAngle: field.labelAngle }
      : {};
  }

  const domain =
    data.domainValues?.[field.field] ??
    Array.from(new Set(data.data.map((row) => row[field.field])));
  const labels = domain
    .slice(0, field.limit ?? domain.length)
    .map((value) => String(value));
  const band = `(width / ${Math.max(1, labels.length)})`;
  const uprightLimit = `${band} - 4`;
  const rotatedLimit = `max(${ROTATED_LABEL_MIN_LIMIT_PX}, height * ${ROTATED_LABEL_MAX_HEIGHT_FRACTION})`;
  const setInConfig = (property: "labelAlign" | "labelBaseline") =>
    CATEGORICAL_X_AXIS_CONFIGS.some(
      (section) => config?.[section]?.[property] !== undefined,
    );

  if (field.labelAngle !== undefined) {
    const upright = field.labelAngle % 180 === 0;
    return {
      labelAngle: field.labelAngle,
      // Vega-Lite aligns a label of a fixed angle on its tick. For a rotated one it may pick the
      // top or bottom baseline, which hangs the label from a corner, so only that is overridden.
      ...(!upright &&
        !setInConfig("labelBaseline") && { labelBaseline: "middle" as const }),
      labelLimit: { expr: upright ? uprightLimit : rotatedLimit },
      labelOverlap: false,
    };
  }

  const widestLabel = Math.ceil(
    Math.max(0, ...labels.map(estimateAxisLabelWidth)) +
      UPRIGHT_LABEL_PADDING_PX,
  );
  const fitsUpright = `${band} >= ${widestLabel}`;
  const fitsAngled = `${band} >= ${ANGLED_LABEL_MIN_BAND_PX}`;
  return {
    labelAngle: {
      expr: `(${fitsUpright} ? 0 : (${fitsAngled} ? -45 : -90))`,
    },
    // The anchor is spelled out because Vega-Lite's default does not hold for an angle driven by
    // a signal: the alignment it derives yields `null` for an upright label, and Vega draws a
    // text mark without an alignment from its left edge. The label then starts at the center of
    // its bar instead of being centered under it.
    ...(!setInConfig("labelAlign") && {
      labelAlign: {
        expr: `${fitsUpright} ? "center" : (${AXIS_ON_TOP} ? "left" : "right")`,
      },
    }),
    ...(!setInConfig("labelBaseline") && {
      labelBaseline: {
        expr: `${fitsUpright} ? (${AXIS_ON_TOP} ? "bottom" : "top") : "middle"`,
      },
    }),
    labelLimit: { expr: `${fitsUpright} ? ${uprightLimit} : ${rotatedLimit}` },
    labelOverlap: false,
  };
}

/**
 * Turns off Vega-Lite's implicit stacking when the axis should not be anchored at zero.
 *
 * Vega-Lite stacks bar and area marks by default, and a stacked scale takes its domain from
 * the stack's start/end fields, which always span zero. `scale.zero: false` is therefore
 * compiled in but has nothing to act on. Un-stacking is only safe for a single series: with a
 * color field the marks would overlap instead of stack, so those charts stay zero-based.
 *
 * Spread this after `createPositionEncoding` so it wins over any default `stack`.
 */
export function createStackOverride(
  field: FieldConfig | undefined,
  colorField: string | undefined,
): { stack?: null } {
  return field?.zeroBasedOrigin === false && !colorField ? { stack: null } : {};
}

/**
 * Centers the bars that share a category's band on the middle of that band, where the axis label
 * is anchored. Such bars are spread side by side by the `xOffset` channel: one per color value,
 * measure or comparison period.
 *
 * The builders size a bar as a share of its band and leave the rest as the gap to its neighbors.
 * With an offset channel, Vega-Lite takes the bar's width from its slot inside the band but the
 * inset that centers it from the whole band. Every bar is then pushed right by more than its own
 * gap, so the group sits right of its label and overhangs the band. The gap is moved into the
 * offset scale's padding instead, which draws the same bars evenly around the middle of the band.
 * It is added to the padding that `config` (the chart's own config, with the widget's `vl_config`)
 * already gives the offset scale, so the bars keep the width they had.
 *
 * A layer without an offset channel is returned as is.
 */
export function centerGroupedBars(
  barLayer: UnitSpec<Field>,
  config: Config<ExprRef | SignalRef> | undefined,
): UnitSpec<Field> {
  const { mark, encoding } = barLayer;
  const xOffset = encoding?.xOffset;
  if (typeof mark !== "object" || mark.type !== "bar") return barLayer;
  if (!xOffset || !("field" in xOffset)) return barLayer;
  const { width } = mark;
  if (typeof width !== "object" || !("band" in width)) return barLayer;

  const inner = config?.scale?.offsetBandPaddingInner ?? 0;
  const outer = config?.scale?.offsetBandPaddingOuter ?? 0;
  // Padding driven by a signal cannot be added to here; such a chart is left as it was.
  if (typeof inner !== "number" || typeof outer !== "number") return barLayer;

  // The bar's gap as a share of a slot's step: the band it is cut from is what the padding leaves.
  const gap = (1 - width.band) * (1 - inner);
  // Rounded so that 1 - 0.9 reads as 0.1 in the spec.
  const round = (value: number) => Number(value.toFixed(4));
  return {
    ...barLayer,
    mark: { ...mark, width: { band: 1 } },
    encoding: {
      ...encoding,
      xOffset: {
        ...xOffset,
        scale: {
          ...xOffset.scale,
          paddingInner: round(inner + gap),
          paddingOuter: round(outer + gap / 2),
        },
      },
    },
  };
}

export function getTemporalAxisLabelExpr(format: string | undefined) {
  if (format !== "%H:%M") return undefined;

  return [
    "timeFormat(datum.value, '%H:%M') === '00:00'",
    "? timeFormat(datum.value, '%b %d')",
    ": timeFormat(datum.value, '%H:%M')",
  ].join(" ");
}

/**
 * The legend of a measure, formatted as the measure's axis and tooltip are. Left to Vega's own
 * number format, a legend read "1.8M" for revenue and "5m" for a rate of 0.5%.
 *
 * Vega-Lite draws a legend that states anything of its own even where the chart's config turns
 * legends off, so a chart whose legend is hidden (`legendOrientation: none`) is given none.
 */
function createMeasureLegend(
  measure: string,
  legendField: FieldConfig | string | undefined,
): { legend?: { formatType: string } } {
  if (isFieldConfig(legendField) && legendField.legendOrientation === "none") {
    return {};
  }
  return { legend: { formatType: sanitizeFieldName(measure) } };
}

/**
 * The size encoding of a measure. `legendField` is the field whose `legendOrientation` places the
 * chart's legends, the size legend among them.
 */
export function createSizeEncoding(
  field: FieldConfig | undefined,
  data: ChartDataResult,
  legendField?: FieldConfig | string,
): MarkPropDef<Field, number> {
  if (!field) return {};
  const metaData = data.fields[field.field];
  return {
    field: sanitizeValueForVega(field.field),
    title: metaData?.displayName || field.field,
    type: "quantitative",
    scale: {
      zero: false,
      range: [40, 400], // Set minimum size to 40 (increased from default ~30)
    },
    ...createMeasureLegend(field.field, legendField),
  };
}

export function createColorEncoding(
  colorField: FieldConfig | string | undefined,
  data: ChartDataResult,
): ColorDef<Field> {
  if (!colorField) return {};
  if (isFieldConfig(colorField)) {
    const metaData = data.fields[colorField.field];

    const colorValues = data.domainValues?.[colorField.field];

    const baseEncoding: ColorDef<Field> = {
      field: sanitizeValueForVega(colorField.field),
      title: metaData?.displayName || colorField.field,
      type: colorField.type === "value" ? "nominal" : colorField.type,
      ...(metaData &&
        "timeUnit" in metaData && { timeUnit: metaData.timeUnit }),
      ...(colorValues?.length && { sort: colorValues }),
      ...(colorField.type === "quantitative" &&
        createMeasureLegend(colorField.field, colorField)),
    };

    // Ideally we would want to use conditional statements to set the color
    // but it's not supported by Vega-Lite yet
    // https://github.com/vega/vega-lite/issues/9497

    let colorMapping: { value: string; color: string }[] | undefined;
    if (isDomainStringArray(colorValues)) {
      colorMapping = getColorForValues(colorValues, colorField.colorMapping);
    }

    if (colorMapping?.length) {
      const domain = colorMapping.map((mapping) => mapping.value);
      // Resolve CSS variables for canvas rendering
      const range = colorMapping.map((mapping) =>
        resolveCSSVariable(mapping.color),
      );

      baseEncoding.scale = {
        domain,
        range,
        type: "ordinal",
      };
    }

    if (colorField.type === "quantitative" && colorField.colorRange) {
      const colorRange = colorField.colorRange;

      if (colorRange.mode === "scheme") {
        // Support palette scheme names
        if (colorRange.scheme === "sequential") {
          // Use our sequential palette (9 colors) as hex for Vega compatibility
          baseEncoding.scale = {
            range: getSequentialColorsAsHex(),
          };
        } else if (colorRange.scheme === "diverging") {
          // Use our diverging palette (11 colors) as hex for Vega compatibility
          baseEncoding.scale = {
            range: getDivergingColorsAsHex(),
          };
        } else {
          // Use Vega's built-in color schemes
          baseEncoding.scale = {
            scheme: colorRange.scheme,
          };
        }
      } else if (colorRange.mode === "gradient") {
        baseEncoding.scale = {
          range: [
            resolveColor(data.theme, colorRange.start),
            resolveColor(data.theme, colorRange.end),
          ],
          type: "linear",
        };
      }
    }

    return baseEncoding;
  }

  if (typeof colorField === "string") {
    const color = resolveColor(data.theme, colorField);
    return { value: color };
  }
  return {};
}

export function createOpacityEncoding(paramName: string) {
  return {
    condition: [
      { param: paramName, empty: false, value: 1 },
      {
        test: `length(data('${paramName}_store')) == 0`,
        value: 0.8,
      },
    ],
    value: 0.2,
  };
}

export function createOrderEncoding(
  field: FieldConfig | undefined,
): OrderFieldDef<Field> {
  if (!field || field.type === "value") return {};
  return {
    field: sanitizeValueForVega(field.field),
    type: field.type,
    sort: "descending",
  };
}

export function createLegendParam(
  paramName: string,
  field: string,
): TopLevelParameter {
  return {
    name: paramName,
    select: {
      type: "point",
      fields: [sanitizeValueForVega(field)],
    },
    bind: "legend",
  };
}

export function createDefaultTooltipEncoding(
  fields: Array<FieldConfig | string | undefined>,
  data: ChartDataResult,
): TooltipValue[] {
  const tooltip: TooltipValue[] = [];

  for (const field of fields) {
    if (!field) continue;

    if (typeof field === "object") {
      if (field.type === "value") continue;
      tooltip.push({
        field: sanitizeValueForVega(field.field),
        title: data.fields[field.field]?.displayName || field.field,
        type: field.type,
        ...(field.type === "quantitative" && {
          formatType: sanitizeFieldName(field.field),
        }),
        ...(field.type === "temporal" && { format: "%b %d, %Y %H:%M" }),
      });
    }
  }

  return tooltip;
}

/**
 * Where a chart's legend goes. Its columns, and how many entries and how much of each label it
 * shows, follow the chart's size: `fitLegends` lays every chart's legends out when it is rendered.
 */
export function getLegendConfig(
  orientation: ChartLegend,
): Config<ExprRef | SignalRef> {
  if (orientation === "none") {
    return {
      legend: {
        disable: true,
      },
    };
  }
  return {
    legend: {
      orient: orientation,
    },
  };
}

export function createConfigWithLegend(
  config: ChartSpec,
  legendField: FieldConfig | string | undefined,
  chartVLConfig: Config<ExprRef | SignalRef> | undefined = undefined,
  defaultLegendPosition: ChartLegend = "top",
): Config<ExprRef | SignalRef> | undefined {
  const vlConfig = createConfig(config, chartVLConfig);

  if (!legendField || typeof legendField === "string") {
    return vlConfig;
  }
  const legendConfig = getLegendConfig(
    legendField.legendOrientation ?? defaultLegendPosition,
  );
  if (!vlConfig) return legendConfig;
  return merge(vlConfig, legendConfig);
}

export function createConfig(
  config: ChartSpec,
  chartVLConfig?: Config<ExprRef | SignalRef> | undefined,
): Config<ExprRef | SignalRef> | undefined {
  const userProvidedConfig = config.vl_config;
  return mergedVlConfig(userProvidedConfig, chartVLConfig);
}

export function createEncoding(
  config: CartesianChartSpec,
  data: ChartDataResult,
): Encoding<Field> {
  return {
    x: createPositionEncoding(config.x, data, "x"),
    y: createPositionEncoding(config.y, data, "y"),
    color: createColorEncoding(config.color, data),
    tooltip: createDefaultTooltipEncoding(
      [config.x, config.y, config.color],
      data,
    ),
  };
}

export function buildHoverPointOverlay(): UnitSpec<Field> {
  return {
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
  };
}

/**
 * Creates a multiValueTooltipChannel for cartesian charts (area, line, bar, stacked-bar)
 * Maps data values based on colorField and includes x-field information
 * In comparison mode, includes both current and previous period values
 */
export function createCartesianMultiValueTooltipChannel(
  config: { x?: FieldConfig; colorField?: string | undefined; yField?: string },
  data: ChartDataResult,
): TooltipValue[] | undefined {
  const { x: xConfig, colorField, yField } = config;

  if (!xConfig || !yField) {
    return undefined;
  }

  const xField = sanitizeValueForVega(xConfig.field);
  const sanitizedYField = sanitizeValueForVega(yField);
  const yFormatType = sanitizeFieldName(yField);

  let multiValueTooltipChannel: TooltipValue[] | undefined;

  // In comparison mode, we need to include both current and previous values
  // The pivot will create fields like "CategoryA", "CategoryA_prev", "CategoryB", "CategoryB_prev"
  if (data.hasComparison && colorField) {
    const tooltipFields: TooltipValue[] = [];
    const domainValues = data.domainValues?.[colorField] as
      | string[]
      | undefined;

    if (domainValues) {
      for (const value of domainValues) {
        const title = sanitizeTitleForVegaTooltip(value);
        // Add current period value
        tooltipFields.push({
          field: sanitizeValueForVega(value),
          title,
          type: "quantitative" as const,
          formatType: yFormatType,
        });

        // Add previous period value
        tooltipFields.push({
          field: sanitizeValueForVega(value) + ComparisonDeltaPreviousSuffix,
          title: title + ComparisonDeltaPreviousSuffix,
          type: "quantitative" as const,
          formatType: yFormatType,
        });
      }
    }

    multiValueTooltipChannel = tooltipFields;
  } else if (data.hasComparison) {
    const yTitle = data.fields[yField]?.displayName || yField;
    multiValueTooltipChannel = [
      {
        field: sanitizedYField,
        title: yTitle,
        type: "quantitative" as const,
        formatType: yFormatType,
      },
      {
        field: sanitizedYField + ComparisonDeltaPreviousSuffix,
        title: yTitle + ComparisonDeltaPreviousSuffix,
        type: "quantitative" as const,
        formatType: yFormatType,
      },
    ];
  } else if (colorField) {
    // Normal mode without comparison
    multiValueTooltipChannel = data.domainValues?.[colorField]?.map(
      (value) => ({
        field: sanitizeValueForVega(value as string),
        title: sanitizeTitleForVegaTooltip(value),
        type: "quantitative" as const,
        formatType: yFormatType,
      }),
    );
  }

  if (multiValueTooltipChannel) {
    multiValueTooltipChannel.unshift({
      field: xField,
      title: data.fields[xConfig.field]?.displayName || xConfig.field,
      type: xConfig?.type === "value" ? "nominal" : xConfig.type,
      ...(xConfig.type === "temporal" && { format: "%b %d, %Y %H:%M" }),
    });

    multiValueTooltipChannel = multiValueTooltipChannel.slice(0, 50);
  }

  return multiValueTooltipChannel;
}

/**
 * Hover target layer shared by the cartesian charts. `xField` is the raw x
 * field name: it is escaped for the encoding here and used as is to look up
 * the chart's domain values, which are keyed by raw names.
 */
export function buildHoverRuleLayer(args: {
  xField?: string;
  defaultTooltip: TooltipValue[];
  multiValueTooltipChannel?: TooltipValue[];
  pivot?: { field: string; value: string; groupby: string[] };
  domainValues?: ChartDomainValues;
  xSort?: ChartSortDirection;
  primaryColor: Color;
  xBand?: number;
  isBarMark?: boolean;
  isDarkMode?: boolean;
  isInteractive?: boolean;
}): UnitSpec<Field> {
  const {
    xField,
    defaultTooltip,
    multiValueTooltipChannel,
    pivot,
    domainValues,
    xSort,
    primaryColor,
    xBand,
    isBarMark = false,
    isDarkMode = false,
    isInteractive = false,
  } = args;
  const sanitizedXField = xField ? sanitizeValueForVega(xField) : undefined;

  const params: SelectionParameter[] = [
    {
      name: "hover",
      select: {
        type: "point",
        encodings: ["x"],
        on: "pointerover",
        clear: "pointerout",
        ...(!isBarMark && { nearest: true }),
      },
    },
  ];

  if (isInteractive) {
    params.push(createBrushParam());
  }

  return {
    transform:
      xField && pivot && multiValueTooltipChannel?.length
        ? [
            {
              pivot: pivot.field,
              value: pivot.value,
              groupby: pivot.groupby,
            },
          ]
        : [],
    mark: {
      type: isBarMark ? "bar" : "rule",
      clip: true,
      opacity: 0.8,
      ...(!isBarMark && { strokeWidth: 1.5 }),
    },
    encoding: {
      x: {
        field: sanitizedXField,
        ...(xBand !== undefined ? { bandPosition: xBand } : {}),
        ...(xSort && xField ? { sort: domainValues?.[xField] } : {}),
      },
      color: {
        condition: [
          {
            param: "hover",
            empty: false,
            value: isBarMark
              ? isDarkMode
                ? BarHighlightColorDark
                : BarHighlightColorLight
              : primaryColor.brighten().css(),
          },
        ],
        value: "transparent",
      },
      tooltip: multiValueTooltipChannel?.length
        ? multiValueTooltipChannel
        : defaultTooltip,
    },
    params,
  };
}
