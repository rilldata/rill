import type { V1SourceV2 } from "@rilldata/web-common/runtime-client";

export function getFileExtension(source: V1SourceV2): string {
  const path = source?.spec?.properties?.path?.toLowerCase();
  if (path?.includes(".csv")) return "CSV";
  if (path?.includes(".parquet")) return "Parquet";
  if (path?.includes(".json")) return "JSON";
  if (path?.includes(".ndjson")) return "JSON";
  return "";
}

export function formatConnectorType(source: V1SourceV2) {
  switch (source?.spec?.sourceConnector) {
    case "s3":
      return "S3";
    case "gcs":
      return "GCS";
    case "https":
      return "http(s)";
    case "local_file":
      return "Local file";
    default:
      return source?.state?.connector ?? "";
  }
}
