#!/usr/bin/env node
// Checks links into the docs site that Docusaurus does not check itself.
//
// Docusaurus fails the build on a broken relative link, but it treats
// `https://docs.rilldata.com/...` URLs as external, and it never sees links
// that React components build at render time (for example the connector cards'
// "YAML Reference" buttons). This script resolves both against the built site:
//
//   1. It indexes every file under docs/dist as a route. Client-redirect pages
//      and the Netlify `_redirects` file are followed, up to MAX_REDIRECT_HOPS.
//   2. It collects every docs.rilldata.com URL in docs pages, blog posts, the
//      project schema and product sources, and fails if the page is missing
//      or the #anchor is not an id on the final page.
//   3. It checks every internal href="/..." in the built HTML the same way.
//   4. It warns about absolute self-links in docs prose, which the Docusaurus
//      link checker cannot see; prefer site-relative links there.
//
// URLs inside code blocks are checked too, so examples must use real pages.
// CI runs this in the docs workflow, which triggers on docs changes only;
// a product-only change is checked on the next docs pull request.
//
// Run it after `npm run build -w docs`:
//   node scripts/check-docs-links.js [--dist docs/dist]

import { existsSync, readdirSync, readFileSync, statSync } from "node:fs";
import { join, relative, sep } from "node:path";
import { fileURLToPath } from "node:url";

const REPO_ROOT = fileURLToPath(new URL("..", import.meta.url));

const DIST_DIR = (() => {
  const i = process.argv.indexOf("--dist");
  return i > 0 ? process.argv[i + 1] : join(REPO_ROOT, "docs/dist");
})();

const MAX_REDIRECT_HOPS = 5;

// Where absolute docs URLs are collected from.
// `prose: true` marks docs pages, where absolute self-links get a warning.
const SOURCE_ROOTS = [
  { dir: "docs/docs", exts: [".md", ".mdx"], prose: true },
  { dir: "docs/blog", exts: [".md", ".mdx"], prose: true },
  { dir: "docs/src", exts: [".js", ".jsx", ".ts", ".tsx"] },
  { dir: "runtime", exts: [".go", ".yaml", ".yml", ".md"] },
  { dir: "cli", exts: [".go", ".yaml", ".yml", ".md"] },
  { dir: "admin", exts: [".go", ".yaml", ".yml", ".html"] },
  { dir: "web-common/src", exts: [".ts", ".js", ".svelte", ".json"] },
  { dir: "web-admin/src", exts: [".ts", ".js", ".svelte", ".json"] },
  { dir: "web-local/src", exts: [".ts", ".js", ".svelte", ".json"] },
];

// Directories and files that are generated or only used by tests.
// Match by directory name, so never add a name that docs pages use (such as "build").
const SKIP_DIRS = new Set([
  "node_modules",
  ".svelte-kit",
  "dist",
  "gen",
  "testdata",
  "tests",
]);
const SKIP_FILE_PATTERN = /(_test\.go|\.spec\.ts|\.test\.ts|\.pb\.go|\.pb\.gw\.go)$/;

// Routes rendered client-side, whose anchors are not in the built HTML.
const SKIP_ANCHOR_PREFIXES = ["/api/"];

// Known broken links, each waiting on its own fix.
// Entries are the link as written (URL or site path, including any #anchor),
// or a bare "#anchor" to allow it on any page.
// Remove an entry together with its fix; entries that no longer match print a warning.
const KNOWN_BROKEN = new Set([
  // Time range descriptions in project.schema.yaml and customization.md;
  // rill-iso-extensions has no #extensions heading.
  "https://docs.rilldata.com/reference/rill-iso-extensions#extensions",
  "/reference/time-syntax/rill-iso-extensions#extensions",
  // security.md links to a removed heading.
  "#testing-policies-in-rill-developer",
  // Docs pages linking to moved pages or removed headings.
  "https://docs.rilldata.com/guide/alerts/alerts/slack",
  "https://docs.rilldata.com/developers/deploy/existing-project",
  "https://docs.rilldata.com/quickstart",
  "https://docs.rilldata.com/developers/build/olap/",
  "https://docs.rilldata.com/developers/build/connectors/source/",
  "https://docs.rilldata.com/reference/project-files/metrics_views",
  "https://docs.rilldata.com/guide/dashboard-101",
  "https://docs.rilldata.com/reference/project-files/explores",
  "https://docs.rilldata.com/developers/build/connectors/olap/clickhouse#connecting-to-clickhouse-cloud",
  // The rill.yaml schema, and the rill.yaml template written by `rill init`.
  "https://docs.rilldata.com/developers/build/rill-project-file#dashboard-defaults",
  "https://docs.rilldata.com/developers/build/rill-project-file#test-access-policies-in-rill-developer",
  // Links emitted by the product.
  "https://docs.rilldata.com/developers/build/connect/#adding-a-remote-source",
  "https://docs.rilldata.com/reference/project-files/dashboards",
  "https://docs.rilldata.com/guides/alerts#configuring-slack-targets",
]);

