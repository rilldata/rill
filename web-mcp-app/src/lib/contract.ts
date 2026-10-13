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
 */

/** A tool result as the host forwards it: the MCP SDK's `CallToolResult`. */
export type ToolResult = AppToolResult;

/** The `widget` value of a result this bundle renders; other values come from newer servers. */
export const CHART_WIDGET_MODE = "chart";

/**
 * The chart types `create_chart` can return. Listed by hand because web-common's
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

/** The validated `create_chart` input; the field configs are Rill's chart spec vocabulary. */
export interface WireChartSpec {
  metrics_view?: string;
  time_range?: WireTimeRange;
  comparison_time_range?: WireTimeRange;
  time_grain?: V1TimeGrain;
  where?: unknown;
  [key: string]: unknown;
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

export interface ChartResult extends ChartDataPayload {
  widget: typeof CHART_WIDGET_MODE;
  chart_type: ChartType;
  spec: WireChartSpec;
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
  if (payload.data !== undefined && !Array.isArray(payload.data)) {
    return { kind: "invalid", reason: "data is not an array", openUrl };
  }
  if (payload.fields !== undefined && !isRecord(payload.fields)) {
    return { kind: "invalid", reason: "fields is not an object", openUrl };
  }
  return { kind: "chart", result: payload as unknown as ChartResult };
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
 * The result with its data section replaced by what `chart_data` returned for `spec`.
 * The section is replaced rather than merged because the backend omits the optional keys
 * it has nothing for (`domain_values`, `totals`, `truncation_warning`, ...), and a value
 * from the previous result must not show through.
 */
export function withChartData(
  result: ChartResult,
  spec: WireChartSpec,
  data: ChartDataPayload,
): ChartResult {
  const { widget, chart_type, message, title, metrics_view, theme, open_url } =
    result;
  return {
    widget,
    chart_type,
    spec,
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
