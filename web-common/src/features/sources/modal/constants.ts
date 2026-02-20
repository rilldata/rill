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

// pre-defined order for sources (superset used for schema registration)
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
  "staging",
];

// Sources supported when OLAP engine is DuckDB
export const DUCKDB_SOURCES = [
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

// Sources supported when OLAP engine is managed ClickHouse
export const CLICKHOUSE_SOURCES = [
  "staging",
  "s3",
  "gcs",
  "azure",
  "https",
  "local_file",
  "sqlite",
  "postgres",
  "mysql",
];

/**
 * Returns the list of supported source connectors for the given OLAP engine.
 * Defaults to DUCKDB_SOURCES when the engine is unknown.
 */
export function getSourcesForOlapEngine(
  olapDriver: string | undefined,
): string[] {
  switch (olapDriver) {
    case "clickhouse":
      return CLICKHOUSE_SOURCES;
    case "duckdb":
    default:
      return DUCKDB_SOURCES;
  }
}

export const OLAP_ENGINES = [
  "clickhouse",
  "motherduck",
  "duckdb",
  "druid",
  "pinot",
  "starrocks",
];

export const ALL_CONNECTORS = [...SOURCES, ...OLAP_ENGINES];
