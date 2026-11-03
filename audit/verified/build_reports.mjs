#!/usr/bin/env node

import fs from "node:fs";
import path from "node:path";
import process from "node:process";

const root = path.resolve(import.meta.dirname, "..");
const verifiedDir = import.meta.dirname;
const partsDir = path.join(verifiedDir, "parts");
const snapshot = "e2e42fb963508b9b9a75e146632e6192805f957d";
const snapshotDate = "2026-07-31";
const allowIncomplete = process.argv.includes("--allow-incomplete");

const rollups = new Set(["00-summary.md", "MUST-FIX.md", "README.md"]);
const expectedSources = fs
  .readdirSync(root)
  .filter((name) => name.endsWith(".md") && !rollups.has(name))
  .sort();

const partFiles = fs.existsSync(partsDir)
  ? fs.readdirSync(partsDir).filter((name) => name.endsWith(".verified.md")).sort()
  : [];

const records = [];
for (const partFile of partFiles) {
  const content = fs.readFileSync(path.join(partsDir, partFile), "utf8");
  for (const rawLine of content.split(/\r?\n/)) {
    const line = rawLine.trim();
    if (!/^(PENDING|FIXED|REMOVE|PARTIAL|NEEDS_TEST|INVALID)\|/.test(line)) continue;
    const fields = line.split("|");
    if (fields.length < 8) {
      throw new Error(`Malformed manifest line in ${partFile}: ${line}`);
    }
    const [status, source, id, originalSeverity, priority, , title, ...evidenceParts] = fields;
    records.push({
      status,
      source,
      id,
      originalSeverity: originalSeverity.toUpperCase(),
      priority: priority.toUpperCase(),
      title,
      evidence: evidenceParts.join("|"),
      partFile,
    });
  }
}

const duplicateKeys = [...new Set(records.map((record) => `${record.source}|${record.id}`))]
  .filter((key) => records.filter((record) => `${record.source}|${record.id}` === key).length > 1);
if (duplicateKeys.length) {
  throw new Error(`Duplicate source IDs in manifests: ${duplicateKeys.join(", ")}`);
}

const coveredSources = [...new Set(records.map((record) => record.source))].sort();
const missingSources = expectedSources.filter((source) => !coveredSources.includes(source));
const unexpectedSources = coveredSources.filter((source) => !expectedSources.includes(source));
if (unexpectedSources.length) {
  throw new Error(`Unexpected sources in manifests: ${unexpectedSources.join(", ")}`);
}
if (missingSources.length && !allowIncomplete) {
  throw new Error(`Missing verified source manifests: ${missingSources.join(", ")}`);
}

const priorityOrder = new Map([["P0", 0], ["P1", 1], ["P2", 2], ["P3", 3], ["-", 4]]);
const statusOrder = new Map([["PENDING", 0], ["PARTIAL", 1], ["NEEDS_TEST", 2], ["FIXED", 3], ["REMOVE", 4], ["INVALID", 5]]);
records.sort((a, b) =>
  (priorityOrder.get(a.priority) ?? 9) - (priorityOrder.get(b.priority) ?? 9) ||
  (statusOrder.get(a.status) ?? 9) - (statusOrder.get(b.status) ?? 9) ||
  a.source.localeCompare(b.source) ||
  a.id.localeCompare(b.id, undefined, { numeric: true }),
);

