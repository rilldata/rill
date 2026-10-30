import type {
  FlintChartSpec,
  FlintEncoding,
  FlintEncodingValue,
} from "@rilldata/web-common/features/custom-viz/flint/compile";
import type {
  FlintFields,
  FlintSemanticType,
} from "@rilldata/web-common/features/custom-viz/flint/semantic-types";
import { ComparisonDeltaPreviousSuffix } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
import { INTEGERS } from "@rilldata/web-common/lib/duckdb-data-types";
import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
import { FormatPreset } from "@rilldata/web-common/lib/number-formatting/humanizer-types";
import { includesCurrencySymbol } from "@rilldata/web-common/lib/number-formatting/utils/d3-format-utils";
import { V1TimeGrain } from "@rilldata/web-common/runtime-client/gen/index.schemas";
import { vlGetTemplateChannels } from "flint-chart/vegalite";
import type {
  ChartDataPayload,
  WireField,
  WireFlintChartSpec,
  WireFlintEncoding,
  WireFlintEncodingValue,
  WireRow,
} from "./contract";

/**
 * Builds what web-common's `compileFlintSpec` needs from a Flint result: the chart spec stripped
 * of what Flint must not see (the card shows the title and subtitle, and `limit` is Rill's own
 * top-N key), the per-field semantics, the measure columns to route through Rill's formatters,
 * and the comparison fold. Pure: no DOM, no host session.
 */

/**
 * Chart types that can show the previous period as a second series, and the channel that
 * separates the series. A bar chart becomes a grouped bar chart so the periods sit side by side
 * rather than stacked; the others color the series.
 */
export const COMPARISON_FOLDS: ReadonlyMap<
  string,
  { chartType: string; channel: "color" | "group" }
> = new Map([
  ["Line Chart", { chartType: "Line Chart", channel: "color" }],
  ["Area Chart", { chartType: "Area Chart", channel: "color" }],
  ["Bar Chart", { chartType: "Grouped Bar Chart", channel: "group" }],
  ["Grouped Bar Chart", { chartType: "Grouped Bar Chart", channel: "group" }],
  ["Lollipop Chart", { chartType: "Lollipop Chart", channel: "color" }],
  ["Slope Chart", { chartType: "Slope Chart", channel: "color" }],
]);

export const COMPARISON_FOLD_CHART_TYPES: ReadonlySet<string> = new Set(
  COMPARISON_FOLDS.keys(),
);

/** The column that names the period of each row once the previous period is folded in. */
export const COMPARISON_SERIES_FIELD = "rill_period";

/** The keys of a Flint encoding object; anything else (Rill's `limit`) is dropped before compiling. */
const FLINT_ENCODING_KEYS = [
  "field",
  "type",
  "aggregate",
  "sortOrder",
  "sortBy",
  "scheme",
] as const satisfies readonly (keyof FlintEncoding)[];

/** What the metrics view's timestamp column is labelled; mirrors web-common's semantic derivation. */
const TIME_DISPLAY_NAME = "Time";

/**
 * Keep in sync with `semantic-types.ts` in web-common and with the server's derivation
 * (`runtime/ai/chartquery`); grains finer than a day have no dedicated Flint type.
 */
const TIME_GRAIN_TO_SEMANTIC_TYPE: Partial<
  Record<V1TimeGrain, FlintSemanticType>
> = {
  [V1TimeGrain.TIME_GRAIN_YEAR]: "Year",
  [V1TimeGrain.TIME_GRAIN_QUARTER]: "YearQuarter",
  [V1TimeGrain.TIME_GRAIN_MONTH]: "YearMonth",
  [V1TimeGrain.TIME_GRAIN_WEEK]: "YearWeek",
  [V1TimeGrain.TIME_GRAIN_DAY]: "Date",
};

export interface FlintCompileInput {
  /** The spec handed to Flint: no title or subtitle, no `limit` keys, comparison folded. */
  spec: FlintChartSpec;
  fields: FlintFields;
  /** The rows Flint compiles: the result's, or their long form when a comparison is folded in. */
  rows: WireRow[];
  /** Measure columns present in the rows; each gets a Rill formatter. */
  measureFields: string[];
  /** The measure whose previous period was folded in, and the channel that separates the periods. */
  comparison?: { channel: "color" | "group"; measure: string };
}

/** The field a single-field encoding binds; undefined for an array (series fold) or nothing. */
export function singleEncodingField(
  value: WireFlintEncodingValue | undefined,
): string | undefined {
  if (value === undefined) return undefined;
  if (typeof value === "string") return value || undefined;
  if (Array.isArray(value)) return undefined;
  return value.field || undefined;
}

