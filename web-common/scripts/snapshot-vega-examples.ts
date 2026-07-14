/**
 * Snapshots a curated subset of the Vega-Lite example gallery into
 * src/features/custom-viz/examples/vega-examples.json.
 *
 * The snapshot is vendored (committed) so the "From Vega-Lite example" gallery works
 * offline and without CSP exceptions. Each example's external data (data/*.json|csv|tsv)
 * is fetched, sampled to at most MAX_ROWS rows, and inlined as data.values, so every
 * gallery entry renders self-contained.
 *
 * Run manually when refreshing the gallery:
 *   npx tsx web-common/scripts/snapshot-vega-examples.ts
 *
 * The Vega-Lite examples are BSD-3-Clause licensed (University of Washington
 * Interactive Data Lab); see the license field in the output file.
 */
import { csvParse, tsvParse, autoType } from "d3-dsv";
import { mkdirSync, writeFileSync } from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";
import * as vega from "vega";
import * as vegaLite from "vega-lite";

const INDEX_URL =
  "https://raw.githubusercontent.com/vega/vega-lite/main/site/_data/examples.json";
const EXAMPLES_BASE = "https://vega.github.io/vega-lite/examples/";
const MAX_ROWS = 200;
const MAX_EXAMPLES = 90;

// Categories to include, in gallery order.
const INCLUDED_CATEGORIES = new Set([
  "Single-View Plots",
  "Composite Marks",
  "Multi-View Displays",
  "Interactive Charts",
]);

// Examples containing these markers depend on features the import flow can't
// handle self-contained (geo projections, images, remote lookups at runtime).
const EXCLUDED_MARKERS = [
  '"projection"',
  '"geoshape"',
  "topojson",
  '"graticule"',
  '"sphere"',
  '"image"',
  ".png",
  ".svg",
];

interface IndexEntry {
  name: string;
  title?: string;
  description?: string;
}

interface GalleryExample {
  name: string;
  title: string;
  description?: string;
  category: string;
  subcategory: string;
  spec: string;
  // Static SVG thumbnail, pre-rendered at snapshot time so the gallery never
  // runs a chart engine at runtime. Width/height attrs are stripped so it
  // scales to its container via the viewBox.
  thumbnail?: string;
}

async function fetchJSON(url: string): Promise<unknown> {
  const res = await fetch(url);
  if (!res.ok) throw new Error(`GET ${url}: ${res.status}`);
  return res.json();
}

async function fetchText(url: string): Promise<string> {
  const res = await fetch(url);
  if (!res.ok) throw new Error(`GET ${url}: ${res.status}`);
  return res.text();
}

function sample(rows: unknown[]): unknown[] {
  if (rows.length <= MAX_ROWS) return rows;
  // Deterministic downsample: keep every nth row.
  const step = rows.length / MAX_ROWS;
  const out: unknown[] = [];
  for (let i = 0; i < MAX_ROWS; i++) {
    out.push(rows[Math.floor(i * step)]);
  }
  return out;
}

async function loadDataset(url: string): Promise<unknown[] | null> {
  const absolute = new URL(url, EXAMPLES_BASE).toString();
  if (url.endsWith(".json")) {
    const data = await fetchJSON(absolute);
    return Array.isArray(data) ? sample(data) : null;
  }
  if (url.endsWith(".csv")) {
    return sample(csvParse(await fetchText(absolute), autoType));
  }
  if (url.endsWith(".tsv")) {
    return sample(tsvParse(await fetchText(absolute), autoType));
  }
  return null;
}

/**
 * Walks the spec and inlines every `data: {url}` as `data: {values}`.
 * Returns false if any dataset can't be inlined (unsupported format).
 */
