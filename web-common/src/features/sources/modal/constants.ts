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

export type BigQueryAuthMethod = "inferred" | "credentials";

export const BIGQUERY_AUTH_OPTIONS: {
  value: BigQueryAuthMethod;
  label: string;
  description: string;
  hint?: string;
}[] = [
  {
    value: "credentials",
    label: "GCP credentials",
    description:
      "Upload a JSON key file for a service account with BigQuery access.",
  },
  {
    value: "inferred",
    label: "Use inferred credentials",
    description:
      "Automatically use credentials from your environment or service account.",
  },
];

// pre-defined order for sources
export const SOURCES = [
  "athena",
  "azure",
  "bigquery",

  "gcs",
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

export const OLAP_ENGINES = [
  "clickhouse",
  "motherduck",
  "duckdb",
  "druid",
  "pinot",
];

export const ALL_CONNECTORS = [...SOURCES, ...OLAP_ENGINES];
