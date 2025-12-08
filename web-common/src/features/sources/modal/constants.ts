import type { MultiStepFormConfig } from "./types";

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

export const multiStepFormConfigs: Record<string, MultiStepFormConfig> = {
  gcs: {
    authOptions: GCS_AUTH_OPTIONS,
    defaultAuthMethod: "credentials",
    clearFieldsByMethod: {
      credentials: ["key_id", "secret"],
      hmac: ["google_application_credentials"],
    },
    excludedKeys: ["google_application_credentials", "key_id", "secret"],
    authFieldGroups: {
      credentials: [
        {
          type: "credentials",
          id: "google_application_credentials",
          optional: false,
          hint: "Upload a JSON key file for a service account with GCS access.",
          accept: ".json",
        },
      ],
      hmac: [
        {
          type: "input",
          id: "key_id",
          label: "Access Key ID",
          placeholder: "Enter your HMAC access key ID",
          optional: false,
          secret: true,
          hint: "HMAC access key ID for S3-compatible authentication",
        },
        {
          type: "input",
          id: "secret",
          label: "Secret Access Key",
          placeholder: "Enter your HMAC secret access key",
          optional: false,
          secret: true,
          hint: "HMAC secret access key for S3-compatible authentication",
        },
      ],
    },
  },
  s3: {
    authOptions: S3_AUTH_OPTIONS,
    defaultAuthMethod: "access_keys",
    clearFieldsByMethod: {
      access_keys: ["aws_role_arn", "aws_role_session_name", "aws_external_id"],
      role: ["aws_access_key_id", "aws_secret_access_key"],
    },
    excludedKeys: [
      "aws_access_key_id",
      "aws_secret_access_key",
      "aws_role_arn",
      "aws_role_session_name",
      "aws_external_id",
    ],
    authFieldGroups: {
      access_keys: [
        {
          type: "input",
          id: "aws_access_key_id",
          label: "Access Key ID",
          placeholder: "Enter AWS access key ID",
          optional: false,
          secret: true,
          hint: "AWS access key ID for the bucket",
        },
        {
          type: "input",
          id: "aws_secret_access_key",
          label: "Secret Access Key",
          placeholder: "Enter AWS secret access key",
          optional: false,
          secret: true,
          hint: "AWS secret access key for the bucket",
        },
      ],
      role: [
        {
          type: "input",
          id: "aws_role_arn",
          label: "Role ARN",
          placeholder: "Enter AWS IAM role ARN",
          optional: false,
          secret: true,
          hint: "Role ARN to assume for accessing the bucket",
        },
        {
          type: "input",
          id: "aws_role_session_name",
          label: "Role session name",
          placeholder: "Optional session name (defaults to rill-session)",
          optional: true,
        },
        {
          type: "input",
          id: "aws_external_id",
          label: "External ID",
          placeholder: "Optional external ID for cross-account access",
          optional: true,
          secret: true,
        },
      ],
    },
  },
  azure: {
    authOptions: AZURE_AUTH_OPTIONS,
    defaultAuthMethod: "account_key",
    clearFieldsByMethod: {
      account_key: [
        "azure_storage_connection_string",
        "azure_storage_sas_token",
      ],
      sas_token: ["azure_storage_connection_string", "azure_storage_key"],
      connection_string: [
        "azure_storage_account",
        "azure_storage_key",
        "azure_storage_sas_token",
      ],
    },
    excludedKeys: [
      "azure_storage_account",
      "azure_storage_key",
      "azure_storage_sas_token",
      "azure_storage_connection_string",
    ],
    authFieldGroups: {
      account_key: [
        {
          type: "input",
          id: "azure_storage_account",
          label: "Storage account",
          placeholder: "Enter Azure storage account",
          optional: false,
          hint: "The name of the Azure storage account",
        },
        {
          type: "input",
          id: "azure_storage_key",
          label: "Access key",
          placeholder: "Enter Azure storage access key",
          optional: false,
          secret: true,
          hint: "Primary or secondary access key for the storage account",
        },
      ],
      sas_token: [
        {
          type: "input",
          id: "azure_storage_account",
          label: "Storage account",
          placeholder: "Enter Azure storage account",
          optional: false,
        },
        {
          type: "input",
          id: "azure_storage_sas_token",
          label: "SAS token",
          placeholder: "Enter Azure SAS token",
          optional: false,
          secret: true,
          hint: "Shared Access Signature token for the storage account",
        },
      ],
      connection_string: [
        {
          type: "input",
          id: "azure_storage_connection_string",
          label: "Connection string",
          placeholder: "Enter Azure storage connection string",
          optional: false,
          secret: true,
          hint: "Full connection string for the storage account",
        },
      ],
    },
  },
};
