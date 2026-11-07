import type { ChartWarning, RawEncodingValue } from "flint-chart/core";
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
}: CompileOptions): CompileResult {
  if (!spec?.chartType) {
    return { spec: null, warnings: [], error: "No chart type provided" };
  }

  // Flint's donut is its pie under another name, with the hole left at 0: a full pie.
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
  }
  labelDerivedEncodings(vegaSpec, fields, measureFields);
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
 */
function spaceTimeAxisTicks(
  spec: Record<string, unknown>,
  rows: Record<string, unknown>[],
  semanticTypes: Record<string, string>,
  size: { width: number; height: number } | undefined,
): void {
  if (!size) return;
  for (const view of layersOf(spec)) {
    for (const [channel, dimension] of [
      ["x", "width"],
      ["y", "height"],
    ] as const) {
      const encoding = encodingsOf(view)[channel];
      if (
        !encoding ||
        Array.isArray(encoding) ||
        encoding.type !== "temporal" ||
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
 * A single donut follows the size of its view; the panels of a faceted one keep the size Flint
 * gives them (and have no `width` or `height` signal of their own), so their hole is fixed too.
 */
function scaleDonutHole(spec: Record<string, unknown>): void {
  const mark = spec.mark as Record<string, unknown> | undefined;
  if (mark?.type !== "arc" || typeof mark.innerRadius !== "number") return;
  const share = mark.innerRadius / 100;
  const encoding = (spec.encoding ?? {}) as Record<string, unknown>;
  const faceted = ["facet", "row", "column"].some((key) => key in encoding);
  mark.innerRadius =
    faceted && typeof spec.width === "number" && typeof spec.height === "number"
      ? (Math.min(spec.width, spec.height) / 2) * share
      : { expr: `min(width, height) / 2 * ${share}` };
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
  authored: Record<string, RawEncodingValue>,
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
 * A quantitative encoding of a derived field takes the display name and formatter of the measure it
 * is titled with. Without such a title it takes those of the measure another layer draws on the
 * same channel, since the layers share that channel's scale and axis. An encoding of a measure keeps
 * its own formatter. And an encoding of a known field without a title gets the field's display
 * name, on the channels whose title is shown.
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
  const views = layersOf(spec);

  // The measure each channel draws, found by field or by Flint's raw-name title in any layer.
  const channelMeasure = new Map<string, string>();
  for (const view of views) {
    for (const [channel, encoding] of Object.entries(encodingsOf(view))) {
      if (Array.isArray(encoding) || channelMeasure.has(channel)) continue;
      const measure = measureOf(encoding.field) ?? measureOf(encoding.title);
      if (measure) channelMeasure.set(channel, measure);
    }
  }

  for (const view of views) {
    for (const [channel, encoding] of Object.entries(encodingsOf(view))) {
      if (Array.isArray(encoding)) continue;
      const measure =
        encoding.type === "quantitative" && !measureOf(encoding.field)
          ? (measureOf(encoding.title) ?? channelMeasure.get(channel))
          : undefined;
      if (measure) {
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
