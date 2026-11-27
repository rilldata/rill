import type { ChartWarning, RawEncodingValue } from "flint-chart/core";
import { assembleVegaLite } from "flint-chart/vegalite";
import { extent, ticks } from "d3-array";
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
 * Flint's own types (from `flint-chart/core`) describe the encodings and the warnings, so a Flint
 * upgrade that changes them fails to type-check here rather than at render time.
 */

/**
 * A Flint chart spec, as authored under `custom_chart.spec` in a component file. A channel binds
 * a field name, an encoding object, or an array of either.
 */
export interface FlintChartSpec {
  chartType?: string;
  encodings?: Record<string, RawEncodingValue>;
  chartProperties?: Record<string, unknown>;
}

export interface CompileOptions {
  spec: FlintChartSpec;
  rows: Record<string, unknown>[];
  fields: FlintFields;
  /** Measured size of the container, used as Flint's hard layout ceiling. */
  size?: { width: number; height: number };
  /** Measure names to route through Rill's `rill_<measure>` Vega expression formatters. */
  measureFields?: string[];
  /**
   * Each measure's formatter by name, as its `rill_<measure>` expression applies it, so the
   * layout can make room for the labels it prints.
   */
  measureFormatters?: Record<string, (value: number) => string>;
}

export interface CompileResult {
  spec: VisualizationSpec | null;
  warnings: ChartWarning[];
  error: string | null;
}

