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
      enum: ["password", "private_key", "dsn"],
      default: "password",
      "x-display": "tabs",
      "x-enum-labels": ["User/Password", "Private Key", "Connection String"],
      "x-ui-only": true,
      "x-tab-group": {
        password: [
          "account",
          "user",
          "password",
          "warehouse",
          "database",
          "schema",
          "role",
        ],
        private_key: [
          "account",
          "user",
          "privateKey",
          "warehouse",
          "database",
          "schema",
          "role",
        ],
        dsn: ["dsn"],
      },
    },
    account: {
      type: "string",
      title: "Account identifier",
      description:
        "Snowflake account identifier (from your Snowflake URL, before .snowflakecomputing.com)",
      "x-placeholder": "abc12345.us-east-1",
      "x-hint": "e.g. abc12345 or abc12345.us-east-1 — don't include https://",
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
      description: "Upload your Snowflake private key file (.pem or .p8)",
      format: "file",
      "x-display": "file",
      "x-file-accept": ".pem,.p8",
      "x-file-encoding": "base64",
      "x-secret": true,
      "x-visible-if": { auth_method: "private_key" },
    },
    warehouse: {
      type: "string",
      title: "Warehouse",
      description: "Compute warehouse",
      "x-placeholder": "your_warehouse",
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
    role: {
      type: "string",
      title: "Role",
      description: "Snowflake role",
      "x-placeholder": "your_role",
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
      if: {
        properties: { auth_method: { const: "password" } },
      },
      then: {
        required: ["account", "user", "password", "database", "warehouse"],
      },
    },
    {
      if: {
        properties: { auth_method: { const: "private_key" } },
      },
      then: {
        required: ["account", "user", "privateKey", "database", "warehouse"],
      },
    },
    {
      if: {
        properties: { auth_method: { const: "dsn" } },
      },
      then: { required: ["dsn"] },
    },
  ],
};
