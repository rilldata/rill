import type { MultiStepFormSchema } from "./types";

export const databricksSchema: MultiStepFormSchema = {
  $schema: "http://json-schema.org/draft-07/schema#",
  type: "object",
  title: "Databricks",
  "x-category": "warehouse",
  "x-form-height": "tall",
  properties: {
    host: {
      type: "string",
      title: "Host",
      description: "",
      "x-placeholder": "adb-12345.azuredatabricks.net",
      "x-hint": "Your Databricks workspace hostname.",
    },
    http_path: {
      type: "string",
      title: "HTTP Path",
      description: "",
      "x-placeholder": "/sql/1.0/warehouses/abc123",
      "x-hint": "The HTTP path to your SQL warehouse or cluster.",
    },
    token: {
      type: "string",
      title: "Access Token",
      description: "",
      "x-placeholder": "dapi...",
      "x-secret": true,
      "x-env-var-name": "DATABRICKS_TOKEN",
      "x-hint":
        "A Databricks personal access token. This will be stored securely.",
    },
    catalog: {
      type: "string",
      title: "Catalog",
      description: "Unity Catalog catalog name",
      "x-placeholder": "main",
    },
    schema: {
      type: "string",
      title: "Schema",
      description: "Default schema within the catalog",
      "x-placeholder": "default",
    },
    sql: {
      type: "string",
      title: "SQL",
      description: "SQL query to run against your warehouse",
      "x-placeholder": "SELECT * FROM my_table",
      "x-step": "explorer",
    },
    name: {
      type: "string",
      title: "Model name",
      description: "Name for the source model",
      pattern: "^[a-zA-Z0-9_]+$",
      "x-placeholder": "my_model",
      "x-step": "explorer",
    },
  },
  required: ["host", "http_path", "token", "sql", "name"],
};
