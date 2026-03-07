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
      "x-env-var-name": "MOTHERDUCK_TOKEN",
    },
    schema_name: {
      type: "string",
      title: "Schema name",
      description: "Default schema to use",
      "x-placeholder": "main",
    },
    mode: {
      type: "string",
      title: "Mode",
      description:
        "Database access mode. 'read' allows only read operations; 'readwrite' enables model creation and table mutations",
      enum: ["read", "readwrite"],
      default: "read",
      "x-display": "select",
      "x-enum-labels": ["Read only", "Read & Write"],
      "x-advanced": true,
    },
  },
  required: ["path", "token", "schema_name"],
};