export function compileFlintSpec({
  spec,
  rows,
  fields,
  size,
  measureFields = [],
  measureFormatters = {},
}: CompileOptions): CompileResult {
  if (!spec?.chartType) {
    return { spec: null, warnings: [], error: "No chart type provided" };
  }

  // Flint's donut is its pie under another name. The hole it gets by default has varied across
  // Flint versions (none before 1.0), so it is stated.
  const chartProperties =
    spec.chartType === "Donut Chart" &&
    spec.chartProperties?.innerRadius === undefined
      ? { ...spec.chartProperties, innerRadius: DEFAULT_DONUT_HOLE }
      : spec.chartProperties;

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
        ...(chartProperties ? { chartProperties } : {}),
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

  const warnings = (assembled._warnings as ChartWarning[] | undefined) ?? [];
  const vegaSpec = stripInternalKeys(assembled);
  const authored = spec.encodings ?? {};
  deferToTheme(vegaSpec, authored);
  orderSeries(vegaSpec);
  applyMeasureFormatters(vegaSpec, measureFields);
  switch (spec.chartType) {
    case "Bump Chart":
      rankBumpChart(vegaSpec, rows, fields, measureFields, size);
      break;
    case "Bar Table":
      formatBarTable(vegaSpec, authored, fields, measureFields, size);
      break;
    case "Violin Plot":
      shapeViolins(vegaSpec, rows, size);
      break;
    case "Rose Chart":
      placeRoseLabels(vegaSpec, rows);
      break;
    case "Marginal Plot":
      fitMarginalPlot(vegaSpec, rows, fields, size, measureFormatters);
      break;
    case "Sparkline":
      sortSparklineRows(vegaSpec);
      break;
  }
  labelDerivedEncodings(vegaSpec, fields, measureFields);
  formatMeasureLegends(vegaSpec);
  labelFoldedSeries(vegaSpec, fields);
  formatDiscreteTimeAxes(vegaSpec, fields.semantic_types);
  formatNormalizedStackAxes(vegaSpec);
  scaleDonutHole(vegaSpec);
  fitSingleViewToContainer(vegaSpec, size);
  fillRowFacetWidth(vegaSpec, rows, size);
  spaceTimeAxisTicks(vegaSpec, rows, fields.semantic_types, size);

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

  if (pinsArcs(spec)) return;

  // Some Flint templates (notably pie/donut) put fixed dimensions directly on the spec, while
  // most use config.view defaults. Both take precedence over responsive container sizing.
  // A pie's outer radius is fixed to those dimensions too; without one it takes half the view.
  delete spec.width;
  delete spec.height;
  for (const view of layersOf(spec)) {
    const mark = view.mark as Record<string, unknown> | undefined;
    if (mark?.type === "arc") delete mark.outerRadius;
  }
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
 * Whether a pie pins its arcs, and the labels it draws around them, to pixel positions in the plot
 * Flint sized, as one that labels its slices does. Refitting that plot would pull them apart.
 */
function pinsArcs(spec: Record<string, unknown>): boolean {
  return layersOf(spec).some((view) => {
    const mark = view.mark as Record<string, unknown> | undefined;
    return mark?.type === "arc" && typeof mark.x === "number";
  });
}

/**
 * Room beside the panels of a row-faceted chart other than the row labels, in pixels at Rill's
 * label sizes and paddings: on the left the row header title, the y-axis labels and their
 * paddings; on the right the last x-axis label, which may run past the panel's edge.
 */
const ROW_FACET_GUTTER = 130;
/** The widest a row's upright label may be before it is cut, in pixels. */
const ROW_FACET_LABEL_LIMIT = 110;
/** A generous width per character of an 11px label, in pixels, for sizing text Vega lays out later. */
const LABEL_CHAR_WIDTH = 6.5;

/**
 * Flint keeps a row-faceted chart's panels near the proportions of a single chart, so a handful
 * of rows in a wide box come out as a narrow column with most of the width unused. In Rill a
 * stack of rows reads like a list of sparklines: each panel spans the box, less the gutter and
 * its longest row label, and each label stands upright beside its panel rather than rotated and
 * cut to the row's height.
 * Panels with a discrete x axis are sized by their band step and are left to Flint.
 */
function fillRowFacetWidth(
  spec: Record<string, unknown>,
  rows: Record<string, unknown>[],
  size: { width: number; height: number } | undefined,
): void {
  if (!size) return;
  const encoding = (spec.encoding ?? {}) as Record<string, EncodingRecord>;
  const facet = (spec.facet ?? {}) as Record<string, EncodingRecord>;
  const row = encoding.row ?? facet.row;
  const hasColumns =
    "column" in encoding ||
    "column" in facet ||
    "facet" in encoding ||
    (spec.facet !== undefined && !("row" in facet) && !("column" in facet));
  if (!row || hasColumns) return;
  // A width in steps belongs to a discrete x axis.
  if (spec.width !== undefined && typeof spec.width !== "number") return;

  const rowField = row.field;
  const longestLabel =
    typeof rowField === "string"
      ? rows.reduce(
          (longest, datum) =>
            Math.max(
              longest,
              String((datum[rowField] as string | number | null) ?? "").length,
            ),
          0,
        )
      : 0;
  const labelWidth = Math.min(
    ROW_FACET_LABEL_LIMIT,
    Math.ceil(longestLabel * LABEL_CHAR_WIDTH),
  );
  const config = (spec.config as Record<string, unknown> | undefined) ?? {};
  spec.config = config;
  const view = (config.view as Record<string, unknown> | undefined) ?? {};
  config.view = view;
  view.continuousWidth = Math.max(
    120,
    Math.round(size.width - ROW_FACET_GUTTER - labelWidth),
  );
  config.headerRow = {
    ...(config.headerRow as Record<string, unknown> | undefined),
    labelAngle: 0,
    // Vega lines up upright left-side header labels against the axis when they are aligned
    // `left`. With `right` each label sat at its own offset, which widened the gutter past
    // its budget, so long labels pushed the panels out of the box.
    labelAlign: "left",
    labelLimit: ROW_FACET_LABEL_LIMIT,
  };
}

/** Room a tick label on a time axis needs along the axis, in pixels. */
const TIME_TICK_SPACING = { x: 70, y: 30 };

/** The axis properties Flint writes when it plans a time axis' ticks. */
const FLINT_TIME_TICK_KEYS = [
  "values",
  "labelExpr",
  "labelOverlap",
  "labelFlush",
  "tickCount",
];

const DAY_MS = 24 * 60 * 60 * 1000;

/** The shortest bucket at each of Flint's temporal types that names a grain, in milliseconds. */
const TIME_GRAIN_MS: Record<string, number> = {
  Year: 365 * DAY_MS,
  YearQuarter: 90 * DAY_MS,
  YearMonth: 28 * DAY_MS,
  YearWeek: 7 * DAY_MS,
  Date: DAY_MS,
};

/**
 * Vega picks a time axis' tick interval from a tick count, which Vega-Lite sets at one tick per
 * 40px. Flint pads a banded time axis by half a bucket at each end, so a day-grain axis starts and
 * ends at noon. Over a week or so the count called for 12-hour ticks, label overlap removal then
 * kept every other label, and all of the ones left read "12 PM". An hourly line over eight days
 * drew the same 12-hour ticks.
 *
 * The count is capped at the number of buckets the axis spans, so no interval is shorter than the
 * field's grain, and at one tick per `TIME_TICK_SPACING`, so every tick has room for its label.
 * The grain comes from the field's semantic type; below a day (`DateTime`) it is the shortest
 * gap between the data's timestamps. Ticks then fall on the grain's boundaries, where a banded
 * axis centers its buckets (except that Vega's weeks start on Sunday, a day before Rill's).
 *
 * Flint plans these ticks itself: it lists the tick dates, labels them with an expression that
 * looks each one up, and moves the scale to UTC when every timestamp is a UTC midnight. That plan
 * is dropped first. It is fitted to Flint's label font and plot width, not the ones the chart is
 * drawn with. Its tick dates are the data's, so an ejected spec would draw no ticks once its
 * dashboard moved to another time range. And a UTC axis would disagree with the labels around it:
 * discrete time axes, tooltips and the MCP widget's zone-shifted dates all read the browser's zone.
 */
function spaceTimeAxisTicks(
  spec: Record<string, unknown>,
  rows: Record<string, unknown>[],
  semanticTypes: Record<string, string>,
  size: { width: number; height: number } | undefined,
): void {
  for (const view of layersOf(spec)) {
    for (const [channel, dimension] of [
      ["x", "width"],
      ["y", "height"],
    ] as const) {
      const encoding = encodingsOf(view)[channel];
      if (!encoding || Array.isArray(encoding) || encoding.type !== "temporal")
        continue;
      const planned = encoding.axis;
      if (
        planned &&
        Array.isArray(planned.values) &&
        typeof planned.labelExpr === "string" &&
        planned.labelFlush === false
      ) {
        for (const key of FLINT_TIME_TICK_KEYS) delete planned[key];
        if (Object.keys(planned).length === 0) delete encoding.axis;
        const scale = encoding.scale as Record<string, unknown> | undefined;
        if (scale?.type === "utc") {
          delete scale.type;
          if (Object.keys(scale).length === 0) delete encoding.scale;
        }
      }
      if (
        !size ||
        typeof encoding.field !== "string" ||
        encoding.bin ||
        encoding.axis === null ||
        encoding.axis?.tickCount !== undefined ||
        encoding.axis?.values !== undefined
      ) {
        continue;
      }
      const field = encoding.field;
      const times = [...new Set(rows.map((row) => timeOf(row[field])))].filter(
        Number.isFinite,
      );
      if (times.length < 2) continue;
      times.sort((a, b) => a - b);

      const domain = (encoding.scale as { domain?: unknown } | undefined)
        ?.domain;
      const [start, end] =
        Array.isArray(domain) && domain.length === 2
          ? domain.map(timeOf)
          : [times[0], times[times.length - 1]];
      const grain =
        TIME_GRAIN_MS[semanticTypes[field]] ??
        times.reduce(
          (shortest, time, i) =>
            i === 0 ? shortest : Math.min(shortest, time - times[i - 1]),
          Infinity,
        );
      // A layer or a facet's panel takes its size from the view that holds it.
      const length =
        [view[dimension], spec[dimension]].find(
          (value): value is number => typeof value === "number",
        ) ?? size[dimension];
      const count = Math.floor(
        Math.min(length / TIME_TICK_SPACING[channel], (end - start) / grain),
      );
      if (!(count > 0)) continue;
      encoding.axis = { ...encoding.axis, tickCount: count };
    }
  }
}

/** A Date, an ISO string or epoch milliseconds as epoch milliseconds; NaN for anything else. */
function timeOf(value: unknown): number {
  if (value instanceof Date) return value.getTime();
  if (typeof value === "number") return value;
  if (typeof value === "string") return Date.parse(value);
  return NaN;
}

/** A donut's hole as a percentage of its outer radius, when the spec does not set one. */
const DEFAULT_DONUT_HOLE = 50;

/**
 * Flint takes a donut's inner radius in pixels, on a 0–100 scale suited to its own small default
 * canvas, so the hole all but vanishes in a chart filling a dashboard panel. Read that setting as
 * a percentage of the outer radius instead, so the donut keeps its proportions at any size.
 * A single donut follows the size of its view. The panels of a faceted one keep the radius Flint
 * gives them (and have no `width` or `height` signal of their own), as does a donut that labels
 * its slices (see `fitSingleViewToContainer`), so their hole is fixed too.
 */
function scaleDonutHole(spec: Record<string, unknown>): void {
  const encoding = (spec.encoding ?? {}) as Record<string, unknown>;
  const faceted = ["facet", "row", "column"].some((key) => key in encoding);
  for (const view of layersOf(spec)) {
    const mark = view.mark as Record<string, unknown> | undefined;
    if (mark?.type !== "arc" || typeof mark.innerRadius !== "number") continue;
    const share = mark.innerRadius / 100;
    mark.innerRadius =
      typeof mark.outerRadius === "number" &&
      (faceted || typeof mark.x === "number")
        ? mark.outerRadius * share
        : { expr: `min(width, height) / 2 * ${share}` };
  }
}

/** Flint's semantic types for a field that already holds ranks, which its bump template expects on y. */
const RANK_SEMANTIC_TYPES = new Set(["Rank", "Score", "Level"]);
/** The field a bump chart's ranks are derived into. */
const BUMP_RANK_FIELD = "__rill_rank";
/** Room each labelled rank needs on a bump chart's axis, in pixels. */
const BUMP_RANK_SPACING = 20;
/** A chart's height outside its plot: the legend, the x axis and their paddings, in pixels. */
const PLOT_CHROME_HEIGHT = 80;

/**
 * Flint's bump template reverses the y axis and expects y to hold ranks already. A Rill measure
 * holds values, so the chart drew them upside down, with $0 at the top. The measure is ranked at
 * each step on x instead (within each facet panel), 1 for the highest, and y draws the rank with
 * a label per rank, or every few ranks when they are crowded; the tooltip keeps the value.
 */
function rankBumpChart(
  spec: Record<string, unknown>,
  rows: Record<string, unknown>[],
  fields: FlintFields,
  measureFields: string[],
  size: { width: number; height: number } | undefined,
): void {
  const encoding = encodingsOf(spec);
  const { x, y } = encoding;
  if (
    !x ||
    !y ||
    Array.isArray(x) ||
    Array.isArray(y) ||
    typeof x.field !== "string" ||
    typeof y.field !== "string" ||
    y.type !== "quantitative" ||
    RANK_SEMANTIC_TYPES.has(fields.semantic_types[y.field])
  ) {
    return;
  }
  const measure = y.field;
  const panels = ["column", "row", "facet"]
    .map((channel) => encoding[channel])
    .filter((def) => def && typeof def.field === "string");
  const groupby = [x.field, ...panels.map((def) => def.field as string)];
  spec.transform = [
    ...((spec.transform as unknown[] | undefined) ?? []),
    { filter: `isValid(datum[${JSON.stringify(measure)}])` },
    {
      window: [{ op: "rank", as: BUMP_RANK_FIELD }],
      sort: [{ field: measure, order: "descending" }],
      groupby,
    },
  ];

  // The lowest rank drawn is the most series ranked at any one step.
  const seriesPerStep = new Map<string, number>();
  for (const row of rows) {
    if (row[measure] === null || row[measure] === undefined) continue;
    const step = JSON.stringify(groupby.map((field) => row[field]));
    seriesPerStep.set(step, (seriesPerStep.get(step) ?? 0) + 1);
  }
  const lowestRank = Math.max(1, ...seriesPerStep.values());
  const plotHeight = Math.max(
    1,
    (typeof spec.height === "number" ? spec.height : (size?.height ?? 300)) -
      PLOT_CHROME_HEIGHT,
  );
  const every = Math.ceil((lowestRank * BUMP_RANK_SPACING) / plotHeight);
  const label = fields.field_display_names[measure] ?? measure;
  encoding.y = {
    field: BUMP_RANK_FIELD,
    type: "quantitative",
    title: `Rank by ${label}`,
    scale: { reverse: true, zero: false, nice: false, padding: 8 },
    axis: {
      format: "d",
      values: Array.from(
        { length: Math.ceil(lowestRank / every) },
        (_, i) => 1 + i * every,
      ),
    },
  };
  (spec.encoding as Record<string, unknown>).tooltip = [
    ...[x, encoding.color, encoding.detail, ...panels]
      .filter((def) => def && typeof def.field === "string")
      .map(({ field, type, title, timeUnit }) => ({
        field,
        type,
        title,
        ...(timeUnit ? { timeUnit } : {}),
      })),
    {
      field: measure,
      type: "quantitative",
      title: label,
      ...(measureFields.includes(measure)
        ? { formatType: sanitizeFieldName(measure) }
        : {}),
    },
    { field: BUMP_RANK_FIELD, type: "quantitative", title: "Rank" },
  ];
}

/** The column Flint aggregates a bar table's values into, per row. */
const BAR_TABLE_VALUE_FIELD = "__bt_val";
/** Vega's padding on each side of a chart, in pixels. */
const VEGA_PADDING = 5;
/** The characters a measure's formatted value takes in a bar table, short of a custom format. */
const BAR_TABLE_VALUE_CHARS = 8;
/** The widest a bar table's value column grows to fit its header, in pixels: Flint's own cap. */
const BAR_TABLE_MAX_COLUMN_WIDTH = 140;
/** Room a bar table's text column leaves around its widest text, in pixels, as Flint pads it. */
const BAR_TABLE_TEXT_PADDING = 12;

/**
 * Flint's bar table is a row of panels: the bars beside the category labels, an optional share
 * column, then the value column. Its headers are the raw field names and its values are in
 * Flint's own number format, so the headers take the fields' display names and the values go
 * through the measure's Rill formatter, with the value column sized for both.
 *
 * Flint sizes the panels to its canvas but leaves out Vega's padding and the headers above the
 * panels, so the table ran past the right and bottom of its box, cutting off the values and the
 * last row. The bars give up the room the padding needs, and the panels' height is fitted to the
 * box. (Vega-Lite fits one dimension of a concatenation, not both; the width is the one sized
 * here, so the canvas runs on to the right with nothing drawn there, and the renderer hides it.)
 */
function formatBarTable(
  spec: Record<string, unknown>,
  authored: Record<string, RawEncodingValue>,
  fields: FlintFields,
  measureFields: string[],
  size: { width: number; height: number } | undefined,
): void {
  const panels = spec.hconcat as Record<string, unknown>[] | undefined;
  if (!Array.isArray(panels) || panels.length < 2) return;
  const bars = panels[0];
  const values = panels[panels.length - 1];
  const category = fieldOf(authored.y);
  const measure = fieldOf(authored.x);

  const barsTitle = bars.title as Record<string, unknown> | undefined;
  if (barsTitle && category) {
    barsTitle.text = fields.field_display_names[category] ?? category;
  }

  const valueEncoding = encodingsOf(values);
  if (measure && measureFields.includes(measure) && valueEncoding.text) {
    valueEncoding.text = {
      field: BAR_TABLE_VALUE_FIELD,
      type: "quantitative",
      formatType: sanitizeFieldName(measure),
    };
    // Flint writes some formats into a string column, which the formatter has no use for.
    values.transform = (
      values.transform as Record<string, unknown>[] | undefined
    )?.filter((transform) => transform.as !== `${BAR_TABLE_VALUE_FIELD}_str`);
  }

  // Flint's estimate of a character's width at the column's font size.
  const charWidth =
    (Number((values.mark as Record<string, unknown> | undefined)?.fontSize) ||
      12) * 0.6;
  const valuesTitle = values.title as Record<string, unknown> | undefined;
  let headerChars = 0;
  if (valuesTitle && measure) {
    const lines = wrapHeader(
      fields.field_display_names[measure] ?? measure,
      Math.floor(
        (BAR_TABLE_MAX_COLUMN_WIDTH - BAR_TABLE_TEXT_PADDING) / charWidth,
      ),
    );
    valuesTitle.text = lines.length === 1 ? lines[0] : lines;
    headerChars = Math.max(...lines.map((line) => line.length));
  }
  values.width = Math.min(
    BAR_TABLE_MAX_COLUMN_WIDTH,
    Math.round(
      Math.max(BAR_TABLE_VALUE_CHARS, headerChars) * charWidth +
        BAR_TABLE_TEXT_PADDING,
    ),
  );
  if (valuesTitle) valuesTitle.limit = values.width;

  // A faceted table's panels are sized per facet; only a single table is fitted to the box.
  const encoding = (spec.encoding ?? {}) as Record<string, unknown>;
  if (!size || "column" in encoding || "row" in encoding) return;
  const barsAxis = encodingsOf(bars).y?.axis;
  const gutter = Number(barsAxis?.labelPadding) || 0;
  const spacing = typeof spec.spacing === "number" ? spec.spacing : 0;
  const used =
    gutter +
    panels.reduce((sum, panel) => sum + (Number(panel.width) || 0), 0) +
    spacing * (panels.length - 1);
  bars.width = Math.max(
    80,
    Math.round(
      (Number(bars.width) || 0) + size.width - 2 * VEGA_PADDING - used,
    ),
  );
  spec.autosize = { type: "fit-y", contains: "padding" };
}

/** A header as one line, or two split at the space nearest its middle when it runs past `maxChars`. */
function wrapHeader(header: string, maxChars: number): string[] {
  if (header.length <= maxChars) return [header];
  const middle = header.length / 2;
  let split = -1;
  for (let i = header.indexOf(" "); i !== -1; i = header.indexOf(" ", i + 1)) {
    if (split === -1 || Math.abs(i - middle) < Math.abs(split - middle)) {
      split = i;
    }
  }
  return split === -1
    ? [header]
    : [header.slice(0, split), header.slice(split + 1)];
}

/** The field of a single-field Flint encoding. */
function fieldOf(value: RawEncodingValue | undefined): string | undefined {
  if (typeof value === "string") return value;
  if (value && !Array.isArray(value)) return value.field;
  return undefined;
}

/** The field each violin's density is scaled by: its own peak. */
const VIOLIN_PEAK_FIELD = "__rill_violin_peak";
/** Room left of a violin plot's panels for the value axis, its labels and its title, in pixels. */
const VIOLIN_AXIS_WIDTH = 80;

/**
 * Flint draws a violin per category, as a horizontal area stacked about its center, in a row of
 * facet panels. Three things went wrong with it:
 * - Rill's theme outlines every area with a line, which Vega-Lite draws in x order. Along a
 *   horizontal area it zig-zagged across each violin and filled half of it with the line color.
 * - Each panel's density axis was rounded out past the widest violin, so the violins sat left of
 *   their panels' centers. A category whose values pile up on one number drew a spike so wide
 *   that the other violins thinned to hairlines. And Flint sizes its panels with no gap between
 *   them and less room for the value axis than Rill's labels take, while its config adds a gap,
 *   so a row of panels ran out of the box; long category labels widened the panels further.
 * - The density was estimated past the data at both ends, so a measure that cannot go negative
 *   drew violins, and a y axis, below 0.
 *
 * So the outline goes; each violin is scaled to its own widest point and centered in its panel;
 * the panels fit the box with no gap, their category labels cut to the panel's width; and the
 * density is estimated over the data's range only.
 */
function shapeViolins(
  spec: Record<string, unknown>,
  rows: Record<string, unknown>[],
  size: { width: number; height: number } | undefined,
): void {
  const transforms = spec.transform as Record<string, unknown>[] | undefined;
  const density = transforms?.[0];
  const encoding = encodingsOf(spec);
  if (
    !transforms ||
    typeof density?.density !== "string" ||
    !Array.isArray(density.as) ||
    !encoding.x
  ) {
    return;
  }
  const measure = density.density;
  const [, densityField] = density.as as string[];

  spec.mark = { ...(spec.mark as Record<string, unknown>), line: false };

  transforms.push(
    {
      joinaggregate: [
        { op: "max", field: densityField, as: VIOLIN_PEAK_FIELD },
      ],
      groupby: density.groupby,
    },
    {
      calculate: `datum.${VIOLIN_PEAK_FIELD} > 0 ? datum[${JSON.stringify(densityField)}] / datum.${VIOLIN_PEAK_FIELD} : 0`,
      as: densityField,
    },
  );
  encoding.x.scale = {
    ...(encoding.x.scale as Record<string, unknown> | undefined),
    nice: false,
    padding: 4,
  };

  const measured = rows
    .map((row) => row[measure])
    .filter((value): value is number => Number.isFinite(value));
  const min = measured.reduce((a, b) => Math.min(a, b), Infinity);
  const max = measured.reduce((a, b) => Math.max(a, b), -Infinity);
  if (min < max) density.extent = [min, max];

  spec.spacing = 0;
  const panels = encoding.facet ?? encoding.column;
  if (!panels || typeof spec.width !== "number") return;
  const columns =
    Number(panels.columns) ||
    new Set(rows.map((row) => row[panels.field as string])).size;
  if (size && columns > 0) {
    spec.width = Math.min(
      spec.width,
      Math.floor((size.width - 2 * VEGA_PADDING - VIOLIN_AXIS_WIDTH) / columns),
    );
  }
  panels.header = { ...(panels.header as object), labelLimit: spec.width };
}

/** Pixels between a rose wedge's tip and its label. */
const ROSE_LABEL_OFFSET = 6;
/** The widest a rose label may be before it is cut, in pixels. */
const ROSE_LABEL_LIMIT = 100;
/** Wedges shorter than this share of the longest wedge's radius are too small to hold a label. */
const ROSE_LABELLED_RADIUS = 0.3;

/**
 * Flint labels each wedge of a (single) rose chart with its category, centered 15px past the
 * wedge's tip, and leaves the wedges the whole radius. Labels at the sides ran off the canvas or
 * back over their wedges, and the labels of short wedges piled up in the middle.
 *
 * Each label is set just past its wedge's tip and aligned away from the center, the radius leaves
 * room for the longest label, and wedges too short to hold a label go without one.
 */
function placeRoseLabels(
  spec: Record<string, unknown>,
  rows: Record<string, unknown>[],
): void {
  const layers = spec.layer as Record<string, unknown>[] | undefined;
  const category = encodingsOf(spec).theta?.field;
  const labels = layers?.[1];
  const radius = labels ? encodingsOf(labels).radius : undefined;
  if (
    !layers ||
    !labels ||
    typeof category !== "string" ||
    typeof radius?.field !== "string"
  ) {
    return;
  }
  const valueField = radius.field;

  const totals = new Map<string, number>();
  for (const row of rows) {
    const key = String((row[category] as string | number | null) ?? "");
    totals.set(key, (totals.get(key) ?? 0) + (Number(row[valueField]) || 0));
  }
  const longestLabel = Math.max(
    0,
    ...[...totals.keys()].map((key) => key.length),
  );
  const labelWidth = Math.min(
    ROSE_LABEL_LIMIT,
    Math.ceil(longestLabel * LABEL_CHAR_WIDTH),
  );
  const range = [
    0,
    {
      expr: `max(0, min(width / 2 - ${ROSE_LABEL_OFFSET + labelWidth + 4}, height / 2 - ${ROSE_LABEL_OFFSET + 18}))`,
    },
  ];
  for (const layer of layers) {
    const layerRadius = encodingsOf(layer).radius;
    if (!layerRadius) continue;
    layerRadius.scale = {
      ...(layerRadius.scale as Record<string, unknown> | undefined),
      range,
    };
  }

  // The label's angle is the middle of its wedge, clockwise from 12 o'clock.
  const angle = `(scale('theta', datum[${JSON.stringify(category)}]) + bandwidth('theta') / 2)`;
  labels.mark = {
    ...(labels.mark as Record<string, unknown>),
    radiusOffset: ROSE_LABEL_OFFSET,
    limit: ROSE_LABEL_LIMIT,
    align: {
      expr: `sin(${angle}) > 0.3 ? 'left' : sin(${angle}) < -0.3 ? 'right' : 'center'`,
    },
    baseline: {
      expr: `cos(${angle}) > 0.3 ? 'bottom' : cos(${angle}) < -0.3 ? 'top' : 'middle'`,
    },
  };
  const longest = Math.max(0, ...totals.values());
  if (longest > 0) {
    // Wedges are drawn on a square-root scale, so a wedge's radius share is the root of its value's.
    labels.transform = [
      ...((labels.transform as unknown[] | undefined) ?? []),
      {
        filter: `datum[${JSON.stringify(valueField)}] >= ${longest * ROSE_LABELLED_RADIUS ** 2}`,
      },
    ];
  }
}

/**
 * Room left of a Marginal Plot's plot other than its y axis's labels, in pixels at Rill's label
 * sizes and paddings: the axis title's line and its padding, and the labels' padding.
 */
const MARGINAL_Y_AXIS_GUTTER = 32;
/**
 * Room below a Marginal Plot's plot for a continuous x axis, in pixels at Rill's label sizes and
 * paddings: a line of labels and the axis title, each with its padding.
 */
const MARGINAL_X_AXIS_HEIGHT = 45;
/** Flint's smallest Marginal Plot, in pixels. */
const MARGINAL_MIN_PLOT = 80;
/** The channels a Marginal Plot's plot draws a legend for, each at its right. */
const MARGINAL_LEGEND_CHANNELS = ["color", "size", "shape", "opacity"];
/** The gap between a Marginal Plot's right margin and its legends, and between two legends. */
const MARGINAL_LEGEND_GAP = 8;
/** Room after a legend symbol's fill for its stroke and the gap before its label, in pixels. */
const LEGEND_SYMBOL_GAP = 6;
/** Room before each legend label for a symbol of Vega's default size, 10px across, and its gap. */
const LEGEND_SYMBOL_WIDTH = 10 + LEGEND_SYMBOL_GAP;
/** Vega-Lite's largest point on a size scale, in square pixels, when the scale names none. */
const DEFAULT_MAX_POINT_SIZE = 361;
/** The font sizes of an axis label and of a legend title in Rill's theme, in pixels. */
const RILL_LABEL_FONT_SIZE = 11;
const RILL_LEGEND_TITLE_FONT_SIZE = 12;
/** A legend's title line and Vega's padding below it, in pixels. */
const LEGEND_TITLE_HEIGHT = RILL_LEGEND_TITLE_FONT_SIZE + 5;
/**
 * A legend entry's row as Vega lays it out, in pixels: a symbol of its default size, 10px across
 * with its stroke, beside a label at Rill's size, with some room for the rounding of its layout.
 */
const LEGEND_ROW_HEIGHT = 14;
/** Vega's gap between a legend's rows, in pixels. */
const LEGEND_ROW_PADDING = 2;
/** The entries Vega lists for a measure on a symbol legend: the ticks of a count of five. */
const LEGEND_TICK_COUNT = 5;

/**
 * A Marginal Plot is a column of the margin over x above a row of the plot and the margin over y,
 * with its legends at the right. Flint sizes the three panels to its canvas less a fixed
 * allowance for the axes and the legends, and the chart is drawn at those sizes: Vega-Lite cannot
 * fit a concatenation to its box. Rill's theme draws the axes larger than that allowance, and
 * Flint's legends, side by side or below the plot, were laid out at its own text sizes, so the
 * guides ran past the box and were cut off.
 *
 * The plot takes the width its y axis, its legends and Vega's padding leave, and with a continuous
 * x axis the height that axis leaves; the margins keep Flint's thickness and run along the plot.
 * The y axis's labels get the width they take once formatted. The legends, whether of
 * categories, sizes, shapes, opacities or a measure's colors, stack in one column at the right,
 * with their labels and titles cut to a third of the box at most, less where the y axis and the
 * plot need it. Where the column would outgrow the box's height, its tallest legends end early,
 * in Vega's "…N entries".
 */
function fitMarginalPlot(
  spec: Record<string, unknown>,
  rows: Record<string, unknown>[],
  fields: FlintFields,
  size: { width: number; height: number } | undefined,
  formatters: Record<string, (value: number) => string>,
): void {
  const [top, row] =
    (spec.vconcat as Record<string, unknown>[] | undefined) ?? [];
  const [plot, right] =
    (row?.hconcat as Record<string, unknown>[] | undefined) ?? [];
  if (!size || !top || !plot || !right) return;
  const encodings = encodingsOf(plot);
  const { x, y } = encodings;
  if (!x || !y) return;

  // Text at a given size, as wide as LABEL_CHAR_WIDTH allows at Rill's label size.
  const textWidth = (chars: number, fontSize: number) =>
    Math.ceil((chars * LABEL_CHAR_WIDTH * fontSize) / RILL_LABEL_FONT_SIZE);
  // The finite values a field takes in the rows, with any more its axis or legend states.
  const numbersOf = (field: string, stated: unknown) =>
    [
      ...rows.map((datum) => datum[field]),
      ...(Array.isArray(stated) ? (stated as unknown[]) : []),
    ]
      .map(Number)
      .filter(Number.isFinite);
  // The widest label a measure's axis or legend shows over those values: the longest of the
  // ticks Vega may pick, as Rill's formatter prints them, or else its largest value in full.
  const numberLabelWidth = (field: string, values: number[]) => {
    const format = formatters[field];
    if (!format) {
      // A share is labelled as a percentage.
      const scale = fields.semantic_types[field] === "Percentage" ? 100 : 1;
      return textWidth(
        numberLabelChars(values.map((value) => value * scale)),
        RILL_LABEL_FONT_SIZE,
      );
    }
    const [low, high] = extent(values);
    if (low === undefined || high === undefined) return 0;
    const candidates = [low, high];
    for (let count = 2; count <= 10; count++) {
      candidates.push(...ticks(low, high, count));
    }
    return textWidth(
      candidates.reduce(
        (longest, value) => Math.max(longest, format(value).length),
        0,
      ),
      RILL_LABEL_FONT_SIZE,
    );
  };

  let yLabelLimit = UPRIGHT_LABEL_LIMIT;
  if (y.type === "quantitative" && typeof y.field === "string") {
    const domain = (y.scale as Record<string, unknown> | undefined)?.domain;
    yLabelLimit = Math.min(
      UPRIGHT_LABEL_LIMIT,
      numberLabelWidth(y.field, numbersOf(y.field, domain)),
    );
  }
  if (y.axis !== null) y.axis = { ...y.axis, labelLimit: yLabelLimit };

  const columnGap = Number(row.spacing) || 0;
  const besidePlot =
    2 * VEGA_PADDING +
    MARGINAL_Y_AXIS_GUTTER +
    yLabelLimit +
    (Number(right.width) || 0) +
    columnGap;
  // The legends' column takes a third of the box at most, and no more than the plot leaves at its
  // smallest. Vega takes a limit of 0 or less as none.
  const room = Math.max(
    2 * LEGEND_SYMBOL_WIDTH,
    Math.min(
      Math.floor(size.width / 3),
      size.width - besidePlot - MARGINAL_MIN_PLOT,
    ) - MARGINAL_LEGEND_GAP,
  );
  const legends: {
    channel: string;
    field: string;
    entries: number;
    keyWidth: number;
    labelWidth: number;
    titleLimit: number;
    titleHeight: number;
    rowHeight: number;
    height: number;
  }[] = [];
  for (const channel of MARGINAL_LEGEND_CHANNELS) {
    const encoding = encodings[channel];
    if (!encoding || encoding.legend === null) continue;
    if (typeof encoding.field !== "string") continue;
    const field = encoding.field;
    const legend = encoding.legend ?? {};

    let keyWidth = LEGEND_SYMBOL_WIDTH;
    let rowHeight = LEGEND_ROW_HEIGHT;
    let labelWidth: number;
    let entries: number;
    if (isDiscrete(encoding)) {
      const labels = new Set(rows.map((datum) => String(datum[field])));
      labelWidth = textWidth(
        [...labels].reduce(
          (longest, label) => Math.max(longest, label.length),
          0,
        ),
        RILL_LABEL_FONT_SIZE,
      );
      entries = labels.size;
    } else {
      const values = numbersOf(field, legend.values);
      labelWidth = numberLabelWidth(field, values);
      const [low = 0, high = 0] = extent(values);
      // A size scale starts at zero.
      entries = Array.isArray(legend.values)
        ? legend.values.length
        : ticks(
            channel === "size" ? Math.min(0, low) : low,
            high,
            LEGEND_TICK_COUNT,
          ).length;
      if (channel === "size") {
        const range = (encoding.scale as Record<string, unknown> | undefined)
          ?.range;
        const largest = Array.isArray(range)
          ? Math.max(...range.map(Number).filter(Number.isFinite))
          : DEFAULT_MAX_POINT_SIZE;
        // A symbol's size is its area.
        const across = Math.ceil(Math.sqrt(largest));
        keyWidth = Math.max(LEGEND_SYMBOL_WIDTH, across + LEGEND_SYMBOL_GAP);
        rowHeight = Math.max(LEGEND_ROW_HEIGHT, across + 2);
      }
    }
    // A title of `null` is none.
    const stated = legend.title !== undefined ? legend.title : encoding.title;
    const title =
      stated !== undefined
        ? stated
        : (fields.field_display_names[field] ?? field);
    const titleChars =
      title === null
        ? 0
        : (Array.isArray(title)
            ? (title as unknown[])
            : [title]
          ).reduce<number>(
            (longest, line) => Math.max(longest, String(line).length),
            0,
          );
    const titleHeight = title === null ? 0 : LEGEND_TITLE_HEIGHT;
    legends.push({
      channel,
      field,
      entries,
      keyWidth,
      labelWidth,
      titleLimit: Math.min(
        room,
        textWidth(titleChars, RILL_LEGEND_TITLE_FONT_SIZE),
      ),
      titleHeight,
      rowHeight,
      height:
        titleHeight +
        entries * rowHeight +
        Math.max(0, entries - 1) * LEGEND_ROW_PADDING,
    });
  }

  // Where the column would outgrow the box's height, the tallest legends give way first, each to
  // an equal share of the height the shorter ones leave.
  let share = Infinity;
  let left =
    size.height -
    2 * VEGA_PADDING -
    MARGINAL_LEGEND_GAP * Math.max(0, legends.length - 1);
  const heights = legends.map((legend) => legend.height).sort((a, b) => a - b);
  for (const [index, height] of heights.entries()) {
    const shorter = heights.length - index;
    if (height * shorter > left) {
      share = left / shorter;
      break;
    }
    left -= height;
  }

  let legendWidth = 0;
  for (const legend of legends) {
    // A measure's colors are listed as symbols too, as Vega draws them already: the points take
    // a color as both fill and stroke, and Vega keeps gradients for legends of one.
    const layout: Record<string, unknown> = {
      type: "symbol",
      orient: "right",
      offset: MARGINAL_LEGEND_GAP,
      titleLimit: legend.titleLimit,
    };
    let labelWidth = legend.labelWidth;
    if (legend.height > share) {
      // Vega lists one entry fewer than its limit, and then how many it left out.
      const limit = Math.max(
        2,
        Math.floor(
          (share - legend.titleHeight + LEGEND_ROW_PADDING) /
            (legend.rowHeight + LEGEND_ROW_PADDING),
        ),
      );
      layout.symbolLimit = limit;
      if (limit < legend.entries) {
        const leftOut = `\u2026${legend.entries - limit + 1} entries`;
        labelWidth = Math.max(
          labelWidth,
          textWidth(leftOut.length, RILL_LABEL_FONT_SIZE),
        );
      }
    }
    const labelLimit = Math.min(room - legend.keyWidth, labelWidth);
    layout.labelLimit = labelLimit;
    legendWidth = Math.max(
      legendWidth,
      MARGINAL_LEGEND_GAP +
        Math.max(legend.keyWidth + labelLimit, legend.titleLimit),
    );
    // Flint states a legend on each panel, and Vega-Lite draws the one they share with the
    // settings of whichever states them first.
    for (const view of layersOf(spec)) {
      const shared = encodingsOf(view)[legend.channel];
      if (shared?.field !== legend.field || shared.legend === null) continue;
      const own = { ...shared.legend };
      dropFlintLegendLayout(own);
      shared.legend = { ...own, ...layout };
    }
  }

  plot.width = Math.max(
    MARGINAL_MIN_PLOT,
    Math.round(size.width - besidePlot - legendWidth),
  );
  top.width = plot.width;

  if (!isDiscrete(x)) {
    const rowGap = Number(spec.spacing) || 0;
    const plotHeight = Math.round(
      size.height -
        2 * VEGA_PADDING -
        MARGINAL_X_AXIS_HEIGHT -
        (Number(top.height) || 0) -
        rowGap,
    );
    plot.height = Math.max(MARGINAL_MIN_PLOT, plotHeight);
    right.height = plot.height;
  }

  // Flint arranges several legends side by side, where they would outgrow the width the plot
  // leaves them.
  const config = (spec.config ?? {}) as Record<string, unknown>;
  config.legend = {
    ...(config.legend as Record<string, unknown> | undefined),
    layout: {
      right: {
        direction: "vertical",
        anchor: "start",
        margin: MARGINAL_LEGEND_GAP,
      },
    },
  };
  spec.config = config;

  // Rill's theme outlines an area with a line, and where a density margin peaks at its panel's
  // outer edge, half of that line's stroke rose past it: above the box, at the top.
  for (const margin of [top, right]) {
    for (const view of layersOf(margin)) {
      if (view.mark === undefined) continue;
      view.mark =
        typeof view.mark === "string"
          ? { type: view.mark, clip: true }
          : { ...(view.mark as Record<string, unknown>), clip: true };
    }
  }

  // The width is fitted to the box, Vega's padding included, as a single chart's is.
  spec.autosize = { type: "fit-x", contains: "padding" };
}

/**
 * The most characters a measure's tick or legend labels take, short of its formatter: the largest
 * value's digits and thousands separators, as a full format such as `$,.0f` prints them, a minus
 * sign, a currency or percent sign or a unit, and a decimal point and two decimals for values
 * under ten. A humanized label such as `$1.2M` takes fewer.
 */
function numberLabelChars(values: number[]): number {
  let largest = 0;
  let negative = false;
  for (const value of values) {
    if (!Number.isFinite(value)) continue;
    largest = Math.max(largest, Math.abs(value));
    negative ||= value < 0;
  }
  const digits = largest < 1 ? 1 : Math.floor(Math.log10(largest)) + 1;
  return (
    digits +
    Math.floor((digits - 1) / 3) +
    (negative ? 1 : 0) +
    1 +
    (largest < 10 ? 3 : 0)
  );
}

/**
 * Removes Flint's non-standard bookkeeping keys (`_warnings`, `_width`, `_height`, `_options`,
 * `_transform`, `_pivot`, `_legendLayout`, `_interactionSemantics`). Vega-Lite rejects unknown
 * top-level properties, so these must not reach the renderer. Matching on the underscore prefix
 * rather than a fixed list keeps this working as Flint adds more of them.
 *
 * Flint also annotates the views nested in a chart for its own interaction runtime and theme
 * stages (`__flintInteractionProvenance`, `__themeSynthetic`). Those go by their prefixes only:
 * a chart's named datasets may start with an underscore too (a bar table's `__bt_displayTable`).
 */
function stripInternalKeys(
  assembled: Record<string, unknown>,
): Record<string, unknown> {
  const spec = Object.fromEntries(
    Object.entries(assembled).filter(([key]) => !key.startsWith("_")),
  );
  for (const view of layersOf(spec)) {
    for (const key of Object.keys(view)) {
      if (key.startsWith("__flint") || key.startsWith("__theme")) {
        delete view[key];
      }
    }
  }
  return spec;
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
 * The legend properties Flint's layout writes on an encoding. Flint lays every legend out in a
 * column at the chart's right, wrapped and sized to fit there. Rill's theme puts legends in a row
 * above the chart instead, where a single column would take the plot's height.
 */
const FLINT_LEGEND_LAYOUT_KEYS = [
  "orient",
  "direction",
  "offset",
  "columns",
  "rowPadding",
  "gridAlign",
  "clipHeight",
  "labelBaseline",
  "labelFontSize",
  "labelLimit",
  "symbolLimit",
  "titleOrient",
  "titlePadding",
  "titleLimit",
  "titleLineHeight",
  "gradientLength",
  // Positions for wrapped labels, and hover text for cut ones, which Vega only shows once the
  // labels are made interactive in the compiled spec.
  "encoding",
];

/** The size Flint sets axis labels at before it shrinks them to fit, in pixels. */
const FLINT_LABEL_FONT_SIZE = 10;
/** The longest an upright category label runs before it is cut, in pixels. */
const UPRIGHT_LABEL_LIMIT = 100;

/** The ink Flint writes value labels in when it has no theme of its own. */
const FLINT_INK = "#121212";

/** The label expression Flint writes to break a legend's labels over lines. */
const FLINT_WRAPPED_LEGEND_LABELS =
  /^indexof\(\[.*\], toString\(datum\.label\)\) < 0 \? datum\.label : /s;

/**
 * Flint emits its own typography (6–10px labels, grey titles), a `tableau10` scheme on every
 * nominal color encoding, and a layout for each legend. A spec's config, scales and legends take
 * precedence over the config the renderer passes, so left in place they would override Rill's
 * theme: dark mode would draw grey-on-grey titles, a project palette would never reach a Flint
 * chart, and legends would stack in a column above the plot. Those keys are removed so the theme
 * applies; a `scheme` the author chose is kept, as are a legend's title, values and format.
 * The order Flint pins a series' colors in goes too, since it drops null values (see below).
 */
function deferToTheme(
  spec: Record<string, unknown>,
  authored: Record<string, RawEncodingValue>,
): void {
  // A chart of concatenated views and a pie that labels its slices are drawn at the size Flint
  // gives them, which leaves room for the legend where Flint lays it out and none above the
  // plot. Their legends keep that layout, at the chart's right.
  const keepsLegendLayout =
    pinsArcs(spec) ||
    ["concat", "hconcat", "vconcat"].some((key) => key in spec);

  const config = spec.config as Record<string, unknown> | undefined;
  if (config) {
    // How Flint arranges several legends on one edge of the chart.
    if (!keepsLegendLayout) {
      delete (config.legend as Record<string, unknown> | undefined)?.layout;
    }
    // Flint sets the names along a crowded band axis smaller, down to 6px, until they lie flat
    // within their bands, and cuts them to the band's width. The theme draws them at its own
    // size, where that width keeps a few letters of each name. They stand upright instead, as
    // they do by default in Vega-Lite.
    const axisX = config.axisX as Record<string, unknown> | undefined;
    if (
      axisX?.labelAngle === 0 &&
      Number(axisX.labelFontSize) < FLINT_LABEL_FONT_SIZE &&
      layersOf(spec).some((view) => {
        const x = encodingsOf(view).x;
        return x && !Array.isArray(x) && isDiscrete(x);
      })
    ) {
      delete axisX.labelExpr;
      delete axisX.labelLineHeight;
      axisX.labelAngle = -90;
      axisX.labelAlign = "right";
      axisX.labelBaseline = "middle";
      axisX.labelLimit = UPRIGHT_LABEL_LIMIT;
    }
    for (const group of STYLED_CONFIG_GROUPS) {
      const entry = config[group] as Record<string, unknown> | undefined;
      if (!entry) continue;
      for (const key of THEME_OWNED_CONFIG_KEYS) delete entry[key];
      if (Object.keys(entry).length === 0) delete config[group];
    }
  }

  for (const view of layersOf(spec)) {
    // Value labels that stand off their marks, above a bar or around a pie, are drawn on the
    // chart's background in Flint's dark ink, which a dark theme hides. They take the theme's
    // text color; labels inside a mark keep the ink Flint chose against the mark.
    const mark = view.mark as Record<string, unknown> | undefined;
    if (
      mark?.type === "text" &&
      mark.color === FLINT_INK &&
      (mark.dy !== undefined || mark.radius !== undefined)
    ) {
      delete mark.color;
      // A pie's labels state the ink as an encoding too, over the slice colors the layer would
      // otherwise inherit. `null` keeps those out without stating a color.
      const encoding = view.encoding as Record<string, unknown> | undefined;
      const color = encoding?.color as Record<string, unknown> | undefined;
      if (encoding && color?.value === FLINT_INK) encoding.color = null;
    }

    for (const [, encoding] of channelsOf(view)) {
      if (Array.isArray(encoding) || !encoding.legend) continue;
      const legend = encoding.legend;
      if (keepsLegendLayout) {
        legend.orient ??= "right";
        continue;
      }
      dropFlintLegendLayout(legend);
      if (Object.keys(legend).length === 0) delete encoding.legend;
    }
  }

  const hasAuthoredScheme = (channel: string) => {
    // Flint draws a `group` encoding on the color channel when the spec binds no color.
    const value =
      channel === "color" && !authored.color
        ? authored.group
        : authored[channel];
    return (
      !!value &&
      typeof value === "object" &&
      !Array.isArray(value) &&
      value.scheme !== undefined
    );
  };
  for (const view of layersOf(spec)) {
    for (const [channel, encoding] of channelsOf(view)) {
      if (Array.isArray(encoding) || !COLOR_CHANNELS.has(channel)) continue;
      const scale = encoding.scale as Record<string, unknown> | undefined;
      if (!scale) continue;
      // Flint lists a series' values as the color scale's domain to fix their order, and leaves
      // out null. A domain is all the scale then colors, so a series of null values, which a
      // dimension often has, was drawn without a color or a legend entry. Vega-Lite's own
      // domain takes every value, in the order the encoding's `sort` gives. A domain with a
      // range (colors the author or the template assigned) is kept.
      if (
        isDiscrete(encoding) &&
        Array.isArray(scale.domain) &&
        scale.range === undefined
      ) {
        delete scale.domain;
      }
      if (!hasAuthoredScheme(channel)) delete scale.scheme;
      if (Object.keys(scale).length === 0) delete encoding.scale;
    }
  }
}

/** Removes the layout Flint gives a legend, its labels broken over lines included. */
function dropFlintLegendLayout(legend: Record<string, unknown>): void {
  for (const key of FLINT_LEGEND_LAYOUT_KEYS) delete legend[key];
  if (Array.isArray(legend.title)) legend.title = legend.title.join(" ");
  if (
    typeof legend.labelExpr === "string" &&
    FLINT_WRAPPED_LEGEND_LABELS.test(legend.labelExpr)
  ) {
    delete legend.labelExpr;
  }
}

/** The channels a series is drawn on apart from its color: the bars of a group, a line's dash, a point's shape. */
const SERIES_CHANNELS = [
  "color",
  "fill",
  "stroke",
  "xOffset",
  "yOffset",
  "strokeDash",
  "shape",
];

/** The columns a series' total, and its place among the series by it, are derived into. */
const SERIES_TOTAL_FIELD = "__rill_series_total";
const SERIES_PLACE_FIELD = "__rill_series_place";

/**
 * Flint lists a chart's series in the order their rows arrive (`sort: null`), in the legend and so
 * in the colors they take. The query of a chart over time is ordered by time, which leaves the
 * order of each bucket's rows to the database: it changed from one load to the next, and every
 * series changed color with it.
 *
 * Where the author gave the series no order (`sortBy`, `sortOrder`), they are ordered by their
 * total of the measure the chart draws, largest first, as a built-in chart orders the values it
 * splits by. A chart with a measure on both axes, or on neither, keeps the rows' order, and so
 * does a pie, whose slices are drawn in that order. So do the series of Flint's own columns.
 *
 * Series of one total are ordered by name. Vega-Lite sorts a scale by a single field, and Vega
 * lists the values that tie on it as their rows arrive, so sorting by the total alone left such
 * series, those of no value among them, to swap colors as before. The view derives each
 * series' place in the order as a column instead, and the scale is sorted by that.
 *
 * A mark that summarizes its rows, such as a box plot, keeps no such column: its series are
 * listed by name alone. So are the series the author typed as ordinal: their values, such as
 * years or levels, have an order of their own, which their totals would scramble. So are the
 * series of a view a facet draws in each of its panels, whatever its axes draw: Vega-Lite derives
 * the column within each panel, where a Sparkline's series, one to a panel, all came first, and it
 * threw on the window that derives it, so the chart was not drawn.
 *
 * A layer draws the measure of the view it is layered in where it does not draw one of its own,
 * as a bullet chart's bars do.
 */
function orderSeries(
  view: Record<string, unknown>,
  inherited: Record<string, EncodingRecord> = {},
  faceted = false,
): void {
  const own = encodingsOf(view);
  const encodings = { ...inherited, ...own };
  orderViewSeries(view, own, encodings, faceted);
  for (const { view: child, layered } of childViews(view)) {
    orderSeries(child, layered ? encodings : {}, faceted || "facet" in view);
  }
}

/**
 * Orders the series a view draws itself, by the measure it draws or is layered over, or by name
 * in the panels of a facet.
 */
function orderViewSeries(
  view: Record<string, unknown>,
  own: Record<string, EncodingRecord>,
  encodings: Record<string, EncodingRecord>,
  faceted: boolean,
): void {
  const mark = view.mark as string | Record<string, unknown> | undefined;
  const markType = typeof mark === "object" ? mark.type : mark;
  if (markType === "arc") return;
  const measures = [encodings.x, encodings.y].filter(
    (encoding) =>
      encoding &&
      !Array.isArray(encoding) &&
      encoding.type === "quantitative" &&
      typeof encoding.field === "string" &&
      !encoding.field.startsWith("__") &&
      !encoding.bin,
  );
  // The series of a facet are listed by name, which needs no measure.
  if (measures.length !== 1 && !faceted) return;
  // The column each field's series are placed in: several channels may draw one field.
  const places = new Map<string, string>();
  for (const channel of SERIES_CHANNELS) {
    // A series drawn by the view this one is layered in is ordered there.
    const encoding = own[channel];
    if (
      !encoding ||
      Array.isArray(encoding) ||
      !isDiscrete(encoding) ||
      encoding.sort !== null ||
      typeof encoding.field !== "string" ||
      encoding.field.startsWith("__")
    ) {
      continue;
    }
    if (
      encoding.type === "ordinal" ||
      COMPOSITE_MARKS.has(markType as string) ||
      faceted
    ) {
      encoding.sort = "ascending";
      continue;
    }
    let place = places.get(encoding.field);
    if (!place) {
      const suffix = places.size ? `_${places.size + 1}` : "";
      const total = SERIES_TOTAL_FIELD + suffix;
      place = SERIES_PLACE_FIELD + suffix;
      places.set(encoding.field, place);
      view.transform = [
        ...((view.transform as unknown[] | undefined) ?? []),
        {
          joinaggregate: [{ op: "sum", field: measures[0].field, as: total }],
          groupby: [encoding.field],
        },
        {
          window: [{ op: "dense_rank", as: place }],
          sort: [
            { field: total, order: "descending" },
            { field: encoding.field, order: "ascending" },
          ],
        },
      ];
    }
    encoding.sort = { field: place, op: "min" };
  }
}

/**
 * Flint stacks a Sparkline's series in rows, in each of its columns of names, trends and values,
 * in the order their rows arrive, and drops any order the author gives them. The query of a chart
 * over time is ordered by time, which leaves the order of each bucket's rows to the database, so a
 * refresh moved the series from row to row. Their rows are listed by name, as their colors are
 * (see `orderSeries`).
 */
function sortSparklineRows(spec: Record<string, unknown>): void {
  for (const { view } of childViews(spec)) {
    const row = (view.facet as { row?: Record<string, unknown> } | undefined)
      ?.row;
    if (row && Array.isArray(row.sort)) row.sort = "ascending";
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
    for (const [, encoding] of channelsOf(view)) {
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

    const record = node as EncodingRecord;
    if (
      typeof record.field === "string" &&
      record.type === "quantitative" &&
      measures.has(record.field)
    ) {
      routeThroughFormatter(record, sanitizeFieldName(record.field));
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

/**
 * The channels whose title is shown as an axis, legend or facet header title; `x2`, `y2` and the
 * tooltip have none.
 */
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
  "facet",
  "column",
  "row",
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
  for (const { view } of childViews(spec)) views.push(...layersOf(view));
  return views;
}

/**
 * The views nested directly in a view, each with whether it is a layer: of them, only a layer
 * draws the encoding of the view it is in.
 */
function childViews(
  spec: Record<string, unknown>,
): { view: Record<string, unknown>; layered: boolean }[] {
  const children: { view: Record<string, unknown>; layered: boolean }[] = [];
  const wrapped = spec.spec;
  if (wrapped && typeof wrapped === "object") {
    children.push({ view: wrapped as Record<string, unknown>, layered: false });
  }
  for (const key of ["layer", "hconcat", "vconcat", "concat"]) {
    const nested = spec[key];
    if (!Array.isArray(nested)) continue;
    for (const view of nested as Record<string, unknown>[]) {
      children.push({ view, layered: key === "layer" });
    }
  }
  return children;
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
 * A view's channels and their encodings, without the channels Flint turns off with `null`
 * (the tooltip of a layer that only decorates, such as a waterfall's connectors).
 */
function channelsOf(view: Record<string, unknown>): [string, EncodingRecord][] {
  return Object.entries(encodingsOf(view)).filter(
    ([, encoding]) => encoding !== null,
  );
}

/**
 * Flint's derived encodings miss the labels and formats Rill gives a measure. A waterfall draws
 * its running sums from columns of its own, titled with the measure; its connector layer carries no
 * title, and its step axis has none either, so Vega-Lite composes one from the field names
 * ("event_time, __wf_lead").
 *
 * A quantitative encoding of a derived field takes the display name and formatter of the measure it
 * is titled with, by name or by display name. Without such a title it takes those of the measure
 * another layer draws on the same channel, since the layers share that channel's scale and axis.
 * So does a layer that states an axis but no type of its own: Vega-Lite takes a layer's axis whole
 * in place of the one the chart states for the channel, and with it went the formatter of a
 * candlestick's prices. An encoding of a measure keeps its own formatter. And an encoding of a
 * known field without a title gets the field's display name, on the channels whose title is shown.
 */
function labelDerivedEncodings(
  spec: Record<string, unknown>,
  fields: FlintFields,
  measureFields: string[],
): void {
  if (!measureFields.length) return;
  const measures = new Set(measureFields);
  const measureOf = (name: unknown) =>
    typeof name === "string" && measures.has(name) ? name : undefined;
  const measureTitled = (title: unknown) =>
    measureOf(title) ??
    measureFields.find((name) => fields.field_display_names[name] === title);
  const views = layersOf(spec);

  // The measure each channel draws, found by field or by the title Flint gives it in any layer.
  const channelMeasure = new Map<string, string>();
  for (const view of views) {
    for (const [channel, encoding] of channelsOf(view)) {
      if (Array.isArray(encoding) || channelMeasure.has(channel)) continue;
      const measure =
        measureOf(encoding.field) ?? measureTitled(encoding.title);
      if (measure) channelMeasure.set(channel, measure);
    }
  }

  for (const view of views) {
    for (const [channel, encoding] of channelsOf(view)) {
      if (Array.isArray(encoding)) continue;
      const derived =
        encoding.type === "quantitative" && !measureOf(encoding.field);
      const measure =
        derived || (encoding.type === undefined && encoding.axis)
          ? ((derived ? measureTitled(encoding.title) : undefined) ??
            channelMeasure.get(channel))
          : undefined;
      if (measure) {
        routeThroughFormatter(encoding, sanitizeFieldName(measure));
        if (derived && encoding.title === measure) {
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

/**
 * A legend of a measure's values, the sizes of a bubble chart or the colors of a heatmap, printed
 * them in Vega's own number format: a rate of 0.5% read "5m" beside an axis that read "0.5%".
 * Vega-Lite takes a legend's format from the legend alone, so each legend of an encoding that was
 * given a measure's formatter states it too. Flint's own `format` goes, as it does on an axis.
 */
function formatMeasureLegends(spec: Record<string, unknown>): void {
  for (const view of layersOf(spec)) {
    for (const [channel, encoding] of channelsOf(view)) {
      if (
        Array.isArray(encoding) ||
        !(LEGEND_CHANNELS.has(channel) || COLOR_CHANNELS.has(channel)) ||
        encoding.type !== "quantitative" ||
        typeof encoding.formatType !== "string" ||
        encoding.legend === null
      ) {
        continue;
      }
      const legend: Record<string, unknown> = {
        ...encoding.legend,
        formatType: encoding.formatType,
      };
      delete legend.format;
      encoding.legend = legend;
    }
  }
}

/**
 * Points an encoding's labels at a Rill measure formatter. Flint's own number formats go: a
 * `format`, which Vega-Lite would pass to the formatter as an argument, and an axis `labelExpr`,
 * which would replace the formatter's labels. One of those is Flint's integer-only expression for
 * counts (`format(datum.value, ',d')`); the integer ticks themselves come from the axis'
 * `tickMinStep`, which stays.
 */
function routeThroughFormatter(
  encoding: EncodingRecord,
  formatType: string,
): void {
  if (encoding.axis) {
    delete encoding.axis.format;
    delete encoding.axis.labelExpr;
    encoding.axis.formatType = formatType;
  }
  delete encoding.format;
  encoding.formatType = formatType;
}

/**
 * Label formats for a time field drawn on a discrete axis or legend, keyed by Flint's temporal
 * types (the ones `semantic-types.ts` emits). `expr` formats an axis or legend label from
 * `datum.value`, which is a Date or an ISO string depending on who supplied the rows; `d3` is
 * the tooltip's time format.
 */
const DISCRETE_TIME_FORMATS: Record<string, { expr: string; d3: string }> = {
  Year: { expr: "timeFormat(toDate(datum.value), '%Y')", d3: "%Y" },
  YearQuarter: {
    expr: "timeFormat(toDate(datum.value), 'Q%q %Y')",
    d3: "Q%q %Y",
  },
  YearMonth: { expr: "timeFormat(toDate(datum.value), '%b %Y')", d3: "%b %Y" },
  YearWeek: { expr: "timeFormat(toDate(datum.value), '%b %d')", d3: "%b %d" },
  Date: { expr: "timeFormat(toDate(datum.value), '%b %d')", d3: "%b %d" },
  DateTime: {
    expr: "timeFormat(toDate(datum.value), '%b %d %H:%M')",
    d3: "%b %d %H:%M",
  },
};

/** The channels Vega-Lite leaves out of the tooltip it builds from a mark's encoding. */
const UNLISTED_TOOLTIP_CHANNELS = new Set([
  "tooltip",
  "href",
  "url",
  "description",
  "key",
]);

/** What a tooltip entry takes from the encoding it lists. */
const TOOLTIP_ENTRY_KEYS = [
  "field",
  "type",
  "title",
  "aggregate",
  "timeUnit",
  "bin",
  "format",
  "formatType",
];

/** Prefix of the columns a tooltip reads a discrete time field from, parsed into dates. */
const TIME_COPY_PREFIX = "__rill_time_";

/** Marks that summarize their data and list that summary in a tooltip of their own. */
const COMPOSITE_MARKS = new Set(["boxplot", "errorbar", "errorband"]);

/**
 * Some templates draw the time field on a discrete scale (a waterfall's steps, a slope chart's two
 * periods, the bars of a chart with only a few time buckets), and Vega-Lite then prints the raw
 * timestamps as category labels. The data keeps its chronological order; only the labels are
 * formatted at the field's grain, on the axis or legend that shows them, and in the tooltip.
 *
 * The tooltip Vega-Lite builds from a mark's encoding has no format to set, so a mark that would
 * get one for such a field is given the same list of fields explicitly, less the columns Flint
 * computes for itself. So is a mark colored over a domain Flint states: Vega-Lite orders a stack
 * by that domain through an index column of its own, which its tooltip would list as well.
 */
function formatDiscreteTimeAxes(
  spec: Record<string, unknown>,
  semanticTypes: Record<string, string>,
): void {
  const formatOf = (field: unknown) =>
    typeof field === "string"
      ? DISCRETE_TIME_FORMATS[semanticTypes[field]]
      : undefined;

  const listTooltips = (
    view: Record<string, unknown>,
    inherited: Record<string, EncodingRecord>,
  ): void => {
    const own = encodingsOf(view);
    const encoding = { ...inherited, ...own };
    const mark = view.mark as string | Record<string, unknown> | undefined;
    const markType = typeof mark === "object" ? mark.type : mark;
    const silenced =
      typeof mark === "object" &&
      (mark.tooltip === null || mark.tooltip === false);
    if (
      mark &&
      !silenced &&
      !COMPOSITE_MARKS.has(markType as string) &&
      !("tooltip" in encoding)
    ) {
      const listed = Object.entries(encoding)
        .filter(
          ([channel, def]) =>
            !UNLISTED_TOOLTIP_CHANNELS.has(channel) &&
            def &&
            !Array.isArray(def) &&
            (typeof def.field === "string"
              ? def.type !== undefined &&
                (!def.field.startsWith("__") || typeof def.title === "string")
              : def.aggregate === "count"),
        )
        .map(([, def]) => def);
      const ordersStack =
        !("order" in encoding) &&
        [encoding.fill, encoding.color].some(
          (def) =>
            def &&
            !Array.isArray(def) &&
            def.type === "nominal" &&
            typeof def.field === "string" &&
            Array.isArray(
              (def.scale as { domain?: unknown } | undefined)?.domain,
            ),
        );
      if (
        ordersStack ||
        listed.some((def) => formatOf(def.field) && isDiscrete(def))
      ) {
        view.encoding = {
          ...own,
          tooltip: listed.map((def) =>
            Object.fromEntries(
              TOOLTIP_ENTRY_KEYS.filter(
                (key) => def[key] !== undefined && def[key] !== null,
              ).map((key) => [key, def[key]]),
            ),
          ),
        };
      }
    }
    // A layer's views read its encoding; a facet's or a concatenation's do not.
    for (const key of ["layer", "spec", "hconcat", "vconcat", "concat"]) {
      const children = view[key] as
        | Record<string, unknown>
        | Record<string, unknown>[]
        | undefined;
      for (const child of [children ?? []].flat()) {
        listTooltips(child, key === "layer" ? encoding : {});
      }
    }
  };
  listTooltips(spec, {});

  // The columns the tooltips read a time field from, by field.
  const copies = new Map<string, string>();
  for (const view of layersOf(spec)) {
    for (const [channel, encoding] of channelsOf(view)) {
      if (Array.isArray(encoding)) {
        for (const entry of encoding as EncodingRecord[]) {
          const format = formatOf(entry.field);
          if (!format || !isDiscrete(entry)) continue;
          // A time format needs a field typed as time, and Vega-Lite parses such a field into
          // dates wherever the chart reads it. The order Flint gives the discrete axis lists
          // the raw values, which the dates would no longer match, so the tooltip reads a copy.
          const field = entry.field as string;
          const copy = copies.get(field) ?? `${TIME_COPY_PREFIX}${copies.size}`;
          copies.set(field, copy);
          const transforms = (view.transform ?? []) as Record<
            string,
            unknown
          >[];
          if (!transforms.some((transform) => transform.as === copy)) {
            view.transform = [
              ...transforms,
              {
                calculate: `toDate(datum[${JSON.stringify(field)}])`,
                as: copy,
              },
            ];
          }
          entry.title ??= field;
          entry.field = copy;
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

/**
 * A 100% stacked chart's measure axis runs from 0 to 1, and for a count measure Flint's integer
 * axis left only "0" and "1" on it: its `tickMinStep` of 1 dropped the ticks between, and the
 * measure's formatter (or Flint's integer label expression) printed the shares as counts. The
 * axis shows the shares as percentages instead; tooltips keep the measure's formatter.
 */
function formatNormalizedStackAxes(spec: Record<string, unknown>): void {
  for (const view of layersOf(spec)) {
    for (const channel of ["x", "y"]) {
      const encoding = encodingsOf(view)[channel];
      if (
        !encoding ||
        Array.isArray(encoding) ||
        encoding.stack !== "normalize" ||
        encoding.axis === null
      ) {
        continue;
      }
      const axis: Record<string, unknown> = { ...encoding.axis, format: ".0%" };
      delete axis.formatType;
      delete axis.labelExpr;
      delete axis.tickMinStep;
      encoding.axis = axis;
    }
  }
}
