import type { View, VisualizationSpec } from "svelte-vega";
import type { Config } from "vega-lite";

/**
 * Lays out the legends of a chart so that they fit it, whoever wrote its Vega-Lite: a built-in
 * chart's builder, the custom chart compiler, or the author of a `vega_spec`.
 *
 * Left to Vega, a legend above or below the plot runs in a single row, and a row longer than the
 * plot is cut off at the chart's edge: a custom chart of two dozen series named three of them.
 * Built-in charts wrapped that row every 140px instead, which stacked short labels in narrow
 * columns beside unused room, and listed up to 40 entries however little height the chart had,
 * so in a small widget the legend took all of it and no plot was drawn.
 *
 * Here a legend takes about a third of its chart at most. Above or below the plot it wraps into
 * as many columns as the plot's width holds, measured from its labels, and stops after the rows
 * a third of the chart's height holds, or after `MIN_ENTRIES` entries where a narrow plot holds
 * fewer in those rows: Vega's "…N entries" stands for the rest. Beside the plot it lists what the
 * plot's height holds, with labels cut so that it takes a third of the chart's width. The legends
 * of categories that stand above and below the plot, one above the other on either side, share
 * that third and those `MIN_ENTRIES` between them.
 *
 * The layout is written as expressions over Vega's `width` and `height` signals, which are the
 * plot's, over the chart's own size, which the renderer keeps in a signal of its own (see
 * `setContainerSize`), and over the widths of each legend's labels, which it keeps in another
 * (see `setLegendLabels`). So it follows a resize, and new data with other labels, without the
 * chart being rebuilt, and `refitOnPlotResize` then fits the plot around the legend they give.
 *
 * A legend that states its own `columns` or `direction` is left as its author laid it out, and so
 * is every legend of a spec whose `config` does. So are the legends of a chart whose plot is not
 * fitted to its container, where `width` or `height` says nothing of the room there is: a chart
 * sized by its bands' step, and a faceted or concatenated one, whose sizes are a panel's.
 *
 * A spec comes back as the same object for as long as its legends are laid out the same: the
 * same spec under the same config.
 */
export function fitLegends(
  spec: VisualizationSpec,
  config: Config | undefined,
): VisualizationSpec {
  // A `vega_spec` that parses to something other than an object is Vega's to report.
  const root = record(spec);
  if (!root || MULTI_VIEW_KEYS.some((key) => key in root)) return spec;
  const encoding = record(root.encoding);
  if (encoding && FACET_CHANNELS.some((key) => key in encoding)) return spec;

  const authored = record(record(root.config)?.legend);
  const themed = record(config?.legend);
  // Giving a legend properties of its own would bring back one its chart has turned off.
  if (authored?.disable ?? themed?.disable) return spec;
  if (authored?.columns !== undefined || authored?.direction !== undefined) {
    return spec;
  }
  // The params the chart declares, whose names its signals take. The renderer sets the container's
  // size by name, so a chart that already has a param of that name is left as it is.
  const named = paramNames(root);
  if (named.has(CONTAINER_SIGNAL)) return spec;
  // The plot's `width` follows the container's where the chart states a width for Vega to fit,
  // and its `height` where the chart is fitted both ways. A chart with neither is drawn at the
  // size of its bands or of Vega-Lite's default view.
  const autosize = root.autosize ?? config?.autosize;
  const fit = record(autosize)?.type ?? autosize;
  // Where the chart's config puts its legends; Vega-Lite's own default is the right.
  const orient = authored?.orient ?? themed?.orient ?? "right";
  const layouts: ViewSpec[] = [];
  const labelled: LabelledLegend[] = [];
  // The font its labels are drawn in, which a theme may set; Vega's own is 10px sans-serif.
  const font = {
    family: String(
      authored?.labelFont ??
        themed?.labelFont ??
        record(root.config)?.font ??
        config?.font ??
        "sans-serif",
    ),
    size: Number(authored?.labelFontSize ?? themed?.labelFontSize) || 10,
  };
  const fitted = fitView(root, () => undefined, {
    layouts,
    labelled,
    named,
    fitsWidth:
      (fit === "fit" || fit === "fit-x") &&
      (typeof root.width === "number" || root.width === "container"),
    fitsHeight: fit === "fit" || fit === "fit-y",
    orient,
    sharing: countCategoryLegends(root, orient),
    font,
  });

  // The labels are measured in the font, which the layouts do not name.
  const key = JSON.stringify({ layouts, font });
  const last = lastFitted.get(root);
  if (last?.key === key) return last.spec;
  const result = (
    layouts.length
      ? {
          ...fitted,
          // The size of the chart's container and the labels of its legends, for the layouts
          // to read once they are known.
          params: [
            ...(Array.isArray(root.params) ? (root.params as unknown[]) : []),
            { name: CONTAINER_SIGNAL, value: {} },
            // Labels a spec holds, as its own rows or as a scale's domain, are known from
            // the start; those of a dataset once the chart is told them.
            ...labelled.map(({ signal, labels }) => ({
              name: signal,
              value: labels({}),
            })),
          ],
        }
      : fitted
  ) as VisualizationSpec;
  lastFitted.set(root, { key, spec: result });
  labelledLegends.set(result as object, labelled);
  return result;
}

