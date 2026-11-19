import type {
  V1ThemeSpec,
  V1TimeGrain,
} from "@rilldata/web-common/runtime-client/gen/index.schemas";
import type { AppToolResult } from "@modelcontextprotocol/ext-apps";
import type { ChartEncoding } from "flint-chart/core";

/**
 * Wire contract between Rill's `create_chart` / `chart_data` MCP tools and this widget
 * (`CreateChartResult` in `runtime/ai/create_chart.go`); the backend emits it as `structuredContent`.
 * The payload is untrusted input: `parseChartPayload` checks the shape before anything renders.
 *
 * A result carries `input`, the effective Flint chart spec plus the metrics view binding the rows
 * were queried with, and the widget compiles it with Flint. Rill's own chat gets a different
 * result (`chart_type` and `spec`, without data), which never reaches this widget.
 */

/** A tool result as the host forwards it: the MCP SDK's `CallToolResult`. */
export type ToolResult = AppToolResult;

/** The `widget` value of a result this bundle renders; other values come from newer servers. */
export const CHART_WIDGET_MODE = "chart";

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

/** A Flint encoding object plus Rill's `limit`, the top-N of a dimension on the channel. */
export interface WireFlintEncoding extends ChartEncoding {
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
  /** The caller's semantic type overrides, per field; echoed so a refresh keeps them. */
  semantic_types?: Record<string, unknown>;
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

/** The `structuredContent` of a `chart_data` result. */
export interface ChartDataResponse extends ChartDataPayload {
  /** The effective input, with the limits the server applied (a facet cap, say). */
  input?: WireFlintInput;
}

/** A `create_chart` result, compiled with Flint from its effective input. */
export interface ChartResult extends ChartDataPayload {
  widget: typeof CHART_WIDGET_MODE;
  chart_type: string;
  input: WireFlintInput;
  message?: string;
  title?: string;
  metrics_view?: WireMetricsView;
  theme?: V1ThemeSpec;
  open_url?: string;
}

export type ParsedPayload =
  | { kind: "chart"; result: ChartResult }
  | { kind: "unsupported_widget"; widget: string; openUrl?: string }
  | { kind: "invalid"; reason: string; openUrl?: string };

export type ParsedChartData =
  | { kind: "data"; response: ChartDataResponse }
  | { kind: "invalid"; reason: string };

/** The keys of a data section that hold an object (per-field maps); `data` holds an array. */
const DATA_SECTION_OBJECT_KEYS = [
  "fields",
  "semantic_types",
  "field_display_names",
  "domain_values",
  "totals",
] as const satisfies readonly (keyof ChartDataPayload)[];

export function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

function optionalString(value: unknown): string | undefined {
  return typeof value === "string" && value.length > 0 ? value : undefined;
}

/**
 * The value when it is an absolute http(s) URL, else undefined: the host opens it as a link,
 * so a `javascript:` or `data:` URL from an untrusted payload must never reach it.
 */
export function httpUrl(value: unknown): string | undefined {
  if (typeof value !== "string") return undefined;
  try {
    const { protocol } = new URL(value);
    return protocol === "https:" || protocol === "http:" ? value : undefined;
  } catch {
    return undefined;
  }
}

/** The rows of a data section, dropping anything that is not an object; no section is no rows. */
export function normalizeRows(data: WireRow[] | undefined): WireRow[] {
  return (data ?? []).filter((row): row is WireRow => isRecord(row));
}

/**
 * Shape check for an untrusted payload.
 * A widget mode other than "chart" is reported so the UI can fall back to "Open in Rill".
 * An `open_url` that is not http(s) is dropped.
 */
export function parseChartPayload(payload: unknown): ParsedPayload {
  if (!isRecord(payload)) {
    return { kind: "invalid", reason: "structuredContent is not an object" };
  }
  const openUrl = httpUrl(payload.open_url);
  if (typeof payload.widget !== "string") {
    return { kind: "invalid", reason: "widget is missing", openUrl };
  }
  if (payload.widget !== CHART_WIDGET_MODE) {
    return { kind: "unsupported_widget", widget: payload.widget, openUrl };
  }
  if (typeof payload.chart_type !== "string") {
    return { kind: "invalid", reason: "chart_type is missing", openUrl };
  }
  const reason =
    invalidFlintInput(payload.input) ?? invalidDataSection(payload);
  if (reason) return { kind: "invalid", reason, openUrl };
  return {
    kind: "chart",
    result: { ...payload, open_url: openUrl } as unknown as ChartResult,
  };
}

/**
 * Shape check for the untrusted `structuredContent` of a `chart_data` result: a data section
 * that carries its rows (`[]` for none), and the effective input when the server sends one.
 */
export function parseChartDataResponse(payload: unknown): ParsedChartData {
  if (!isRecord(payload)) {
    return { kind: "invalid", reason: "structuredContent is not an object" };
  }
  if (!Array.isArray(payload.data)) {
    return { kind: "invalid", reason: "data is not an array" };
  }
  const reason =
    invalidDataSection(payload) ??
    (payload.input === undefined
      ? undefined
      : invalidFlintInput(payload.input));
  if (reason) return { kind: "invalid", reason };
  return { kind: "data", response: payload as ChartDataResponse };
}

/** The reason a data section is unusable, or undefined when its keys have the expected shapes. */
function invalidDataSection(
  payload: Record<string, unknown>,
): string | undefined {
  if (payload.data !== undefined && !Array.isArray(payload.data)) {
    return "data is not an array";
  }
  for (const key of DATA_SECTION_OBJECT_KEYS) {
    if (payload[key] !== undefined && !isRecord(payload[key])) {
      return `${key} is not an object`;
    }
  }
  return undefined;
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
 * The result with its data section replaced by what `chart_data` returned for `queried`,
 * the input the data answers.
 * The section is replaced rather than merged because the backend omits the optional keys
 * it has nothing for (`domain_values`, `totals`, `truncation_warning`, ...), and a value
 * from the previous result must not show through.
 */
export function withChartData(
  result: ChartResult,
  queried: WireFlintInput,
  data: ChartDataPayload,
): ChartResult {
  const { widget, chart_type, message, title, metrics_view, theme, open_url } =
    result;
  return {
    widget,
    chart_type,
    input: queried,
    message,
    title,
    metrics_view,
    theme,
    open_url,
    ...data,
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