async function inlineData(node: unknown): Promise<boolean> {
  if (Array.isArray(node)) {
    for (const child of node) {
      if (!(await inlineData(child))) return false;
    }
    return true;
  }
  if (!node || typeof node !== "object") return true;

  const obj = node as Record<string, unknown>;
  const data = obj.data as Record<string, unknown> | undefined;
  if (data && typeof data.url === "string") {
    const values = await loadDataset(data.url);
    if (!values) return false;
    const format = data.format as Record<string, unknown> | undefined;
    obj.data = {
      values,
      // Preserve parse-relevant format hints (e.g. {type: "json", property: ...})
      // are not needed once values are inlined; drop the format entirely except
      // explicit field parses.
      ...(format?.parse ? { format: { parse: format.parse } } : {}),
    };
  }

  for (const [key, value] of Object.entries(obj)) {
    if (key === "data") continue;
    if (!(await inlineData(value))) return false;
  }
  return true;
}

/**
 * Renders a static SVG thumbnail for the example, headless. Returns undefined
 * (and records why) when the spec can't render outside a browser.
 */
async function renderThumbnail(
  spec: Record<string, unknown>,
  name: string,
  skipped: string[],
): Promise<string | undefined> {
  try {
    const compiled = vegaLite.compile(
      spec as unknown as vegaLite.TopLevelSpec,
    ).spec;
    const view = new vega.View(vega.parse(compiled), { renderer: "none" });
    const svg = await view.toSVG();
    view.finalize();
    // Strip fixed dimensions so the SVG scales to its container via the viewBox.
    return svg.replace(
      /^<svg([^>]*?)\swidth="[^"]*"\sheight="[^"]*"/,
      '<svg$1 preserveAspectRatio="xMidYMid meet"',
    );
  } catch (error) {
    skipped.push(`${name} thumbnail (${(error as Error).message})`);
    return undefined;
  }
}

async function main() {
  const index = (await fetchJSON(INDEX_URL)) as Record<
    string,
    Record<string, IndexEntry[]>
  >;

  const examples: GalleryExample[] = [];
  const skipped: string[] = [];
  // The upstream index lists some examples under multiple subcategories;
  // keep only the first occurrence so names stay unique (they key UI lists).
  const seen = new Set<string>();

  for (const [category, subcategories] of Object.entries(index)) {
    if (!INCLUDED_CATEGORIES.has(category)) continue;
    for (const [subcategory, entries] of Object.entries(subcategories)) {
      for (const entry of entries) {
        if (examples.length >= MAX_EXAMPLES) break;
        if (seen.has(entry.name)) {
          skipped.push(`${entry.name} (duplicate listing)`);
          continue;
        }
        seen.add(entry.name);
        try {
          const rawSpec = await fetchText(
            `${EXAMPLES_BASE}${entry.name}.vl.json`,
          );
          if (
            EXCLUDED_MARKERS.some((marker) =>
              rawSpec.toLowerCase().includes(marker.toLowerCase()),
            )
          ) {
            skipped.push(`${entry.name} (excluded feature)`);
            continue;
          }
          const spec = JSON.parse(rawSpec) as Record<string, unknown>;
          delete spec.description;
          if (!(await inlineData(spec))) {
            skipped.push(`${entry.name} (unsupported data format)`);
            continue;
          }
          examples.push({
            name: entry.name,
            title: entry.title ?? entry.name,
            description: entry.description,
            category,
            subcategory,
            spec: JSON.stringify(spec),
            thumbnail: await renderThumbnail(spec, entry.name, skipped),
          });
        } catch (error) {
          skipped.push(`${entry.name} (${(error as Error).message})`);
        }
      }
    }
  }

  const out = {
    license:
      "Examples from the Vega-Lite project (https://vega.github.io/vega-lite/examples/), BSD-3-Clause, University of Washington Interactive Data Lab.",
    snapshotSource: INDEX_URL,
    examples,
  };

  const outPath = join(
    dirname(fileURLToPath(import.meta.url)),
    "../src/features/custom-viz/examples/vega-examples.json",
  );
  mkdirSync(dirname(outPath), { recursive: true });
  writeFileSync(outPath, JSON.stringify(out, null, 2) + "\n");

  console.log(`Wrote ${examples.length} examples to ${outPath}`);
  console.log(`Skipped ${skipped.length}:`);
  for (const reason of skipped) console.log(`  - ${reason}`);
}

void main();
