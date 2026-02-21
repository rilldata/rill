import type { MultiStepFormSchema } from "./types";

export const databricksSchema: MultiStepFormSchema = {
  $schema: "http://json-schema.org/draft-07/schema#",
  type: "object",
  title: "Databricks",
  "x-category": "warehouse",
  "x-form-height": "tall",
  properties: {
    account: {
      type: "string",
      title: "Host",
      description: "",
      "x-placeholder": "adb-12345.azuredatabricks.net",
      "x-hint": "",
    },
    http_path: {
      type: "string",
      title: "HTTP Path",
      description: "",
      "x-placeholder": "",
    },
    token: {
      type: "string",
      title: "Access Token",
      description: "",
      "x-placeholder": "Access Token",
      "x-secret": true,
      "x-env-var-name": "DATABRICKS_TOKEN",
      "x-visible-if": { auth_method: "token" },
    },
    catalog: {
      type: "string",
      title: "Catalog",
      description: "Databricks catalog",
      "x-placeholder": "",
    },
    schema: {
      type: "string",
      title: "Schema",
      description: "Default schema",
      "x-placeholder": "public",
    },
    sql: {
      type: "string",
      title: "SQL",
      description: "SQL query to run against your warehouse",
      "x-placeholder": "Input SQL",
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
  required: ["sql", "name"],
  allOf: [
    {
      if: {
        properties: { auth_method: { const: "token" } },
      },
      then: {
        required: ["host", "http_path", "token"],
      },
    },
  ],
};
