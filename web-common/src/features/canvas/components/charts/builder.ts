import type { ChartDataResult } from "@rilldata/web-common/features/canvas/components/charts/selector";
import type { ChartConfig } from "@rilldata/web-common/features/canvas/components/charts/types";
import type {
  ColorDef,
  Field,
  PositionDef,
} from "vega-lite/build/src/channeldef";
import type { Encoding } from "vega-lite/build/src/encoding";
import type { TopLevelUnitSpec } from "vega-lite/build/src/spec/unit";

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

function createXEncoding(
  config: ChartConfig,
  data: ChartDataResult,
): PositionDef<Field> {
  if (!config.x) return {};
  return {
    field: config.x.field,
    title: data.fields[config.x.field]?.displayName || config.x.field,
    type: config.x.type,
    ...(config.x.timeUnit && { timeUnit: config.x.timeUnit }),
    ...(!config.x.showAxisTitle && { axis: { title: null } }),
  };
}

function createYEncoding(
  config: ChartConfig,
  data: ChartDataResult,
): PositionDef<Field> {
  if (!config.y) return {};
  return {
    field: config.y.field,
    title: data.fields[config.y.field]?.displayName || config.y.field,
    type: config.y.type,
    ...(config.y.timeUnit && { timeUnit: config.y.timeUnit }),
    ...(!config.y.showAxisTitle && { axis: { title: null } }),
  };
}

function createColorEncoding(
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

export function createEncoding(
  config: ChartConfig,
  data: ChartDataResult,
): Encoding<Field> {
  return {
    x: createXEncoding(config, data),
    y: createYEncoding(config, data),
    color: createColorEncoding(config, data),
  };
}
