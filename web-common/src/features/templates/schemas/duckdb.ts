import type { MultiStepFormSchema } from "./types";

export const duckdbSchema: MultiStepFormSchema = {
  $schema: "http://json-schema.org/draft-07/schema#",
  type: "object",
  title: "DuckDB",
  "x-category": "olap",
  properties: {
    path: {
      type: "string",
      title: "Path",
      description: "Path to external DuckDB database",
      "x-placeholder": "/path/to/main.db",
    },
  },
  required: ["path"],
};
