#!/usr/bin/env node
// i18n guard with two independent checks:
//
// 1. Catalog integrity (exact, always fatal). Scans every locale file under
//    `web-common/src/lib/i18n/messages/` and reports duplicate keys, keys
//    missing from any locale (against the union of keys across all locales),
//    empty message texts, and parameter mismatches (each locale is checked
//    against the superset of `input` parameters seen for that key). Messages
//    may be plain strings or variant arrays (declarations/selectors/match, as
//    compiled by paraglide); variant messages are also checked for internal
//    consistency: selectors, match keys, and placeholders must resolve to
//    declared inputs or locals.
//
// 2. Hardcoded-string heuristic (warning by default). Scans `.svelte` markup
//    (not <script>/<style>) in already-migrated areas for visible text nodes
//    and a fixed set of human-facing attributes. Intentionally a lightweight,
//    dependency-free heuristic, not a parser: it runs at WARNING level
//    (exit 0) by default so occasional false positives are tolerable. Pass
//    `--strict` to make findings fatal (exit 1) — the final i18n migration
//    chunk flips the quality pipeline to strict once every listed area is
//    clean. Suppress a specific line with an `i18n-ignore` comment on it or
//    the line above.
//
// Usage: node scripts/i18n-guard.js [--strict]

import { globSync, readFileSync, readdirSync } from "node:fs";
import { join, relative } from "node:path";

const strict = process.argv.includes("--strict");

// ---------------------------------------------------------------------------
// Catalog integrity
// ---------------------------------------------------------------------------

const MESSAGES_DIR = "web-common/src/lib/i18n/messages";
// Keys that are part of the file format, not messages.
const NON_MESSAGE_KEYS = new Set(["$schema"]);

const PLACEHOLDER_RE = /\{(\w+)\}/g;
const INPUT_DECL_RE = /^input\s+(\w+)$/;
const LOCAL_DECL_RE = /^local\s+(\w+)\s*=\s*(\w+)\s*:/;

// Top-level keys read from the raw text (JSON.parse silently drops duplicate
// keys, so duplicates must be detected before parsing).
function topLevelKeys(raw) {
  const keys = [];
  let depth = 0;
  let i = 0;
  while (i < raw.length) {
    const ch = raw[i];
    if (ch === '"') {
      const start = ++i;
      while (i < raw.length && raw[i] !== '"') {
        if (raw[i] === "\\") i++;
        i++;
      }
      const str = raw.slice(start, i);
      i++;
      if (depth === 1 && /^\s*:/.test(raw.slice(i))) keys.push(str);
    } else {
      if (ch === "{" || ch === "[") depth++;
      else if (ch === "}" || ch === "]") depth--;
      i++;
    }
  }
  return keys;
}

function placeholders(text) {
  return [...String(text).matchAll(PLACEHOLDER_RE)].map((m) => m[1]);
}

// Returns the set of parameter (input) names for a message, pushing any
// internal-consistency problems onto `errors`. A plain string's parameters
// are its placeholders; a variant message's parameters are its declared
// inputs.
function messageInputs(value, where, errors) {
  if (typeof value === "string") {
    if (!value.trim()) errors.push(`${where}: empty message text`);
    return new Set(placeholders(value));
  }

  if (!Array.isArray(value)) {
    errors.push(`${where}: message must be a string or a variant array`);
    return new Set();
  }

  const inputs = new Set();
  value.forEach((variant, idx) => {
    const at = value.length === 1 ? where : `${where}[${idx}]`;
    if (typeof variant !== "object" || variant === null) {
      errors.push(`${at}: variant must be an object`);
      return;
    }

    const locals = new Set();
    for (const decl of variant.declarations ?? []) {
      const input = INPUT_DECL_RE.exec(decl);
      const local = LOCAL_DECL_RE.exec(decl);
      if (input) {
        inputs.add(input[1]);
      } else if (local) {
        locals.add(local[1]);
        if (!inputs.has(local[2])) {
          errors.push(`${at}: local "${local[1]}" reads undeclared input "${local[2]}"`);
        }
      } else {
        errors.push(`${at}: unrecognized declaration "${decl}"`);
      }
    }

    const selectors = variant.selectors ?? [];
    for (const sel of selectors) {
      if (!locals.has(sel) && !inputs.has(sel)) {
        errors.push(`${at}: selector "${sel}" is not a declared input or local`);
      }
    }

    const match = variant.match ?? {};
    if (Object.keys(match).length === 0) {
      errors.push(`${at}: variant has no match branches`);
    }
    for (const [branch, text] of Object.entries(match)) {
      for (const cond of branch.split(/[,\s]+/).filter(Boolean)) {
        const name = cond.split("=")[0];
        if (!selectors.includes(name)) {
          errors.push(`${at}: match branch "${branch}" uses unknown selector "${name}"`);
        }
      }
      if (typeof text !== "string" || !text.trim()) {
        errors.push(`${at}: empty text for match branch "${branch}"`);
      }
      for (const ref of placeholders(text)) {
        if (!inputs.has(ref) && !locals.has(ref)) {
          errors.push(`${at}: placeholder "{${ref}}" is not a declared input or local`);
        }
      }
    }
  });
  return inputs;
}

