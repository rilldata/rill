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
    },
    attach: {
      type: "string",
      title: "Attach (Advanced)",
      description: "Attach to an existing DuckDB database with options (alternative to path)",
      "x-placeholder": "'ducklake:metadata.ducklake' AS my_ducklake(DATA_PATH 'datafiles')",
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
    },
  },
  required: ["path"],
};
