import {
  AZURE_AUTH_OPTIONS,
  GCS_AUTH_OPTIONS,
  S3_AUTH_OPTIONS,
} from "./constants";
import type { MultiStepFormConfig } from "./types";

export const multiStepFormConfigs: Record<string, MultiStepFormConfig> = {
  gcs: {
    authOptions: GCS_AUTH_OPTIONS,
    defaultAuthMethod: "credentials",
    clearFieldsByMethod: {
      public: ["google_application_credentials", "key_id", "secret"],
      credentials: ["key_id", "secret"],
      hmac: ["google_application_credentials"],
    },
    excludedKeys: [
      "google_application_credentials",
      "key_id",
      "secret",
      "name",
    ],
    authFieldGroups: {
      public: [],
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
      access_keys: [],
    },
    excludedKeys: [
      "aws_access_key_id",
      "aws_secret_access_key",
      "region",
      "endpoint",
      "name",
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
        {
          type: "input",
          id: "region",
          label: "Region",
          placeholder: "us-east-1",
          optional: true,
          hint: "Rill uses your default AWS region unless you set it explicitly.",
        },
        {
          type: "input",
          id: "endpoint",
          label: "Endpoint",
          placeholder: "https://s3.example.com",
          optional: true,
          hint: "Override the S3 endpoint (for S3-compatible services like R2/MinIO).",
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
      "name",
    ],
    authFieldGroups: {
      connection_string: [
        {
          type: "input",
          id: "azure_storage_connection_string",
          label: "Connection string",
          placeholder: "Enter Azure storage connection string",
          optional: false,
          secret: true,
          hint: "Paste an Azure Storage connection string",
        },
      ],
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
    },
  },
};
