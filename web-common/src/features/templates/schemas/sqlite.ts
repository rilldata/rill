import type { MultiStepFormSchema } from "./types";

export const sqliteSchema: MultiStepFormSchema = {
  $schema: "http://json-schema.org/draft-07/schema#",
  type: "object",
  properties: {
    auth_method: {
      type: "string",
      title: "Database location",
      enum: ["local_file", "remote_url"],
      default: "local_file",
      description: "Choose where the SQLite database is located",
      "x-display": "radio",
      "x-enum-labels": ["Local File", "Remote URL"],
      "x-enum-descriptions": [
        "Path to a local SQLite database file on disk.",
        "URL to a remote SQLite database file.",
      ],
      "x-grouped-fields": {
        local_file: ["db"],
        remote_url: ["db"],
      },
    },
    db: {
      type: "string",
      title: "Database Path",
      description: "Path or URL to the SQLite database file",
      "x-placeholder": "/path/to/database.db",
      "x-visible-if": { auth_method: "local_file" },
    },
  },
  allOf: [
    {
      if: { properties: { auth_method: { const: "local_file" } } },
      then: { required: ["db"] },
    },
    {
      if: { properties: { auth_method: { const: "remote_url" } } },
      then: { required: ["db"] },
    },
  ],
};
