import type { MultiStepFormSchema } from "./types";

export const lanceSchema: MultiStepFormSchema = {
  $schema: "http://json-schema.org/draft-07/schema#",
  type: "object",
  title: "Lance",
  "x-category": "fileStore",
  "x-form-height": "tall",
  properties: {
    storage_type: {
      type: "string",
      title: "Storage backend",
      enum: ["local", "s3", "gcs", "azure"],
      default: "local",
      "x-display": "select",
      "x-select-style": "rich",
      "x-enum-labels": [
        "Local",
        "Amazon S3",
        "Google Cloud Storage",
        "Azure Blob Storage",
      ],
      "x-enum-descriptions": [
        "Read Lance datasets from a local directory",
        "Read Lance datasets from an S3 bucket",
        "Read Lance datasets from a GCS bucket",
        "Read Lance datasets from Azure Blob Storage",
      ],
      "x-ui-only": true,
      "x-required-driver": {
        s3: "s3",
        gcs: "gcs",
        azure: "azure",
      },
      "x-grouped-fields": {
        s3: ["s3_info", "s3_path"],
        gcs: ["gcs_info", "gcs_path"],
        azure: ["azure_info", "azure_path"],
        local: ["local_path"],
      },
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
      title: "Lance dataset URI",
      description: "S3 path to the Lance dataset",
      pattern: "^s3://[^/]+(/.*)?$",
      errorMessage: {
        pattern: "Must be an S3 URI (e.g. s3://bucket/path/to/dataset.lance)",
      },
      "x-placeholder": "s3://bucket/path/to/dataset.lance",
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
      title: "Lance dataset URI",
      description: "GCS path to the Lance dataset",
      pattern: "^gs://[^/]+(/.*)?$",
      errorMessage: {
        pattern: "Must be a GCS URI (e.g. gs://bucket/path/to/dataset.lance)",
      },
      "x-placeholder": "gs://bucket/path/to/dataset.lance",
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
      title: "Lance dataset URI",
      description: "Azure path to the Lance dataset",
      pattern: "^az://.+",
      errorMessage: {
        pattern:
          "Must be an Azure URI (e.g. az://container/path/to/dataset.lance)",
      },
      "x-placeholder": "az://container/path/to/dataset.lance",
      "x-step": "source",
    },
    local_path: {
      type: "string",
      title: "Lance dataset path",
      description: "Local filesystem path to the Lance dataset",
      "x-placeholder": "/path/to/dataset.lance",
      "x-step": "source",
    },
    name: {
      type: "string",
      title: "Model name",
      description: "Name for the source model",
      pattern: "^[a-zA-Z0-9_]+$",
      "x-placeholder": "my_lance_model",
      "x-step": "source",
    },
  },
  required: ["name"],
  allOf: [
    {
      if: { properties: { storage_type: { const: "s3" } } },
      then: { required: ["s3_path"] },
    },
    {
      if: { properties: { storage_type: { const: "gcs" } } },
      then: { required: ["gcs_path"] },
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