/**
 * Tells a chart the labels of its legends, which `fitLegends` lays them out by: those its spec
 * states, or else those in the rows of `datasets`, or of the spec, that it draws. They are kept in
 * signals, so that new data with other labels lays the legends out again without the chart being
 * rebuilt. Until a chart is told the labels of its datasets, each column of a legend that lists
 * them is taken to be as wide as its labels may run.
 *
 * Returns whether any of them changed, for the view to be run again. Its plot is then fitted
 * around the legends they lay out: Vega fits it only once the view is resized.
 */
export function setLegendLabels(
  view: View,
  spec: VisualizationSpec,
  datasets: Record<string, unknown>,
): boolean {
  let changed = false;
  for (const { signal, labels } of labelledLegends.get(spec as object) ?? []) {
    const value = labels(datasets);
    try {
      if (JSON.stringify(view.signal(signal)) === JSON.stringify(value)) {
        continue;
      }
      view.signal(signal, value);
      changed = true;
    } catch {
      // The view of another spec, as one being replaced is, reads no such labels.
    }
  }
  if (changed) view.resize();
  return changed;
}

/**
 * Tells a chart the size of its container, which `fitLegends` sizes its legends against. Until a
 * chart is told, as one drawn outside `VegaLiteRenderer` never is, a legend above or below the
 * plot takes `DEFAULT_ROWS` rows and one beside it cuts its labels at `LABEL_LIMIT`.
 */
export function setContainerSize(
  view: View,
  width: number,
  height: number,
): void {
  // Vega offers no way to ask whether a signal exists; setting one that does not throws. A chart
  // with no legend to lay out has none.
  try {
    view.signal(CONTAINER_SIGNAL, { width, height });
  } catch {
    // Nothing reads the size.
  }
}

/** The signal a chart's container size is kept in, as `{ width, height }`. */
const CONTAINER_SIGNAL = "rill_container";
/** The prefix of the signals the labels of a chart's legends are kept in, as `LegendLabels`. */
const LABELS_SIGNAL = "rill_legend_labels";
/** The share of its chart's height that a legend above or below the plot may take. */
const HEIGHT_SHARE = 0.3;
/**
 * The share of its chart's width that a legend beside the plot may take. A little more than of
 * its height: what does not fit here is cut from every label, not left to a last entry.
 */
const WIDTH_SHARE = 0.35;

/** Each spec as it was last laid out, with the layouts that gave it. */
const lastFitted = new WeakMap<
  ViewSpec,
  { key: string; spec: VisualizationSpec }
>();
/** The legends of each spec `fitLegends` gave that are laid out by their labels. */
const labelledLegends = new WeakMap<object, LabelledLegend[]>();

/** The widest a legend's label runs before it is cut, in pixels, where its legend is crowded. */
const LABEL_LIMIT = 140;
/** The widest a label runs where every entry of its legend still fits, in pixels. */
const FULL_LABEL_LIMIT = 240;
/** The narrowest a label is cut to, in pixels, however narrow the plot. */
const MIN_LABEL_LIMIT = 40;
/** A legend entry's symbol and the gap before its label, in pixels. */
const SYMBOL_WIDTH = 16;
/** Vega's gap between the columns of a legend, in pixels. */
const COLUMN_PADDING = 10;
/** Vega's gap between the rows of a legend, in pixels. */
const ROW_PADDING = 2;
/** A legend's title line and Vega's padding below it, in pixels at the theme's title size. */
const TITLE_HEIGHT = 17;
/** The fewest and the most rows a legend above or below the plot takes, whatever its chart's height. */
const MIN_ROWS = 2;
const MAX_ROWS = 10;
/** The rows it takes while the chart's height is not known. */
const DEFAULT_ROWS = 4;
/** The entries such a legend lists whatever rows they take: a handful of series are all named. */
const MIN_ENTRIES = 6;
/** The most columns a legend's width is worked out for; a wider plot lays out no more. */
const MAX_COLUMNS = 12;
/**
 * Room for an entry of a legend of a measure's values, in pixels: a symbol as large as a size
 * scale draws them, a formatted number and the gap to the next entry.
 */
