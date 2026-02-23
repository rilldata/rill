import { extractFileExtension } from "@rilldata/web-common/features/entity-management/file-path-utils";
import type {
  V1ConnectorDriver,
  V1Source,
} from "@rilldata/web-common/runtime-client";
import { sanitizeEntityName } from "../entity-management/name-utils";

// Helper text that we put at the top of every Model YAML file
function sourceModelFileTop(driverName: string) {
  return `# Model YAML
# Reference documentation: https://docs.rilldata.com/developers/build/connectors/data-source/${driverName}

type: model
materialize: true`;
}

export function compileLocalFileSourceYAML(path: string) {
  return `${sourceModelFileTop("local_file")}\n\nconnector: duckdb\nsql: "${buildDuckDbQuery(path)}"`;
}

export function buildDuckDbQuery(
  path: string | undefined,
  options?: { defaultToJson?: boolean },
): string {
  const safePath = typeof path === "string" ? path : "";
  const extension = extractFileExtension(safePath);
  if (extensionContainsParts(extension, [".csv", ".tsv", ".txt"])) {
    return `select * from read_csv('${safePath}', auto_detect=true, ignore_errors=1, header=true)`;
  } else if (extensionContainsParts(extension, [".parquet"])) {
    return `select * from read_parquet('${safePath}')`;
  } else if (extensionContainsParts(extension, [".json", ".ndjson"])) {
    return `select * from read_json('${safePath}', auto_detect=true, format='auto')`;
  }

  if (options?.defaultToJson) {
    return `select * from read_json('${safePath}', auto_detect=true, format='auto')`;
  }

  return `select * from '${safePath}'`;
}

/**
 * Checks if a file extension '.v1.parquet.gz' contains parts like '.parquet'
 */
function extensionContainsParts(
  fileExtension: string,
  extensionParts: Array<string>,
) {
  for (const extension of extensionParts) {
    if (fileExtension.includes(extension)) return true;
  }
  return false;
}

export function inferSourceName(connector: V1ConnectorDriver, path: string) {
  if (
    !path ||
    path.endsWith("/") ||
    (connector.name === "gcs" && !path.startsWith("gs://")) ||
    (connector.name === "s3" && !path.startsWith("s3://")) ||
    (connector.name === "https" &&
      !path.startsWith("https://") &&
      !path.startsWith("http://"))
  )
    return;

  const slug = path
    .split("/")
    .filter((s: string) => s.length > 0)
    .pop();

  if (!slug) return;

  const fileName = slug.split(".").shift();

  if (!fileName) return;

  return sanitizeEntityName(fileName);
}

export function inferModelNameFromSQL(sql: string): string | undefined {
  if (!sql) return;
  const match = sql.match(/\bFROM\s+([^\s;,()]+)/i);
  if (!match) return;
  // Take the last segment if schema-qualified (e.g. schema.table)
  const raw = match[1]
    .replace(/[`"[\]]/g, "")
    .split(".")
    .pop();
  if (!raw) return;
  return sanitizeEntityName(raw);
}

export function getFileTypeFromPath(fileName) {
  if (!fileName.includes(".")) return "";
  const fileType = fileName.split(/[#?]/)[0].split(".").pop();

  if (!fileType) return "";

  if (fileType === "gz") {
    return fileName.split(".").slice(-2).shift();
  }

  return fileType;
}

export function getFileExtension(source: V1Source): string {
  const path = String(source?.spec?.properties?.path).toLowerCase();
  if (path?.includes(".csv")) return "CSV";
  if (path?.includes(".parquet")) return "Parquet";
  if (path?.includes(".json")) return "JSON";
  if (path?.includes(".ndjson")) return "JSON";
  return "";
}

export function formatConnectorType(source: V1Source) {
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
