import { TIME_GRAIN } from "@rilldata/web-common/lib/time/config";
import type { V1TimeGrain } from "@rilldata/web-common/runtime-client/gen/index.schemas";
import type {
  WireField,
  WireFieldKind,
  WireFlintChartSpec,
  WireFlintEncoding,
  WireFlintInput,
  WireTimeRange,
} from "./contract";
import {
  grainsForRange,
  previousPeriod,
  TOP_N_PRESETS,
  type ControlOptions,
  type ControlState,
  type TopNTarget,
} from "./controls";
import {
  comparisonChannel,
  encodingFields,
  singleEncodingField,
} from "./flint-input";

/**
 * The data controls (time grain, top-N, comparison) on the Flint path. They patch the effective
 * input the server echoed, which the widget sends back to the app-only `chart_data` tool, and
 * describe the view for the model.
 */

/** Channels whose dimension gets a top-N query on the server, in the order the control picks them. */
const TOP_N_CHANNELS = ["x", "y", "color", "group", "column", "row"];
/** Small multiples stay small: the server caps a facet at 8 panels (`MaxFacetValues`). */
const FACET_CHANNELS: ReadonlySet<string> = new Set(["column", "row"]);
const FACET_TOP_N_PRESETS: readonly number[] = [4, 6, 8];

export function flintFieldKind(
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
export function flintTemporalField(
  chartSpec: WireFlintChartSpec,
  fields: Record<string, WireField> | undefined,
  timeDimension?: string,
): string | undefined {
  for (const channel of ["x", "y"]) {
    const field = singleEncodingField(chartSpec.encodings?.[channel]);
    if (flintFieldKind(field, fields, timeDimension) === "time") return field;
  }
  return undefined;
}

/**
 * The dimension Flint colors the series by: the one on color, else a grouped bar chart's
 * clusters on group.
 */
export function flintSeriesField(
  chartSpec: WireFlintChartSpec,
  fields: Record<string, WireField> | undefined,
  timeDimension?: string,
): string | undefined {
  for (const channel of ["color", "group"]) {
    const field = singleEncodingField(chartSpec.encodings?.[channel]);
    if (flintFieldKind(field, fields, timeDimension) === "dimension") {
      return field;
    }
  }
  return undefined;
}

/** The first dimension channel a top-N limit applies to, with the limit the server put in effect. */
export function flintTopNTarget(
  chartSpec: WireFlintChartSpec,
  fields: Record<string, WireField> | undefined,
  timeDimension?: string,
): TopNTarget | undefined {
  for (const channel of TOP_N_CHANNELS) {
    const value = chartSpec.encodings?.[channel];
    const field = singleEncodingField(value);
    if (flintFieldKind(field, fields, timeDimension) !== "dimension") continue;
    const limit =
      typeof value === "object" && !Array.isArray(value)
        ? value.limit
        : undefined;
    return { channel, limit };
  }
  return undefined;
}

export function flintControlOptions(
  chartSpec: WireFlintChartSpec,
  fields: Record<string, WireField> | undefined,
  timeDimension: string | undefined,
  range: WireTimeRange | undefined,
  smallestGrain?: V1TimeGrain,
): ControlOptions | undefined {
  const temporal = flintTemporalField(chartSpec, fields, timeDimension);
  const target = flintTopNTarget(chartSpec, fields, timeDimension);
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
export function patchFlintInput(
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
    const target = flintTopNTarget(chartSpec, fields, timeDimension);
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
export function describeFlintView(
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
