import type { MultiStepFormSchema } from "./types";

export const postgresSchema: MultiStepFormSchema = {
  $schema: "http://json-schema.org/draft-07/schema#",
  type: "object",
  properties: {
    auth_method: {
      type: "string",
      title: "Authentication method",
      enum: ["parameters", "connection_string"],
      default: "parameters",
      description: "Choose how to connect to PostgreSQL",
      "x-display": "radio",
      "x-enum-labels": ["Username & Password", "Connection String"],
      "x-enum-descriptions": [
        "Provide individual connection parameters (host, port, database, username, password).",
        "Provide a complete PostgreSQL connection string (DSN).",
      ],
      "x-grouped-fields": {
        parameters: ["host", "port", "database", "user", "password", "sslmode"],
        connection_string: ["dsn"],
      },
    },
    host: {
      type: "string",
      title: "Host",
      description: "Database server hostname or IP address",
      "x-placeholder": "localhost",
      "x-visible-if": { auth_method: "parameters" },
    },
    port: {
      type: "number",
      title: "Port",
      description: "Database server port",
      default: 5432,
      "x-placeholder": "5432",
      "x-visible-if": { auth_method: "parameters" },
    },
    database: {
      type: "string",
      title: "Database",
      description: "Database name",
      "x-placeholder": "my_database",
      "x-visible-if": { auth_method: "parameters" },
    },
    user: {
      type: "string",
      title: "Username",
      description: "Database user",
      "x-placeholder": "postgres",
      "x-visible-if": { auth_method: "parameters" },
    },
    password: {
      type: "string",
      title: "Password",
      description: "Database password",
      "x-placeholder": "Enter password",
      "x-secret": true,
      "x-visible-if": { auth_method: "parameters" },
    },
    sslmode: {
      type: "string",
      title: "SSL Mode",
      description: "SSL connection mode",
      enum: ["disable", "require", "verify-ca", "verify-full"],
      default: "prefer",
      "x-display": "select",
      "x-visible-if": { auth_method: "parameters" },
    },
    dsn: {
      type: "string",
      title: "Connection String",
      description: "PostgreSQL connection string (DSN)",
      "x-placeholder": "postgres://user:password@host:5432/dbname?sslmode=require",
      "x-secret": true,
      "x-visible-if": { auth_method: "connection_string" },
    },
  },
  allOf: [
    {
      if: { properties: { auth_method: { const: "parameters" } } },
      then: { required: ["host", "database", "user", "password"] },
    },
    {
      if: { properties: { auth_method: { const: "connection_string" } } },
      then: { required: ["dsn"] },
    },
  ],
};
