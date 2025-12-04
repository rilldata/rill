import {
  sanitizeFieldName,
  sanitizeValueForVega,
} from "@rilldata/web-common/components/vega/util";
import type { CartesianChartSpec } from "@rilldata/web-common/features/components/charts";
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
  getColorForValues,
  isDomainStringArray,
  isFieldConfig,
  mergedVlConfig,
  resolveColor,
  resolveCSSVariable,
  sanitizeSortFieldForVega,
} from "@rilldata/web-common/features/components/charts/util";
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
import type { Config } from "vega-lite";
import type {
  ColorDef,
  Field,
  PositionFieldDef,
} from "vega-lite/build/src/channeldef";
import type { Encoding } from "vega-lite/build/src/encoding";
import type { TopLevelParameter } from "vega-lite/build/src/spec/toplevel";
import type { TopLevelUnitSpec, UnitSpec } from "vega-lite/build/src/spec/unit";
import type { ExprRef, SignalRef } from "vega-typings";

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

export function createPositionEncoding(
  field: FieldConfig | undefined,
  data: ChartDataResult,
): PositionFieldDef<Field> {
  if (!field || field.type === "value") return {};
  const metaData = data.fields[field.field];
  return {
    field: sanitizeValueForVega(field.field),
    title: metaData?.displayName || field.field,
    type: field.type,
    ...(metaData && "timeUnit" in metaData && { timeUnit: metaData.timeUnit }),
    ...(field.sort &&
      field.type !== "temporal" && {
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
    axis: {
      ...(field.labelAngle !== undefined && { labelAngle: field.labelAngle }),
      ...(field.type === "quantitative" && {
        formatType: sanitizeFieldName(field.field),
      }),
      ...(metaData && "format" in metaData && { format: metaData.format }),
      ...(!field.showAxisTitle && { title: null }),
    },
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

export function createOrderEncoding(field: FieldConfig | undefined) {
  if (!field || field.type === "value") return {};
  return {
    field: sanitizeValueForVega(field.field),
    type: field.type,
    order: "descending",
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

export function getLegendConfig(
  orientation: ChartLegend,
): Config<ExprRef | SignalRef> {
  let columns: number | ExprRef = 1;
  let symbolLimit: number | ExprRef = 40;
  if (orientation === "top" || orientation === "bottom") {
    columns = { expr: "floor(width / 140)" };
  } else if (orientation === "right" || orientation === "left") {
    symbolLimit = { expr: "floor(height / 13 )" };
  }
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
      columns: columns,
      labelLimit: 140,
      symbolLimit: symbolLimit,
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
    x: createPositionEncoding(config.x, data),
    y: createPositionEncoding(config.y, data),
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
 */
export function createCartesianMultiValueTooltipChannel(
  config: { x?: FieldConfig; colorField?: string; yField?: string },
  data: ChartDataResult,
): TooltipValue[] | undefined {
  const { x: xConfig, colorField, yField } = config;

  if (!colorField || !xConfig || !yField) {
    return undefined;
  }

  const xField = sanitizeValueForVega(xConfig.field);
  const sanitizedYField = sanitizeValueForVega(yField);

  let multiValueTooltipChannel: TooltipValue[] | undefined;

  multiValueTooltipChannel = data.domainValues?.[colorField]?.map((value) => ({
    field: sanitizeValueForVega(value as string),
    type: "quantitative" as const,
    formatType: sanitizeFieldName(sanitizedYField),
  }));

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
  } = args;

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
      opacity: 0.6,
      ...(!isBarMark && { strokeWidth: 5 }),
    },
    encoding: {
      x: {
        field: xField,
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
    params: [
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
    ],
  };
}
