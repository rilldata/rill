export type ClickHouseConnectorType =
  | "rill-managed"
  | "self-hosted"
  | "clickhouse-cloud";

export const CONNECTOR_TYPE_OPTIONS: {
  value: ClickHouseConnectorType;
  label: string;
}[] = [
  { value: "rill-managed", label: "Rill-managed ClickHouse" },
  { value: "self-hosted", label: "Self-hosted ClickHouse" },
  { value: "clickhouse-cloud", label: "ClickHouse Cloud" },
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

// FIXME: rename non-olap connectors to A_BETTER_NAME
export const NON_OLAP_CONNECTORS = [
  "bigquery",
  "athena",
  "redshift",
  "duckdb",
  "postgres",
  "mysql",
  "snowflake",
  "salesforce",
];

export const OLAP_ENGINES = ["clickhouse", "motherduck", "druid", "pinot"];
// sources: ImplementsObjectStore, ImplementsFileStore,
// connectors: ImplementsOLAP, ImplementsWarehouse, ImplementsSQLStore
export const SOURCES = [...DUCKDB_REWRITE_CONNECTORS, ...NON_OLAP_CONNECTORS];

export const ALL_CONNECTORS = [...SOURCES, ...OLAP_ENGINES];
