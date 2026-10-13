import type { ChartType } from "@rilldata/web-common/features/components/charts/types";
import type {
  V1ThemeSpec,
  V1TimeGrain,
} from "@rilldata/web-common/runtime-client/gen/index.schemas";
import type { AppToolResult } from "@modelcontextprotocol/ext-apps";

/**
 * Wire contract between Rill's `create_chart` / `chart_data` MCP tools and this widget
 * (`CreateChartResult` in `runtime/ai/create_chart.go`); the backend emits it as `structuredContent`.
 * The payload is untrusted input: `parseChartPayload` checks the shape before anything renders.
 *
 * Two result shapes exist. A Flint result carries `input`, the effective Flint chart spec plus the
 * metrics view binding the rows were queried with, and the widget compiles it with Flint. A Rill
 * result carries `spec`, Rill's own chart config for one of the 11 built-in chart types, and renders
 * through web-common's chart components; it is what older runtimes answer with.
 */

/** A tool result as the host forwards it: the MCP SDK's `CallToolResult`. */
export type ToolResult = AppToolResult;

/** The `widget` value of a result this bundle renders; other values come from newer servers. */
export const CHART_WIDGET_MODE = "chart";

/**
 * The Rill chart types `create_chart` can return. Listed by hand because web-common's
 * `CHART_CONFIG` cannot be imported here: it pulls the icons and the query providers into
 * the bundle.
 */
export const CHART_TYPES: ChartType[] = [
  "bar_chart",
  "line_chart",
  "area_chart",
  "stacked_bar",
  "stacked_bar_normalized",
  "donut_chart",
  "pie_chart",
  "funnel_chart",
  "heatmap",
  "combo_chart",
  "scatter_plot",
];

const CARTESIAN_CHART_TYPES: ReadonlySet<ChartType> = new Set([
  "bar_chart",
  "line_chart",
  "area_chart",
  "stacked_bar",
  "stacked_bar_normalized",
]);

export interface WireTimeRange {
  start: string;
  end: string;
  time_zone?: string;
}

export type WireFieldKind = "measure" | "dimension" | "time";

export interface WireField {
  name: string;
  kind: WireFieldKind;
  display_name?: string;
  description?: string;
  type?: string;
  format_preset?: string;
  format_d3?: string;
  format_d3_locale?: Record<string, unknown>;
  lower_is_better?: boolean;
}

export interface WireMetricsView {
  name: string;
  display_name?: string;
  time_dimension?: string;
  smallest_time_grain?: V1TimeGrain;
}

export type WireRow = Record<string, unknown>;

/** The validated `create_chart` input of a Rill result; the field configs are Rill's chart spec vocabulary. */
export interface WireChartSpec {
  metrics_view?: string;
  time_range?: WireTimeRange;
  comparison_time_range?: WireTimeRange;
  time_grain?: V1TimeGrain;
  where?: unknown;
  [key: string]: unknown;
}

/** A Flint encoding object: Flint's six keys plus Rill's `limit`, the top-N of a dimension on the channel. */
export interface WireFlintEncoding {
  field?: string;
  type?: "quantitative" | "nominal" | "ordinal" | "temporal";
  aggregate?: "count" | "sum" | "average" | "mean";
  sortOrder?: "ascending" | "descending";
  sortBy?: string;
  scheme?: string;
  limit?: number;
}

/** A channel binds a field name, an encoding object, or an array of either (Flint's series fold). */
export type WireFlintEncodingValue =
  | string
  | WireFlintEncoding
  | (string | WireFlintEncoding)[];

/** Flint's `chart_spec` as the server echoes it, with the effective limits filled in. */
export interface WireFlintChartSpec {
  chartType: string;
  title?: string;
  subtitle?: string;
  encodings: Record<string, WireFlintEncodingValue>;
  chartProperties?: Record<string, unknown>;
}

/** Rill's data binding: what the rows were queried with, in Flint's `data` slot. */
export interface WireDataBinding {
  metrics_view: string;
  time_range: WireTimeRange;
  comparison_time_range?: WireTimeRange;
  where?: unknown;
  time_grain?: V1TimeGrain;
}

/** The effective input of a Flint result; `chart_data` takes the same shape. */
export interface WireFlintInput {
  chart_spec: WireFlintChartSpec;
  data: WireDataBinding;
}

/** Data fields shared by `create_chart` and the app-only `chart_data` tool. */
export interface ChartDataPayload {
  time_grain?: V1TimeGrain;
  /**
   * Rows keyed by field name; `[]` when the query returned none. The key is missing only
   * when the host did not forward the data section (see `hasChartData`).
   */
  data?: WireRow[];
  fields?: Record<string, WireField>;
  /** Flint semantic type per field, derived by the server from the metrics view and the grain. */
  semantic_types?: Record<string, string>;
  /** Display name per field for Flint's axis and legend titles. */
  field_display_names?: Record<string, string>;
  domain_values?: Record<string, unknown[]>;
  totals?: Record<string, number>;
  resolved_time_range?: WireTimeRange;
  resolved_comparison_time_range?: WireTimeRange;
  /** Set when any budget cut the data: a top-N cap on a nominal channel or the row limit. */
  truncated?: boolean;
  /** Set when the row limit cut the rows; a time series keeps its newest buckets. */
  rows_truncated?: boolean;
  /** The server's explanation of what was cut, written for the model. */
  truncation_warning?: string;
}

