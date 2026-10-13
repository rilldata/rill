import { assembleVegaLite } from "flint-chart/vegalite";
import type { VisualizationSpec } from "svelte-vega";
import { sanitizeFieldName } from "@rilldata/web-common/components/vega/util";
import type { FlintFields } from "./semantic-types";

/**
 * Compiles a Flint chart spec into a Vega-Lite spec.
 *
 * Flint is a compiler, not a renderer: it derives scales, axes, formats, sorting, stacking and
 * layout from the data and the declared semantics, then emits a Vega-Lite spec. That spec goes
 * through the same VegaLiteRenderer, theme and measure formatters as the rest of Rill's charts.
 *
 * Flint's types are intentionally not imported: only values are re-exported from the
 * `flint-chart/vegalite` subpath, and the package is pre-1.0 and moving quickly, so the narrow
 * shapes this module actually depends on are declared locally.
 */

/** The complete Flint encoding vocabulary. */
export interface FlintEncoding {
  field?: string;
  type?: "quantitative" | "nominal" | "ordinal" | "temporal";
  aggregate?: "count" | "sum" | "average" | "mean";
  sortOrder?: "ascending" | "descending";
  sortBy?: string;
  scheme?: string;
}

/** A channel binds to a field name, an encoding object, or an array of either. */
export type FlintEncodingValue =
  | string
  | FlintEncoding
  | (string | FlintEncoding)[];

/** A Flint chart spec, as authored under `custom_chart.spec` in a component file. */
export interface FlintChartSpec {
  chartType?: string;
  encodings?: Record<string, FlintEncodingValue>;
  chartProperties?: Record<string, unknown>;
}

export interface FlintWarning {
  severity: "info" | "warning" | "error";
  code: string;
  message: string;
  channel?: string;
  field?: string;
}

export interface CompileOptions {
  spec: FlintChartSpec;
  rows: Record<string, unknown>[];
  fields: FlintFields;
  /** Measured size of the container, used as Flint's hard layout ceiling. */
  size?: { width: number; height: number };
  /** Measure names to route through Rill's `rill_<measure>` Vega expression formatters. */
  measureFields?: string[];
}

export interface CompileResult {
  spec: VisualizationSpec | null;
  warnings: FlintWarning[];
  error: string | null;
}

export function compileFlintSpec({
  spec,
  rows,
  fields,
  size,
  measureFields = [],
}: CompileOptions): CompileResult {
  if (!spec?.chartType) {
    return { spec: null, warnings: [], error: "No chart type provided" };
  }

  let assembled: Record<string, unknown>;
  try {
    assembled = assembleVegaLite({
      data: { values: rows },
      semantic_types: fields.semantic_types,
      field_display_names: fields.field_display_names,
      chart_spec: {
        chartType: spec.chartType,
        encodings: spec.encodings ?? {},
        // Flint treats baseSize as a target and canvasSize as a ceiling it may not exceed.
        // Passing the measured box as both pins the chart to its canvas slot.
        ...(size ? { baseSize: size, canvasSize: size } : {}),
        ...(spec.chartProperties
          ? { chartProperties: spec.chartProperties }
          : {}),
      },
      options: {
        // Off by default in Flint.
        addTooltips: true,
        // Flint reserves ~35px beside a facet grid for its y-axis labels and ~22px below it;
        // Rill's larger labels, their padding, the row headers and the legend on top need
        // more, or the grid overflows its box.
        facetFixedPadding: { width: 100, height: 100 },
        // Flint drops categories once their facet panels would be under 60px; rows of a
        // chart read fine down to 40px, so fewer categories are left out of a short box.
        minSubplotSize: 40,
      },
    }) as Record<string, unknown>;
  } catch (e) {
    return {
      spec: null,
      warnings: [],
      error: e instanceof Error ? e.message : String(e),
    };
  }

  const warnings = (assembled._warnings as FlintWarning[] | undefined) ?? [];
  const vegaSpec = stripInternalKeys(assembled);
  deferToTheme(vegaSpec, spec.encodings ?? {});
  applyMeasureFormatters(vegaSpec, measureFields);
  labelDerivedEncodings(vegaSpec, fields, measureFields);
  labelFoldedSeries(vegaSpec, fields);
  formatDiscreteTimeAxes(vegaSpec, fields.semantic_types);
  scaleDonutHole(vegaSpec);
  fitSingleViewToContainer(vegaSpec, size);
  fillRowFacetWidth(vegaSpec, size);

  return { spec: vegaSpec as VisualizationSpec, warnings, error: null };
}

