import type { MultiStepFormSchema } from "./types";

export const sqliteSchema: MultiStepFormSchema = {
  $schema: "http://json-schema.org/draft-07/schema#",
  type: "object",
  properties: {
    db: {
      type: "string",
      title: "Database Path",
      description: "Path to the SQLite database file",
      "x-placeholder": "/path/to/database.db",
      "x-step": "connector",
    },
    table: {
      type: "string",
      title: "Table",
      description: "SQLite table name",
      "x-placeholder": "my_table",
      "x-step": "source",
    },
    name: {
      type: "string",
      title: "Source Name",
      description: "Name for this source",
      "x-placeholder": "my_sqlite_source",
      "x-step": "source",
    },
  },
  required: ["db", "table", "name"],
};
