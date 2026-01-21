import type { MultiStepFormSchema } from "./types";

export const clickhousecloudSchema: MultiStepFormSchema = {
  $schema: "http://json-schema.org/draft-07/schema#",
  type: "object",
  properties: {
    connector_type: {
      type: "string",
      default: "clickhouse-cloud",
      "x-internal": true,
      "x-visible-if": {
        _never_matches: "true",
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
      "x-step": "connector",
    },
    dsn: {
      type: "string",
      title: "Connection string",
      description:
        "DSN connection string (use instead of individual host/port/user settings)",
      "x-placeholder":
        "clickhouse://your-instance.clickhouse.cloud:8443?username=default&password=password&secure=true",
      "x-secret": true,
      "x-step": "connector",
    },
    host: {
      type: "string",
      title: "Host",
      description: "Hostname of your ClickHouse Cloud instance",
      "x-placeholder": "your-instance.clickhouse.cloud",
      "x-hint":
        "Your ClickHouse Cloud hostname (e.g., abc123.us-east-1.aws.clickhouse.cloud)",
      "x-step": "connector",
    },
    port: {
      type: "string",
      title: "Port",
      description: "Port number (8443 for ClickHouse Cloud with HTTPS)",
      pattern: "^\\d+$",
      errorMessage: { pattern: "Port must be a number" },
      default: "8443",
      "x-placeholder": "8443",
      "x-step": "connector",
    },
    username: {
      type: "string",
      title: "Username",
      description: "Username to connect to ClickHouse Cloud",
      default: "default",
      "x-placeholder": "default",
      "x-step": "connector",
    },
    password: {
      type: "string",
      title: "Password",
      description: "Password to connect to ClickHouse Cloud",
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
      description: "Use SSL to connect (required for ClickHouse Cloud)",
      default: true,
      "x-step": "connector",
    },
  },
  required: [],
  allOf: [
    {
      if: {
        properties: {
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