/**
 * Makes the renderer's measured width and height describe the whole chart, including axes and
 * legends, rather than only its plot area.
 *
 * Vega-Lite's default `pad` autosize treats `view.width()` / `view.height()` as the plot size and
 * adds guides outside it. That is why charts with a bottom legend stayed a fixed number of pixels
 * taller than their canvas slot while charts with a right-side legend happened to fit. `fit`
 * subtracts those guides from the plot; `contains: "padding"` makes the requested dimensions
 * include Vega's outer padding instead of adding it beyond the container.
 *
 * The view is pinned to the measured size rather than left unsized. A discrete axis without a
 * width falls back to Vega-Lite's default 20px step, and Vega-Lite drops `fit` on that axis, so a
 * ten-bar chart rendered 200px wide in any container. A numeric size keeps `fit` in effect; the
 * renderer then sets the live size as the container resizes.
 *
 * Vega-Lite does not support fit autosizing for composed or faceted views; Flint already sizes
 * their subplots to its canvas, so leave those specs untouched.
 */
function fitSingleViewToContainer(
  spec: Record<string, unknown>,
  size: { width: number; height: number } | undefined,
): void {
  const multiViewKeys = ["facet", "repeat", "concat", "hconcat", "vconcat"];
  if (multiViewKeys.some((key) => key in spec)) return;
  const encoding = spec.encoding as Record<string, unknown> | undefined;
  if (encoding && ["facet", "row", "column"].some((key) => key in encoding)) {
    return;
  }

  // Some Flint templates (notably pie/donut) put fixed dimensions directly on the spec, while
  // most use config.view defaults. Both take precedence over responsive container sizing.
  delete spec.width;
  delete spec.height;
  const config = spec.config as Record<string, unknown> | undefined;
  const view = config?.view as Record<string, unknown> | undefined;
  if (view) {
    delete view.continuousWidth;
    delete view.continuousHeight;
    if (Object.keys(view).length === 0) delete config?.view;
  }

  if (size) {
    spec.width = size.width;
    spec.height = size.height;
  }
  spec.autosize = { type: "fit", contains: "padding" };
}

/**
 * Room on the left of a row-faceted chart for the y-axis labels and the row header title, in
 * pixels, at Rill's label sizes and paddings.
 */
const ROW_FACET_GUTTER = 110;
/** The widest a row's upright label may be before it is cut, in pixels. */
const ROW_FACET_LABEL_LIMIT = 110;

/**
 * Flint keeps a row-faceted chart's panels near the proportions of a single chart, so a handful
 * of rows in a wide box come out as a narrow column with most of the width unused. In Rill a
 * stack of rows reads like a list of sparklines: each panel spans the box, less the gutter, and
 * its label stands upright beside it rather than rotated and cut to the row's height.
 * Panels with a discrete x axis are sized by their band step and are left to Flint.
 */
function fillRowFacetWidth(
  spec: Record<string, unknown>,
  size: { width: number; height: number } | undefined,
): void {
  if (!size) return;
  const encoding = (spec.encoding ?? {}) as Record<string, unknown>;
  const facet = (spec.facet ?? {}) as Record<string, unknown>;
  const hasRow = "row" in encoding || "row" in facet;
  const hasColumns =
    "column" in encoding ||
    "column" in facet ||
    "facet" in encoding ||
    (spec.facet !== undefined && !("row" in facet) && !("column" in facet));
  if (!hasRow || hasColumns) return;
  // A width in steps belongs to a discrete x axis.
  if (spec.width !== undefined && typeof spec.width !== "number") return;

  const config = (spec.config as Record<string, unknown> | undefined) ?? {};
  spec.config = config;
  const view = (config.view as Record<string, unknown> | undefined) ?? {};
  config.view = view;
  view.continuousWidth = Math.max(
    120,
    Math.round(size.width - ROW_FACET_GUTTER - ROW_FACET_LABEL_LIMIT),
  );
  config.headerRow = {
    ...(config.headerRow as Record<string, unknown> | undefined),
    labelAngle: 0,
    labelAlign: "right",
    labelLimit: ROW_FACET_LABEL_LIMIT,
  };
}

