import type { MultiStepFormSchema } from "./types";

export const sqliteSchema: MultiStepFormSchema = {
  $schema: "http://json-schema.org/draft-07/schema#",
  type: "object",
  title: "SQLite",
  "x-category": "sqlStore",
  "x-olap": {
    duckdb: { formType: "connector" },
    clickhouse: { formType: "source" },
  },
  properties: {
    db: {
      type: "string",
      title: "Database file",
      description: "Path to SQLite db file",
      "x-placeholder": "/path/to/sqlite.db",
    },
    table: {
      type: "string",
      title: "Table",
      description: "SQLite table name",
      "x-placeholder": "table",
    },
    name: {
      type: "string",
      title: "Source name",
      description: "Name of the source",
      "x-placeholder": "my_new_source",
    },
  },
  required: ["db", "table", "name"],
};
