/**
 * Snapshots a curated subset of the Vega-Lite example gallery into
 * src/features/custom-viz/examples/.
 *
 * The snapshot is vendored (committed) so the "From Vega-Lite example" gallery works
 * offline and without CSP exceptions. Each example's external data (data/*.json|csv|tsv)
 * is fetched, sampled to at most MAX_ROWS rows, and inlined as data.values, so every
 * gallery entry renders self-contained.
 *
 * Output is split so the gallery grid stays cheap to open:
 *   vega-examples-index.json  metadata for every example (loaded when the dialog opens)
 *   thumbnails/<name>.png     the official preview image from the Vega editor repo
 *   specs/<name>.json         one spec per example, with its inlined sample data
 *                             (code-split by the bundler, fetched on selection)
 *
 * Thumbnails are the images the Vega editor itself ships, downloaded as-is rather
 * than rendered here: they are the canonical look of each example, they cover every
 * entry (including dense ones a headless render struggles with), and reusing them
 * keeps this script to fetching and writing files.
 *
 * Run manually when refreshing the gallery:
 *   npx tsx web-common/scripts/snapshot-vega-examples.ts
 *
 * Fetched specs, datasets, and images are cached under node_modules/.cache/vega-examples
 * so re-runs after a config tweak cost no network. Delete that directory to force a
 * fresh fetch.
 *
 * The Vega-Lite examples and the editor's preview images are both BSD-3-Clause
 * licensed (University of Washington Interactive Data Lab); see the index's license.
 */
import { csvParse, tsvParse, autoType } from "d3-dsv";
import { createHash } from "node:crypto";
import {
  existsSync,
  mkdirSync,
  readFileSync,
  rmSync,
  writeFileSync,
} from "node:fs";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";
import { parameterizeExampleSpec } from "../src/features/custom-viz/examples/example-spec";

const INDEX_URL =
  "https://raw.githubusercontent.com/vega/vega-lite/main/site/_data/examples.json";
const EXAMPLES_BASE = "https://vega.github.io/vega-lite/examples/";
// The Vega editor's own preview images, named <example>.vl.png.
const THUMBNAIL_BASE =
  "https://raw.githubusercontent.com/vega/editor/master/public/images/examples/vl/";
const MAX_ROWS = 200;
// Per subcategory rather than a global cap: a global cap is consumed by whichever
// subcategories the upstream index happens to list first, which silently starved
// whole categories of the gallery.
const MAX_PER_SUBCATEGORY = 12;
// Parallel fetches. Kept modest to stay friendly to the upstream static hosts.
const FETCH_CONCURRENCY = 6;
// A drop below this share of examples having a thumbnail means the upstream image
// set moved (renamed, relocated); fail rather than silently ship icon placeholders.
const MIN_THUMBNAIL_RATE = 0.95;