const VALUE_ENTRY_WIDTH = 80;
/** The label Vega ends a legend with once it leaves entries out, at its widest. */
const LEFT_OUT_LABEL = "…999 entries";

/** The channels Vega-Lite draws a legend for. */
const LEGEND_CHANNELS = [
  "color",
  "fill",
  "stroke",
  "strokeDash",
  "shape",
  "size",
  "opacity",
  "fillOpacity",
  "strokeOpacity",
  "strokeWidth",
];
/** The channels whose legend of a measure is a gradient rather than a list of symbols. */
const GRADIENT_CHANNELS = new Set(["color", "fill", "stroke"]);
const MULTI_VIEW_KEYS = ["facet", "repeat", "concat", "hconcat", "vconcat"];
const FACET_CHANNELS = ["facet", "row", "column"];

/** A view of a Vega-Lite spec as this module reads it; other keys pass through untouched. */
type ViewSpec = Record<string, unknown>;
/** The rows a view draws, from the datasets its chart is given. */
type RowsSource = (datasets: Record<string, unknown>) => unknown[] | undefined;

/** A legend laid out by its labels, and how they are read. */
interface LabelledLegend {
  /** The signal the labels are kept in. */
  signal: string;
  labels: (datasets: Record<string, unknown>) => LegendLabels | null;
}

/**
 * What a legend's layout reads of its labels, widths in pixels: how many there are, and how wide
 * the plot is to be for each number of their columns from two on, with the labels cut where the
 * legend is crowded, and in full where that cuts any.
 */
interface LegendLabels {
  count: number;
  crowded: number[];
  full: number[] | null;
}

interface LabelFont {
  family: string;
  size: number;
}

interface FitContext {
  /** Every layout given, in the order the legends are met. */
  layouts: ViewSpec[];
  /** The legends among them laid out by their labels. */
  labelled: LabelledLegend[];
  /** The names the chart's params take, and those given to the signals of `labelled` so far. */
  named: Set<string>;
  /** Whether the plot's `width` and `height` are fitted to the chart's container. */
  fitsWidth: boolean;
  fitsHeight: boolean;
  orient: unknown;
  /** How many legends of categories the chart draws above and below the plot, sharing its height. */
  sharing: number;
  font: LabelFont;
}

/**
 * A view with its legends laid out, and its layers' with it. The view itself is returned when
 * none of them changes, so that a chart is not rebuilt for a spec that reads the same.
 */
function fitView(
  view: ViewSpec,
  inheritedRows: RowsSource,
  context: FitContext,
): ViewSpec {
  const rows: RowsSource = (datasets) =>
    rowsOf(view.data, datasets) ?? inheritedRows(datasets);
  let fitted = view;

  const encoding = record(view.encoding);
  if (encoding) {
    const layouts = LEGEND_CHANNELS.flatMap((channel) => {
      const definition = record(encoding[channel]);
      const layout =
        definition && legendLayout(channel, definition, rows, context);
      return definition && layout ? [{ channel, definition, layout }] : [];
    });
    if (layouts.length) {
      const fittedEncoding = { ...encoding };
      for (const { channel, definition, layout } of layouts) {
        context.layouts.push(layout);
        fittedEncoding[channel] = {
          ...definition,
          // What the legend states for itself is kept.
          legend: { ...layout, ...record(definition.legend) },
        };
      }
      fitted = { ...view, encoding: fittedEncoding };
    }
  }

  if (Array.isArray(view.layer)) {
    const layers = view.layer as unknown[];
    const fittedLayers = layers.map((layer) => {
      const layerView = record(layer);
      return layerView ? fitView(layerView, rows, context) : layer;
    });
    if (fittedLayers.some((layer, i) => layer !== layers[i])) {
      fitted = { ...fitted, layer: fittedLayers };
    }
  }
  return fitted;
}

