import type { MultiStepFormSchema } from "./types";

export const clickhouseSchema: MultiStepFormSchema = {
  $schema: "http://json-schema.org/draft-07/schema#",
  type: "object",
  properties: {
    connector_type: {
      type: "string",
      title: "Connection type",
      enum: ["self-hosted", "clickhouse-cloud", "rill-managed"],
      default: "self-hosted",
      "x-display": "radio",
      "x-enum-labels": [
        "Self-hosted ClickHouse",
        "ClickHouse Cloud",
        "Rill-managed ClickHouse",
      ],
      "x-step": "connector",
    },
    dsn: {
      type: "string",
      title: "Connection string",
      description:
        "DSN connection string (use instead of individual host/port/user settings)",
      "x-placeholder":
        "clickhouse://localhost:9000?username=default&password=password",
      "x-step": "connector",
    },
    managed: {
      type: "boolean",
      title: "Managed",
      description:
        "Use a managed ClickHouse instance (handled automatically by Rill)",
      default: false,
      "x-step": "connector",
    },
    host: {
      type: "string",
      title: "Host",
      description: "Hostname or IP address of the ClickHouse server",
      "x-placeholder":
        "your-instance.clickhouse.cloud or your.clickhouse.server.com",
      "x-hint":
        "Your ClickHouse hostname (e.g., your-instance.clickhouse.cloud or your-server.com)",
      "x-step": "connector",
    },
    port: {
      type: "string",
      title: "Port",
      description: "Port number of the ClickHouse server",
      pattern: "^\\d+$",
      errorMessage: { pattern: "Port must be a number" },
      default: "9000",
      "x-placeholder": "9000",
      "x-step": "connector",
    },
    username: {
      type: "string",
      title: "Username",
      description: "Username to connect to the ClickHouse server",
      default: "default",
      "x-placeholder": "default",
      "x-step": "connector",
    },
    password: {
      type: "string",
      title: "Password",
      description: "Password to connect to the ClickHouse server",
      "x-placeholder": "Database password",
      "x-secret": true,
      "x-step": "connector",
    },
    database: {
      type: "string",
      title: "Database",
      description: "Name of the ClickHouse database to connect to",
      default: "default",
      "x-placeholder": "default",
      "x-step": "connector",
    },
    cluster: {
      type: "string",
      title: "Cluster",
      description:
        "Cluster name. If set, models are created as distributed tables.",
      "x-placeholder": "Cluster name",
      "x-step": "connector",
    },
    ssl: {
      type: "boolean",
      title: "SSL",
      description: "Use SSL to connect to the ClickHouse server",
      default: true,
      "x-step": "connector",
    },
  },
  required: ["connector_type"],
  allOf: [
    {
      if: { properties: { connector_type: { const: "rill-managed" } } },
      then: {
        required: ["managed"],
        properties: {
          managed: { const: true },
        },
      },
    },
    {
      if: { properties: { connector_type: { const: "self-hosted" } } },
      then: {
        required: ["host", "username"],
        properties: {
          managed: { const: false },
          ssl: { default: true },
        },
      },
    },
    {
      if: { properties: { connector_type: { const: "clickhouse-cloud" } } },
      then: {
        required: ["host", "username", "ssl"],
        properties: {
          managed: { const: false },
          port: { default: "8443" },
          ssl: { const: true },
        },
      },
    },
  ],
};
