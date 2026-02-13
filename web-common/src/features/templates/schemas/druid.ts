import { olapExplorerFields } from "./olap-explorer-fields";
import type { MultiStepFormSchema } from "./types";

export const druidSchema: MultiStepFormSchema = {
  $schema: "http://json-schema.org/draft-07/schema#",
  type: "object",
  title: "Apache Druid",
  "x-category": "olap",
  properties: {
    connection_mode: {
      type: "string",
      title: "Connection method",
      enum: ["parameters", "dsn"],
      default: "parameters",
      "x-display": "tabs",
      "x-enum-labels": ["Enter parameters", "Enter connection string"],
      "x-ui-only": true,
      "x-tab-group": {
        parameters: ["host", "port", "username", "password", "ssl"],
        dsn: ["dsn"],
      },
    },
    dsn: {
      type: "string",
      title: "Connection string",
      description:
        "Full Druid SQL/Avatica endpoint, e.g. https://host:8888/druid/v2/sql/avatica-protobuf?authentication=BASIC&avaticaUser=user&avaticaPassword=pass",
      "x-placeholder":
        "https://example.com/druid/v2/sql/avatica-protobuf?authentication=BASIC&avaticaUser=user&avaticaPassword=pass",
      "x-secret": true,
      "x-env-var-name": "DRUID_DSN",
    },
    host: {
      type: "string",
      title: "Host",
      description: "Druid host or IP",
      "x-placeholder": "localhost",
    },
    port: {
      type: "string",
      title: "Port",
      description: "Druid port",
      pattern: "^\\d+$",
      errorMessage: { pattern: "Port must be a number" },
      "x-placeholder": "8888",
    },
    username: {
      type: "string",
      title: "Username",
      description: "Druid username",
      "x-placeholder": "default",
    },
    password: {
      type: "string",
      title: "Password",
      description: "Druid password",
      "x-placeholder": "password",
      "x-secret": true,
      "x-env-var-name": "DRUID_PASSWORD",
    },
    ssl: {
      type: "boolean",
      title: "SSL",
      description: "Use SSL for the connection",
      default: true,
    },
    ...olapExplorerFields("Druid"),
  },
  required: [],
  oneOf: [
    {
      title: "Use connection string",
      required: ["dsn"],
    },
    {
      title: "Use individual parameters",
      required: ["host", "ssl"],
    },
  ],
};