/** Every field an encoding binds, arrays included. */
export function encodingFields(
  value: WireFlintEncodingValue | undefined,
): string[] {
  if (value === undefined) return [];
  const values = Array.isArray(value) ? value : [value];
  return values
    .map((entry) => (typeof entry === "string" ? entry : entry.field))
    .filter((field): field is string => !!field);
}

function stripEncoding(
  value: string | WireFlintEncoding,
): string | FlintEncoding {
  if (typeof value === "string") return value;
  const stripped: FlintEncoding = {};
  for (const key of FLINT_ENCODING_KEYS) {
    if (value[key] !== undefined) {
      (stripped as Record<string, unknown>)[key] = value[key];
    }
  }
  return stripped;
}

/**
 * The chart spec as Flint should see it: without the title, the subtitle and Rill's `limit` keys,
 * and without the channels the chart type does not accept. The server takes `detail` on every
 * chart type as a Rill extension (a row-grain dimension that gets no top-N query); for a type
 * whose template has no `detail` channel the rows are already at that grain and the binding is dropped.
 */
export function stripFlintChartSpec(
  chartSpec: WireFlintChartSpec,
): FlintChartSpec {
  const accepted = templateChannels(chartSpec.chartType);
  const encodings: Record<string, FlintEncodingValue> = {};
  for (const [channel, value] of Object.entries(chartSpec.encodings ?? {})) {
    if (value === undefined || value === null) continue;
    if (accepted && !accepted.has(channel)) continue;
    encodings[channel] = Array.isArray(value)
      ? value.map(stripEncoding)
      : stripEncoding(value);
  }
  return {
    chartType: chartSpec.chartType,
    encodings,
    ...(chartSpec.chartProperties &&
    Object.keys(chartSpec.chartProperties).length > 0
      ? { chartProperties: { ...chartSpec.chartProperties } }
      : {}),
  };
}

/** The channels a Flint chart type accepts, or undefined for a type Flint does not know. */
function templateChannels(chartType: string): Set<string> | undefined {
  try {
    const channels = vlGetTemplateChannels(chartType);
    return channels.length > 0 ? new Set(channels) : undefined;
  } catch {
    return undefined;
  }
}

/**
 * Picks a Flint type from how a measure is formatted rather than from its data type, the way
 * web-common's `semantic-types.ts` does for custom viz: the format preset encodes the intent.
 */
function measureSemanticType(field: WireField): FlintSemanticType {
  switch (field.format_preset as FormatPreset | undefined) {
    case FormatPreset.CURRENCY_USD:
    case FormatPreset.CURRENCY_EUR:
      return "Amount";
    case FormatPreset.PERCENTAGE:
      return "Percentage";
    case FormatPreset.INTERVAL:
      return "Duration";
  }
  if (field.format_d3) {
    if (includesCurrencySymbol(field.format_d3)) return "Amount";
    if (field.format_d3.includes("%")) return "Percentage";
  }
  // The wire type has no CODE_ prefix; web-common's set is keyed by the runtime's codes.
  if (field.type && INTEGERS.has(`CODE_${field.type}`)) return "Count";
  return "Number";
}

function timeSemanticType(grain: V1TimeGrain | undefined): FlintSemanticType {
  if (!grain) return "DateTime";
  return TIME_GRAIN_TO_SEMANTIC_TYPE[grain] ?? "DateTime";
}

/**
 * Semantic types and display names derived from the result's field metadata. The server sends
 * both already; this is the fallback for a result without them, and it fills the fields the
 * server's maps do not mention.
 */
export function deriveFlintFields(
  fields: Record<string, WireField>,
  grain: V1TimeGrain | undefined,
): FlintFields {
  const semantic_types: Record<string, FlintSemanticType> = {};
  const field_display_names: Record<string, string> = {};
  for (const [name, field] of Object.entries(fields)) {
    switch (field.kind) {
      case "measure":
        semantic_types[name] = measureSemanticType(field);
        break;
      case "time":
        semantic_types[name] = timeSemanticType(grain);
        break;
      default:
        semantic_types[name] = "Category";
    }
    field_display_names[name] =
      field.kind === "time"
        ? TIME_DISPLAY_NAME
        : field.display_name || field.name || name;
  }
  return { semantic_types, field_display_names };
}

