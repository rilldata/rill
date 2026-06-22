#!/usr/bin/env node
// Heuristic guard against hardcoded user-facing strings in already-migrated
// areas of the frontend. It scans `.svelte` markup (not <script>/<style>) for
// visible text nodes and a fixed set of human-facing attributes.
//
// This is intentionally a lightweight, dependency-free heuristic, not a parser:
// it runs at WARNING level (exit 0) by default so occasional false positives
// are tolerable. Pass `--strict` to make findings fatal (exit 1) — the final
// i18n migration chunk flips the quality pipeline to strict once every listed
// area is clean.
//
// Suppress a specific line with an `i18n-ignore` comment on it or the line above.
//
// Usage: node scripts/i18n-guard.js [--strict]

import { globSync, readFileSync } from "node:fs";
import { relative } from "node:path";

// Directories whose strings have been migrated to paraglide. Each migration
// chunk appends its directories here; the guard only polices these areas so
// unmigrated code does not drown the signal.
const MIGRATED_GLOBS = [
  "web-common/src/layout/**/*.svelte",
  "web-common/src/features/welcome/**/*.svelte",
  "web-common/src/features/onboarding/**/*.svelte",
  "web-common/src/features/help/**/*.svelte",
  // web-admin organization overview page (chunk A). Literal `[organization]`
  // brackets are a glob character class, so match the segment with a wildcard.
  "web-admin/src/routes/*organization*/+page.svelte",
  "web-admin/src/features/organizations/OrganizationHero.svelte",
  "web-admin/src/features/organizations/OrganizationTabs.svelte",
  "web-admin/src/features/projects/ProjectCards.svelte",
  "web-admin/src/features/projects/ProjectCard.svelte",
  "web-admin/src/features/projects/ProjectCardActions.svelte",
  // web-admin project overview page (chunk B).
  "web-admin/src/routes/*organization*/*project*/+page.svelte",
  "web-admin/src/routes/*organization*/*project*/+layout.svelte",
  "web-admin/src/features/projects/ProjectTabs.svelte",
  "web-admin/src/features/projects/ProjectBuilding.svelte",
  "web-admin/src/features/projects/RedeployProjectCTA.svelte",
  "web-admin/src/features/dashboards/listing/DashboardsTable.svelte",
  "web-admin/src/features/dashboards/listing/DashboardsTableCompositeCell.svelte",
  "web-admin/src/features/dashboards/listing/LastRefreshedDate.svelte",
  // explore dashboard route + errored state (chunk C).
  "web-admin/src/routes/*organization*/*project*/explore/*dashboard*/+page.svelte",
  "web-admin/src/routes/*organization*/*project*/explore/*dashboard*/+layout.svelte",
  "web-admin/src/features/dashboards/DashboardErrored.svelte",
  // web-common explores feature (chunk D).
  "web-common/src/features/explores/**/*.svelte",
  // web-common dashboard shell/chrome (chunk E).
  "web-common/src/features/dashboards/*.svelte",
  "web-common/src/features/dashboards/workspace/*.svelte",
  "web-common/src/features/dashboards/toolbars/*.svelte",
  "web-common/src/features/dashboards/tab-bar/*.svelte",
  "web-common/src/features/dashboards/listing/*.svelte",
  "web-common/src/features/dashboards/errors/*.svelte",
  // web-common dashboard filters + dimension search (chunk F).
  "web-common/src/features/dashboards/filters/**/*.svelte",
  "web-common/src/features/dashboards/dimension-search/**/*.svelte",
];

// Human-facing attributes worth translating. Attributes like `class`, `id`,
// `href`, `name`, etc. are deliberately excluded.
const TEXT_ATTRS = ["placeholder", "title", "aria-label", "alt", "label"];

const strict = process.argv.includes("--strict");

function stripBlocks(src) {
  // Replace <script> and <style> bodies with blank lines so line numbers and
  // offsets are preserved while their contents are ignored.
  return src.replace(
    /<(script|style)\b[^>]*>[\s\S]*?<\/\1>/gi,
    (m) => m.replace(/[^\n]/g, " "),
  );
}

function lineOf(src, index) {
  return src.slice(0, index).split("\n").length;
}

function isIgnored(src, line) {
  const lines = src.split("\n");
  const current = lines[line - 1] ?? "";
  const previous = lines[line - 2] ?? "";
  return current.includes("i18n-ignore") || previous.includes("i18n-ignore");
}

// A segment is human-facing text worth translating when, after removing
// mustache expressions, it still contains real words. We skip identifiers,
// URLs, paths, and symbol/number-only content to keep the signal high.
function looksLikeCopy(raw) {
  const text = raw.replace(/\{[^}]*\}/g, "").trim();
  if (!text) return false;
  if (!/[A-Za-z]{2,}/.test(text)) return false; // needs at least one real word
  if (/^[a-z0-9_\-./:]+$/.test(text)) return false; // identifier / url / path
  if (/^[A-Z0-9_]+$/.test(text)) return false; // CONSTANT_CASE
  return true;
}

const findings = [];

for (const pattern of MIGRATED_GLOBS) {
  for (const file of globSync(pattern)) {
    const original = readFileSync(file, "utf8");
    const src = stripBlocks(original);
    const rel = relative(process.cwd(), file);

    // Visible text nodes: content between a closing `>` and the next `<`.
    for (const match of src.matchAll(/>([^<>]+)</g)) {
      if (!looksLikeCopy(match[1])) continue;
      const line = lineOf(src, match.index);
      if (isIgnored(original, line)) continue;
      findings.push(`${rel}:${line}  text: ${match[1].trim()}`);
    }

    // Human-facing attributes with a static string value.
    const attrRe = new RegExp(`\\b(${TEXT_ATTRS.join("|")})="([^"]+)"`, "g");
    for (const match of src.matchAll(attrRe)) {
      if (!looksLikeCopy(match[2])) continue;
      const line = lineOf(src, match.index);
      if (isIgnored(original, line)) continue;
      findings.push(`${rel}:${line}  ${match[1]}: ${match[2].trim()}`);
    }
  }
}

if (findings.length === 0) {
  console.log("i18n-guard: no hardcoded strings found in migrated areas.");
  process.exit(0);
}

const label = strict ? "ERROR" : "WARNING";
console.log(
  `i18n-guard: ${findings.length} hardcoded string(s) in migrated areas (${label}):`,
);
for (const f of findings) console.log(`  ${f}`);
console.log(
  "\nWrap these in paraglide messages (see web-common/src/lib/i18n/README.md), " +
    "or add an `i18n-ignore` comment if intentional.",
);

process.exit(strict ? 1 : 0);
