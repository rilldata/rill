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
    },
  },
  required: ["db"],
};
