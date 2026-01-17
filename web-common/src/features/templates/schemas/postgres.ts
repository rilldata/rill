import type { MultiStepFormSchema } from "./types";

export const postgresSchema: MultiStepFormSchema = {
  $schema: "http://json-schema.org/draft-07/schema#",
  type: "object",
  properties: {
    connection_mode: {
      type: "string",
      title: "Connection method",
      enum: ["parameters", "dsn"],
      default: "parameters",
      "x-display": "tabs",
      "x-enum-labels": ["Enter parameters", "Enter connection string"],
      "x-internal": true,
      "x-tab-group": {
        parameters: ["host", "port", "user", "password", "dbname", "sslmode"],
        dsn: ["dsn"],
      },
    },
    dsn: {
      type: "string",
      title: "Postgres connection string",
      description:
        "e.g. postgresql://user:password@host:5432/dbname?sslmode=require",
      "x-placeholder": "postgresql://postgres:postgres@localhost:5432/postgres",
      "x-secret": true,
      "x-hint":
        "Use a DSN or provide host/user/password/dbname below (but not both).",
    },
    host: {
      type: "string",
      title: "Host",
      description: "Postgres server hostname or IP",
      "x-placeholder": "localhost",
    },
    port: {
      type: "string",
      title: "Port",
      description: "Postgres server port",
      pattern: "^\\d+$",
      errorMessage: { pattern: "Port must be a number" },
      default: "5432",
      "x-placeholder": "5432",
    },
    user: {
      type: "string",
      title: "Username",
      description: "Postgres user",
      "x-placeholder": "postgres",
    },
    password: {
      type: "string",
      title: "Password",
      description: "Postgres password",
      "x-placeholder": "your_password",
      "x-secret": true,
    },
    dbname: {
      type: "string",
      title: "Database",
      description: "Database name",
      "x-placeholder": "postgres",
    },
    sslmode: {
      type: "string",
      title: "SSL mode",
      description: "Use disable, allow, prefer, require",
      enum: ["disable", "allow", "prefer", "require"],
      "x-placeholder": "require",
    },
  },
  required: [],
  oneOf: [
    {
      title: "Use DSN",
      required: ["dsn"],
      not: {
        anyOf: [
          { required: ["database_url"] },
          { required: ["host"] },
          { required: ["port"] },
          { required: ["user"] },
          { required: ["password"] },
          { required: ["dbname"] },
          { required: ["sslmode"] },
        ],
      },
    },
    {
      title: "Use Database URL",
      required: ["database_url"],
      not: {
        anyOf: [
          { required: ["dsn"] },
          { required: ["host"] },
          { required: ["port"] },
          { required: ["user"] },
          { required: ["password"] },
          { required: ["dbname"] },
          { required: ["sslmode"] },
        ],
      },
    },
    {
      title: "Use individual parameters",
      required: ["host", "user", "dbname"],
    },
  ],
  allOf: [
    {
      if: { properties: { connection_mode: { const: "dsn" } } },
      then: { required: ["dsn"] },
      else: { required: ["host", "user", "dbname"] },
    },
  ],
};
