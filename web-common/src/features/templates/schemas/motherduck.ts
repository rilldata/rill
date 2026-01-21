import type { MultiStepFormSchema } from "./types";

export const motherduckSchema: MultiStepFormSchema = {
  $schema: "http://json-schema.org/draft-07/schema#",
  type: "object",
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
  },
  required: ["path", "token", "schema_name"],
};
