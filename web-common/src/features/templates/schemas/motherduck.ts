import type { MultiStepFormSchema } from "./types";

export const motherduckSchema: MultiStepFormSchema = {
  $schema: "http://json-schema.org/draft-07/schema#",
  type: "object",
  properties: {
    path: {
      type: "string",
      title: "Database Path",
      description: "Path to MotherDuck database (must be prefixed with 'md:')",
      "x-placeholder": "md:my_db",
      "x-step": "connector",
    },
    token: {
      type: "string",
      title: "MotherDuck Token",
      description: "Your MotherDuck authentication token",
      "x-placeholder": "Enter your MotherDuck token",
      "x-secret": true,
      "x-step": "connector",
    },
    schema_name: {
      type: "string",
      title: "Schema Name",
      description: "Default schema used by the MotherDuck database",
      "x-placeholder": "main",
      "x-step": "connector",
    },
    mode: {
      type: "string",
      title: "Connection Mode",
      description: "Database access mode",
      enum: ["read", "readwrite"],
      default: "read",
      "x-display": "radio",
      "x-enum-labels": ["Read-only", "Read-write"],
      "x-enum-descriptions": [
        "Only read operations are allowed (recommended for security)",
        "Enable model creation and table mutations",
      ],
      "x-step": "connector",
    },
    sql: {
      type: "string",
      title: "SQL Query",
      description: "SQL query to extract data from MotherDuck",
      "x-placeholder": "SELECT * FROM my_table;",
      "x-step": "source",
      "x-visible-if": { mode: "readwrite" },
    },
    name: {
      type: "string",
      title: "Model Name",
      description: "Name for the source model",
      pattern: "^[a-zA-Z_][a-zA-Z0-9_]*$",
      "x-placeholder": "my_model",
      "x-step": "source",
      "x-visible-if": { mode: "readwrite" },
    },
    explorer_table: {
      type: "string",
      title: "Select a table",
      description: "Select a table to generate metrics from",
      "x-step": "explorer",
      "x-visible-if": { mode: "read" },
    },
  },
  allOf: [
    {
      if: { properties: { mode: { const: "readwrite" } } },
      then: { required: ["path", "token", "schema_name", "sql", "name"] },
    },
    {
      if: { properties: { mode: { const: "read" } } },
      then: { required: ["path", "token", "schema_name"] },
    },
  ],
};
