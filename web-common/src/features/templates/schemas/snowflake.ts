import type { MultiStepFormSchema } from "./types";

export const snowflakeSchema: MultiStepFormSchema = {
  $schema: "http://json-schema.org/draft-07/schema#",
  type: "object",
  title: "Snowflake",
  "x-category": "warehouse",
  "x-form-height": "tall",
  properties: {
    auth_method: {
      type: "string",
      title: "Authentication method",
      enum: ["password", "private_key", "dsn"],
      default: "password",
      "x-display": "radio",
      "x-enum-labels": [
        "Username/Password",
        "Private Key",
        "Connection String",
      ],
      "x-enum-descriptions": [
        "Authenticate with your Snowflake username and password.",
        "Authenticate using a private key with SNOWFLAKE_JWT authenticator.",
        "Use a full Snowflake connection string (DSN).",
      ],
      "x-ui-only": true,
      "x-grouped-fields": {
        password: [
          "account",
          "user",
          "password",
          "database",
          "schema",
          "warehouse",
          "role",
        ],
        private_key: [
          "account",
          "user",
          "privateKey",
          "database",
          "schema",
          "warehouse",
          "role",
        ],
        dsn: ["dsn"],
      },
    },
    dsn: {
      type: "string",
      title: "Connection string",
      description:
        "Full Snowflake DSN, e.g. <user>@<account>/<db>/<schema>?warehouse=<warehouse>&role=<role>",
      "x-placeholder":
        "<username>@<account_identifier>/<database>/<schema>?warehouse=<warehouse>&role=<role>",
      "x-secret": true,
      "x-hint":
        "Include authenticator and privateKey query params for JWT if needed.",
      "x-visible-if": { auth_method: "dsn" },
    },
    account: {
      type: "string",
      title: "Account identifier",
      description:
        "Snowflake account identifier (from your Snowflake URL, before .snowflakecomputing.com)",
      "x-placeholder": "abc12345",
    },
    user: {
      type: "string",
      title: "Username",
      description: "Snowflake username",
      "x-placeholder": "your_username",
    },
    password: {
      type: "string",
      title: "Password",
      description: "Snowflake password",
      "x-placeholder": "your_password",
      "x-secret": true,
      "x-visible-if": { auth_method: "password" },
    },
    privateKey: {
      type: "string",
      title: "Private key",
      description: "URL-safe base64 or PEM private key",
      "x-placeholder": "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----",
      "x-secret": true,
      "x-visible-if": { auth_method: "private_key" },
    },
    database: {
      type: "string",
      title: "Database",
      description: "Snowflake database",
      "x-placeholder": "your_database",
    },
    schema: {
      type: "string",
      title: "Schema",
      description: "Default schema",
      "x-placeholder": "public",
    },
    warehouse: {
      type: "string",
      title: "Warehouse",
      description: "Compute warehouse",
      "x-placeholder": "your_warehouse",
    },
    role: {
      type: "string",
      title: "Role",
      description: "Snowflake role",
      "x-placeholder": "your_role",
    },
    sql: {
      type: "string",
      title: "SQL",
      description: "SQL query to run against your warehouse",
      "x-placeholder": "Input SQL",
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
  allOf: [
    {
      if: { properties: { auth_method: { const: "password" } } },
      then: { required: ["account", "user", "password", "database", "warehouse"] },
    },
    {
      if: { properties: { auth_method: { const: "private_key" } } },
      then: { required: ["account", "user", "privateKey", "database", "warehouse"] },
    },
    {
      if: { properties: { auth_method: { const: "dsn" } } },
      then: { required: ["dsn"] },
    },
  ],
};
