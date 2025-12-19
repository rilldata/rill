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

export type GCSAuthMethod = "public" | "credentials" | "hmac";

export const GCS_AUTH_OPTIONS: {
  value: GCSAuthMethod;
  label: string;
  description: string;
  hint?: string;
}[] = [
  {
    value: "credentials",
    label: "GCP credentials",
    description:
      "Upload a JSON key file for a service account with GCS access.",
  },
  {
    value: "hmac",
    label: "HMAC keys",
    description:
      "Use HMAC access key and secret for S3-compatible authentication.",
  },
  {
    value: "public",
    label: "Public",
    description: "Access publicly readable buckets without credentials.",
  },
];

export type S3AuthMethod = "access_keys" | "public";

export const S3_AUTH_OPTIONS: {
  value: S3AuthMethod;
  label: string;
  description: string;
  hint?: string;
}[] = [
  {
    value: "access_keys",
    label: "Access keys",
    description: "Use AWS access key ID and secret access key.",
  },
  {
    value: "public",
    label: "Public",
    description: "Access publicly readable buckets without credentials.",
  },
];

export type AzureAuthMethod =
  | "account_key"
  | "sas_token"
  | "connection_string"
  | "public";

export const AZURE_AUTH_OPTIONS: {
  value: AzureAuthMethod;
  label: string;
  description: string;
  hint?: string;
}[] = [
  {
    value: "connection_string",
    label: "Connection String",
    description: "Alternative for cloud deployment",
  },
  {
    value: "account_key",
    label: "Storage Account Key",
    description: "Recommended for cloud deployment",
  },
  {
    value: "sas_token",
    label: "Shared Access Signature (SAS) Token",
    description: "Most secure, fine-grained control",
  },
  {
    value: "public",
    label: "Public",
    description: "Access publicly readable blobs without credentials.",
  },
];

// pre-defined order for sources
export const SOURCES = [
  "athena",
  "azure",
  "bigquery",
  "gcs",
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

// Connectors that support multi-step forms (connector -> source)
export const MULTI_STEP_CONNECTORS = [
  "gcs",
  "s3",
  "azure",
  "https",
  "postgres",
  "mysql",
  "snowflake",
  "bigquery",
  "redshift",
  "athena",
  "duckdb",
  "motherduck",
  "druid",
  "pinot",
];

export const FORM_HEIGHT_TALL = "max-h-[38.5rem] min-h-[38.5rem]";
export const FORM_HEIGHT_DEFAULT = "max-h-[34.5rem] min-h-[34.5rem]";
export const TALL_FORM_CONNECTORS = new Set([
  "clickhouse",
  "snowflake",
  "salesforce",
]);
