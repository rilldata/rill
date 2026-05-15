import type { MultiStepFormSchema } from "./types";

export const druidSchema: MultiStepFormSchema = {
  $schema: "http://json-schema.org/draft-07/schema#",
  type: "object",
  title: "Druid",
  "x-category": "olap",
  "x-form-width": "wide",
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
    max_open_conns: {
      type: "number",
      title: "Max open connections",
      description:
        "Maximum number of open database connections (0 for default)",
      "x-placeholder": "20",
      "x-advanced": true,
    },
    skip_version_check: {
      type: "boolean",
      title: "Skip version check",
      description: "Skip the Druid version compatibility check",
      "x-advanced": true,
    },
    skip_query_priority: {
      type: "boolean",
      title: "Skip query priority",
      description: "Skip passing query priority to Druid",
      "x-advanced": true,
    },
    log_queries: {
      type: "boolean",
      title: "Log queries",
      description: "Enable SQL query logging for debugging",
      "x-advanced": true,
    },
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