/** Measure columns present in the rows: the measures plus any `<measure>_prev` companions. */
export function measureFieldsInRows(
  fields: Record<string, WireField>,
  rows: WireRow[],
): string[] {
  const sample = rows.slice(0, 50);
  const present = (name: string) => sample.some((row) => name in row);
  const result: string[] = [];
  for (const [name, field] of Object.entries(fields)) {
    if (field.kind !== "measure") continue;
    result.push(name);
    const previous = name + ComparisonDeltaPreviousSuffix;
    if (present(previous)) result.push(previous);
  }
  return result;
}

/**
 * The channel whose single measure can be folded with its previous period: `y`, or `x` when the
 * measure sits on x (a horizontal layout). Undefined when the chart type does not fold, the
 * channel binds several fields, or the series channel is already taken.
 */
export function comparisonChannel(
  chartSpec: WireFlintChartSpec,
  fields: Record<string, WireField>,
): { channel: string; measure: string } | undefined {
  const fold = COMPARISON_FOLDS.get(chartSpec.chartType);
  if (!fold) return undefined;
  const taken = (channel: string) =>
    encodingFields(chartSpec.encodings?.[channel]).length > 0;
  // A grouped bar chart has no color channel, so a colored bar chart cannot become one.
  if (taken(fold.channel) || (fold.channel === "group" && taken("color"))) {
    return undefined;
  }
  for (const channel of ["y", "x"]) {
    const field = singleEncodingField(chartSpec.encodings?.[channel]);
    if (field && fields[field]?.kind === "measure") {
      return { channel, measure: field };
    }
  }
  return undefined;
}

/**
 * The rows in long form: one row per period, the measure holding the period's value and
 * `rill_period` naming it. The previous-period column is dropped.
 */
function foldRows(
  rows: WireRow[],
  measure: string,
  labels: { current: string; previous: string },
): WireRow[] {
  const previous = measure + ComparisonDeltaPreviousSuffix;
  const folded: WireRow[] = [];
  for (const row of rows) {
    const { [previous]: previousValue, ...rest } = row;
    folded.push({ ...rest, [COMPARISON_SERIES_FIELD]: labels.current });
    folded.push({
      ...rest,
      [measure]: previousValue ?? null,
      [COMPARISON_SERIES_FIELD]: labels.previous,
    });
  }
  return folded;
}

export function buildFlintCompileInput(
  chartSpec: WireFlintChartSpec,
  result: Pick<
    ChartDataPayload,
    "fields" | "semantic_types" | "field_display_names"
  >,
  rows: WireRow[],
  timeGrain: V1TimeGrain | undefined,
): FlintCompileInput {
  const wireFields = result.fields ?? {};
  const derived = deriveFlintFields(wireFields, timeGrain);
  const fields: FlintFields = {
    semantic_types: {
      ...derived.semantic_types,
      ...(result.semantic_types as Record<string, FlintSemanticType>),
    },
    field_display_names: {
      ...derived.field_display_names,
      ...result.field_display_names,
    },
  };
  const spec = stripFlintChartSpec(chartSpec);
  let measureFields = measureFieldsInRows(wireFields, rows);

  // Fold the previous period in as a second series: the rows go long, the series channel
  // names the period, and a bar chart becomes a grouped bar chart so the periods sit side by side.
  let comparison: FlintCompileInput["comparison"];
  const target = comparisonChannel(chartSpec, wireFields);
  const previous = target && target.measure + ComparisonDeltaPreviousSuffix;
  if (target && previous && measureFields.includes(previous)) {
    const fold = COMPARISON_FOLDS.get(chartSpec.chartType)!;
    const current =
      fields.field_display_names[target.measure] ?? target.measure;
    const labels = {
      current,
      previous:
        fields.field_display_names[previous] ??
        m.mcp_chart_previous_period_name({ name: current }),
    };
    rows = foldRows(rows, target.measure, labels);
    spec.chartType = fold.chartType;
    (spec.encodings ??= {})[fold.channel] = COMPARISON_SERIES_FIELD;
    if (fold.chartType === "Area Chart") {
      // Layered rather than stacked areas: the periods are alternatives, not parts of a whole.
      spec.chartProperties = { stackMode: "layered", ...spec.chartProperties };
    }
    fields.semantic_types[COMPARISON_SERIES_FIELD] = "Category";
    fields.field_display_names[COMPARISON_SERIES_FIELD] =
      m.mcp_chart_series_period();
    measureFields = measureFields.filter((name) => name !== previous);
    comparison = { channel: fold.channel, measure: target.measure };
  }

  return { spec, fields, rows, measureFields, comparison };
}
