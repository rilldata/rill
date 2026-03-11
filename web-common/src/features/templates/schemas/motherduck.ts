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
      type: "boolean",
      title: "Enable write mode",
      description:
        "Read-write mode allows Rill to drop, create, and modify tables, not just query them",
      default: false,
      "x-display": "toggle",
      "x-yaml-value": "readwrite",
      "x-advanced": true,
    },
  },
  required: ["path", "token", "schema_name"],
};