const DOCS_URL_PATTERN = /https?:\/\/docs\.rilldata\.com(\/[^\s"'`<>)\]\\|]*)?/g;
const TEMPLATE_PATTERN = /\$\{|\{\{|%[sdv]/;

function main() {
  if (!existsSync(DIST_DIR)) {
    console.error(
      `ERROR: ${DIST_DIR} not found. Build the docs first: npm run build -w docs`,
    );
    process.exit(2);
  }

  const site = indexSite(DIST_DIR);
  const errors = [];
  const templates = [];
  const selfLinks = [];
  const usedKnown = new Set();

  const check = (link, where) => {
    if (KNOWN_BROKEN.has(link)) {
      usedKnown.add(link);
      return;
    }
    const problem = resolve(site, link);
    if (problem) errors.push(`${where}: ${link}\n    ${problem}`);
  };

  // Absolute docs URLs in sources.
  for (const root of SOURCE_ROOTS) {
    for (const file of walk(join(REPO_ROOT, root.dir), root.exts)) {
      const rel = relative(REPO_ROOT, file).split(sep).join("/");
      const lines = readFileSync(file, "utf8").split("\n");
      let inFence = false;
      lines.forEach((line, i) => {
        if (root.prose && /^\s*(```|~~~)/.test(line)) inFence = !inFence;
        // Commented-out markup is not rendered.
        if (/^\s*<!--/.test(line)) return;
        for (const match of line.matchAll(DOCS_URL_PATTERN)) {
          const url = match[0].replace(/[.,;:!?'*_]+$/, "");
          const where = `${rel}:${i + 1}`;
          if (TEMPLATE_PATTERN.test(url)) {
            templates.push(`${where}: ${line.trim().slice(0, 160)}`);
            continue;
          }
          check(url, where);
          if (root.prose && !inFence) selfLinks.push(`${where}: ${url}`);
        }
      });
    }
  }

  // Internal hrefs in the built HTML, including ones built by components.
  const hrefSources = new Map();
  for (const [route, page] of site.pages) {
    for (const match of page.html.matchAll(/\shref="(\/[^"]*|#[^"]+)"/g)) {
      let link = decodeEntities(match[1]);
      if (link.startsWith("//")) continue;
      if (link.startsWith("#")) link = route + link;
      if (!hrefSources.has(link)) hrefSources.set(link, []);
      hrefSources.get(link).push(route);
    }
  }
  for (const [link, routes] of hrefSources) {
    const hash = link.includes("#") ? link.slice(link.indexOf("#")) : "";
    if (KNOWN_BROKEN.has(hash)) {
      usedKnown.add(hash);
      continue;
    }
    const more = routes.length > 1 ? ` (and ${routes.length - 1} more)` : "";
    check(link, `built page ${routes[0]}${more}`);
  }

  if (templates.length > 0) {
    console.log(
      `\nDocs URLs built from templates (not checked; review by hand when pages move):`,
    );
    for (const t of templates) console.log(`  ${t}`);
  }

  if (selfLinks.length > 0) {
    console.log(
      `\nWarning: absolute docs.rilldata.com links in docs prose bypass the Docusaurus link check; prefer site-relative links:`,
    );
    for (const s of selfLinks) console.log(`  ${s}`);
  }

  const stale = [...KNOWN_BROKEN].filter((k) => !usedKnown.has(k));
  if (stale.length > 0) {
    console.log(
      `\nWarning: KNOWN_BROKEN entries no longer found; remove them from scripts/check-docs-links.js:`,
    );
    for (const s of stale) console.log(`  ${s}`);
  }

  if (errors.length > 0) {
    console.error(`\nBroken docs links (${errors.length}):`);
    for (const e of errors) console.error(`  ${e}`);
    console.error(
      `\nFix: point each link at an existing page and heading id, or add a redirect in docs/docusaurus.config.js when a page moves.`,
    );
    process.exit(1);
  }

  console.log(
    `\nDocs links OK (${site.pages.size} pages, ${site.redirects.size} redirects, ${hrefSources.size} distinct internal hrefs).`,
  );
}

// Builds the route index: pages (with their ids), redirects and static files.
function indexSite(distDir) {
  const pages = new Map();
  const redirects = new Map();
  const files = new Set();

  for (const file of walk(distDir)) {
    const rel = "/" + relative(distDir, file).split(sep).join("/");
    files.add(rel);
    if (!rel.endsWith(".html")) continue;

    const route = normalizePath(rel.replace(/(\/index)?\.html$/, "")) || "/";
    const html = readFileSync(file, "utf8");
    const refresh = html.match(
      /<meta http-equiv="refresh" content="0; url=([^"]+)"/,
    );
    if (refresh) {
      redirects.set(route, decodeEntities(refresh[1]));
      continue;
    }
    const ids = new Set();
    for (const m of html.matchAll(/\sid="([^"]+)"/g)) {
      ids.add(decodeEntities(m[1]));
    }
    pages.set(route, { html, ids });
  }

  // Server-side redirects served by Netlify.
  const netlify = join(distDir, "_redirects");
  if (existsSync(netlify)) {
    for (const line of readFileSync(netlify, "utf8").split("\n")) {
      const [from, to] = line.trim().split(/\s+/);
      if (!from || from.startsWith("#") || !to) continue;
      redirects.set(normalizePath(from), to);
    }
  }

  return { pages, redirects, files };
}

// Returns a description of the problem, or undefined if the link resolves.
function resolve(site, link) {
  let path = link.replace(/^https?:\/\/docs\.rilldata\.com/, "") || "/";
  let hash = "";
  const hashAt = path.indexOf("#");
  if (hashAt >= 0) {
    hash = safeDecode(path.slice(hashAt + 1));
    path = path.slice(0, hashAt);
  }
  path = normalizePath(safeDecode(path.split("?")[0])) || "/";

  for (let hop = 0; hop <= MAX_REDIRECT_HOPS; hop++) {
    const page = site.pages.get(path);
    if (page) {
      if (!hash || SKIP_ANCHOR_PREFIXES.some((p) => path.startsWith(p))) {
        return undefined;
      }
      return page.ids.has(hash)
        ? undefined
        : `anchor #${hash} not found on ${path}`;
    }
    if (site.files.has(path)) return undefined;

    const target = site.redirects.get(path);
    if (!target) return `page ${path} not found`;
    if (/^https?:\/\//.test(target) && !target.includes("docs.rilldata.com")) {
      return undefined;
    }
    path =
      normalizePath(target.replace(/^https?:\/\/docs\.rilldata\.com/, "")) ||
      "/";
  }
  return `more than ${MAX_REDIRECT_HOPS} redirects`;
}

function normalizePath(path) {
  return path.replace(/\/+$/, "");
}

function safeDecode(s) {
  try {
    return decodeURIComponent(s);
  } catch {
    return s;
  }
}

function decodeEntities(s) {
  return s
    .replace(/&amp;/g, "&")
    .replace(/&quot;/g, '"')
    .replace(/&#x27;|&#39;/g, "'")
    .replace(/&lt;/g, "<")
    .replace(/&gt;/g, ">");
}

// Yields every file under dir; with exts, only matching source files.
function* walk(dir, exts) {
  if (!existsSync(dir)) return;
  for (const entry of readdirSync(dir)) {
    const full = join(dir, entry);
    if (statSync(full).isDirectory()) {
      if (!exts || !SKIP_DIRS.has(entry)) yield* walk(full, exts);
    } else if (
      !exts ||
      (exts.some((e) => entry.endsWith(e)) && !SKIP_FILE_PATTERN.test(entry))
    ) {
      yield full;
    }
  }
}

main();