/**
 * Flint takes a donut's inner radius in pixels, on a 0–100 scale suited to its own small default
 * canvas, so the hole all but vanishes in a chart filling a dashboard panel. Read that setting as
 * a percentage of the outer radius instead, so the donut keeps its proportions at any size.
 */
function scaleDonutHole(spec: Record<string, unknown>): void {
  const mark = spec.mark as Record<string, unknown> | undefined;
  if (mark?.type !== "arc" || typeof mark.innerRadius !== "number") return;
  mark.innerRadius = {
    expr: `min(width, height) / 2 * ${mark.innerRadius / 100}`,
  };
}

/**
 * Removes Flint's non-standard bookkeeping keys (`_warnings`, `_width`, `_height`, `_options`,
 * `_transform`, `_pivot`). Vega-Lite rejects unknown top-level properties, so these must not reach
 * the renderer. Matching on the underscore prefix rather than a fixed list keeps this working as
 * Flint adds more of them.
 */
function stripInternalKeys(
  assembled: Record<string, unknown>,
): Record<string, unknown> {
  return Object.fromEntries(
    Object.entries(assembled).filter(([key]) => !key.startsWith("_")),
  );
}

/** Flint's config keys that style text; Rill's theme owns typography and colors. */
const THEME_OWNED_CONFIG_KEYS = [
  "labelColor",
  "titleColor",
  "labelFont",
  "titleFont",
  "labelFontSize",
  "titleFontSize",
  "labelFontWeight",
  "titleFontWeight",
];

/** The config groups Flint styles; `view`, `facet` and the label limits in these are kept. */
const STYLED_CONFIG_GROUPS = [
  "axis",
  "axisX",
  "axisY",
  "legend",
  "header",
  "headerRow",
  "headerColumn",
  "headerFacet",
];

/** The channels drawn from a color scale. */
const COLOR_CHANNELS = new Set(["color", "fill", "stroke"]);

/**
 * Flint emits its own typography (8–10px labels, grey titles) and a `tableau10` scheme on every
 * nominal color encoding. A spec's config and scales take precedence over the config the renderer
 * passes, so left in place they would override Rill's theme: dark mode would draw grey-on-grey
 * titles and a project palette would never reach a Flint chart. Those keys are removed so the
 * theme applies; a `scheme` the author chose is kept.
 */
function deferToTheme(
  spec: Record<string, unknown>,
  authored: Record<string, FlintEncodingValue>,
): void {
  const config = spec.config as Record<string, unknown> | undefined;
  if (config) {
    for (const group of STYLED_CONFIG_GROUPS) {
      const entry = config[group] as Record<string, unknown> | undefined;
      if (!entry) continue;
      for (const key of THEME_OWNED_CONFIG_KEYS) delete entry[key];
      if (Object.keys(entry).length === 0) delete config[group];
    }
  }

  const hasAuthoredScheme = (channel: string) => {
    const value = authored[channel];
    return (
      !!value &&
      typeof value === "object" &&
      !Array.isArray(value) &&
      value.scheme !== undefined
    );
  };
  for (const view of layersOf(spec)) {
    for (const [channel, encoding] of Object.entries(encodingsOf(view))) {
      if (Array.isArray(encoding) || !COLOR_CHANNELS.has(channel)) continue;
      if (hasAuthoredScheme(channel)) continue;
      const scale = encoding.scale as Record<string, unknown> | undefined;
      if (!scale || scale.scheme === undefined) continue;
      delete scale.scheme;
      if (Object.keys(scale).length === 0) delete encoding.scale;
    }
  }
}

/** The series key of Flint's multi-measure fold; its values are the folded fields' names. */
const SERIES_KEY = "__flint_series_key";

/**
 * A chart with several measures on one channel is drawn from a fold whose series are keyed by
 * the measures' names, which is what the legend and the tooltip then show. The keys take the
 * measures' display names, in the data and in any explicit order the spec gives the series.
 */
function labelFoldedSeries(
  spec: Record<string, unknown>,
  fields: FlintFields,
): void {
  const values = (
    spec.data as { values?: Record<string, unknown>[] } | undefined
  )?.values;
  if (!values?.length || !(SERIES_KEY in values[0])) return;
  const label = (value: unknown) =>
    typeof value === "string"
      ? (fields.field_display_names[value] ?? value)
      : value;

  for (const row of values) row[SERIES_KEY] = label(row[SERIES_KEY]);
  for (const view of layersOf(spec)) {
    for (const encoding of Object.values(encodingsOf(view))) {
      if (Array.isArray(encoding) || encoding.field !== SERIES_KEY) continue;
      if (Array.isArray(encoding.sort))
        encoding.sort = encoding.sort.map(label);
      const scale = encoding.scale as Record<string, unknown> | undefined;
      if (Array.isArray(scale?.domain)) scale.domain = scale.domain.map(label);
    }
  }
}

