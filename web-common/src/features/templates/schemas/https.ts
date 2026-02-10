import type { MultiStepFormSchema } from "./types";

export const httpsSchema: MultiStepFormSchema = {
  $schema: "http://json-schema.org/draft-07/schema#",
  type: "object",
  title: "HTTP(S)",
  "x-category": "fileStore",
  "x-olap": {
    duckdb: { formType: "source" },
    clickhouse: { formType: "source" },
  },
  properties: {
    path: {
      type: "string",
      title: "Path",
      description: "HTTP/HTTPS URL to the remote file",
      pattern: "^https?://.+",
      errorMessage: {
        pattern: "Path must start with http:// or https://",
      },
      "x-placeholder": "https://example.com/file.csv",
      "x-step": "source",
    },
    name: {
      type: "string",
      title: "Source name",
      description: "Name of the source",
      "x-placeholder": "my_new_source",
      "x-step": "source",
    },
  },
  required: ["path", "name"],
};
