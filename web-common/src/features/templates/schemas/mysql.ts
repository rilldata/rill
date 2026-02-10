import type { MultiStepFormSchema } from "./types";

export const mysqlSchema: MultiStepFormSchema = {
  $schema: "http://json-schema.org/draft-07/schema#",
  type: "object",
  title: "MySQL",
  "x-category": "sqlStore",
  "x-olap": {
    duckdb: { formType: "connector", hiddenFields: ["table"] },
    clickhouse: {
      formType: "source",
      hiddenFields: ["ssl-mode", "sql", "connection_mode", "dsn"],
    },
  },
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
          "ssl-mode",
        ],
        dsn: ["dsn"],
      },
    },
    dsn: {
      type: "string",
      title: "MySQL connection string",
      description:
        "Full DSN, e.g. mysql://user:password@host:3306/database?ssl-mode=REQUIRED",
      "x-placeholder": "mysql://user:password@host:3306/database",
      "x-secret": true,
      "x-hint":
        "Use DSN or fill host/user/password/database below (not both at once).",
    },
    host: {
      type: "string",
      title: "Host",
      description: "MySQL server hostname or IP",
      "x-placeholder": "localhost",
    },
    port: {
      type: "string",
      title: "Port",
      description: "MySQL server port",
      pattern: "^\\d+$",
      errorMessage: { pattern: "Port must be a number" },
      default: "3306",
      "x-placeholder": "3306",
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
      description: "MySQL user",
      "x-placeholder": "mysql",
    },
    password: {
      type: "string",
      title: "Password",
      description: "MySQL password",
      "x-placeholder": "your_password",
      "x-secret": true,
    },
    "ssl-mode": {
      type: "string",
      title: "SSL mode",
      description: "Use DISABLED, PREFERRED, or REQUIRED",
      enum: ["DISABLED", "PREFERRED", "REQUIRED"],
      "x-placeholder": "PREFERRED",
    },
    table: {
      type: "string",
      title: "Table",
      description: "Name of the table to read from",
      "x-placeholder": "my_table",
      "x-step": "explorer",
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
  required: ["sql", "table", "name"],
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
          { required: ["ssl-mode"] },
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