/**
 * What the legend an encoding draws lists: a field's categories or a measure's values. Undefined
 * for an encoding that draws no legend, or draws a gradient.
 */
function entriesOf(
  channel: string,
  definition: ViewSpec,
): "categories" | "values" | undefined {
  if (definition.legend === null) return undefined;
  // A constant or a condition alone draws no legend.
  if (definition.field === undefined && definition.datum === undefined) {
    return undefined;
  }
  const type =
    definition.type ??
    (definition.timeUnit || definition.bin || definition.aggregate
      ? undefined
      : "nominal");
  if (type === "nominal" || type === "ordinal") return "categories";
  const listsValues =
    (type === "quantitative" || type === "temporal") &&
    (!GRADIENT_CHANNELS.has(channel) ||
      record(definition.legend)?.type === "symbol");
  return listsValues ? "values" : undefined;
}

/**
 * How many legends of categories a chart and its layers draw above and below the plot: one for
 * each field listed on a side, as Vega-Lite draws the channels that encode one field as one
 * legend.
 */
function countCategoryLegends(root: ViewSpec, defaultOrient: unknown): number {
  const listed = new Set<string>();
  const visit = (view: ViewSpec) => {
    const encoding = record(view.encoding) ?? {};
    for (const channel of LEGEND_CHANNELS) {
      const definition = record(encoding[channel]);
      if (!definition || entriesOf(channel, definition) !== "categories") {
        continue;
      }
      const orient = record(definition.legend)?.orient ?? defaultOrient;
      if (orient === "top" || orient === "bottom") {
        listed.add(JSON.stringify([orient, definition.field]));
      }
    }
    if (Array.isArray(view.layer)) {
      for (const layer of view.layer as unknown[]) {
        const layerView = record(layer);
        if (layerView) visit(layerView);
      }
    }
  };
  visit(root);
  return listed.size;
}

/** The names of the params a view and its layers declare, each of which Vega makes a signal of. */
function paramNames(view: ViewSpec, names = new Set<string>()): Set<string> {
  for (const param of Array.isArray(view.params) ? view.params : []) {
    const name = record(param)?.name;
    if (typeof name === "string") names.add(name);
  }
  for (const layer of Array.isArray(view.layer) ? view.layer : []) {
    const layerView = record(layer);
    if (layerView) paramNames(layerView, names);
  }
  return names;
}

