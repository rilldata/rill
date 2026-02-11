import type { MultiStepFormSchema } from "./types";

export const motherduckSchema: MultiStepFormSchema = {
  $schema: "http://json-schema.org/draft-07/schema#",
  type: "object",
  title: "MotherDuck",
  "x-category": "olap",
  properties: {
    path: {
      type: "string",
      title: "Path",
      description: "MotherDuck database path (prefix with md:)",
      "x-placeholder": "md:my_db",
    },
    token: {
      type: "string",
      title: "Token",
      description: "MotherDuck token",
      "x-placeholder": "your_motherduck_token",
      "x-secret": true,
    },
    schema_name: {
      type: "string",
      title: "Schema name",
      description: "Default schema to use",
      "x-placeholder": "main",
    },
    sql: {
      type: "string",
      title: "SQL",
      description: "SQL query to run against MotherDuck",
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
  required: ["path", "token", "schema_name"],
};
