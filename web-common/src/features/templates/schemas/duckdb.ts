import type { MultiStepFormSchema } from "./types";

export const duckdbSchema: MultiStepFormSchema = {
  $schema: "http://json-schema.org/draft-07/schema#",
  type: "object",
  properties: {
    path: {
      type: "string",
      title: "Path",
      description: "Path to external DuckDB database",
      "x-placeholder": "/path/to/main.db",
    },
    sql: {
      type: "string",
      title: "SQL",
      description: "SQL query to run against your database",
      "x-display": "textarea",
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
  required: ["path", "sql", "name"],
};
