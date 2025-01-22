import type { ChartDataResult } from "@rilldata/web-common/features/canvas/components/charts/selector";
import type {
  ChartConfig,
  TooltipValue,
} from "@rilldata/web-common/features/canvas/components/charts/types";
import type { VisualizationSpec } from "svelte-vega";
import type {
  ColorDef,
  Field,
  PositionDef,
} from "vega-lite/build/src/channeldef";
import type { Encoding } from "vega-lite/build/src/encoding";
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
  mark: "line" | "bar" | "point",
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
  config: ChartConfig,
  data: ChartDataResult,
): PositionDef<Field> {
  if (!config.x) return {};
  return {
    field: config.x.field,
    title: data.fields[config.x.field]?.displayName || config.x.field,
    type: config.x.type,
    ...(config.x.timeUnit && { timeUnit: config.x.timeUnit }),
    axis: {
      ...(config.x.type === "quantitative" && {
        formatType: config.x.field,
      }),
      ...(!config.x.showAxisTitle && { title: null }),
    },
  };
}

export function createYEncoding(
  config: ChartConfig,
  data: ChartDataResult,
): PositionDef<Field> {
  if (!config.y) return {};
  return {
    field: config.y.field,
    title: data.fields[config.y.field]?.displayName || config.y.field,
    type: config.y.type,
    axis: {
      ...(config.y.type === "quantitative" && {
        formatType: config.y.field,
      }),
      ...(!config.y.showAxisTitle && { title: null }),
    },
    ...(config.y.timeUnit && { timeUnit: config.y.timeUnit }),
  };
}

export function createColorEncoding(
  config: ChartConfig,
  data: ChartDataResult,
): ColorDef<Field> {
  if (!config.color) return {};
  if (typeof config.color === "object") {
    return {
      field: config.color.field,
      title: data.fields[config.color.field]?.displayName || config.color.field,
      type: config.color.type,
      ...(config.color.timeUnit && { timeUnit: config.color.timeUnit }),
    };
  }
  if (typeof config.color === "string") {
    return { value: config.color };
  }
  return {};
}

export function createDefaultTooltipEncoding(
  config: ChartConfig,
  data: ChartDataResult,
) {
  const tooltip: TooltipValue[] = [];

  if (config.x) {
    tooltip.push({
      field: config.x.field,
      title: data.fields[config.x.field]?.displayName || config.x.field,
      type: config.x.type,
      ...(config.x.type === "quantitative" && {
        formatType: config.x.field,
      }),
    });
  }
  if (config.y) {
    tooltip.push({
      field: config.y.field,
      title: data.fields[config.y.field]?.displayName || config.y.field,
      type: config.y.type,
      ...(config.y.type === "quantitative" && {
        formatType: config.y.field,
      }),
    });
  }
  if (typeof config.color === "object" && config.color.field) {
    tooltip.push({
      field: config.color.field,
      title: data.fields[config.color.field]?.displayName || config.color.field,
      type: config.color.type,
    });
  }

  return tooltip;
}

export function createEncoding(
  config: ChartConfig,
  data: ChartDataResult,
): Encoding<Field> {
  return {
    x: createXEncoding(config, data),
    y: createYEncoding(config, data),
    color: createColorEncoding(config, data),
    tooltip: createDefaultTooltipEncoding(config, data),
  };
}
