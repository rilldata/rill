import type { MultiStepFormSchema } from "./types";

export const duckdbSchema: MultiStepFormSchema = {
  $schema: "http://json-schema.org/draft-07/schema#",
  type: "object",
  properties: {
    auth_method: {
      type: "string",
      title: "Connection type",
      enum: ["self-managed", "rill-managed"],
      default: "self-managed",
      description: "Choose how to connect to DuckDB",
      "x-display": "radio",
      "x-enum-labels": ["Self-managed", "Rill-managed"],
      "x-enum-descriptions": [
        "Connect to your own self-hosted DuckDB server.",
        "Use a managed DuckDB instance hosted by Rill.",
      ],
      "x-grouped-fields": {
        "self-managed": ["path", "attach", "mode"],
        "rill-managed": ["managed"],
      },
      "x-step": "connector",
    },
    path: {
      type: "string",
      title: "Database Path",
      description: "Path to external DuckDB database file",
      "x-placeholder": "/path/to/main.db",
      "x-step": "connector",
    },
    attach: {
      type: "string",
      title: "Attach",
      description:
        "Attach to an existing DuckDB database with options (alternative to path)",
      "x-placeholder":
        "'ducklake:metadata.ducklake' AS my_ducklake(DATA_PATH 'datafiles')",
      "x-step": "connector",
      "x-advanced": true,
    },
    mode: {
      type: "string",
      title: "Connection Mode",
      description: "Database access mode",
      enum: ["read", "readwrite"],
      default: "readwrite",
      "x-display": "radio",
      "x-enum-labels": ["Read-only", "Read-write"],
      "x-enum-descriptions": [
        "Only read operations are allowed (recommended for security)",
        "Enable model creation and table mutations",
      ],
      "x-step": "connector",
    },
    managed: {
      type: "boolean",
      title: "Managed",
      description: "Enable managed mode for the ClickHouse server",
      default: true,
      "x-readonly": true,
      "x-hint": "Enable managed mode to manage the server automatically",
      "x-step": "connector",
      "x-visible-if": { auth_method: "rill-managed" },
    },
    sql: {
      type: "string",
      title: "SQL Query",
      description: "SQL query to extract data from DuckDB",
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
      then: { required: ["path", "sql", "name"] },
    },
    {
      if: { properties: { mode: { const: "read" } } },
      then: { required: ["path"] },
    },
  ],
};
