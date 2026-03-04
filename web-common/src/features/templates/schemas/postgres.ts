import Postgres from "../../../components/icons/connectors/Postgres.svelte";
import PostgresIcon from "../../../components/icons/connectors/PostgresIcon.svelte";
import type { MultiStepFormSchema } from "./types";

export const postgresSchema: MultiStepFormSchema = {
  $schema: "http://json-schema.org/draft-07/schema#",
  type: "object",
  title: "PostgreSQL",
  "x-category": "sqlStore",
  "x-icon": Postgres,
  "x-small-icon": PostgresIcon,
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
      "x-env-var-name": "POSTGRES_DSN",
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
      "x-env-var-name": "POSTGRES_PASSWORD",
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
