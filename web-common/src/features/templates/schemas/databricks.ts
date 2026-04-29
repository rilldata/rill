import type { MultiStepFormSchema } from "./types";

export const databricksSchema: MultiStepFormSchema = {
  $schema: "http://json-schema.org/draft-07/schema#",
  type: "object",
  title: "Databricks",
  "x-category": "warehouse",
  "x-form-height": "tall",
  "x-form-width": "wide",
  properties: {
    connection_mode: {
      type: "string",
      title: "Connection method",
      enum: ["parameters", "dsn"],
      default: "parameters",
      "x-display": "tabs",
      "x-enum-labels": ["Enter parameters", "Enter connection string"],
      "x-ui-only": true,
      "x-tab-group": {
        parameters: ["host", "http_path", "token", "catalog", "schema"],
        dsn: ["dsn"],
      },
    },
    dsn: {
      type: "string",
      title: "Connection string",
      description: "Databricks connection string.",
      "x-placeholder":
        "token:<token>@<host>:443/<http_path>?catalog=<catalog>&schema=<schema>",
      "x-secret": true,
    },
    host: {
      type: "string",
      title: "Host",
      description: "Databricks SQL warehouse hostname.",
      "x-placeholder": "dbc-xxxxxxxx-xxxx.cloud.databricks.com",
    },
    http_path: {
      type: "string",
      title: "HTTP path",
      description: "HTTP path for the SQL warehouse.",
      "x-placeholder": "/sql/1.0/warehouses/xxxxxxxxxxxxxxxx",
    },
    token: {
      type: "string",
      title: "Access token",
      description: "Databricks personal access token.",
      "x-placeholder": "dapi...",
      "x-secret": true,
    },
    catalog: {
      type: "string",
      title: "Catalog",
      description:
        "Unity Catalog name (optional; defaults to the workspace default).",
      "x-placeholder": "main",
    },
    schema: {
      type: "string",
      title: "Schema",
      description:
        "Schema within the catalog (optional; defaults to the workspace default).",
      "x-placeholder": "default",
    },
  },
  required: [],
  allOf: [
    {
      if: {
        properties: { connection_mode: { const: "dsn" } },
      },
      then: { required: ["dsn"] },
      else: { required: ["host", "http_path", "token"] },
    },
  ],
};
