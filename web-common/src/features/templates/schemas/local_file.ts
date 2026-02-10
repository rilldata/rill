import type { MultiStepFormSchema } from "./types";

export const localFileSchema: MultiStepFormSchema = {
  $schema: "http://json-schema.org/draft-07/schema#",
  type: "object",
  title: "Local File",
  "x-category": "fileStore",
  "x-olap": {
    duckdb: { formType: "source" },
    clickhouse: { formType: "source" },
  },
  properties: {
    path: {
      type: "string",
      title: "Path",
      description: "Local file path or glob (relative to project root)",
      "x-placeholder": "data/*.parquet",
      "x-step": "source",
    },
    name: {
      type: "string",
      title: "Source name",
      description: "Name for the source",
      "x-placeholder": "my_new_source",
      "x-step": "source",
    },
  },
  required: ["path", "name"],
};
