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

export const OLAP_ENGINES = ["clickhouse", "druid", "pinot"];

// pre-defined order for sources
export const SOURCES = [
  "athena",
  "azure",
  "bigquery",
  "duckdb",
  "gcs",
  "motherduck",
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

export const ALL_CONNECTORS = [...SOURCES, ...OLAP_ENGINES];