// Canonical, deduplicated mapping from the historical 42-item must-fix rollup
// to the independently verified source manifests. Five REVIEW_FINDINGS_TODO
// additions have no exact source-file finding, so their current evidence is
// maintained here as explicit supplemental records.
const mustFixMap = [
  ["MF-01", "web-admin.md", "web-admin-01"],
  ["MF-02", "ai.md", "ai-01"],
  ["MF-03", "drivers-connectors-ai.md", "drivers-connectors-ai-01"],
  ["MF-04", "drivers-olap.md", "drivers-olap-01"],
  ["MF-05", "runtime-pkg.md", "runtime-pkg-13"],
  ["MF-06", "runtime-pkg.md", "runtime-pkg-02"],
  ["MF-07", "runtime-core.md", "runtime-core-02"],
  ["MF-08", "ai.md", "ai-02"],
  ["MF-09", "admin-billing-provision.md", "abp-02"],
  ["MF-10", "frontend-shared.md", "fs-shared-01"],
  ["MF-11", "frontend-runtime-client.md", "frontend-runtime-client-01"],
  ["MF-12", "drivers-object-store.md", "drivers-object-store-01"],
  ["MF-13", "admin-auth-rbac.md", "admin-auth-rbac-03"],
  ["MF-14", "admin-projects-deploy.md", "admin-projects-deploy-01"],
  ["MF-15", "drivers-object-store.md", "drivers-object-store-02"],
  ["MF-16", "drivers-object-store.md", "drivers-object-store-05"],
  ["MF-17", "ai.md", "ai-03"],
  ["MF-18", "cli.md", "cli-05"],
  ["MF-19", "cli.md", "cli-04"],
  ["MF-20", null, null],
  ["MF-21", null, null],
  ["MF-22", "drivers-olap.md", "drivers-olap-02"],
  ["MF-23", "drivers-olap.md", "drivers-olap-03"],
  ["MF-24", "frontend-runtime-client.md", "frontend-runtime-client-08"],
  ["MF-25", "deep-cloud-deploy.md", "deep-cloud-deploy-14"],
  ["MF-26", "reconcilers.md", "reconcilers-06"],
  ["MF-27", null, null],
  ["MF-28", "deep-connectors.md", "deep-connectors-01"],
  ["MF-29", "deep-connectors.md", "deep-connectors-02"],
  ["MF-30", "deep-connectors.md", "deep-connectors-03"],
  ["MF-31", "deep-connectors.md", "deep-connectors-04"],
  ["MF-32", "deep-connectors.md", "deep-connectors-05"],
  ["MF-33", "deep-audit-canvas.md", "CANV-1"],
  ["MF-34", "deep-audit-canvas.md", "CANV-2"],
  ["MF-35", "deep-audit-pivot.md", "PIVO-1"],
  ["MF-36", "deep-audit-pivot.md", "PIVO-2"],
  ["MF-37", "deep-audit-pivot.md", "PIVO-3"],
  ["MF-38", "deep-audit-state-management.md", "STAT-1"],
  ["MF-39", "deep-audit-time-dimension-detail.md", "TIME-1"],
  ["MF-40", "deep-audit-time-dimension-detail.md", "TIME-2"],
  ["MF-41", "deep-audit-time-dimension-detail.md", "TIME-3"],
  ["MF-42", "deep-metrics-sql.md", "deep-metrics-sql-08"],
];

const supplementalMustFix = new Map([
  ["MF-20", {
    status: "PENDING", source: "REVIEW_FINDINGS_TODO.md", id: "controller-delete-error", originalSeverity: "HIGH", priority: "P0",
    title: "Failed delete reconciliation still hard-deletes the catalog resource",
    evidence: "runtime/controller.go:1459-1470 logs the delete error and still calls catalog.delete",
  }],
  ["MF-21", {
    status: "PENDING", source: "REVIEW_FINDINGS_TODO.md", id: "controller-late-completion", originalSeverity: "HIGH", priority: "P1",
    title: "Late reconcile completion can block after controller shutdown",
    evidence: "runtime/controller.go:320-350 stops draining after 30s; invoke sends unconditionally at :1335",
  }],
  ["MF-27", {
    status: "PENDING", source: "REVIEW_FINDINGS_TODO.md", id: "embed-singleton-stale-context", originalSeverity: "HIGH", priority: "P0",
    title: "Embed singleton can retain a previous security context",
    evidence: "embed-store.ts:30-53 is process-global; embed/+layout.ts:7-10 initializes only when absent",
  }],
]);

const oldMustFixContent = fs.readFileSync(path.join(root, "MUST-FIX.md"), "utf8");
const oldMustFixTitles = new Map();
for (const line of oldMustFixContent.split(/\r?\n/)) {
  const match = line.match(/^- \[ \] \*\*(MF-\d+) — (.+?)\*\*/);
  if (match) oldMustFixTitles.set(match[1], match[2].replace(/\.$/, ""));
}

