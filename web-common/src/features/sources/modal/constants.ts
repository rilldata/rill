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
 * - Model types (type: "model"): Local/embedded databases like DuckDB, MotherDuck, etc. (maybeRewriteToDuckDb)
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
