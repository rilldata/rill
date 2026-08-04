/**
 * Generates the custom viz gallery from Flint's own chart registry and example set.
 *
 * Run manually after bumping the flint-chart dependency:
 *
 *   npx tsx web-common/scripts/generate-flint-catalog.ts
 *
 * The gallery used to be scraped from the Vega-Lite examples site. It is now derived from the
 * installed package, so it cannot drift from the version we render with.
 *
 * Flint's example set is its rendering test suite: alongside the curated examples behind its own
 * gallery (https://microsoft.github.io/flint-chart/#/gallery/vegalite) it carries a combinatorial
 * matrix (the same chart at 20/100/1800 points, with and without color) and deliberately adversarial
 * cases (axis overflow, degenerate single-series, warning paths). A picker wants neither, so this
 * takes the curated examples first and drops anything tagged adversarial; see pickExamples.
 *
 * Each example is compiled and rendered to an SVG thumbnail here, at build time, so the gallery
 * grid stays instant and needs no chart engine. The SVGs are inlined into the index as data URIs
 * rather than written as sibling files: web-common is a library the apps alias in from outside their
 * own root ("@rilldata/web-common": "/../web-common/src" in web-local/vite.config.ts, with
 * server.fs.allow scoped to "."), so an asset URL Vite resolves against web-common 404s when
 * web-local or web-admin serves the dialog. A data URI belongs to the module, so it just works.
 *
 * Field roles are resolved from each example's declared semantic types via Flint's registry, then
 * mapped onto Rill's param types (measure / dimension / time_dimension).
 */

import { createRequire } from "node:module";
import { readFileSync, rmSync, writeFileSync } from "node:fs";
import { join } from "node:path";
import * as vega from "vega";
import { compile } from "vega-lite";
import {
  assembleVegaLite,
  vlAllTemplateDefs,
  vlGetTemplateDef,
  vlTemplateDefs,
} from "flint-chart/vegalite";
import { getRegistryEntry, SemanticTypes } from "flint-chart/core";
import { TEST_GENERATORS } from "flint-chart/test-data";

const require = createRequire(import.meta.url);

const EXAMPLES_DIR = join(
  import.meta.dirname,
  "../src/features/custom-viz/examples",
);
const OUT_PATH = join(EXAMPLES_DIR, "flint-charts-index.json");

/** Thumbnail render box. Small enough that Flint drops dense examples down to a readable summary. */
const THUMBNAIL_SIZE = { width: 240, height: 150 };

/** Keeps one chart type from flooding the grid, and bounds the vendored SVG payload. */
const MAX_PER_CHART_TYPE = 3;

/**
 * Tags marking an example as a rendering test rather than a starter: it either exceeds a layout
 * budget on purpose, exercises a degenerate shape, or asserts a warning fires. Showing one in the
 * picker would offer a chart that looks broken as a starting point.
 */
const ADVERSARIAL_TAGS = new Set([
  "stress",
  "overflow",
  "cutoff",
  "overstretch",
  "crowded",
  "very-large",
  "edge",
  "edge-case",
  "degenerate",
  "single-row",
  "warning",
  "dtype-conversion",
  "sparse",
  "ambiguous",
  "high-cardinality",
  "redundant-color",
  "redundant-group",
  "dodge-global",
  "dodge-local",
]);

/**
 * Titles from the internal suite that name a matrix cell rather than a chart, e.g. "N(4)×Q
 * +color(N,3) (12 pts)" or "Strip ▸ X ×20". Deprioritized behind prose titles: the derived title
 * carries the card, but the original still shows in tooltips and feeds search.
 */
