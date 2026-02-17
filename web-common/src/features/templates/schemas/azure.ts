import type { MultiStepFormSchema } from "./types";

export const azureSchema: MultiStepFormSchema = {
  $schema: "http://json-schema.org/draft-07/schema#",
  type: "object",
  title: "Azure Blob Storage",
  "x-category": "objectStore",
  properties: {
    auth_method: {
      type: "string",
      title: "Authentication method",
      enum: ["connection_string", "account_key", "sas_token", "public"],
      default: "connection_string",
      description: "Choose how to authenticate to Azure Blob Storage",
      "x-display": "radio",
      "x-enum-labels": [
        "Connection String",
        "Storage Account Key",
        "SAS Token",
        "Public",
      ],
      "x-enum-descriptions": [
        "Provide a full Azure Storage connection string.",
        "Provide the storage account name and access key.",
        "Provide the storage account name and SAS token.",
        "Access publicly readable blobs without credentials.",
      ],
      "x-ui-only": true,
      "x-grouped-fields": {
        connection_string: ["azure_storage_connection_string"],
        account_key: ["azure_storage_account", "azure_storage_key"],
        sas_token: ["azure_storage_account", "azure_storage_sas_token"],
        public: [],
      },
      "x-step": "connector",
    },
    azure_storage_connection_string: {
      type: "string",
      title: "Connection string",
      description: "Paste an Azure Storage connection string",
      "x-placeholder": "Enter Azure storage connection string",
      "x-secret": true,
      "x-env-var-name": "AZURE_STORAGE_CONNECTION_STRING",
      "x-step": "connector",
      "x-visible-if": { auth_method: "connection_string" },
    },
    azure_storage_account: {
      type: "string",
      title: "Storage account",
      description: "The name of the Azure storage account",
      "x-placeholder": "Enter Azure storage account",
      "x-step": "connector",
      "x-visible-if": { auth_method: ["account_key", "sas_token"] },
    },
    azure_storage_key: {
      type: "string",
      title: "Access key",
      description: "Primary or secondary access key for the storage account",
      "x-placeholder": "Enter Azure storage access key",
      "x-secret": true,
      "x-env-var-name": "AZURE_STORAGE_KEY",
      "x-step": "connector",
      "x-visible-if": { auth_method: "account_key" },
    },
    azure_storage_sas_token: {
      type: "string",
      title: "SAS token",
      description:
        "Shared Access Signature token for the storage account (starting with ?sv=)",
      "x-placeholder": "Enter Azure SAS token",
      "x-secret": true,
      "x-env-var-name": "AZURE_STORAGE_SAS_TOKEN",
      "x-step": "connector",
      "x-visible-if": { auth_method: "sas_token" },
    },
    path: {
      type: "string",
      title: "Blob URI",
      description:
        "URI to the Azure blob container or directory (e.g., https://<account>.blob.core.windows.net/container)",
      pattern: "^azure://.+",
      errorMessage: {
        pattern: "Must be an Azure URI (e.g. azure://container/path)",
      },
      "x-placeholder": "azure://container/path",
      "x-step": "source",
    },
    name: {
      type: "string",
      title: "Model name",
      description: "Name for the source model",
      pattern: "^[a-zA-Z0-9_]+$",
      "x-placeholder": "my_model",
      "x-step": "source",
    },
  },
  required: ["path", "name"],
  allOf: [
    {
      if: { properties: { auth_method: { const: "connection_string" } } },
      then: { required: ["azure_storage_connection_string"] },
    },
    {
      if: { properties: { auth_method: { const: "account_key" } } },
      then: { required: ["azure_storage_account", "azure_storage_key"] },
    },
    {
      if: { properties: { auth_method: { const: "sas_token" } } },
      then: {
        required: ["azure_storage_account", "azure_storage_sas_token"],
      },
    },
  ],
};
