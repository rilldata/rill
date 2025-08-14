export const CONNECTOR_TYPE_OPTIONS: {
  value: boolean;
  label: string;
}[] = [
  { value: true, label: "Rill-managed ClickHouse" },
  { value: false, label: "Self-managed ClickHouse" },
];

export const CONNECTION_TAB_OPTIONS: { value: string; label: string }[] = [
  { value: "parameters", label: "Enter parameters" },
  { value: "dsn", label: "Enter connection string" },
];

/**
 * Data source types supported by the application.
 *
 * Note: Source types are categorized into two groups:
 * - Connector types (type: "connector"): External services like BigQuery, Athena, etc.
 * - Model types (type: "model"): maybeRewriteToDuckDb
 *
 * This categorization affects how the source is handled in the application.
 */
export const SOURCES = [
  "gcs",
  "s3",
  "azure",
  "bigquery",
  "athena",
  "redshift",
  "duckdb",
  "motherduck",
  "postgres",
  "mysql",
  "sqlite",
  "snowflake",
  "salesforce",
  "local_file",
  "https",
];

export const OLAP_ENGINES = ["clickhouse", "druid", "pinot"];

export const SORT_ORDER = [...SOURCES, ...OLAP_ENGINES];

/**
 * Connectors that are automatically converted to DuckDB to leverage its native
 * file reading capabilities and extensions.
 */
export const DUCKDB_NATIVE_CONNECTORS = [
  "s3",
  "gcs",
  "https",
  "azure",
  "local_file",
  "sqlite",
] as const;