const SHORTHAND_TITLE = /[×▸]\s*\d|^[A-Z]\(|^(?:EC|Phase \d):|\(\d+ pts/;

/** Skip a thumbnail that would bloat the index; the example is dropped rather than shipped blank. */
const MAX_THUMBNAIL_BYTES = 60_000;

/**
 * MIT requires the copyright and permission notice to travel with substantial portions of the
 * software. The example titles, descriptions and channel layouts here, and the thumbnails rendered
 * from the sample data, are all derived from flint-chart, so the notice is recorded in the
 * generated index alongside them.
 */
const LICENSE_NOTICE =
  "Chart examples and thumbnails derived from flint-chart (https://github.com/microsoft/flint-chart), MIT License, Copyright (c) Microsoft Corporation.";
const SOURCE_URL = "https://github.com/microsoft/flint-chart";

type Role = "measure" | "dimension" | "time_dimension";

/**
 * Chart types excluded because Rill never derives the semantics they need: our metrics view mapping
 * emits Category for every non-temporal dimension, so geographic channels would silently degrade.
 */
const EXCLUDED_CHART_TYPES = new Set(["Map", "Choropleth"]);

/** Semantic types Rill cannot produce; any example using one is skipped. */
const UNSUPPORTED_SEMANTIC_TYPES = new Set(["Latitude", "Longitude"]);

/** Param names by channel, describing the field's role in the chart. */
const CHANNEL_PARAM_NAMES: Record<string, string> = {
  x: "x_axis",
  y: "y_axis",
  x2: "x_axis_end",
  y2: "y_axis_end",
  angle: "angle",
  radius: "radius",
  color: "color",
  size: "size",
  shape: "shape",
  opacity: "opacity",
  detail: "detail",
  order: "order",
  group: "group",
  row: "row",
  column: "column",
  metric: "metric",
  value: "value",
  goal: "goal",
  id: "region",
  open: "open",
  high: "high",
  low: "low",
  close: "close",
};

/**
 * Row-count default by how much space each mark needs. Flint drops rows that overflow the axis
 * budget and reports it as a warning, so these stay under it.
 */
const DENSE_CHARTS = new Set([
  "Scatter Plot",
  "Regression",
  "Connected Scatter Plot",
  "Strip Plot",
  "Heatmap",
  "Histogram",
  "Density Plot",
  "ECDF Plot",
  "Boxplot",
  "Violin Plot",
]);
const TIME_SERIES_CHARTS = new Set([
  "Line Chart",
  "Area Chart",
  "Streamgraph",
  "Sparkline",
  "Range Area Chart",
  "Bump Chart",
  "Candlestick Chart",
]);

interface FlintExample {
  title: string;
  description?: string;
  tags?: string[];
  chartType: string;
  data: Record<string, unknown>[];
  metadata: Record<string, { type: string; semanticType: string }>;
  encodingMap: Record<string, { fieldID: string } | { fieldID: string }[]>;
  chartProperties?: Record<string, unknown>;
}

interface GalleryExample {
  name: string;
  /** Slug of the chart type, used as the generated component's file name. */
  chartSlug: string;
  chartType: string;
  category: string;
  /** Flint's own label for the example. Not shown on the card; it feeds search and the tooltip. */
  title: string;
  description: string;
  tags: string[];
  bindings: { channel: string; param: string; role: Role }[];
  chartProperties?: Record<string, unknown>;
  /** The rendered SVG as a data URI, ready to drop into an img src. */
  thumbnail: string;
  defaultLimit: number;
}

const registeredSemanticTypes = new Set(Object.keys(SemanticTypes));

/**
 * Maps an example field onto a Rill param type.
 *
 * Flint's registry is the source of truth where it knows the type, but its example set uses several
 * domain-flavoured names (GDP, Team, Value, ...) that are not registered and would otherwise all
 * degrade to categorical. For those, the field's own data type decides.
 */
function roleFor(semanticType: string, dataType: string): Role | null {
  if (UNSUPPORTED_SEMANTIC_TYPES.has(semanticType)) return null;

  if (!registeredSemanticTypes.has(semanticType)) {
    return dataType === "number" ? "measure" : "dimension";
  }

  const entry = getRegistryEntry(semanticType) as
    | { t0?: string; aggRole?: string }
    | undefined;
  if (!entry) return dataType === "number" ? "measure" : "dimension";

  // A temporal field is an axis unless it is a quantity of time (Duration is additive).
  if (entry.t0 === "Temporal") {
    return entry.aggRole === "dimension" ? "time_dimension" : "measure";
  }
  if (entry.t0 === "Measure") return "measure";
  // Score is intensive and numeric; Rank is an ordinal label.
  if (entry.aggRole === "intensive" || entry.aggRole === "additive") {
    return "measure";
  }
  return "dimension";
}

function categoryOf(chart: string): string {
  for (const [category, defs] of Object.entries(
    vlTemplateDefs as unknown as Record<string, { chart: string }[]>,
  )) {
    if (defs.some((def) => def.chart === chart)) return category;
  }
  return "Other";
}

function defaultLimit(chartType: string): number {
  if (TIME_SERIES_CHARTS.has(chartType)) return 2000;
  if (DENSE_CHARTS.has(chartType)) return 100;
  return 10;
}

function slugify(value: string): string {
  return value
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "_")
    .replace(/^_|_$/g, "");
}

/**
 * Ranks an example by how well it works as a starter. Flint's curated gallery examples come first,
 * then the ones built on named real-world datasets (Palmer Penguins, the Keeling Curve), then
 * anything else it pins to a gallery, then the remaining synthetic shapes.
 */
function starterRank(example: FlintExample, fromGalleryGenerator: boolean): number {
  const tags = example.tags ?? [];
  if (fromGalleryGenerator) return 0;
  if (tags.includes("real")) return 1;
  if (tags.includes("gallery") || tags.includes("gallery-pin")) return 2;
  return SHORTHAND_TITLE.test(example.title) ? 4 : 3;
}

/**
 * Picks the examples to ship for one chart type: drops the adversarial cases and the ones encoding a
 * channel this type does not declare, dedupes examples registered under more than one generator key,
 * then keeps the best-ranked few.
 */
function pickExamples(
  candidates: { example: FlintExample; fromGalleryGenerator: boolean }[],
  declaredChannels: string[],
): { example: FlintExample; fromGalleryGenerator: boolean }[] {
  const seenTitles = new Set<string>();
  return candidates
    .filter(
      ({ example }) =>
        !(example.tags ?? []).some((tag) => ADVERSARIAL_TAGS.has(tag)),
    )
    // Some examples in the cross-backend suites encode a channel this chart type does not declare
    // (a grouped bar with both color and group). Flint's renderer drops it, so binding it would
    // declare a param that never reaches the chart. Filtering here lets a valid example take the
    // slot instead of leaving the chart type one card short.
    .filter(({ example }) =>
      Object.keys(example.encodingMap).every((channel) =>
        declaredChannels.includes(channel),
      ),
    )
    .sort(
      (a, b) =>
        starterRank(a.example, a.fromGalleryGenerator) -
        starterRank(b.example, b.fromGalleryGenerator),
    )
    .filter(({ example }) => {
      if (seenTitles.has(example.title)) return false;
      seenTitles.add(example.title);
      return true;
    })
    .slice(0, MAX_PER_CHART_TYPE);
}

/** Compiles an example to Vega-Lite and rasterizes it to a standalone SVG string. */
async function renderThumbnail(
  example: FlintExample,
  encodings: Record<string, { field: string }>,
): Promise<string> {
  const semantic_types = Object.fromEntries(
    Object.entries(example.metadata).map(([field, meta]) => [
      field,
      meta.semanticType,
    ]),
  );

  const spec = assembleVegaLite({
    data: { values: example.data },
    semantic_types,
    chart_spec: {
      chartType: example.chartType,
      encodings,
      baseSize: THUMBNAIL_SIZE,
      canvasSize: THUMBNAIL_SIZE,
      ...(example.chartProperties
        ? { chartProperties: example.chartProperties }
        : {}),
    },
  }) as Record<string, unknown>;

  // Flint's bookkeeping keys are not valid Vega-Lite.
  for (const key of Object.keys(spec)) {
    if (key.startsWith("_")) delete spec[key];
  }

  // A thumbnail reads as a shape, not a chart you can look values up in. Suppressing tick labels,
  // titles and legends keeps the mark geometry legible at 240x150 and drops the text elements that
  // otherwise dominate the SVG.
  spec.config = {
    ...((spec.config as Record<string, unknown>) ?? {}),
    axis: { labels: false, title: null, ticks: false, domainColor: "#d0d5dd" },
    legend: { disable: true },
    // Faceted examples (violin grids, small multiples) label each panel separately.
    header: { labels: false, title: null },
  };
  // Flint sets some axis titles on the encoding itself, which wins over config.
  stripEncodingTitles(spec);

  const view = new vega.View(vega.parse(compile(spec as never).spec), {
    renderer: "none",
  });
  return minifySvg(await view.toSVG());
}

/** Nulls out per-encoding axis titles so a thumbnail carries no text of its own. */
function stripEncodingTitles(node: unknown): void {
  if (Array.isArray(node)) {
    node.forEach(stripEncodingTitles);
    return;
  }
  if (!node || typeof node !== "object") return;
  const record = node as Record<string, unknown>;
  if (typeof record.field === "string" && "title" in record) {
    record.title = null;
  }
  Object.values(record).forEach(stripEncodingTitles);
}

/**
 * Vega emits full float precision, which dominates the size of a dense thumbnail: a 500-point
 * scatter is mostly digits past the decimal point that no one can see at 240x150. Rounding to two
 * places roughly halves the payload and is visually lossless at this scale.
 */
function minifySvg(svg: string): string {
  return svg.replace(/-?\d+\.\d{3,}/g, (match) =>
    String(Math.round(Number(match) * 100) / 100),
  );
}

/**
 * Wraps an SVG as a data URI for an img src.
 *
 * Percent-encoding only the characters that would terminate the URI or the surrounding attribute
 * keeps the markup readable and roughly a third smaller than base64, which inflates by 4/3 and
 * defeats gzip on text this repetitive.
 */
function svgDataURI(svg: string): string {
  const escaped = svg
    .replace(/%/g, "%25")
    .replace(/#/g, "%23")
    .replace(/"/g, "%22")
    .replace(/&/g, "%26")
    .replace(/</g, "%3C")
    .replace(/>/g, "%3E")
    // A raw newline is legal in an attribute but not in a URI.
    .replace(/\s*\n\s*/g, " ");
  return `data:image/svg+xml,${escaped}`;
}

async function main() {
  // flint-chart does not export ./package.json, so resolve it via a subpath we can import.
  const { version: flintVersion } = JSON.parse(
    readFileSync(
      join(require.resolve("flint-chart/vegalite"), "../../../package.json"),
      "utf8",
    ),
  ) as { version: string };

  // Drop the sibling SVG directory earlier revisions emitted; the thumbnails are inlined now.
  rmSync(join(EXAMPLES_DIR, "thumbnails"), { recursive: true, force: true });

  const chartTypes = new Set(
    (vlAllTemplateDefs as { chart: string }[]).map((def) => def.chart),
  );

  const examples: GalleryExample[] = [];
  const skipped: string[] = [];
  const usedNames = new Set<string>();
  let thumbnailBytes = 0;

  // Group every generated example by its own chart type rather than by generator key. Flint's
  // curated examples live under "Gallery: Bar"-style keys that are not chart type names, so keying
  // off the generator would drop exactly the ones a picker wants.
  const candidates = new Map<
    string,
    { example: FlintExample; fromGalleryGenerator: boolean }[]
  >();

  for (const [key, generate] of Object.entries(TEST_GENERATORS)) {
    let generated: FlintExample[];
    try {
      generated = (generate as () => FlintExample[])();
    } catch (error) {
      skipped.push(`${key} — generator threw: ${String(error)}`);
      continue;
    }

    for (const example of generated) {
      // Other backends and internal layout suites.
      if (!chartTypes.has(example.chartType)) continue;
      if (EXCLUDED_CHART_TYPES.has(example.chartType)) continue;

      const list = candidates.get(example.chartType) ?? [];
      list.push({ example, fromGalleryGenerator: key.startsWith("Gallery:") });
      candidates.set(example.chartType, list);
    }
  }

  for (const chartType of EXCLUDED_CHART_TYPES) {
    skipped.push(`${chartType} — excluded: needs semantics Rill never derives`);
  }

  for (const [chartType, forChartType] of candidates) {
    const category = categoryOf(chartType);
    const { channels: declaredChannels } = vlGetTemplateDef(chartType) as {
      channels: string[];
    };

    for (const { example } of pickExamples(forChartType, declaredChannels)) {
      // Resolve every encoded channel to a Rill param role; skip the example if any field
      // uses semantics Rill cannot produce, rather than shipping a dead-end starter.
      const bindings: GalleryExample["bindings"] = [];
      const encodings: Record<string, { field: string }> = {};
      let usable = true;

      for (const [channel, encoding] of Object.entries(example.encodingMap)) {
        // A channel bound to several fields (static-series fold) has no single param.
        if (Array.isArray(encoding)) {
          usable = false;
          break;
        }
        const field = encoding.fieldID;
        const meta = example.metadata[field];
        if (!meta) {
          usable = false;
          break;
        }
        const role = roleFor(meta.semanticType, meta.type);
        if (!role) {
          usable = false;
          break;
        }
        bindings.push({
          channel,
          param: CHANNEL_PARAM_NAMES[channel] ?? channel,
          role,
        });
        encodings[channel] = { field };
      }

      if (!usable || bindings.length === 0) continue;

      let svg: string;
      try {
        svg = await renderThumbnail(example, encodings);
      } catch (error) {
        skipped.push(
          `${chartType} / ${example.title} — render failed: ${String(error)}`,
        );
        continue;
      }
      if (svg.length > MAX_THUMBNAIL_BYTES) {
        skipped.push(
          `${chartType} / ${example.title} — thumbnail too large (${svg.length} bytes)`,
        );
        continue;
      }

      let name = `${slugify(chartType)}__${slugify(example.title)}`;
      for (let i = 2; usedNames.has(name); i++) {
        name = `${slugify(chartType)}__${slugify(example.title)}_${i}`;
      }
      usedNames.add(name);

      thumbnailBytes += svg.length;

      examples.push({
        name,
        chartSlug: slugify(chartType),
        chartType,
        category,
        title: example.title,
        description: example.description ?? "",
        tags: example.tags ?? [],
        bindings,
        ...(example.chartProperties
          ? { chartProperties: example.chartProperties }
          : {}),
        thumbnail: svgDataURI(svg),
        defaultLimit: defaultLimit(chartType),
      });
    }
  }

  writeFileSync(
    OUT_PATH,
    `${JSON.stringify({ license: LICENSE_NOTICE, source: SOURCE_URL, flintVersion, examples }, null, 2)}\n`,
  );

  const byChartType = new Set(examples.map((e) => e.chartType));
  console.log(
    `Wrote ${examples.length} examples across ${byChartType.size} chart types to ${OUT_PATH}`,
  );
  console.log(
    `Thumbnails: ${examples.length} SVGs inlined, ${(thumbnailBytes / 1024).toFixed(0)} KB total`,
  );
  if (skipped.length) {
    console.log(`Skipped ${skipped.length}:`);
    for (const entry of skipped.slice(0, 20)) console.log(`  - ${entry}`);
    if (skipped.length > 20) console.log(`  ... and ${skipped.length - 20} more`);
  }
}

await main();