interface ChartResultBase extends ChartDataPayload {
  widget: typeof CHART_WIDGET_MODE;
  message?: string;
  title?: string;
  metrics_view?: WireMetricsView;
  theme?: V1ThemeSpec;
  open_url?: string;
}

/** A result rendered with web-common's chart components from Rill's chart config. */
export interface RillChartResult extends ChartResultBase {
  chart_type: ChartType;
  spec: WireChartSpec;
  input?: undefined;
}

/** A result compiled with Flint from its effective input. */
export interface FlintChartResult extends ChartResultBase {
  chart_type: string;
  input: WireFlintInput;
  spec?: undefined;
}

export type ChartResult = RillChartResult | FlintChartResult;

export type ParsedPayload =
  | { kind: "chart"; result: ChartResult }
  | { kind: "unsupported_widget"; widget: string; openUrl?: string }
  | { kind: "invalid"; reason: string; openUrl?: string };

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

function optionalString(value: unknown): string | undefined {
  return typeof value === "string" && value.length > 0 ? value : undefined;
}

export function isChartType(value: unknown): value is ChartType {
  return typeof value === "string" && CHART_TYPES.includes(value as ChartType);
}

export function isCartesianChartType(chartType: ChartType): boolean {
  return CARTESIAN_CHART_TYPES.has(chartType);
}

export function isFlintResult(result: ChartResult): result is FlintChartResult {
  return result.input !== undefined;
}

/** The rows of a data section, dropping anything that is not an object; no section is no rows. */
export function normalizeRows(data: WireRow[] | undefined): WireRow[] {
  return (data ?? []).filter((row): row is WireRow => isRecord(row));
}

/**
 * Shape check for an untrusted payload.
 * A widget mode other than "chart" is reported so the UI can fall back to "Open in Rill".
 */
export function parseChartPayload(payload: unknown): ParsedPayload {
  if (!isRecord(payload)) {
    return { kind: "invalid", reason: "structuredContent is not an object" };
  }
  const openUrl = optionalString(payload.open_url);
  if (typeof payload.widget !== "string") {
    return { kind: "invalid", reason: "widget is missing", openUrl };
  }
  if (payload.widget !== CHART_WIDGET_MODE) {
    return { kind: "unsupported_widget", widget: payload.widget, openUrl };
  }
  if (payload.input !== undefined) {
    const reason = invalidFlintInput(payload.input);
    if (reason) return { kind: "invalid", reason, openUrl };
    if (typeof payload.chart_type !== "string") {
      return { kind: "invalid", reason: "chart_type is missing", openUrl };
    }
  } else {
    if (!isChartType(payload.chart_type)) {
      return {
        kind: "invalid",
        reason: `unknown chart_type ${JSON.stringify(payload.chart_type)}`,
        openUrl,
      };
    }
    if (!isRecord(payload.spec)) {
      return { kind: "invalid", reason: "spec is not an object", openUrl };
    }
  }
  if (payload.data !== undefined && !Array.isArray(payload.data)) {
    return { kind: "invalid", reason: "data is not an array", openUrl };
  }
  if (payload.fields !== undefined && !isRecord(payload.fields)) {
    return { kind: "invalid", reason: "fields is not an object", openUrl };
  }
  return { kind: "chart", result: payload as unknown as ChartResult };
}

/** The reason a Flint `input` is unusable, or undefined when it has what the compiler needs. */
function invalidFlintInput(input: unknown): string | undefined {
  if (!isRecord(input)) return "input is not an object";
  const chartSpec = input.chart_spec;
  if (!isRecord(chartSpec)) return "input.chart_spec is not an object";
  if (!optionalString(chartSpec.chartType)) {
    return "input.chart_spec.chartType is missing";
  }
  if (!isRecord(chartSpec.encodings)) {
    return "input.chart_spec.encodings is not an object";
  }
  if (!isRecord(input.data)) return "input.data is not an object";
  return undefined;
}

/**
 * True when the result carries its data section. The backend always sends `data`, as `[]`
 * for a query without rows; the key is missing only when the host did not forward the
 * section (e.g. a large result written to a file).
 */
export function hasChartData(result: ChartDataPayload): boolean {
  return Array.isArray(result.data);
}

/**
 * The result with its data section replaced by what `chart_data` returned for `patched`,
 * the spec (Rill result) or the input (Flint result) the data was requested with.
 * The section is replaced rather than merged because the backend omits the optional keys
 * it has nothing for (`domain_values`, `totals`, `truncation_warning`, ...), and a value
 * from the previous result must not show through.
 */
export function withChartData(
  result: ChartResult,
  patched: WireChartSpec | WireFlintInput,
  data: ChartDataPayload,
): ChartResult {
  const { widget, chart_type, message, title, metrics_view, theme, open_url } =
    result;
  const base = {
    widget,
    message,
    title,
    metrics_view,
    theme,
    open_url,
    ...data,
  };
  if (isFlintResult(result)) {
    return { ...base, chart_type, input: patched as WireFlintInput };
  }
  return {
    ...base,
    chart_type: result.chart_type,
    spec: patched as WireChartSpec,
  };
}

/** Concatenated text blocks of a tool result; used for error messages. */
export function textContent(result: ToolResult): string {
  return (result.content ?? [])
    .map((block) =>
      block.type === "text" && typeof block.text === "string" ? block.text : "",
    )
    .filter(Boolean)
    .join("\n");
}
