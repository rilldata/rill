import type { MultiStepFormSchema } from "./types";

export const sqliteSchema: MultiStepFormSchema = {
  $schema: "http://json-schema.org/draft-07/schema#",
  type: "object",
  properties: {
    db: {
      type: "string",
      title: "Database file",
      description: "Path to SQLite db file",
      "x-placeholder": "/path/to/sqlite.db",
      "x-step": "source",
    },
    table: {
      type: "string",
      title: "Table",
      description: "SQLite table name",
      "x-placeholder": "table",
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
  required: ["db", "table", "name"],
};
