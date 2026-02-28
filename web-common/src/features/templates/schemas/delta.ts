import type { MultiStepFormSchema } from "./types";

export const deltaSchema: MultiStepFormSchema = {
  $schema: "http://json-schema.org/draft-07/schema#",
  type: "object",
  title: "Delta Lake",
  "x-category": "fileStore",
  "x-form-height": "tall",
  properties: {
    storage_type: {
      type: "string",
      title: "Storage backend",
      enum: ["local", "s3", "azure"],
      default: "local",
      "x-display": "select",
      "x-select-style": "rich",
      "x-enum-labels": ["Local", "Amazon S3", "Azure Blob Storage"],
      "x-enum-descriptions": [
        "Read Delta tables from a local directory",
        "Read Delta tables from an S3 bucket",
        "Read Delta tables from Azure Blob Storage",
      ],
      "x-ui-only": true,
      "x-grouped-fields": {
        s3: ["s3_info", "s3_path"],
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
      title: "Delta table URI",
      description: "S3 path to the Delta table directory",
      pattern: "^s3://[^/]+(/.*)?$",
      errorMessage: {
        pattern: "Must be an S3 URI (e.g. s3://bucket/path/to/delta_table)",
      },
      "x-placeholder": "s3://bucket/path/to/delta_table",
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
      title: "Delta table URI",
      description: "Azure path to the Delta table directory",
      pattern: "^azure://.+",
      errorMessage: {
        pattern:
          "Must be an Azure URI (e.g. azure://container/path/to/delta_table)",
      },
      "x-placeholder": "azure://container/path/to/delta_table",
      "x-step": "source",
    },
    local_path: {
      type: "string",
      title: "Delta table path",
      description: "Local filesystem path to the Delta table directory",
      "x-placeholder": "/path/to/delta_table",
      "x-step": "source",
    },
    name: {
      type: "string",
      title: "Model name",
      description: "Name for the source model",
      pattern: "^[a-zA-Z0-9_]+$",
      "x-placeholder": "my_delta_model",
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
      if: { properties: { storage_type: { const: "azure" } } },
      then: { required: ["azure_path"] },
    },
    {
      if: { properties: { storage_type: { const: "local" } } },
      then: { required: ["local_path"] },
    },
  ],
};
