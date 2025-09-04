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
  "azure",
  "bigquery",
  "duckdb",
  "postgres",
  "mysql",
  "postgres",
  "redshift",
  "s3",
  "salesforce",
  "snowflake",
  "sqlite",
  "https",
  "local_file",
];

export const OLAP_ENGINES = ["clickhouse", "motherduck", "druid", "pinot"];
// sources: ImplementsObjectStore, ImplementsFileStore,
// connectors: ImplementsOLAP, ImplementsWarehouse, ImplementsSQLStore
export const SOURCES = [...DUCKDB_REWRITE_CONNECTORS, ...NON_OLAP_CONNECTORS];

export const ALL_CONNECTORS = [...SOURCES, ...OLAP_ENGINES];
