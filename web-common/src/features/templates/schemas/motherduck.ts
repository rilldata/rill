import type { MultiStepFormSchema } from "./types";

export const motherduckSchema: MultiStepFormSchema = {
  $schema: "http://json-schema.org/draft-07/schema#",
  type: "object",
  properties: {
    path: {
      type: "string",
      title: "Database Path",
      description: "Path to MotherDuck database (must be prefixed with 'md:')",
      "x-placeholder": "md:my_db",
    },
    token: {
      type: "string",
      title: "MotherDuck Token",
      description: "Your MotherDuck authentication token",
      "x-placeholder": "Enter your MotherDuck token",
      "x-secret": true,
    },
    schema_name: {
      type: "string",
      title: "Schema Name",
      description: "Default schema used by the MotherDuck database",
      "x-placeholder": "main",
    },
    mode: {
      type: "string",
      title: "Connection Mode",
      description: "Database access mode",
      enum: ["read", "readwrite"],
      default: "read",
      "x-display": "radio",
      "x-enum-labels": ["Read-only", "Read-write"],
      "x-enum-descriptions": [
        "Only read operations are allowed (recommended for security)",
        "Enable model creation and table mutations",
      ],
    },
  },
  required: ["path", "token", "schema_name"],
};
