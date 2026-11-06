import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";
import { V1TimeGrain } from "@rilldata/web-common/runtime-client/gen/index.schemas";
import { DateTime } from "luxon";
import type {
  WireField,
  WireFieldKind,
  WireFlintChartSpec,
  WireFlintEncoding,
  WireFlintInput,
  WireTimeRange,
} from "./contract";
import {
  comparisonChannel,
  encodingFields,
  singleEncodingField,
} from "./flint-input";

/**
 * Interactive controls (fullscreen only): time grain, top-N and comparison. Each change
 * re-queries through the host with the app-only `chart_data` tool, so these helpers only patch
 * the effective input the server echoed and describe the view for the model.
 */

/** The top-N limits offered for an axis or a series split; the server caps them at 50 and 20. */
export const TOP_N_PRESETS: readonly number[] = [5, 10, 20];
/** Channels whose dimension gets a top-N query on the server, in the order the control picks them. */
const TOP_N_CHANNELS = ["x", "y", "color", "group", "column", "row"];
/** Small multiples stay small: the server caps a facet at 8 panels (`MaxFacetValues`). */
const FACET_CHANNELS: ReadonlySet<string> = new Set(["column", "row"]);
const FACET_TOP_N_PRESETS: readonly number[] = [4, 6, 8];

export interface ControlState {
  grain?: V1TimeGrain;
  /** The limit of the chart's top-N channel; see `topNTarget`. */
  topN?: number;
  comparison?: boolean;
}

