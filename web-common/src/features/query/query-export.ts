import type {
  V1StructType,
  V1QueryResolverResponseDataItem,
} from "@rilldata/web-common/runtime-client";

function triggerDownload(content: string, filename: string, mimeType: string) {
  const blob = new Blob([content], { type: mimeType });
  const url = URL.createObjectURL(blob);
  const a = document.createElement("a");
  a.href = url;
  a.download = filename;
  document.body.appendChild(a);
  a.click();
  document.body.removeChild(a);
  URL.revokeObjectURL(url);
}

function getTimestamp(): string {
  return new Date().toISOString().replace(/[:.]/g, "-").slice(0, 19);
}

function escapeCSVField(value: unknown): string {
  if (value === null || value === undefined) return "";
  const str = String(value);
  if (str.includes(",") || str.includes('"') || str.includes("\n")) {
    return `"${str.replace(/"/g, '""')}"`;
  }
  return str;
}

export function downloadResultsAsCSV(
  schema: V1StructType | null,
  data: V1QueryResolverResponseDataItem[] | null,
) {
  if (!schema?.fields || !data?.length) return;

  const columns = schema.fields.map((f) => f.name ?? "");
  const header = columns.map(escapeCSVField).join(",");
  const rows = data.map((row) =>
    columns.map((col) => escapeCSVField(row[col])).join(","),
  );

  const csv = [header, ...rows].join("\n");
  triggerDownload(csv, `query-results-${getTimestamp()}.csv`, "text/csv");
}

export function downloadResultsAsJSON(
  schema: V1StructType | null,
  data: V1QueryResolverResponseDataItem[] | null,
) {
  if (!schema?.fields || !data?.length) return;

  const json = JSON.stringify(data, null, 2);
  triggerDownload(
    json,
    `query-results-${getTimestamp()}.json`,
    "application/json",
  );
}
