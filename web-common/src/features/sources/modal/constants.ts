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

export const ALL_CONNECTORS = [...SOURCES, ...OLAP_ENGINES];
