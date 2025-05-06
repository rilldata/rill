import {
  sanitizeFieldName,
  sanitizeValueForVega,
} from "@rilldata/web-common/components/vega/util";
import type { ChartSpec } from "@rilldata/web-common/features/canvas/components/charts";
import type { CartesianChartSpec } from "@rilldata/web-common/features/canvas/components/charts/cartesian-charts/CartesianChart";
import type {
  FieldConfig,
  TooltipValue,
} from "@rilldata/web-common/features/canvas/components/charts/types";
import { mergedVlConfig } from "@rilldata/web-common/features/canvas/components/charts/util";
import type { VisualizationSpec } from "svelte-vega";
import type { Config } from "vega-lite";
import type {
  ColorDef,
  Field,
  PositionDef,
} from "vega-lite/build/src/channeldef";
import type { Encoding } from "vega-lite/build/src/encoding";
import type { TopLevelParameter } from "vega-lite/build/src/spec/toplevel";
import type { TopLevelUnitSpec } from "vega-lite/build/src/spec/unit";
import type { ExprRef, SignalRef } from "vega-typings";
import type { ChartDataResult } from "./types";

export function createMultiLayerBaseSpec() {
  const baseSpec: VisualizationSpec = {
    $schema: "https://vega.github.io/schema/vega-lite/v5.json",
    width: "container",
    data: { name: "metrics-view" },
    autosize: { type: "fit" },
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
    mark,
    width: "container",
    data: { name: "metrics-view" },
    autosize: { type: "fit" },
  };
}

export function createPositionEncoding(
  field: FieldConfig | undefined,
  data: ChartDataResult,
): PositionDef<Field> {
  if (!field) return {};
  const metaData = data.fields[field.field];
  return {
    field: sanitizeValueForVega(field.field),
    title: metaData?.displayName || field.field,
    type: field.type,
    ...(metaData && "timeUnit" in metaData && { timeUnit: metaData.timeUnit }),
    ...(field.sort && field.type !== "temporal" && { sort: field.sort }),
    ...(field.type === "quantitative" &&
      field.zeroBasedOrigin !== true && {
        scale: {
          zero: false,
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
  if (typeof colorField === "object") {
    const metaData = data.fields[colorField.field];

    return {
      field: sanitizeValueForVega(colorField.field),
      title: metaData?.displayName || colorField.field,
      type: colorField.type,
      ...(metaData &&
        "timeUnit" in metaData && { timeUnit: metaData.timeUnit }),
    };
  }
  if (typeof colorField === "string") {
    return { value: colorField };
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