/**
 * Routes measure axis and tooltip labels through Rill's registered `rill_<measure>` formatters.
 *
 * Flint picks a number format from the semantic type alone, so it has no way to honour a measure's
 * `format_d3` or currency/percent preset. Rill registers a Vega expression function per measure
 * (see Chart.svelte); pointing `formatType` at it restores measure formatting. Flint's own `format`
 * is removed where we override, since Vega-Lite would otherwise pass it to the formatter as an argument.
 */
function applyMeasureFormatters(
  spec: Record<string, unknown>,
  measureFields: string[],
): void {
  if (!measureFields.length) return;
  const measures = new Set(measureFields);

  const visit = (node: unknown): void => {
    if (Array.isArray(node)) {
      node.forEach(visit);
      return;
    }
    if (!node || typeof node !== "object") return;

    const record = node as Record<string, unknown>;
    if (
      typeof record.field === "string" &&
      record.type === "quantitative" &&
      measures.has(record.field)
    ) {
      const formatType = sanitizeFieldName(record.field);
      const axis = record.axis as Record<string, unknown> | undefined;
      if (axis) {
        delete axis.format;
        axis.formatType = formatType;
      }
      delete record.format;
      record.formatType = formatType;
    }

    Object.values(record).forEach(visit);
  };

  visit(spec);
}

/** An encoding definition as the post-processing steps read it; other keys pass through untouched. */
type EncodingRecord = Record<string, unknown> & {
  field?: unknown;
  type?: unknown;
  title?: unknown;
  format?: unknown;
  formatType?: unknown;
  axis?: Record<string, unknown> | null;
  legend?: Record<string, unknown> | null;
};

/** The channels whose title is shown as an axis or legend title; `x2`, `y2` and the tooltip have none. */
const TITLED_CHANNELS = new Set([
  "x",
  "y",
  "color",
  "size",
  "shape",
  "opacity",
  "strokeDash",
  "theta",
  "radius",
]);

/** The channels drawn with a legend, where a label expression goes on `legend` rather than `axis`. */
const LEGEND_CHANNELS = new Set([
  "color",
  "size",
  "shape",
  "opacity",
  "strokeDash",
]);

/**
 * A view and every view nested in it: the view a facet or repeat wraps, the layers of a layered
 * chart and the panels of a concatenated one (a bar table is a concatenation whose panels each
 * carry the category axis).
 */
function layersOf(spec: Record<string, unknown>): Record<string, unknown>[] {
  const views = [spec];
  const wrapped = spec.spec;
  if (wrapped && typeof wrapped === "object") {
    views.push(...layersOf(wrapped as Record<string, unknown>));
  }
  for (const key of ["layer", "hconcat", "vconcat", "concat"]) {
    const children = spec[key];
    if (!Array.isArray(children)) continue;
    for (const child of children as Record<string, unknown>[]) {
      views.push(...layersOf(child));
    }
  }
  return views;
}

function encodingsOf(
  view: Record<string, unknown>,
): Record<string, EncodingRecord> {
  const encoding = view.encoding;
  return encoding && typeof encoding === "object"
    ? (encoding as Record<string, EncodingRecord>)
    : {};
}

/**
 * Flint's derived encodings keep raw names where Rill's labels and formats belong. A waterfall titles
 * its running sums with the measure's name, its connector layer carries no title, and its step axis
 * has none either, so Vega-Lite composes one from the field names ("event_time, __wf_lead").
 *
 * A quantitative encoding titled with a measure name takes the measure's display name and formatter.
 * In a layered chart the other layers' encodings on that channel share its scale and axis, so they
 * take the formatter too. And an encoding of a known field without a title gets the field's display
 * name, on the channels whose title is shown.
 */
