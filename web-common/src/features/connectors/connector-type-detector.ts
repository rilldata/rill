/**
 * Connector type detection utilities.
 *
 * Provides centralized logic for detecting cloud storage connector types
 * from file paths, URLs, and SQL content. Used by the resource graph and
 * resource graph components.
 */

/**
 * Known cloud storage connector types that can be detected from paths/URLs.
 */
export type CloudConnectorType =
  | "s3"
  | "gcs"
  | "azure"
  | "https"
  | "local_file";

/**
 * Cloud storage URL prefixes mapped to their connector types.
 * Uses startsWith for prefix matching.
 */
const PATH_PREFIX_MAP: Record<string, CloudConnectorType> = {
  "s3://": "s3",
  "s3a://": "s3",
  "gs://": "gcs",
  "gcs://": "gcs",
  "azure://": "azure",
  "az://": "azure",
  "abfs://": "azure",
  "abfss://": "azure",
  "https://": "https",
  "http://": "https",
};

/**
 * Common data file extensions for HTTP URL detection.
 * Used to differentiate data URLs from documentation links.
 */
const DATA_FILE_EXTENSIONS = [
  ".parquet",
  ".csv",
  ".json",
  ".ndjson",
  ".jsonl",
  ".xlsx",
  ".xls",
  ".tsv",
];

/**
 * DuckDB read functions that indicate local file access.
 */
const DUCKDB_READ_FUNCTIONS = [
  "read_parquet(",
  "read_csv(",
  "read_json(",
  "read_ndjson(",
];

/**
 * Detect connector type from a path/URL prefix.
 *
 * Checks if the path starts with a known cloud storage prefix.
 * Use this for direct paths like "s3://bucket/file.parquet".
 *
 * @param path - The path or URL to check
 * @returns The detected connector type, or undefined if no match
 *
 * @example
 * detectConnectorFromPath("s3://bucket/file.parquet") // "s3"
 * detectConnectorFromPath("gs://bucket/file.csv") // "gcs"
 * detectConnectorFromPath("azure://container/file.json") // "azure"
 * detectConnectorFromPath("https://example.com/data.csv") // "https"
 * detectConnectorFromPath("/local/file.csv") // undefined
 */
export function detectConnectorFromPath(
  path: string | null | undefined,
): CloudConnectorType | undefined {
  if (!path) return undefined;

  const lowerPath = path.toLowerCase();

  for (const [prefix, connectorType] of Object.entries(PATH_PREFIX_MAP)) {
    if (lowerPath.startsWith(prefix)) {
      return connectorType;
    }
  }

  return undefined;
}

/**
 * Detect connector type from content that may contain embedded URLs.
 *
 * Searches the content for cloud storage URL patterns anywhere in the string.
 * Use this for SQL queries or other content that may embed URLs.
 *
 * For HTTP URLs, only matches if the URL contains a data file extension
 * to avoid false positives from documentation links.
 *
 * Also detects DuckDB read functions that indicate local file access.
 *
 * @param content - The content to search (e.g., SQL query)
 * @returns The detected connector type, or undefined if no match
 *
 * @example
 * detectConnectorFromContent("SELECT * FROM read_parquet('s3://bucket/file.parquet')") // "s3"
 * detectConnectorFromContent("SELECT * FROM read_json('https://api.example.com/data.json')") // "https"
 * detectConnectorFromContent("SELECT * FROM read_csv('/local/file.csv')") // "local_file"
 * detectConnectorFromContent("SELECT * FROM other_model") // undefined
 */
export function detectConnectorFromContent(
  content: string | null | undefined,
): CloudConnectorType | undefined {
  if (!content) return undefined;

  const normalized = content.toLowerCase();

  // Check for cloud storage URL patterns (using includes for embedded URLs)
  if (normalized.includes("s3://") || normalized.includes("s3a://")) {
    return "s3";
  }
  if (normalized.includes("gs://") || normalized.includes("gcs://")) {
    return "gcs";
  }
  if (
    normalized.includes("azure://") ||
    normalized.includes("az://") ||
    normalized.includes("abfs://") ||
    normalized.includes("abfss://")
  ) {
    return "azure";
  }

  // For HTTP(S), only match if it looks like a data file URL
  const httpMatch = normalized.match(/https?:\/\/[^\s'"]+/g);
  if (httpMatch) {
    for (const url of httpMatch) {
      if (DATA_FILE_EXTENSIONS.some((ext) => url.includes(ext))) {
        return "https";
      }
    }
  }

  // Check for DuckDB read functions (indicates local file access)
  for (const func of DUCKDB_READ_FUNCTIONS) {
    if (normalized.includes(func)) {
      return "local_file";
    }
  }

  return undefined;
}

/**
 * Detect connector type by checking multiple sources in priority order.
 *
 * Convenience function that tries path detection first, then content detection.
 * Useful when you have both a direct path and SQL content to check.
 *
 * @param path - A direct path/URL to check first
 * @param content - Content to search if path doesn't match
 * @returns The detected connector type, or undefined if no match
 */
export function detectConnector(
  path: string | null | undefined,
  content: string | null | undefined,
): CloudConnectorType | undefined {
  return detectConnectorFromPath(path) ?? detectConnectorFromContent(content);
}