/** The layout of the legend an encoding draws, or undefined where it draws none to lay out. */
function legendLayout(
  channel: string,
  definition: ViewSpec,
  rows: RowsSource,
  context: FitContext,
): ViewSpec | undefined {
  const {
    orient: defaultOrient,
    sharing,
    font,
    fitsWidth,
    fitsHeight,
  } = context;
  const lists = entriesOf(channel, definition);
  const legend = record(definition.legend) ?? {};
  if (
    !lists ||
    legend.columns !== undefined ||
    legend.direction !== undefined
  ) {
    return undefined;
  }
  const listsCategories = lists === "categories";
  const rowHeight = font.size + ROW_PADDING;

  const orient = legend.orient ?? defaultOrient;
  if (orient === "left" || orient === "right") {
    // A measure's few values fit beside any plot.
    if (!listsCategories || !fitsHeight) return undefined;
    const share = `floor(${CONTAINER_SIGNAL}.width * ${WIDTH_SHARE}) - ${SYMBOL_WIDTH}`;
    return {
      columns: 1,
      labelLimit: {
        expr: `${CONTAINER_SIGNAL}.width ? clamp(${share}, ${MIN_LABEL_LIMIT}, ${LABEL_LIMIT}) : ${LABEL_LIMIT}`,
      },
      symbolLimit: { expr: `max(2, floor(height / ${rowHeight}))` },
    };
  }
  if ((orient !== "top" && orient !== "bottom") || !fitsWidth) return undefined;

  const titleLimit = { expr: `max(${MIN_LABEL_LIMIT}, width)` };
  if (!listsCategories) {
    return {
      columns: {
        expr: `max(1, floor((width + ${COLUMN_PADDING}) / ${VALUE_ENTRY_WIDTH}))`,
      },
      titleLimit,
    };
  }

  // The legends above and below the plot share the height that one alone takes: each listing
  // `MIN_ENTRIES`, two in a short widget left its plot no height at all. Each has a title of
  // its own, and the further titles come out of that height as well.
  const furtherTitles = (TITLE_HEIGHT * (sharing - 1)) / rowHeight;
  // A name and the count of the rest, at the least.
  const minEntries = Math.max(
    2,
    Math.floor((MIN_ENTRIES - furtherTitles) / sharing),
  );
  const minRows = Math.max(1, Math.floor(MIN_ROWS / sharing));
  const defaultRows = Math.max(minRows, Math.round(DEFAULT_ROWS / sharing));
  // The rows its share of the chart's height holds under the legend's title.
  const room = `${CONTAINER_SIGNAL}.height * ${HEIGHT_SHARE / sharing} - ${TITLE_HEIGHT}`;
  const maxRows = `(${CONTAINER_SIGNAL}.height ? clamp(floor((${room}) / ${rowHeight}), ${minRows}, ${MAX_ROWS}) : ${defaultRows})`;
  const entries = (columns: string) =>
    `max(${maxRows} * ${columns}, ${minEntries})`;

  // Until its labels are known, or where they cannot be read, each column is taken to be as wide
  // as a label may run.
  const column = SYMBOL_WIDTH + LABEL_LIMIT + COLUMN_PADDING;
  const unlabelled = `max(1, floor((width + ${COLUMN_PADDING}) / ${column}))`;
  if (!readsLabels(definition, legend)) {
    return {
      columns: { expr: unlabelled },
      labelLimit: {
        expr: `max(${MIN_LABEL_LIMIT}, min(${LABEL_LIMIT}, width - ${SYMBOL_WIDTH}))`,
      },
      symbolLimit: { expr: entries(unlabelled) },
      titleLimit,
    };
  }

  // A name no param of the chart takes: Vega refuses a second signal of one.
  let n = 0;
  while (context.named.has(`${LABELS_SIGNAL}_${n}`)) n++;
  const signal = `${LABELS_SIGNAL}_${n}`;
  context.named.add(signal);
  context.labelled.push({
    signal,
    labels: (datasets) => {
      const labels = labelsOf(definition, legend, rows(datasets));
      if (!labels) return null;
      const widths: LabelWidths = {
        count: labels.length,
        fewestListed: (columns) => Math.max(minRows * columns, minEntries),
        // Sorted once, for every number of columns and every label limit tried.
        widest: labels
          .map((label) => labelWidth(label, font))
          .sort((a, b) => b - a)
          .slice(0, MAX_COLUMNS),
        leftOut: labelWidth(LEFT_OUT_LABEL, font),
      };
      return {
        count: widths.count,
        crowded: widthsForColumns(widths, LABEL_LIMIT),
        full:
          widths.widest[0] > LABEL_LIMIT
            ? widthsForColumns(widths, FULL_LABEL_LIMIT)
            : null,
      };
    },
  });
  const crowded = columnsByWidth(`${signal}.crowded`);
  const inFull = columnsByWidth(`${signal}.full`);
  // Where every entry fits with its label in full, none is cut short: three long names in a wide
  // chart read better on two rows than as three stubs on one.
  const roomy = `${signal} && ${signal}.full && ${entries(inFull)} >= ${signal}.count`;
  const columns = `(${signal} ? (${roomy} ? ${inFull} : ${crowded}) : ${unlabelled})`;
  const labelLimit = `(${roomy} ? ${FULL_LABEL_LIMIT} : ${LABEL_LIMIT})`;
  return {
    columns: { expr: columns },
    // A single column is no wider than the plot it stands over.
    labelLimit: {
      expr: `max(${MIN_LABEL_LIMIT}, min(${labelLimit}, width - ${SYMBOL_WIDTH}))`,
    },
    // Vega lists one entry fewer than this, then how many it left out.
    symbolLimit: { expr: entries(columns) },
    titleLimit,
  };
}

/** A legend's labels, as far as they set the widths of its columns. Widths are in pixels. */
interface LabelWidths {
  /** How many labels the legend has. */
  count: number;
  /** The fewest entries the legend lists in a number of columns, however short its chart. */
  fewestListed: (columns: number) => number;
  /** Its widest labels, widest first: one for each column a legend may have. */
  widest: number[];
  /** The label that stands for the entries left out. */
  leftOut: number;
}

/**
 * The width of plot a legend's labels take in each number of columns from two on, cut at a limit.
 *
 * Vega sizes each column to its widest label, and which labels share a column follows the
 * legend's order, which is the data's. The widest label of each column is at most one of the
 * widest labels overall, so their sum holds for any order.
 */
