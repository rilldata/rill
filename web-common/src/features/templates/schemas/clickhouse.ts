import type { MultiStepFormSchema } from "./types";

export const clickhouseSchema: MultiStepFormSchema = {
  $schema: "http://json-schema.org/draft-07/schema#",
  type: "object",
  properties: {
    connector_type: {
      type: "string",
      title: "Connection type",
      enum: ["rill-managed", "self-hosted", "clickhouse-cloud"],
      default: "self-hosted",
      "x-display": "radio",
      "x-enum-labels": [
        "Rill-managed ClickHouse",
        "Self-hosted ClickHouse",
        "ClickHouse Cloud",
      ],
      "x-internal": true,
      "x-grouped-fields": {
        "rill-managed": ["managed"],
        "self-hosted": ["connection_mode"],
        "clickhouse-cloud": ["connection_mode"],
      },
      "x-step": "connector",
    },
    connection_mode: {
      type: "string",
      enum: ["parameters", "dsn"],
      default: "parameters",
      "x-display": "tabs",
      "x-enum-labels": ["Enter parameters", "Enter connection string"],
      "x-internal": true,
      "x-tab-group": {
        parameters: [
          "host",
          "port",
          "username",
          "password",
          "database",
          "cluster",
          "ssl",
        ],
        dsn: ["dsn"],
      },
      "x-visible-if": {
        connector_type: ["self-hosted", "clickhouse-cloud"],
      },
      "x-step": "connector",
    },
    dsn: {
      type: "string",
      title: "Connection string",
      description:
        "DSN connection string (use instead of individual host/port/user settings)",
      "x-placeholder":
        "clickhouse://localhost:9000?username=default&password=password",
      "x-visible-if": {
        connector_type: ["self-hosted", "clickhouse-cloud"],
      },
      "x-step": "connector",
    },
    managed: {
      type: "boolean",
      title: "Managed",
      description:
        "This option uses ClickHouse as an OLAP engine with Rill-managed infrastructure. No additional configuration is required - Rill will handle the setup and management of your ClickHouse instance.",
      default: false,
      "x-informational": true,
      "x-visible-if": {
        connector_type: "rill-managed",
      },
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
      "x-visible-if": {
        connector_type: ["self-hosted", "clickhouse-cloud"],
      },
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
      "x-visible-if": {
        connector_type: ["self-hosted", "clickhouse-cloud"],
      },
      "x-step": "connector",
    },
    username: {
      type: "string",
      title: "Username",
      description: "Username to connect to the ClickHouse server",
      default: "default",
      "x-placeholder": "default",
      "x-visible-if": {
        connector_type: ["self-hosted", "clickhouse-cloud"],
      },
      "x-step": "connector",
    },
    password: {
      type: "string",
      title: "Password",
      description: "Password to connect to the ClickHouse server",
      "x-placeholder": "Database password",
      "x-secret": true,
      "x-visible-if": {
        connector_type: ["self-hosted", "clickhouse-cloud"],
      },
      "x-step": "connector",
    },
    database: {
      type: "string",
      title: "Database",
      description: "Name of the ClickHouse database to connect to",
      default: "default",
      "x-placeholder": "default",
      "x-visible-if": {
        connector_type: ["self-hosted", "clickhouse-cloud"],
      },
      "x-step": "connector",
    },
    cluster: {
      type: "string",
      title: "Cluster",
      description:
        "Cluster name. If set, models are created as distributed tables.",
      "x-placeholder": "Cluster name",
      "x-visible-if": {
        connector_type: ["self-hosted", "clickhouse-cloud"],
      },
      "x-step": "connector",
    },
    ssl: {
      type: "boolean",
      title: "SSL",
      description: "Use SSL to connect to the ClickHouse server",
      default: true,
      "x-visible-if": {
        connector_type: ["self-hosted", "clickhouse-cloud"],
      },
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
      if: {
        properties: {
          connector_type: { const: "self-hosted" },
          connection_mode: { const: "parameters" },
        },
      },
      then: {
        required: ["host", "username"],
        properties: {
          managed: { const: false },
          ssl: { default: true },
        },
      },
    },
    {
      if: {
        properties: {
          connector_type: { const: "self-hosted" },
          connection_mode: { const: "dsn" },
        },
      },
      then: {
        required: ["dsn"],
        properties: {
          managed: { const: false },
        },
      },
    },
    {
      if: {
        properties: {
          connector_type: { const: "clickhouse-cloud" },
          connection_mode: { const: "parameters" },
        },
      },
      then: {
        required: ["host", "username", "ssl"],
        properties: {
          managed: { const: false },
          port: { default: "8443" },
          ssl: { const: true },
        },
      },
    },
    {
      if: {
        properties: {
          connector_type: { const: "clickhouse-cloud" },
          connection_mode: { const: "dsn" },
        },
      },
      then: {
        required: ["dsn"],
        properties: {
          managed: { const: false },
        },
      },
    },
  ],
};
