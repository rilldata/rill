import { V1ExportFormat } from "../../../client";

export function exportFormatToPrettyString(format: V1ExportFormat): string {
  switch (format) {
    case V1ExportFormat.EXPORT_FORMAT_UNSPECIFIED:
      return "Unspecified Format";
    case V1ExportFormat.EXPORT_FORMAT_CSV:
      return "CSV";
    case V1ExportFormat.EXPORT_FORMAT_XLSX:
      return "Excel (XLSX)";
    case V1ExportFormat.EXPORT_FORMAT_PARQUET:
      return "Parquet";
    default:
      return "Unknown";
  }
}

export function formatNextRunOn(nextRunOn: string): string {
  return new Date(nextRunOn).toLocaleString("en-US", {
    month: "short",
    day: "numeric",
    year: "numeric",
    hour: "numeric",
    minute: "numeric",
    hour12: true,
  });
}
