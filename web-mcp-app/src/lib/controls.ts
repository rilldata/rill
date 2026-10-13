import type { CartesianChartSpec } from "@rilldata/web-common/features/components/charts/cartesian/CartesianChartProvider";
import {
  isHorizontal,
  toVerticalSpec,
} from "@rilldata/web-common/features/components/charts/cartesian/orientation";
import type { ChartType } from "@rilldata/web-common/features/components/charts/types";
import { isFieldConfig } from "@rilldata/web-common/features/components/charts/util";
import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";
import { V1TimeGrain } from "@rilldata/web-common/runtime-client/gen/index.schemas";
import { DateTime } from "luxon";
import {
  isCartesianChartType,
  type WireChartSpec,
  type WireTimeRange,
} from "./contract";

/**
 * Interactive controls (fullscreen only): time grain, top-N and comparison.
 * Each change re-queries through the host with the app-only `chart_data` tool,
 * so these helpers only patch the chart spec and describe what is shown.
 */

const TOP_N_PRESETS = [5, 10, 20];

export interface ControlState {
  grain?: V1TimeGrain;
  /** The limit of the chart's top-N channel; see `topNTarget`. */
  topN?: number;
  comparison?: boolean;
}

export interface ControlOptions {
  /** Grains the user can pick; empty when the chart has no temporal x axis. */
  grains: V1TimeGrain[];
  /** Whether a nominal x or color channel can take a top-N limit. */
  topN: boolean;
  /** Whether period-over-period comparison applies (cartesian charts only). */
  comparison: boolean;
}

const GRAIN_ORDER: V1TimeGrain[] = [
  V1TimeGrain.TIME_GRAIN_MILLISECOND,
  V1TimeGrain.TIME_GRAIN_SECOND,
  V1TimeGrain.TIME_GRAIN_MINUTE,
  V1TimeGrain.TIME_GRAIN_HOUR,
  V1TimeGrain.TIME_GRAIN_DAY,
  V1TimeGrain.TIME_GRAIN_WEEK,
  V1TimeGrain.TIME_GRAIN_MONTH,
  V1TimeGrain.TIME_GRAIN_QUARTER,
  V1TimeGrain.TIME_GRAIN_YEAR,
];

const SELECTABLE_GRAINS: V1TimeGrain[] = [
  V1TimeGrain.TIME_GRAIN_HOUR,
  V1TimeGrain.TIME_GRAIN_DAY,
  V1TimeGrain.TIME_GRAIN_WEEK,
  V1TimeGrain.TIME_GRAIN_MONTH,
  V1TimeGrain.TIME_GRAIN_QUARTER,
  V1TimeGrain.TIME_GRAIN_YEAR,
];

/** A grain is offered only when the range splits into a readable number of buckets. */
export const MIN_GRAIN_BUCKETS = 2;
export const MAX_GRAIN_BUCKETS = 1000;

/** The field on the x axis when it is temporal (cartesian charts and combo charts). */
export function temporalXField(
  chartType: ChartType,
  spec: WireChartSpec,
): string | undefined {
  const x = isCartesianChartType(chartType)
    ? toVerticalSpec(spec as unknown as CartesianChartSpec).x
    : chartType === "combo_chart"
      ? (spec as { x?: unknown }).x
      : undefined;
  return isFieldConfig(x) && x.type === "temporal" ? x.field : undefined;
}

export function bucketCount(range: WireTimeRange, grain: V1TimeGrain): number {
  const unit = TIME_GRAIN[grain]?.unit;
  const start = DateTime.fromISO(range.start, { zone: "UTC" });
  const end = DateTime.fromISO(range.end, { zone: "UTC" });
  if (!unit || !start.isValid || !end.isValid) return 0;
  return Math.ceil(end.diff(start, unit).as(unit));
}

export function availableGrains(
  chartType: ChartType,
  spec: WireChartSpec,
  range: WireTimeRange | undefined,
  smallestGrain?: V1TimeGrain,
): V1TimeGrain[] {
  if (!range || !temporalXField(chartType, spec)) return [];
  const minimumOrder = smallestGrain ? GRAIN_ORDER.indexOf(smallestGrain) : -1;
  return SELECTABLE_GRAINS.filter((grain) => {
    if (GRAIN_ORDER.indexOf(grain) < minimumOrder) return false;
    const count = bucketCount(range, grain);
    return count >= MIN_GRAIN_BUCKETS && count <= MAX_GRAIN_BUCKETS;
  });
}