const mustFixRecords = mustFixMap.map(([mfId, source, findingId]) => {
  const supplemental = supplementalMustFix.get(mfId);
  const verified = supplemental ?? records.find((record) => record.source === source && record.id === findingId);
  if (!verified) {
    if (allowIncomplete) return null;
    throw new Error(`Missing verified record for ${mfId}: ${source}|${findingId}`);
  }
  return {
    ...verified,
    mfId,
    title: oldMustFixTitles.get(mfId) ?? verified.title,
  };
}).filter(Boolean);

function tableEscape(value) {
  return String(value).replaceAll("|", "\\|").replaceAll("\n", " ");
}

function sourceLink(record) {
  if (!record.partFile) return `[${tableEscape(record.id)}](./REVIEW-FINDINGS-DELTA.md)`;
  return `[${tableEscape(record.id)}](./parts/${encodeURI(record.partFile)})`;
}

function writeReport(filename, title, intro, selected) {
  const lines = [
    `# ${title}`,
    "",
    `Snapshot: \`${snapshot}\` (${snapshotDate}).`,
    "",
    intro,
    "",
    `Count: **${selected.length}**.`,
    "",
    "| Finding | Source | Original severity | Status | Current evidence |",
    "|---|---|---:|---|---|",
  ];
  for (const record of selected) {
    lines.push(
      `| ${sourceLink(record)} — ${tableEscape(record.title)} | \`${tableEscape(record.source)}\` | ${tableEscape(record.originalSeverity)} | ${tableEscape(record.status)} | ${tableEscape(record.evidence)} |`,
    );
  }
  if (!selected.length) lines.push("| — | — | — | — | No findings in this category. |");
  lines.push("");
  fs.writeFileSync(path.join(verifiedDir, filename), `${lines.join("\n")}\n`);
}

const isProposal = (record) =>
  record.source === "product-ideas.md" ||
  record.source === "plg_improvement_ideas.md" ||
  record.id.startsWith("NFI-");
const supplementalRecords = [...supplementalMustFix.values()].map((record) => ({ ...record, partFile: null }));
const active = [
  ...records.filter((record) => record.status === "PENDING" && !isProposal(record)),
  ...supplementalRecords.filter((record) => record.status === "PENDING"),
].sort((a, b) =>
  (priorityOrder.get(a.priority) ?? 9) - (priorityOrder.get(b.priority) ?? 9) ||
  a.source.localeCompare(b.source) ||
  a.id.localeCompare(b.id, undefined, { numeric: true }),
);
const proposals = records.filter((record) =>
  isProposal(record) && (record.status === "PENDING" || record.status === "PARTIAL" || record.status === "NEEDS_TEST"),
);
for (const [priority, filename, title, intro] of [
  ["P0", "P0-NOW.md", "P0 — Must fix now", "Active security, data-loss, corruption, crash, and broad-outage risks. Fixed findings are deliberately excluded."],
  ["P1", "P1-NEXT.md", "P1 — High priority next", "Major correctness and reliability defects that should follow the P0 queue."],
  ["P2", "P2-PLANNED.md", "P2 — Planned fixes", "Confirmed lower-impact bugs, performance problems, and bounded robustness work."],
  ["P3", "P3-LATER.md", "P3 — Later backlog", "Low-impact hardening, cleanup, and product improvements for opportunistic work."],
]) {
  writeReport(filename, title, intro, active.filter((record) => record.priority === priority));
}

writeReport(
  "FIXED-REMOVED.md",
  "Fixed findings removed from the active queues",
  "These source findings are verified fixed, deleted with the affected code, or refuted on the snapshot and do not appear in any active priority file.",
  records.filter((record) => record.status === "FIXED" || record.status === "REMOVE" || record.status === "INVALID"),
);
writeReport(
  "NEEDS-VERIFICATION.md",
  "Partial, dormant, or test-dependent findings",
  "These findings are not cleanly actionable until the named reachability or regression test is completed.",
  records.filter((record) => !isProposal(record) && (record.status === "PARTIAL" || record.status === "NEEDS_TEST")),
);
writeReport(
  "IDEAS-BACKLOG.md",
  "Verified product and PLG proposal backlog",
  "These are roadmap proposals rather than live defects. Their status records whether the underlying capability is absent or partially delivered on the snapshot.",
  proposals,
);

