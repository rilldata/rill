export const CONNECTOR_TYPE_OPTIONS: {
  value: boolean;
  label: string;
}[] = [
  { value: true, label: "Rill-managed ClickHouse" },
  { value: false, label: "Self-hosted ClickHouse" },
];

export const CONNECTION_TAB_OPTIONS: { value: string; label: string }[] = [
  { value: "parameters", label: "Enter parameters" },
  { value: "dsn", label: "Enter connection string" },
];

// Connectors that get rewritten to DuckDB
const DUCKDB_REWRITE_CONNECTORS = [
  "s3",
  "gcs",
  "https",
  "azure",
  "local_file",
  "sqlite",
];

export const OLAP_ENGINES = ["clickhouse", "druid", "pinot"];

// FIXME: rename non-olap connectors to A_BETTER_NAME
export const NON_OLAP_CONNECTORS = [
  "bigquery",
  "athena",
  "redshift",
  "duckdb",
  "motherduck",
  "postgres",
  "mysql",
  "snowflake",
  "salesforce",
];

// sources: ImplementsObjectStore, ImplementsFileStore,
// connectors: ImplementsOLAP, ImplementsWarehouse, ImplementsSQLStore
export const SOURCES = [...DUCKDB_REWRITE_CONNECTORS, ...NON_OLAP_CONNECTORS];

export const ALL_CONNECTORS = [...SOURCES, ...OLAP_ENGINES];