export interface TopNTarget {
  /** Key of the spec channel that holds the nominal field (the raw channel, not normalized). */
  channel: string;
  /**
   * The limit in effect. A result's spec carries the limit the backend applied, after its
   * defaults and caps, so this is only undefined for a spec that never went through it.
   */
  limit: number | undefined;
}

/** The nominal channel a top-N limit applies to: the category axis first, else the color split. */
export function topNTarget(
  chartType: ChartType,
  spec: WireChartSpec,
): TopNTarget | undefined {
  const nominal = (channel: string): TopNTarget | undefined => {
    const field = spec[channel];
    return isFieldConfig(field) && field.type === "nominal"
      ? { channel, limit: field.limit }
      : undefined;
  };
  switch (chartType) {
    case "bar_chart":
    case "line_chart":
    case "area_chart":
    case "stacked_bar":
    case "stacked_bar_normalized": {
      const categoryChannel = isHorizontal(
        spec as unknown as CartesianChartSpec,
      )
        ? "y"
        : "x";
      return nominal(categoryChannel) ?? nominal("color");
    }
    case "combo_chart":
    case "heatmap":
      return nominal("x");
    case "donut_chart":
    case "pie_chart":
    case "scatter_plot":
      return nominal("color");
    case "funnel_chart":
      return nominal("stage");
  }
}

/** The values the top-N control offers: the presets plus the current limit, ascending. */
export function topNChoices(current: number | undefined): number[] {
  const choices = new Set(TOP_N_PRESETS);
  if (current !== undefined) choices.add(current);
  return Array.from(choices).sort((a, b) => a - b);
}

export function supportsComparison(chartType: ChartType): boolean {
  return isCartesianChartType(chartType);
}

export function controlOptions(
  chartType: ChartType,
  spec: WireChartSpec,
  range: WireTimeRange | undefined,
  smallestGrain?: V1TimeGrain,
): ControlOptions | undefined {
  const options: ControlOptions = {
    grains: availableGrains(chartType, spec, range, smallestGrain),
    topN: topNTarget(chartType, spec) !== undefined,
    comparison: supportsComparison(chartType),
  };
  if (options.grains.length < 2 && !options.topN && !options.comparison) {
    return undefined;
  }
  return options;
}

/** The period of the same length immediately before `range`. */
export function previousPeriod(range: WireTimeRange): WireTimeRange {
  const start = DateTime.fromISO(range.start, { zone: "UTC" });
  const end = DateTime.fromISO(range.end, { zone: "UTC" });
  const length = end.diff(start);
  return {
    start: start.minus(length).toUTC().toISO() ?? range.start,
    end: range.start,
    ...(range.time_zone && { time_zone: range.time_zone }),
  };
}

/** Returns a new spec with the control state applied; the input is left untouched. */
export function patchSpec(
  spec: WireChartSpec,
  chartType: ChartType,
  state: ControlState,
  range: WireTimeRange | undefined,
): WireChartSpec {
  const patched: WireChartSpec = JSON.parse(JSON.stringify(spec));
  if (state.grain) {
    patched.time_grain = state.grain;
  }
  if (state.topN !== undefined) {
    const target = topNTarget(chartType, spec);
    if (target) {
      const field = patched[target.channel];
      if (isFieldConfig(field)) {
        patched[target.channel] = { ...field, limit: state.topN };
      }
    }
  }
  if (state.comparison === true && supportsComparison(chartType)) {
    const base = range ?? spec.time_range;
    if (!patched.comparison_time_range && base) {
      patched.comparison_time_range = previousPeriod(base);
    }
  } else if (state.comparison === false) {
    delete patched.comparison_time_range;
  }
  return patched;
}

/**
 * One line for the model about what the chart currently shows (model-facing, English).
 * Every control is stated, including the ones the chart does not offer, so the model
 * never has to infer a setting from the caption. The top-N is the limit in effect, so
 * "top values n/a" only appears for a chart without a top-N channel.
 */
export function describeView(
  title: string,
  controls: ControlState,
  timeRange?: string,
): string {
  const grain = controls.grain
    ? (TIME_GRAIN[controls.grain]?.unit ?? controls.grain)
    : "n/a";
  const parts = [
    `time grain ${grain}`,
    controls.topN === undefined
      ? "top values n/a"
      : `top ${controls.topN} values`,
    `comparison ${controls.comparison ? "on" : "off"}`,
  ];
  if (timeRange) parts.push(timeRange);
  // i18n-ignore: text for the model, not shown to the user
  return `Chart now shows ${title}: ${parts.join(", ")}.`;
}