function widthsForColumns(labels: LabelWidths, limit: number): number[] {
  const leftOut = Math.min(labels.leftOut, limit);
  const widths: number[] = [];
  // The widest labels, one for each column so far. Cutting them at the limit keeps their order.
  let labelsWidth = Math.min(labels.widest[0] ?? 0, limit);
  for (let columns = 2; columns <= labels.widest.length; columns++) {
    const narrowest = Math.min(labels.widest[columns - 1], limit);
    labelsWidth += narrowest;
    // Once entries are left out, the label that says so takes a place of its own, in place of
    // the narrowest of these where it is the wider. How many rows the chart's height allows is
    // not known here, so it is counted wherever the fewest would leave entries out.
    const leavesOut = labels.count > labels.fewestListed(columns);
    widths.push(
      labelsWidth +
        (leavesOut ? Math.max(0, leftOut - narrowest) : 0) +
        columns * SYMBOL_WIDTH +
        (columns - 1) * COLUMN_PADDING,
    );
  }
  return widths;
}

/**
 * An expression for the most columns of a legend that fit across the plot's `width`, given an
 * expression for the widths they take (see `widthsForColumns`). A legend has as many columns as
 * it has labels, up to `MAX_COLUMNS`: the widths of more are undefined, which no width reaches.
 */
function columnsByWidth(widths: string): string {
  let expr = "1";
  for (let columns = 2; columns <= MAX_COLUMNS; columns++) {
    expr = `(width >= ${widths}[${columns - 2}] ? ${columns} : ${expr})`;
  }
  return expr;
}

/**
 * Whether a legend of categories has labels that can be read off the chart. A legend that formats
 * its labels, or a field the chart derives, has none.
 */
function readsLabels(definition: ViewSpec, legend: ViewSpec): boolean {
  return !(
    legend.labelExpr !== undefined ||
    legend.format !== undefined ||
    legend.formatType !== undefined ||
    definition.timeUnit ||
    definition.bin ||
    definition.aggregate
  );
}

/**
 * The labels of a legend of categories that `readsLabels`: the values the legend or its scale
 * states, or else the field's in the rows the chart draws.
 */
function labelsOf(
  definition: ViewSpec,
  legend: ViewSpec,
  rows: unknown[] | undefined,
): string[] | undefined {
  const stated = legend.values ?? record(definition.scale)?.domain;
  if (Array.isArray(stated)) {
    return stated.every((value) => typeof value !== "object")
      ? stated.map(String)
      : undefined;
  }
  const field = definition.field;
  // Rows need not be objects: Vega-Lite also draws a list of numbers or strings.
  const first = record(rows?.[0]);
  if (typeof field !== "string" || !rows || !first || !(field in first)) {
    return undefined;
  }
  return [...new Set(rows.map((row) => String(record(row)?.[field])))];
}

/** The rows a view's `data` names or holds, when it does either. */
function rowsOf(
  data: unknown,
  datasets: Record<string, unknown>,
): unknown[] | undefined {
  const source = record(data);
  if (!source) return undefined;
  const rows =
    source.values ??
    (typeof source.name === "string" ? datasets[source.name] : undefined);
  return Array.isArray(rows) ? (rows as unknown[]) : undefined;
}

/** A generous width per character, as a share of the font size, where no canvas measures text. */
const CHAR_WIDTH_PER_FONT_SIZE = 0.6;
let measureContext: CanvasRenderingContext2D | null | undefined;

/**
 * The width a legend's label is drawn at, in pixels, rounded up with a pixel to spare. Measured
 * on a canvas where there is one, and estimated per character elsewhere (jsdom), so that the
 * layout is the same on every run there.
 */
function labelWidth(text: string, font: LabelFont): number {
  if (measureContext === undefined) {
    measureContext =
      typeof document === "undefined"
        ? null
        : document.createElement("canvas").getContext("2d");
  }
  let width = text.length * font.size * CHAR_WIDTH_PER_FONT_SIZE;
  if (measureContext) {
    measureContext.font = `${font.size}px ${font.family}`;
    width = measureContext.measureText(text).width;
  }
  return Math.ceil(width) + 1;
}

function record(value: unknown): Record<string, unknown> | undefined {
  return value && typeof value === "object" && !Array.isArray(value)
    ? (value as Record<string, unknown>)
    : undefined;
}