// Categories to include, in gallery order. Composite Marks and Multi-View Displays
// are included but rarely parameterize (composite specs fall back to a verbatim
// sample-data import); Interactive Charts are excluded because their selections do
// not survive rebinding params to a different metrics view.
const INCLUDED_CATEGORIES = new Set([
  "Single-View Plots",
  "Composite Marks",
  "Multi-View Displays",
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

interface GalleryEntry {
  name: string;
  title: string;
  description?: string;
  category: string;
  subcategory: string;
  // File name of the preview image under thumbnails/, or undefined when the
  // upstream editor repo has no image for this example.
  thumbnail?: string;
  // Whether the deterministic import parameterizes this example (vs. falling
  // back to a verbatim sample-data import). Precomputed so the gallery grid can
  // show it without loading every spec.
  parameterizes: boolean;
}

const CACHE_DIR = join(
  dirname(fileURLToPath(import.meta.url)),
  "../../node_modules/.cache/vega-examples",
);

/** Fetches a URL as text, memoized on disk so re-runs cost no network. */
async function fetchTextCached(url: string): Promise<string> {
  const cachePath = join(CACHE_DIR, createHash("sha256").update(url).digest("hex"));
  if (existsSync(cachePath)) return readFileSync(cachePath, "utf8");

  const res = await fetch(url);
  if (!res.ok) throw new Error(`GET ${url}: ${res.status}`);
  const text = await res.text();
  mkdirSync(CACHE_DIR, { recursive: true });
  writeFileSync(cachePath, text);
  return text;
}

async function fetchJSONCached(url: string): Promise<unknown> {
  return JSON.parse(await fetchTextCached(url));
}

/** Fetches a URL as bytes, memoized on disk. Used for the preview images. */
async function fetchBinaryCached(url: string): Promise<Buffer> {
  const cachePath = join(
    CACHE_DIR,
    createHash("sha256").update(url).digest("hex"),
  );
  if (existsSync(cachePath)) return readFileSync(cachePath);

  const res = await fetch(url);
  if (!res.ok) throw new Error(`GET ${url}: ${res.status}`);
  const bytes = Buffer.from(await res.arrayBuffer());
  mkdirSync(CACHE_DIR, { recursive: true });
  writeFileSync(cachePath, bytes);
  return bytes;
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
    const data = await fetchJSONCached(absolute);
    return Array.isArray(data) ? sample(data) : null;
  }
  if (url.endsWith(".csv")) {
    return sample(csvParse(await fetchTextCached(absolute), autoType));
  }
  if (url.endsWith(".tsv")) {
    return sample(tsvParse(await fetchTextCached(absolute), autoType));
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
      // Parse-relevant format hints (e.g. {type: "json", property: ...}) are not
      // needed once values are inlined; drop the format except explicit parses.
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
 * Downloads the Vega editor's preview image for an example, returning the file
 * name to write under thumbnails/. Returns undefined (and records why) when the
 * upstream repo has no image for this example.
 */
async function fetchThumbnail(
  name: string,
  skipped: string[],
): Promise<{ fileName: string; bytes: Buffer } | undefined> {
  const fileName = `${name}.png`;
  try {
    return {
      fileName,
      bytes: await fetchBinaryCached(`${THUMBNAIL_BASE}${name}.vl.png`),
    };
  } catch (error) {
    skipped.push(`${name} thumbnail (${(error as Error).message})`);
    return undefined;
  }
}

/** Runs tasks with bounded concurrency, preserving input order in the results. */
async function mapWithConcurrency<T, R>(
  items: T[],
  limit: number,
  task: (item: T) => Promise<R>,
): Promise<R[]> {
  const results = new Array<R>(items.length);
  let next = 0;
  await Promise.all(
    Array.from({ length: Math.min(limit, items.length) }, async () => {
      for (let i = next++; i < items.length; i = next++) {
        results[i] = await task(items[i]);
      }
    }),
  );
  return results;
}

interface Candidate {
  entry: IndexEntry;
  category: string;
  subcategory: string;
}

interface BuiltExample {
  entry: GalleryEntry;
  spec: Record<string, unknown>;
  thumbnail?: { fileName: string; bytes: Buffer };
}

async function buildExample(
  candidate: Candidate,
  skipped: string[],
): Promise<BuiltExample | null> {
  const { entry, category, subcategory } = candidate;
  try {
    const rawSpec = await fetchTextCached(
      `${EXAMPLES_BASE}${entry.name}.vl.json`,
    );
    if (
      EXCLUDED_MARKERS.some((marker) =>
        rawSpec.toLowerCase().includes(marker.toLowerCase()),
      )
    ) {
      skipped.push(`${entry.name} (excluded feature)`);
      return null;
    }
    const spec = JSON.parse(rawSpec) as Record<string, unknown>;
    delete spec.description;
    if (!(await inlineData(spec))) {
      skipped.push(`${entry.name} (unsupported data format)`);
      return null;
    }

    // parameterizeExampleSpec mutates nothing, but it is the same function the
    // app runs at import time, so the flag can never drift from actual behavior.
    const parameterizes = parameterizeExampleSpec(spec) !== null;
    const thumbnail = await fetchThumbnail(entry.name, skipped);

    return {
      entry: {
        name: entry.name,
        title: entry.title ?? entry.name,
        description: entry.description,
        category,
        subcategory,
        thumbnail: thumbnail?.fileName,
        parameterizes,
      },
      spec,
      thumbnail,
    };
  } catch (error) {
    skipped.push(`${entry.name} (${(error as Error).message})`);
    return null;
  }
}

async function main() {
  const index = (await fetchJSONCached(INDEX_URL)) as Record<
    string,
    Record<string, IndexEntry[]>
  >;

  const skipped: string[] = [];
  // The upstream index lists some examples under multiple subcategories;
  // keep only the first occurrence so names stay unique (they key UI lists).
  const seen = new Set<string>();
  const candidates: Candidate[] = [];

  for (const [category, subcategories] of Object.entries(index)) {
    if (!INCLUDED_CATEGORIES.has(category)) continue;
    for (const [subcategory, entries] of Object.entries(subcategories)) {
      let taken = 0;
      for (const entry of entries) {
        if (taken >= MAX_PER_SUBCATEGORY) {
          skipped.push(`${entry.name} (over ${subcategory} cap)`);
          continue;
        }
        if (seen.has(entry.name)) {
          skipped.push(`${entry.name} (duplicate listing)`);
          continue;
        }
        seen.add(entry.name);
        candidates.push({ entry, category, subcategory });
        taken++;
      }
    }
  }

  const built = await mapWithConcurrency(
    candidates,
    FETCH_CONCURRENCY,
    (candidate) => buildExample(candidate, skipped),
  );
  const examples = built.filter(
    (result): result is BuiltExample => result !== null,
  );

  const withThumbnail = examples.filter((e) => e.thumbnail).length;
  const thumbnailRate = examples.length ? withThumbnail / examples.length : 0;

  const outDir = join(
    dirname(fileURLToPath(import.meta.url)),
    "../src/features/custom-viz/examples",
  );
  const specsDir = join(outDir, "specs");
  const thumbnailsDir = join(outDir, "thumbnails");
  // Rewrite both directories from scratch so examples dropped by a refresh don't
  // linger as orphaned files.
  rmSync(specsDir, { recursive: true, force: true });
  rmSync(thumbnailsDir, { recursive: true, force: true });
  mkdirSync(specsDir, { recursive: true });
  mkdirSync(thumbnailsDir, { recursive: true });

  for (const { entry, spec, thumbnail } of examples) {
    writeFileSync(join(specsDir, `${entry.name}.json`), JSON.stringify(spec));
    if (thumbnail) {
      writeFileSync(join(thumbnailsDir, thumbnail.fileName), thumbnail.bytes);
    }
  }

  writeFileSync(
    join(outDir, "vega-examples-index.json"),
    JSON.stringify(
      {
        license:
          "Examples from the Vega-Lite project (https://vega.github.io/vega-lite/examples/) and preview images from the Vega editor (https://github.com/vega/editor), BSD-3-Clause, University of Washington Interactive Data Lab.",
        snapshotSource: INDEX_URL,
        thumbnailSource: THUMBNAIL_BASE,
        examples: examples.map(({ entry }) => entry),
      },
      null,
      2,
    ) + "\n",
  );

  const parameterized = examples.filter((e) => e.entry.parameterizes).length;
  const thumbnailBytes = examples.reduce(
    (sum, e) => sum + (e.thumbnail?.bytes.length ?? 0),
    0,
  );
  console.log(`Wrote ${examples.length} examples to ${outDir}`);
  console.log(
    `  ${parameterized} parameterize, ${examples.length - parameterized} import verbatim`,
  );
  console.log(
    `  ${withThumbnail} thumbnails (${(thumbnailBytes / 1024 / 1024).toFixed(2)} MB)`,
  );
  console.log(`Skipped ${skipped.length}:`);
  for (const reason of skipped) console.log(`  - ${reason}`);

  if (thumbnailRate < MIN_THUMBNAIL_RATE) {
    throw new Error(
      `only ${(thumbnailRate * 100).toFixed(0)}% of examples have a thumbnail ` +
        `(expected at least ${MIN_THUMBNAIL_RATE * 100}%) — the upstream image set may have moved`,
    );
  }
}

void main();
