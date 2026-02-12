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
  "clickhousecloud",
  "motherduck",
  "duckdb",
  "druid",
  "pinot",
  "starrocks",
];

export const AI_CONNECTORS = ["claude", "openai", "gemini"];

export const ALL_CONNECTORS = [...SOURCES, ...OLAP_ENGINES, ...AI_CONNECTORS];