const mustFixPending = mustFixRecords.filter((record) => record.status === "PENDING");
const mustFixActive = mustFixPending.filter((record) => record.priority === "P0" || record.priority === "P1");
const mustFixDeprioritized = mustFixPending.filter((record) => record.priority !== "P0" && record.priority !== "P1");
const mustFixRemoved = mustFixRecords.filter((record) => record.status === "FIXED" || record.status === "REMOVE" || record.status === "INVALID");
const mustFixUncertain = mustFixRecords.filter((record) => record.status === "PARTIAL" || record.status === "NEEDS_TEST");
if (!allowIncomplete && mustFixRecords.length !== 42) {
  throw new Error(`Must-fix rollup has ${mustFixRecords.length} records, expected 42`);
}

function mustFixSource(record) {
  if (record.partFile) return `[${tableEscape(record.source)}](./parts/${encodeURI(record.partFile)})`;
  return `[REVIEW_FINDINGS_TODO.md delta](./REVIEW-FINDINGS-DELTA.md)`;
}

const mustFixLines = [
  "# Current must-fix and high-priority backlog",
  "",
  `Snapshot: \`${snapshot}\` (${snapshotDate}).`,
  "",
  `This is the clean, deduplicated successor to the historical 42-item list. It contains only current high-priority work in the active tables. **${mustFixActive.length} remain P0/P1; ${mustFixDeprioritized.length} are still pending but were reclassified to P2/P3; ${mustFixRemoved.length} were verified fixed/invalid and removed${mustFixUncertain.length ? `; ${mustFixUncertain.length} require a targeted test` : ""}.**`,
  "",
  "Priority is based on current impact: P0 is an active security/data-loss/corruption/crash boundary; P1 is major correctness or reliability work.",
  "",
];

for (const priority of ["P0", "P1"]) {
  const selected = mustFixActive.filter((record) => record.priority === priority);
  if (!selected.length) continue;
  mustFixLines.push(
    `## ${priority} (${selected.length})`,
    "",
    "| ID | Finding | Verified source | Current code evidence |",
    "|---|---|---|---|",
  );
  for (const record of selected) {
    mustFixLines.push(`| ${record.mfId} | ${tableEscape(record.title)} | ${mustFixSource(record)} | ${tableEscape(record.evidence)} |`);
  }
  mustFixLines.push("");
}

if (mustFixDeprioritized.length) {
  mustFixLines.push(
    `## Deprioritized from the high-priority rollup (${mustFixDeprioritized.length})`,
    "",
    "These findings remain valid and appear in the planned/later source queues, but current-code review no longer supports treating them as must-fix blockers.",
    "",
    "| ID | Current priority | Finding | Verified source | Current code evidence |",
    "|---|---:|---|---|---|",
  );
  for (const record of mustFixDeprioritized) {
    mustFixLines.push(`| ${record.mfId} | ${record.priority} | ${tableEscape(record.title)} | ${mustFixSource(record)} | ${tableEscape(record.evidence)} |`);
  }
  mustFixLines.push("");
}

if (mustFixUncertain.length) {
  mustFixLines.push(
    `## Requires a targeted test (${mustFixUncertain.length})`,
    "",
    "| ID | Finding | Verified source | Current evidence |",
    "|---|---|---|---|",
  );
  for (const record of mustFixUncertain) {
    mustFixLines.push(`| ${record.mfId} | ${tableEscape(record.title)} | ${mustFixSource(record)} | ${tableEscape(record.evidence)} |`);
  }
  mustFixLines.push("");
}

mustFixLines.push(
  `## Removed since the prior 42-item snapshot (${mustFixRemoved.length})`,
  "",
  "| ID | Finding | Disposition | Evidence |",
  "|---|---|---|---|",
);
for (const record of mustFixRemoved) {
  mustFixLines.push(`| ${record.mfId} | ${tableEscape(record.title)} | ${record.status} | ${tableEscape(record.evidence)} |`);
}
if (!mustFixRemoved.length) mustFixLines.push("| — | — | — | No removals. |");
mustFixLines.push(
  "",
  "The original IDs are retained so links and remediation history remain stable. Full source-level inventories, including lower priorities and duplicates, are linked above and from [README.md](./README.md).",
  "",
);
fs.writeFileSync(path.join(verifiedDir, "MUST-FIX.md"), mustFixLines.join("\n"));