function checkCatalogs() {
  const errors = [];
  const files = readdirSync(MESSAGES_DIR)
    .filter((f) => f.endsWith(".json"))
    .sort();
  if (files.length === 0) {
    errors.push(`${MESSAGES_DIR}: no locale files found`);
    return errors;
  }

  const catalogs = new Map(); // file -> { keys: Set, inputs: Map<key, Set> }
  for (const file of files) {
    const raw = readFileSync(join(MESSAGES_DIR, file), "utf8");

    const rawKeys = topLevelKeys(raw).filter((k) => !NON_MESSAGE_KEYS.has(k));
    const seen = new Set();
    for (const key of rawKeys) {
      if (seen.has(key)) errors.push(`${file}: duplicate key "${key}"`);
      seen.add(key);
    }

    const messages = JSON.parse(raw);
    const inputs = new Map();
    for (const [key, value] of Object.entries(messages)) {
      if (NON_MESSAGE_KEYS.has(key)) continue;
      inputs.set(key, messageInputs(value, `${file}: ${key}`, errors));
    }
    catalogs.set(file, { keys: seen, inputs });
  }

  // Key parity: every locale must define the union of keys across locales.
  const allKeys = new Set();
  for (const { keys } of catalogs.values()) {
    for (const key of keys) allKeys.add(key);
  }
  for (const [file, { keys }] of catalogs) {
    const missing = [...allKeys].filter((k) => !keys.has(k)).sort();
    for (const key of missing) errors.push(`${file}: missing key "${key}"`);
  }

  // Parameter parity: each locale must accept the superset of inputs seen for
  // a key, so no locale breaks when a caller passes every declared input.
  for (const key of allKeys) {
    const superset = new Set();
    for (const { inputs } of catalogs.values()) {
      for (const name of inputs.get(key) ?? []) superset.add(name);
    }
    if (superset.size === 0) continue;
    for (const [file, { inputs }] of catalogs) {
      if (!inputs.has(key)) continue; // already reported as missing key
      const missing = [...superset].filter((p) => !inputs.get(key).has(p)).sort();
      if (missing.length > 0) {
        errors.push(
          `${file}: ${key}: missing parameter(s) ${missing.map((p) => `"{${p}}"`).join(", ")}`,
        );
      }
    }
  }

  return errors;
}

const catalogErrors = checkCatalogs();
if (catalogErrors.length === 0) {
  console.log("i18n-guard: message catalogs are consistent.");
} else {
  console.log(`i18n-guard: ${catalogErrors.length} catalog error(s):`);
  for (const e of catalogErrors) console.log(`  ${e}`);
}

// ---------------------------------------------------------------------------
// Hardcoded-string heuristic
// ---------------------------------------------------------------------------

// Directories whose strings have been migrated to paraglide. Each migration
// chunk appends its directories here; the guard only polices these areas so
// unmigrated code does not drown the signal.
const MIGRATED_GLOBS = [
  // web-admin organization overview page. Literal `[organization]` brackets are
  // a glob character class, so match the segment with a wildcard.
  "web-admin/src/routes/*organization*/+page.svelte",
];

// Human-facing attributes worth translating. Attributes like `class`, `id`,
// `href`, `name`, etc. are deliberately excluded.
const TEXT_ATTRS = ["placeholder", "title", "aria-label", "alt", "label"];

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
} else {
  const label = strict ? "ERROR" : "WARNING";
  console.log(
    `i18n-guard: ${findings.length} hardcoded string(s) in migrated areas (${label}):`,
  );
  for (const f of findings) console.log(`  ${f}`);
  console.log(
    "\nWrap these in paraglide messages (see web-common/src/lib/i18n/README.md), " +
      "or add an `i18n-ignore` comment if intentional.",
  );
}

process.exit(catalogErrors.length > 0 || (strict && findings.length > 0) ? 1 : 0);
