import type { ChartDataResult } from "@rilldata/web-common/features/canvas/components/charts/selector";
import type {
  ChartConfig,
  FieldConfig,
  TooltipValue,
} from "@rilldata/web-common/features/canvas/components/charts/types";
import { sanitizeFieldName } from "@rilldata/web-common/features/canvas/components/charts/util";
import { sanitizeValueForVega } from "@rilldata/web-common/features/templates/charts/utils";
import type { VisualizationSpec } from "svelte-vega";
import type {
  ColorDef,
  Field,
  PositionDef,
} from "vega-lite/build/src/channeldef";
import type { Encoding } from "vega-lite/build/src/encoding";
import type { TopLevelParameter } from "vega-lite/build/src/spec/toplevel";
import type { TopLevelUnitSpec } from "vega-lite/build/src/spec/unit";

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
  mark: "line" | "bar" | "point" | "area" | "arc",
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

export function createXEncoding(
  xField: FieldConfig | undefined,
  data: ChartDataResult,
): PositionDef<Field> {
  if (!xField) return {};
  const metaData = data.fields[xField.field];
  return {
    field: sanitizeValueForVega(xField.field),
    title: metaData?.displayName || xField.field,
    type: xField.type,
    ...(metaData && "timeUnit" in metaData && { timeUnit: metaData.timeUnit }),
    ...(xField.sort && xField.type !== "temporal" && { sort: xField.sort }),
    axis: {
      ...(xField.type === "quantitative" && {
        formatType: sanitizeFieldName(xField.field),
      }),
      ...(metaData && "format" in metaData && { format: metaData.format }),
      ...(!xField.showAxisTitle && { title: null }),
    },
  };
}

export function createYEncoding(
  yField: FieldConfig | undefined,
  data: ChartDataResult,
): PositionDef<Field> {
  if (!yField) return {};
  const metaData = data.fields[yField.field];
  return {
    field: sanitizeValueForVega(yField.field),
    title: metaData?.displayName || yField.field,
    type: yField.type,
    ...(yField.zeroBasedOrigin !== true && {
      scale: {
        zero: false,
      },
    }),
    axis: {
      ...(yField.type === "quantitative" && {
        formatType: sanitizeFieldName(yField.field),
      }),
      ...(!yField.showAxisTitle && { title: null }),
      ...(metaData && "format" in metaData && { format: metaData.format }),
    },
    ...(metaData && "timeUnit" in metaData && { timeUnit: metaData.timeUnit }),
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
  xField: FieldConfig | undefined,
  yField: FieldConfig | undefined,
  colorField: FieldConfig | string | undefined,
  data: ChartDataResult,
) {
  const tooltip: TooltipValue[] = [];

  if (xField) {
    tooltip.push({
      field: sanitizeValueForVega(xField.field),
      title: data.fields[xField.field]?.displayName || xField.field,
      type: xField.type,
      ...(xField.type === "quantitative" && {
        formatType: sanitizeFieldName(xField.field),
      }),
      ...(xField.type === "temporal" && { format: "%b %d, %Y %H:%M" }),
    });
  }
  if (yField) {
    tooltip.push({
      field: sanitizeValueForVega(yField.field),
      title: data.fields[yField.field]?.displayName || yField.field,
      type: yField.type,
      ...(yField.type === "quantitative" && {
        formatType: sanitizeFieldName(yField.field),
      }),
      ...(yField.type === "temporal" && { format: "%b %d, %Y %H:%M" }),
    });
  }
  if (typeof colorField === "object" && colorField.field) {
    tooltip.push({
      field: sanitizeValueForVega(colorField.field),
      title: data.fields[colorField.field]?.displayName || colorField.field,
      type: colorField.type,
    });
  }

  return tooltip;
}

export function createEncoding(
  config: ChartConfig,
  data: ChartDataResult,
): Encoding<Field> {
  return {
    x: createXEncoding(config.x, data),
    y: createYEncoding(config.y, data),
    color: createColorEncoding(config.color, data),
    tooltip: createDefaultTooltipEncoding(
      config.x,
      config.y,
      config.color,
      data,
    ),
  };
}
