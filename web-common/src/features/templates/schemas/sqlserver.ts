import type { MultiStepFormSchema } from "./types";

export const sqlserverSchema: MultiStepFormSchema = {
  $schema: "http://json-schema.org/draft-07/schema#",
  type: "object",
  title: "SQL Server",
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
        parameters: [
          "host",
          "port",
          "database",
          "user",
          "password",
          "encrypt_connection",
          "tls",
        ],
        dsn: ["dsn"],
      },
    },
    dsn: {
      type: "string",
      title: "SQL Server connection string",
      description:
        "Full DSN, e.g. msql://user:password@host:0000/database?ssl-mode=REQUIRED",
      "x-placeholder": "msql://user:password@host:0000/database",
      "x-secret": true,
      "x-env-var-name": "SQLSERVER_DSN",
      "x-hint":
        "Use DSN or fill host/user/password/database below (not both at once).",
    },
    host: {
      type: "string",
      title: "Host",
      description: "SQL Server server hostname or IP",
      "x-placeholder": "localhost",
    },
    port: {
      type: "string",
      title: "Port",
      description: "SQL Server server port",
      pattern: "^\\d+$",
      errorMessage: { pattern: "Port must be a number" },
      default: "0000",
    },
    database: {
      type: "string",
      title: "Database",
      description: "Database name",
      "x-placeholder": "my_database",
    },
    user: {
      type: "string",
      title: "Username",
      description: "SQL Server user",
      "x-placeholder": "msql",
    },
    password: {
      type: "string",
      title: "Password",
      description: "SQL Server password",
      "x-placeholder": "your_password",
      "x-secret": true,
      "x-env-var-name": "SQLSERVER_PASSWORD",
    },
    "encrypt_connection": {
      type: "boolean",
      title: "Encrypt connection",
      description: "",
    },
    "tls": {
      type: "boolean",
      title: "TLS mode",
      description: "",
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
          { required: ["database"] },
          { required: ["user"] },
          { required: ["password"] },
          { required: ["port"] },
          { required: ["encrypt_connection"] },
          { required: ["tls"] },
        ],
      },
    },
    {
      title: "Use individual parameters",
      required: ["host", "database", "user"],
    },
  ],
  allOf: [
    {
      if: { properties: { connection_mode: { const: "dsn" } } },
      then: { required: ["dsn"] },
      else: { required: ["host", "database", "user"] },
    },
  ],
};
