import type { MultiStepFormSchema } from "./types";

export const oracleSchema: MultiStepFormSchema = {
  $schema: "http://json-schema.org/draft-07/schema#",
  type: "object",
  title: "Oracle",
  "x-category": "sqlStore",
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
        parameters: ["host", "port", "service_name", "user", "password"],
        dsn: ["dsn"],
      },
    },
    dsn: {
      type: "string",
      title: "Oracle connection string",
      description:
        "Full DSN, e.g. oracle://user:password@host:1521/service_name",
      "x-placeholder": "oracle://user:password@host:1521/service_name",
      "x-secret": true,
      "x-env-var-name": "ORACLE_DSN",
      "x-hint":
        "Use DSN or fill host/user/password/service name below (not both at once).",
    },
    host: {
      type: "string",
      title: "Host",
      description: "Oracle server hostname or IP",
      "x-placeholder": "localhost",
    },
    port: {
      type: "string",
      title: "Port",
      description: "Oracle listener port",
      pattern: "^\\d+$",
      errorMessage: { pattern: "Port must be a number" },
      default: "1521",
      "x-placeholder": "1521",
    },
    service_name: {
      type: "string",
      title: "Service Name",
      description: "Oracle service name or SID",
      "x-placeholder": "ORCLPDB1",
    },
    user: {
      type: "string",
      title: "Username",
      description: "Oracle user",
      "x-placeholder": "system",
    },
    password: {
      type: "string",
      title: "Password",
      description: "Oracle password",
      "x-placeholder": "your_password",
      "x-secret": true,
      "x-env-var-name": "ORACLE_PASSWORD",
    },
    sql: {
      type: "string",
      title: "SQL",
      description: "SQL query to run against your database",
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
  required: ["sql", "name"],
  oneOf: [
    {
      title: "Use DSN",
      required: ["dsn"],
      not: {
        anyOf: [
          { required: ["host"] },
          { required: ["service_name"] },
          { required: ["user"] },
          { required: ["password"] },
          { required: ["port"] },
        ],
      },
    },
    {
      title: "Use individual parameters",
      required: ["host", "service_name", "user"],
    },
  ],
  allOf: [
    {
      if: { properties: { connection_mode: { const: "dsn" } } },
      then: { required: ["dsn"] },
      else: { required: ["host", "service_name", "user"] },
    },
  ],
};
