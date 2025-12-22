import { manual } from "prismjs";
import type { MultiStepFormSchema } from "./types";

export const clickhouseSchema: MultiStepFormSchema = {
  $schema: "http://json-schema.org/draft-07/schema#",
  type: "object",
  properties: {
    auth_method: {
      type: "string",
      title: "Connection type",
      enum: ["self-managed", "rill-managed"],
      default: "self-managed",
      description: "Choose how to connect to ClickHouse",
      "x-display": "radio",
      "x-enum-labels": ["Self-managed", "Rill-managed"],
      "x-enum-descriptions": [
        "Connect to your own self-hosted ClickHouse server.",
        "Use a managed ClickHouse instance (starts embedded ClickHouse in development).",
      ],
      "x-grouped-fields": {
        "self-managed": ["connection_method"],
        "rill-managed": [],
      },
      "x-step": "connector",
    },
    connection_method: {
      type: "string",
      title: "Connection method",
      enum: ["parameters", "connection_string"],
      default: "parameters",
      description: "Choose how to provide connection details",
      "x-display": "tabs",
      "x-enum-labels": ["Enter parameters", "Enter connection string"],
      "x-grouped-fields": {
        parameters: [
          "host",
          "port",
          "username",
          "password",
          "database",
          "ssl",
          "cluster",
          "mode",
        ],
        connection_string: ["dsn", "mode"],
      },
      "x-step": "connector",
      "x-visible-if": { auth_method: "self-managed" },
    },
    host: {
      type: "string",
      title: "Host",
      description: "Hostname or IP address of the ClickHouse server",
      "x-placeholder": "your-server.clickhouse.com",
      "x-step": "connector",
      "x-visible-if": {
        auth_method: "self-managed",
        connection_method: "parameters",
      },
    },
    port: {
      type: "number",
      title: "Port",
      description: "Port number of the ClickHouse server",
      default: 9000,
      "x-placeholder": "9000",
      "x-hint":
        "Default: 9000 (native TCP), 8123 (HTTP). Secure: 9440 (TCP+TLS), 8443 (HTTPS)",
      "x-step": "connector",
      "x-visible-if": {
        auth_method: "self-managed",
        connection_method: "parameters",
      },
    },
    username: {
      type: "string",
      title: "Username",
      description: "Username to connect to the ClickHouse server",
      "x-placeholder": "default",
      "x-step": "connector",
      "x-visible-if": {
        auth_method: "self-managed",
        connection_method: "parameters",
      },
    },
    password: {
      type: "string",
      title: "Password",
      description: "Password to connect to the ClickHouse server",
      "x-placeholder": "Enter password",
      "x-secret": true,
      "x-step": "connector",
      "x-visible-if": {
        auth_method: "self-managed",
        connection_method: "parameters",
      },
    },
    database: {
      type: "string",
      title: "Database",
      description: "Name of the ClickHouse database to connect to",
      default: "default",
      "x-placeholder": "default",
      "x-step": "connector",
      "x-visible-if": {
        auth_method: "self-managed",
        connection_method: "parameters",
      },
    },
    ssl: {
      type: "boolean",
      title: "SSL",
      description: "Use SSL to connect to the ClickHouse server",
      default: true,
      "x-hint": "Enable SSL for secure connections",
      "x-step": "connector",
      "x-visible-if": {
        auth_method: "self-managed",
        connection_method: "parameters",
      },
    },
    cluster: {
      type: "string",
      title: "Cluster",
      description: "Cluster name for distributed tables",
      "x-placeholder": "Cluster name",
      "x-hint":
        "If set, Rill will create models as distributed tables in the cluster",
      "x-step": "connector",
      "x-visible-if": {
        auth_method: "self-managed",
        connection_method: "parameters",
      },
      "x-advanced": true,
    },
    dsn: {
      type: "string",
      title: "Connection String",
      description: "ClickHouse connection string (DSN)",
      "x-placeholder": "clickhouse://username:password@host:port/database",
      "x-secret": true,
      "x-step": "connector",
      "x-visible-if": {
        auth_method: "self-managed",
        connection_method: "connection_string",
      },
    },
    mode: {
      type: "string",
      title: "Connection Mode",
      description: "Database access mode",
      enum: ["read", "readwrite"],
      default: "read",
      "x-display": "radio",
      "x-enum-labels": ["Read-only", "Read-write"],
      "x-enum-descriptions": [
        "Only read operations are allowed (recommended for security)",
        "Enable model creation and table mutations",
      ],
      "x-step": "connector",
      "x-visible-if": { auth_method: "self-managed" },
    },
    managed: {
      type: "boolean",
      title: "Managed",
      description: "Enable managed mode for the ClickHouse server",
      default: true,
      "x-readonly": true,
      "x-hint": "Enable managed mode to manage the server automatically",
      "x-step": "connector",
      "x-visible-if": { auth_method: "rill-managed" },
    },
    database_whitelist: {
      type: "string",
      title: "Database Whitelist",
      description: "List of allowed databases",
      "x-placeholder": "db1,db2",
      "x-step": "connector",
      "x-visible-if": { connection_method: "self-managed" },
      "x-advanced": true,
    },
    optimize_temporary_tables_before_partition_replace: {
      type: "string",
      title: "Optimize Temporary Tables",
      description: "Optimize temporary tables before partition replace",
      "x-placeholder": "true",
      "x-advanced": true,
      "x-step": "connector",
      "x-visible-if": { connection_method: "self-managed" }
    },
    log_queries: {
      type: "string",
      title: "Log Queries",
      description: "Log all queries executed by Rill",
      "x-placeholder": "false",
      "x-advanced": true,
      "x-step": "connector",
      "x-visible-if": { connection_method: "self-managed" }
    },
    query_settings_override: {
      type: "object",
      title: "Query Settings Override",
      description: "Override default query settings",
      "x-placeholder": "key1=value1,key2=value2",
      "x-advanced": true,
      "x-step": "connector",
      "x-visible-if": { connection_method: "self-managed" }
    },
    query_settings: {
      type: "object",
      title: "Query Settings",
      description: "Custom query settings",
      "x-placeholder": "key1=value1,key2=value2",
      "x-advanced": true,
      "x-step": "connector",
      "x-visible-if": { connection_method: "self-managed" }
    },
    embed_port: {
      type: "string",
      title: "Embed Port",
      description: "Port number for embedding the ClickHouse Cloud server",
      "x-placeholder": "8443",
      "x-step": "connector",
      "x-advanced": true,
      "x-visible-if": { connection_method: "self-managed" }
    },
    can_scale_to_zero: {
      type: "string",
      title: "Can Scale to Zero",
      description: "Enable scaling to zero",
      "x-placeholder": "false",
      "x-advanced": true,
      "x-step": "connector",
      "x-visible-if": { connection_method: "self-managed" }
    },
    max_open_conns: {
      type: "string",
      title: "Max Open Connections",
      description: "Maximum number of open connections",
      "x-placeholder": "100",
      "x-advanced": true,
      "x-step": "connector",
      "x-visible-if": { connection_method: "self-managed" }
    },
    max_idle_conns: {
      type: "string",
      title: "Max Idle Connections",
      description: "Maximum number of idle connections",
      "x-placeholder": "10",
      "x-advanced": true,
      "x-step": "connector",
      "x-visible-if": { connection_method: "self-managed" }
    },
    dial_timeout: {
      type: "string",
      title: "Dial Timeout",
      description: "Timeout for establishing a connection",
      "x-placeholder": "30s",
      "x-advanced": true,
      "x-step": "connector",
      "x-visible-if": { connection_method: "self-managed" }
    },
    conn_max_lifetime: {
      type: "string",
      title: "Connection Max Lifetime",
      description: "Maximum lifetime of a connection",
      "x-placeholder": "30m",
      "x-advanced": true,
      "x-step": "connector",
      "x-visible-if": { connection_method: "self-managed" }
    },
    read_timeout: {
      type: "string",
      title: "Read Timeout",
      description: "Timeout for reading from the connection",
      "x-placeholder": "30s",
      "x-advanced": true,
      "x-step": "connector",
      "x-visible-if": { connection_method: "self-managed" }
    },
    sql: {
      type: "string",
      title: "SQL Query",
      description: "SQL query to extract data from ClickHouse",
      "x-placeholder": "SELECT * FROM my_table;",
      "x-step": "source",
      "x-visible-if": { mode: "readwrite" },
    },
    name: {
      type: "string",
      title: "Model Name",
      description: "Name for the source model",
      pattern: "^[a-zA-Z_][a-zA-Z0-9_]*$",
      "x-placeholder": "my_model",
      "x-step": "source",
      "x-visible-if": { mode: "readwrite" },
    },
  },
  allOf: [
    {
      if: {
        properties: {
          auth_method: { const: "self-managed" },
          connection_method: { const: "parameters" },
        },
      },
      then: { required: ["host", "username"] },
    },
    {
      if: {
        properties: {
          auth_method: { const: "self-managed" },
          connection_method: { const: "connection_string" },
        },
      },
      then: { required: ["dsn"] },
    },
    {
      if: { properties: { mode: { const: "readwrite" } } },
      then: { required: ["managed", "sql", "name"] },
    },
  ],
};
