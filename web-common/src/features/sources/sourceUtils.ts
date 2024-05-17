import {
  extractFileExtension,
  sanitizeEntityName,
} from "@rilldata/web-common/features/sources/extract-file-name";
import type { V1ConnectorDriver } from "@rilldata/web-common/runtime-client";

export function compileCreateSourceYAML(
  values: Record<string, unknown>,
  connectorName: string,
) {
  const topOfFile = `# Source YAML
# Reference documentation: https://docs.rilldata.com/reference/project-files/sources

type: source`;

  switch (connectorName) {
    case "s3":
    case "gcs":
    case "https":
    case "azure":
    case "local_file":
      connectorName = "duckdb";
      values.sql = buildDuckDbQuery(values.path as string);
      delete values.path;
      break;
    case "sqlite":
      connectorName = "duckdb";
      values.sql = `SELECT * FROM sqlite_scan('${values.db as string}', '${
        values.table as string
      }');`;
      delete values.db;
      delete values.table;
      break;
    case "duckdb": {
      const db = values.db as string;
      if (db.startsWith("md:")) {
        connectorName = "motherduck";
        values.db = db.replace("md:", "");
      }
      break;
    }
  }

  const compiledKeyValues = Object.entries(values)
    .filter(([key]) => key !== "sourceName")
    .map(([key, value]) => `${key}: "${value}"`)
    .join("\n");

  return `${topOfFile}\n\nconnector: "${connectorName}"\n` + compiledKeyValues;
}

function buildDuckDbQuery(path: string): string {
  const extension = extractFileExtension(path);
  if (extensionContainsParts(extension, [".csv", ".tsv", ".txt"])) {
    return `select * from read_csv('${path}', auto_detect=true, ignore_errors=1, header=true)`;
  } else if (extensionContainsParts(extension, [".parquet"])) {
    return `select * from read_parquet('${path}')`;
  } else if (extensionContainsParts(extension, [".json", ".ndjson"])) {
    return `select * from read_json('${path}', auto_detect=true, format='auto')`;
  }

  return `select * from '${path}'`;
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

export function getFileTypeFromPath(fileName) {
  if (!fileName.includes(".")) return "";
  const fileType = fileName.split(/[#?]/)[0].split(".").pop();

  if (!fileType) return "";

  if (fileType === "gz") {
    return fileName.split(".").slice(-2).shift();
  }

  return fileType;
}
