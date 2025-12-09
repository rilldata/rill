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

export type GCSAuthMethod = "credentials" | "hmac";

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
];

export type S3AuthMethod = "access_keys" | "role";

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
    value: "role",
    label: "Assume role",
    description:
      "Assume an AWS IAM role using your local or provided credentials.",
  },
];

export type AzureAuthMethod = "account_key" | "sas_token" | "connection_string";

export const AZURE_AUTH_OPTIONS: {
  value: AzureAuthMethod;
  label: string;
  description: string;
  hint?: string;
}[] = [
  {
    value: "account_key",
    label: "Access key",
    description: "Authenticate with storage account name and access key.",
  },
  {
    value: "sas_token",
    label: "SAS token",
    description: "Authenticate with storage account name and SAS token.",
  },
  {
    value: "connection_string",
    label: "Connection string",
    description: "Authenticate with a full Azure storage connection string.",
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
export const MULTI_STEP_CONNECTORS = ["gcs", "s3", "azure"];

export const FORM_HEIGHT_TALL = "max-h-[38.5rem] min-h-[38.5rem]";
export const FORM_HEIGHT_DEFAULT = "max-h-[34.5rem] min-h-[34.5rem]";
export const TALL_FORM_CONNECTORS = new Set([
  "clickhouse",
  "snowflake",
  "salesforce",
]);