export interface ControlOptions {
  /** Grains the user can pick; empty when the chart binds no time field on x or y. */
  grains: V1TimeGrain[];
  /** The top-N limits to offer besides the one in effect; empty when no channel takes a limit. */
  topN: readonly number[];
  /** Whether the chart can draw the previous period as a second series (see `comparisonChannel`). */
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

export function bucketCount(range: WireTimeRange, grain: V1TimeGrain): number {
  const unit = TIME_GRAIN[grain]?.unit;
  const start = DateTime.fromISO(range.start, { zone: "UTC" });
  const end = DateTime.fromISO(range.end, { zone: "UTC" });
  if (!unit || !start.isValid || !end.isValid) return 0;
  return Math.ceil(end.diff(start, unit).as(unit));
}

/** The selectable grains for a range, no finer than the metrics view's smallest grain. */
export function grainsForRange(
  range: WireTimeRange,
  smallestGrain?: V1TimeGrain,
): V1TimeGrain[] {
  const minimumOrder = smallestGrain ? GRAIN_ORDER.indexOf(smallestGrain) : -1;
  return SELECTABLE_GRAINS.filter((grain) => {
    if (GRAIN_ORDER.indexOf(grain) < minimumOrder) return false;
    const count = bucketCount(range, grain);
    return count >= MIN_GRAIN_BUCKETS && count <= MAX_GRAIN_BUCKETS;
  });
}

export interface TopNTarget {
  /** The channel that binds the dimension. */
  channel: string;
  /**
   * The limit in effect. A result's spec carries the limit the backend applied, after its
   * defaults and caps, so this is only undefined for a spec that never went through it.
   */
  limit: number | undefined;
}

/** The values the top-N control offers: the presets plus the current limit, ascending. */
export function topNChoices(
  current: number | undefined,
  presets: readonly number[],
): number[] {
  const choices = new Set(presets);
  if (current !== undefined) choices.add(current);
  return Array.from(choices).sort((a, b) => a - b);
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

export function fieldKind(
  name: string | undefined,
  fields: Record<string, WireField> | undefined,
  timeDimension?: string,
): WireFieldKind | undefined {
  if (!name) return undefined;
  const kind = fields?.[name]?.kind;
  if (kind) return kind;
  return name === timeDimension ? "time" : undefined;
}

/** The time field bound to the x axis, else to the y axis. */
export function temporalFieldOf(
  chartSpec: WireFlintChartSpec,
  fields: Record<string, WireField> | undefined,
  timeDimension?: string,
): string | undefined {
  for (const channel of ["x", "y"]) {
    const field = singleEncodingField(chartSpec.encodings?.[channel]);
    if (fieldKind(field, fields, timeDimension) === "time") return field;
  }
  return undefined;
}

/**
 * The dimension Flint colors the series by: the one on color, else a grouped bar chart's
 * clusters on group.
 */
export function seriesFieldOf(
  chartSpec: WireFlintChartSpec,
  fields: Record<string, WireField> | undefined,
  timeDimension?: string,
): string | undefined {
  for (const channel of ["color", "group"]) {
    const field = singleEncodingField(chartSpec.encodings?.[channel]);
    if (fieldKind(field, fields, timeDimension) === "dimension") {
      return field;
    }
  }
  return undefined;
}

/** The first dimension channel a top-N limit applies to, with the limit the server put in effect. */
export function topNTarget(
  chartSpec: WireFlintChartSpec,
  fields: Record<string, WireField> | undefined,
  timeDimension?: string,
): TopNTarget | undefined {
  for (const channel of TOP_N_CHANNELS) {
    const value = chartSpec.encodings?.[channel];
    const field = singleEncodingField(value);
    if (fieldKind(field, fields, timeDimension) !== "dimension") continue;
    const limit =
      typeof value === "object" && !Array.isArray(value)
        ? value.limit
        : undefined;
    return { channel, limit };
  }
  return undefined;
}

export function controlOptions(
  chartSpec: WireFlintChartSpec,
  fields: Record<string, WireField> | undefined,
  timeDimension: string | undefined,
  range: WireTimeRange | undefined,
  smallestGrain?: V1TimeGrain,
): ControlOptions | undefined {
  const temporal = temporalFieldOf(chartSpec, fields, timeDimension);
  const target = topNTarget(chartSpec, fields, timeDimension);
  const options: ControlOptions = {
    grains: range && temporal ? grainsForRange(range, smallestGrain) : [],
    topN: !target
      ? []
      : FACET_CHANNELS.has(target.channel)
        ? FACET_TOP_N_PRESETS
        : TOP_N_PRESETS,
    comparison: comparisonChannel(chartSpec, fields ?? {}) !== undefined,
  };
  if (
    options.grains.length < 2 &&
    options.topN.length === 0 &&
    !options.comparison
  ) {
    return undefined;
  }
  return options;
}

/**
 * The `chart_data` arguments for a control state: the current chart spec with the top-N limit on
 * its channel, the data binding with the grain and the comparison range, and the caller's
 * semantic type overrides. A spec that cannot draw a comparison is sent without one, since the
 * server would reject it. Inputs are left untouched.
 */
export function patchInput(
  input: WireFlintInput,
  chartSpec: WireFlintChartSpec,
  state: ControlState,
  range: WireTimeRange | undefined,
  fields: Record<string, WireField> | undefined,
  timeDimension?: string,
): WireFlintInput {
  const patched: WireFlintInput = JSON.parse(
    JSON.stringify({
      chart_spec: chartSpec,
      data: input.data,
      semantic_types: input.semantic_types,
    }),
  );
  if (state.grain) {
    patched.data.time_grain = state.grain;
  }
  if (state.topN !== undefined) {
    const target = topNTarget(chartSpec, fields, timeDimension);
    if (target) {
      const value = patched.chart_spec.encodings[target.channel];
      const encoding: WireFlintEncoding =
        typeof value === "string"
          ? { field: value }
          : (value as WireFlintEncoding);
      patched.chart_spec.encodings[target.channel] = {
        ...encoding,
        limit: state.topN,
      };
    }
  }
  const foldable = comparisonChannel(chartSpec, fields ?? {}) !== undefined;
  if (state.comparison === true && foldable) {
    const base = range ?? input.data.time_range;
    if (!patched.data.comparison_time_range && base) {
      patched.data.comparison_time_range = previousPeriod(base);
    }
  } else if (state.comparison === false || !foldable) {
    delete patched.data.comparison_time_range;
  }
  return patched;
}

/**
 * One line for the model about what the chart currently shows (model-facing, English): the chart
 * type and its bindings, then every data control, including the ones the chart does not offer.
 */
export function describeView(
  title: string,
  chartSpec: WireFlintChartSpec,
  controls: ControlState,
  timeRange?: string,
): string {
  const bindings = Object.entries(chartSpec.encodings ?? {})
    .map(([channel, value]) => [channel, encodingFields(value)] as const)
    .filter(([, fields]) => fields.length > 0)
    .map(([channel, fields]) => `${channel} = ${fields.join(" + ")}`)
    .join(", ");
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
  return `Chart now shows ${title}: ${chartSpec.chartType} with ${bindings || "no bindings"}; ${parts.join(", ")}.`;
}