const counts = (items, key) => Object.fromEntries(
  [...new Set(items.map((item) => item[key]))].sort().map((value) => [value, items.filter((item) => item[key] === value).length]),
);
const statusCounts = counts(records, "status");
const priorityCounts = counts(active, "priority");
const proposalPriorityCounts = counts(proposals, "priority");
const coverageRows = expectedSources.map((source) => {
  const sourceRecords = records.filter((record) => record.source === source);
  const part = sourceRecords[0]?.partFile;
  const statusSummary = ["PENDING", "FIXED", "REMOVE", "PARTIAL", "NEEDS_TEST", "INVALID"]
    .map((status) => [status, sourceRecords.filter((record) => record.status === status).length])
    .filter(([, count]) => count)
    .map(([status, count]) => `${status.toLowerCase()}: ${count}`)
    .join(" · ");
  return part
    ? `| [${source}](./parts/${encodeURI(part)}) | ${sourceRecords.length} | ${statusSummary} |`
    : `| ${source} | 0 | MISSING |`;
});

const readme = [
  "# Verified audit backlog",
  "",
  `Code snapshot: \`${snapshot}\` (latest \`origin/main\` on ${snapshotDate}).`,
  "",
  "The original audit files are preserved as historical evidence. This directory is the refreshed view: every numbered source finding is rechecked against current code and merged PR history, fixed items are removed from active queues, and unresolved work is grouped by current priority.",
  "",
  "## Active queues",
  "",
  `- [P0 — Must fix now](./P0-NOW.md): **${priorityCounts.P0 ?? 0}**`,
  `- [P1 — High priority next](./P1-NEXT.md): **${priorityCounts.P1 ?? 0}**`,
  `- [P2 — Planned fixes](./P2-PLANNED.md): **${priorityCounts.P2 ?? 0}**`,
  `- [P3 — Later backlog](./P3-LATER.md): **${priorityCounts.P3 ?? 0}**`,
  `- [Fixed / removed](./FIXED-REMOVED.md): **${(statusCounts.FIXED ?? 0) + (statusCounts.REMOVE ?? 0) + (statusCounts.INVALID ?? 0)}**`,
  `- [Partial / needs verification](./NEEDS-VERIFICATION.md): **${records.filter((record) => !isProposal(record) && (record.status === "PARTIAL" || record.status === "NEEDS_TEST")).length}**`,
  `- [Product / PLG proposals](./IDEAS-BACKLOG.md): **${proposals.length}**`,
  "",
  "The updated historical 42-item ship-blocker rollup is in [MUST-FIX.md](./MUST-FIX.md). The P0/P1 files are the exhaustive current-code inventories and also include unique supplemental findings from `REVIEW_FINDINGS_TODO.md`.",
  "",
  "## Disposition summary",
  "",
  "| Disposition | Count |",
  "|---|---:|",
  ...Object.entries(statusCounts).sort().map(([status, count]) => `| ${status} | ${count} |`),
  `| **Total classified** | **${records.length}** |`,
  "",
  "## Source coverage",
  "",
  `Covered: **${coveredSources.length}/${expectedSources.length}** substantive audit files.${missingSources.length ? ` Missing: ${missingSources.map((source) => `\`${source}\``).join(", ")}.` : ""}`,
  "",
  "| Source | Classified findings | Disposition |",
  "|---|---:|---|",
  ...coverageRows,
  "",
  "## Rebuild",
  "",
  "Run `node audit/verified/build_reports.mjs` after changing a per-source manifest. The build fails on missing sources, duplicate IDs, malformed records, or unexpected source names.",
  "",
];
fs.writeFileSync(path.join(verifiedDir, "README.md"), readme.join("\n"));

console.log(JSON.stringify({
  snapshot,
  expectedSources: expectedSources.length,
  coveredSources: coveredSources.length,
  missingSources,
  totalRecords: records.length,
  statusCounts,
  priorityCounts,
  proposalPriorityCounts,
}, null, 2));
