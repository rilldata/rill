import type { MultiStepFormSchema } from "./types";

export const snowflakeSchema: MultiStepFormSchema = {
  $schema: "http://json-schema.org/draft-07/schema#",
  type: "object",
  properties: {
    dsn: {
      type: "string",
      title: "Snowflake connection string",
      description:
        "Full Snowflake DSN, e.g. <user>@<account>/<db>/<schema>?warehouse=<warehouse>&role=<role>",
      "x-placeholder":
        "<username>@<account_identifier>/<database>/<schema>?warehouse=<warehouse>&role=<role>",
      "x-secret": true,
      "x-hint":
        "Use a full DSN or fill the fields below (not both). Include authenticator and privateKey for JWT if needed.",
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
      description: "Snowflake password (use JWT private key if password auth is disabled)",
      "x-placeholder": "your_password",
      "x-secret": true,
    },
    privateKey: {
      type: "string",
      title: "Private key (JWT)",
      description:
        "URL-safe base64 or PEM private key for SNOWFLAKE_JWT authenticator",
      "x-display": "textarea",
      "x-secret": true,
    },
    authenticator: {
      type: "string",
      title: "Authenticator",
      description: "Override authenticator (e.g., SNOWFLAKE_JWT)",
      "x-placeholder": "SNOWFLAKE_JWT",
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
    parallel_fetch_limit: {
      type: "number",
      title: "Parallel fetch limit",
      description: "Limit concurrent fetches (leave empty for default)",
    },
    log_queries: {
      type: "boolean",
      title: "Log queries",
      description: "Enable SQL query logging (debugging)",
      default: false,
    },
  },
  required: [],
};

