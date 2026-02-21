import type { MultiStepFormSchema } from "./types";

export const icebergSchema: MultiStepFormSchema = {
  $schema: "http://json-schema.org/draft-07/schema#",
  type: "object",
  title: "Apache Iceberg",
  "x-category": "fileStore",
  "x-form-height": "tall",
  properties: {
    storage_type: {
      type: "string",
      title: "Storage backend",
      enum: ["local", "gcs", "s3", "azure"],
      default: "local",
      "x-display": "select",
      "x-select-style": "rich",
      "x-enum-labels": [
        "Local",
        "Google Cloud Storage",
        "Amazon S3",
        "Azure Blob Storage",
      ],
      "x-enum-descriptions": [
        "Read Iceberg tables from a local directory",
        "Read Iceberg tables from a GCS bucket",
        "Read Iceberg tables from an S3 bucket",
        "Read Iceberg tables from Azure Blob Storage",
      ],
      "x-ui-only": true,
      "x-grouped-fields": {
        gcs: ["gcs_info", "gcs_path"],
        s3: ["s3_info", "s3_path"],
        azure: ["azure_info", "azure_path"],
        local: ["local_path"],
      },
      "x-step": "source",
    },
    gcs_info: {
      type: "boolean",
      title: "GCS Connector Required",
      default: true,
      "x-informational": true,
      "x-ui-only": true,
      "x-step": "source",
    },
    gcs_path: {
      type: "string",
      title: "Iceberg table URI",
      description: "GCS path to the Iceberg table directory",
      pattern: "^gs://[^/]+(/.*)?$",
      errorMessage: {
        pattern: "Must be a GCS URI (e.g. gs://bucket/path/to/iceberg_table)",
      },
      "x-placeholder": "gs://bucket/path/to/iceberg_table",
      "x-step": "source",
    },
    s3_info: {
      type: "boolean",
      title: "S3 Connector Required",
      default: true,
      "x-informational": true,
      "x-ui-only": true,
      "x-step": "source",
    },
    s3_path: {
      type: "string",
      title: "Iceberg table URI",
      description: "S3 path to the Iceberg table directory",
      pattern: "^s3://[^/]+(/.*)?$",
      errorMessage: {
        pattern: "Must be an S3 URI (e.g. s3://bucket/path/to/iceberg_table)",
      },
      "x-placeholder": "s3://bucket/path/to/iceberg_table",
      "x-step": "source",
    },
    azure_info: {
      type: "boolean",
      title: "Azure Connector Required",
      default: true,
      "x-informational": true,
      "x-ui-only": true,
      "x-step": "source",
    },
    azure_path: {
      type: "string",
      title: "Iceberg table URI",
      description: "Azure path to the Iceberg table directory",
      pattern: "^azure://.+",
      errorMessage: {
        pattern:
          "Must be an Azure URI (e.g. azure://container/path/to/iceberg_table)",
      },
      "x-placeholder": "azure://container/path/to/iceberg_table",
      "x-step": "source",
    },
    local_path: {
      type: "string",
      title: "Iceberg table path",
      description: "Local filesystem path to the Iceberg table directory",
      "x-placeholder": "/path/to/iceberg_table",
      "x-step": "source",
    },
    name: {
      type: "string",
      title: "Model name",
      description: "Name for the source model",
      pattern: "^[a-zA-Z0-9_]+$",
      "x-placeholder": "my_iceberg_model",
      "x-step": "source",
    },
  },
  required: ["name"],
  allOf: [
    {
      if: { properties: { storage_type: { const: "gcs" } } },
      then: { required: ["gcs_path"] },
    },
    {
      if: { properties: { storage_type: { const: "s3" } } },
      then: { required: ["s3_path"] },
    },
    {
      if: { properties: { storage_type: { const: "azure" } } },
      then: { required: ["azure_path"] },
    },
    {
      if: { properties: { storage_type: { const: "local" } } },
      then: { required: ["local_path"] },
    },
  ],
};
