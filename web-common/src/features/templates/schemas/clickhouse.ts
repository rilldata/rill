import ClickHouse from "../../../components/icons/connectors/ClickHouse.svelte";
import ClickHouseIcon from "../../../components/icons/connectors/ClickHouseIcon.svelte";
import type { MultiStepFormSchema } from "./types";

export const clickhouseSchema: MultiStepFormSchema = {
  "x-icon": ClickHouse,
  "x-small-icon": ClickHouseIcon,
  $schema: "http://json-schema.org/draft-07/schema#",
  type: "object",
  title: "ClickHouse",
  "x-category": "olap",
  "x-form-height": "tall",
  "x-form-width": "wide",
  "x-button-labels": {
    deployment_type: {
      playground: { idle: "Connect", loading: "Connecting..." },
      "rill-managed": { idle: "Connect", loading: "Connecting..." },
    },
  },
  properties: {
    deployment_type: {
      type: "string",
      title: "Connection type",
      enum: ["cloud", "playground", "self-managed"], // removed rill-managed until SQL support is ready
      default: "cloud",
      "x-display": "select",
      "x-select-style": "rich",
      "x-enum-labels": [
        "ClickHouse Cloud",
        "ClickHouse Playground",
        "Self Managed",
      ],
      "x-enum-descriptions": [
        "Connect to your ClickHouse Cloud instance",
        "Free public instance for testing and demos",
        "Connect to your own self-hosted server",
      ],
      "x-ui-only": true,
      "x-grouped-fields": {
        cloud: ["connection_mode"],
        playground: ["playground_info"],
        "self-managed": ["connection_mode"],
        "rill-managed": ["managed"],
      },
      "x-step": "connector",
    },
    connection_mode: {
      type: "string",
      enum: ["parameters", "dsn"],
      default: "parameters",
      "x-display": "tabs",
      "x-enum-labels": ["Enter parameters", "Enter connection string"],
      "x-ui-only": true,
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
        deployment_type: ["cloud", "self-managed"],
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
      "x-secret": true,
      "x-env-var-name": "CLICKHOUSE_DSN",
      "x-visible-if": {
        deployment_type: ["cloud", "self-managed"],
      },
      "x-step": "connector",
    },
    managed: {
      type: "boolean",
      title: "Managed",
      description:
        "This option uses ClickHouse as an OLAP engine with Rill-managed infrastructure. No additional configuration is required - Rill will handle the setup and management of your ClickHouse instance.",
      default: true,
      "x-informational": true,
      "x-visible-if": {
        deployment_type: "rill-managed",
      },
      "x-step": "connector",
    },
    playground_info: {
      type: "boolean",
      title: "Playground",
      description:
        'Connect to ClickHouse\'s free public <a href="https://play.clickhouse.com/play?user=play" target="_blank" class="text-primary-600 hover:underline">playground instance</a>. This is a read-only demo environment with sample datasets, perfect for testing Rill\'s ClickHouse integration without any setup. No credentials required.',
      default: true,
      "x-informational": true,
      "x-ui-only": true,
      "x-visible-if": {
        deployment_type: "playground",
      },
      "x-step": "connector",
    },
    host: {
      type: "string",
      title: "Host",
      description: "Hostname or IP address of the ClickHouse server",
      "x-placeholder": "your.clickhouse.server.com",
      "x-hint": "Your ClickHouse hostname (e.g., your-server.com)",
      "x-visible-if": {
        deployment_type: ["cloud", "self-managed"],
      },
      "x-step": "connector",
    },
    port: {
      type: "string",
      title: "Port",
      description:
        "Port number of the ClickHouse server. Common ports: 8443 (HTTPS), 9440 (Native TLS, secure), 8123 (HTTP, insecure), 9000 (Native TCP, insecure)",
      pattern: "^\\d+$",
      errorMessage: { pattern: "Port must be a number" },
      default: "8443",
      "x-placeholder": "8443",
      "x-visible-if": {
        deployment_type: ["cloud", "self-managed"],
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
        deployment_type: ["cloud", "self-managed"],
      },
      "x-step": "connector",
    },
    password: {
      type: "string",
      title: "Password",
      description: "Password to connect to the ClickHouse server",
      "x-placeholder": "Database password",
      "x-secret": true,
      "x-env-var-name": "CLICKHOUSE_PASSWORD",
      "x-visible-if": {
        deployment_type: ["cloud", "self-managed"],
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
        deployment_type: ["cloud", "self-managed"],
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
        deployment_type: ["cloud", "self-managed"],
      },
      "x-step": "connector",
    },
    ssl: {
      type: "boolean",
      title: "SSL",
      description: "Use SSL to connect to the ClickHouse server",
      default: true,
      "x-visible-if": {
        deployment_type: ["cloud", "self-managed"],
      },
      "x-step": "connector",
    },
  },
  required: ["deployment_type"],
  allOf: [
    {
      if: { properties: { deployment_type: { const: "rill-managed" } } },
      then: {
        required: ["managed"],
        properties: {
          managed: { const: true },
        },
      },
    },
    {
      if: { properties: { deployment_type: { const: "playground" } } },
      then: {
        properties: {
          managed: { const: false },
          host: { const: "play.clickhouse.com" },
          port: { const: "9440" },
          username: { const: "play" },
          password: { const: "" },
          database: { const: "default" },
          ssl: { const: true },
        },
      },
    },
    {
      if: {
        properties: {
          deployment_type: { const: "cloud" },
          connection_mode: { const: "parameters" },
        },
      },
      then: {
        required: ["host", "username", "port"],
        properties: {
          managed: { const: false },
          ssl: { const: true },
        },
      },
    },
    {
      if: {
        properties: {
          deployment_type: { const: "cloud" },
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
          deployment_type: { const: "self-managed" },
          connection_mode: { const: "parameters" },
        },
      },
      then: {
        required: ["host", "username"],
        properties: {
          managed: { const: false },
          port: { default: "9000" },
          ssl: { default: true },
        },
      },
    },
    {
      if: {
        properties: {
          deployment_type: { const: "self-managed" },
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