function labelDerivedEncodings(
  spec: Record<string, unknown>,
  fields: FlintFields,
  measureFields: string[],
): void {
  if (!measureFields.length) return;
  const measures = new Set(measureFields);
  const views = layersOf(spec);

  // The measure each channel draws, found by field or by Flint's raw-name title in any layer.
  const channelMeasure = new Map<string, string>();
  for (const view of views) {
    for (const [channel, encoding] of Object.entries(encodingsOf(view))) {
      if (Array.isArray(encoding) || channelMeasure.has(channel)) continue;
      const measure = [encoding.field, encoding.title].find(
        (name): name is string =>
          typeof name === "string" && measures.has(name),
      );
      if (measure) channelMeasure.set(channel, measure);
    }
  }

  for (const view of views) {
    for (const [channel, encoding] of Object.entries(encodingsOf(view))) {
      if (Array.isArray(encoding)) continue;
      const measure = channelMeasure.get(channel);
      if (
        measure &&
        encoding.type === "quantitative" &&
        encoding.field !== measure
      ) {
        routeThroughFormatter(encoding, sanitizeFieldName(measure));
        if (encoding.title === measure) {
          encoding.title = fields.field_display_names[measure] ?? measure;
        }
      }
      if (
        encoding.title === undefined &&
        TITLED_CHANNELS.has(channel) &&
        typeof encoding.field === "string" &&
        fields.field_display_names[encoding.field]
      ) {
        encoding.title = fields.field_display_names[encoding.field];
      }
    }
  }
}

function routeThroughFormatter(
  encoding: EncodingRecord,
  formatType: string,
): void {
  if (encoding.axis) {
    delete encoding.axis.format;
    encoding.axis.formatType = formatType;
  }
  delete encoding.format;
  encoding.formatType = formatType;
}

/**
 * Label formats for a time field drawn on a discrete axis or legend, keyed by Flint's temporal
 * types (the ones `semantic-types.ts` emits). `expr` formats an axis or legend label from
 * `datum.value`, which is a Date or an ISO string depending on who supplied the rows; `d3` is
 * the tooltip's time format. Vega's `timeFormat` has no quarter directive, so that one is spelled out.
 */
const DISCRETE_TIME_FORMATS: Record<string, { expr: string; d3: string }> = {
  Year: { expr: "timeFormat(toDate(datum.value), '%Y')", d3: "%Y" },
  YearQuarter: {
    expr: "'Q' + (quarter(toDate(datum.value)) + 1) + ' ' + year(toDate(datum.value))",
    d3: "%b %Y",
  },
  YearMonth: { expr: "timeFormat(toDate(datum.value), '%b %Y')", d3: "%b %Y" },
  YearWeek: { expr: "timeFormat(toDate(datum.value), '%b %d')", d3: "%b %d" },
  Date: { expr: "timeFormat(toDate(datum.value), '%b %d')", d3: "%b %d" },
  DateTime: {
    expr: "timeFormat(toDate(datum.value), '%b %d %H:%M')",
    d3: "%b %d %H:%M",
  },
};

/**
 * Some templates draw the time field on a discrete scale (a waterfall's steps, a slope chart's two
 * periods), and Vega-Lite then prints the raw timestamps as category labels. The data keeps its
 * chronological order; only the labels are formatted at the field's grain, on the axis or legend
 * that shows them, and in the tooltip.
 */
function formatDiscreteTimeAxes(
  spec: Record<string, unknown>,
  semanticTypes: Record<string, string>,
): void {
  const formatOf = (field: unknown) =>
    typeof field === "string"
      ? DISCRETE_TIME_FORMATS[semanticTypes[field]]
      : undefined;

  for (const view of layersOf(spec)) {
    for (const [channel, encoding] of Object.entries(encodingsOf(view))) {
      if (Array.isArray(encoding)) {
        for (const entry of encoding as EncodingRecord[]) {
          const format = formatOf(entry.field);
          if (!format || !isDiscrete(entry)) continue;
          entry.type = "temporal";
          entry.format = format.d3;
          delete entry.formatType;
        }
        continue;
      }
      const format = formatOf(encoding.field);
      if (!format || !isDiscrete(encoding)) continue;
      if (LEGEND_CHANNELS.has(channel)) {
        if (encoding.legend !== null) {
          encoding.legend = { ...encoding.legend, labelExpr: format.expr };
        }
      } else if (channel === "x" || channel === "y") {
        if (encoding.axis !== null) {
          encoding.axis = { ...encoding.axis, labelExpr: format.expr };
        }
      }
    }
  }
}

function isDiscrete(encoding: EncodingRecord): boolean {
  return encoding.type === "nominal" || encoding.type === "ordinal";
}
