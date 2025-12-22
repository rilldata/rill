import type { MultiStepFormSchema } from "./types";

export const duckdbSchema: MultiStepFormSchema = {
  $schema: "http://json-schema.org/draft-07/schema#",
  type: "object",
  properties: {
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
