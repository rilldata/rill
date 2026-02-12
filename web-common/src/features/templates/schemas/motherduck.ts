import { olapExplorerFields } from "./olap-explorer-fields";
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
    ...olapExplorerFields("MotherDuck"),
  },
  required: ["path", "token", "schema_name"],
};
