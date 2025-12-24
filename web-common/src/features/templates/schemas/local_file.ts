import type { MultiStepFormSchema } from "./types";

export const localFileSchema: MultiStepFormSchema = {
  $schema: "http://json-schema.org/draft-07/schema#",
  type: "object",
  properties: {
    path: {
      type: "string",
      title: "Path",
      description: "Path or URL to file",
      "x-placeholder": "/path/to/file.csv",
      "x-step": "source",
    },
    format: {
      type: "string",
      title: "Format",
      description: "File format. Inferred from extension if not set.",
      enum: ["csv", "parquet", "json", "ndjson"],
      "x-display": "select",
      "x-enum-labels": ["CSV", "Parquet", "JSON", "NDJSON"],
      "x-placeholder": "csv",
      "x-step": "source",
    },
    name: {
      type: "string",
      title: "Source name",
      description: "Name for the source model",
      pattern: "^[a-zA-Z_][a-zA-Z0-9_]*$",
      "x-placeholder": "my_local_source",
      "x-step": "source",
    },
  },
  required: ["path", "name"],
};
