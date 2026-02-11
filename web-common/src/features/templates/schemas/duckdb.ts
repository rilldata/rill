import type { MultiStepFormSchema } from "./types";

export const duckdbSchema: MultiStepFormSchema = {
  $schema: "http://json-schema.org/draft-07/schema#",
  type: "object",
  title: "DuckDB",
  "x-category": "olap",
  "x-button-labels": {
    connector_type: {
      "rill-managed": { idle: "Connect", loading: "Connecting..." },
    },
  },
  properties: {
    connector_type: {
      type: "string",
      title: "Connection type",
      enum: ["rill-managed", "self-hosted"],
      default: "rill-managed",
      "x-display": "select",
      "x-select-style": "rich",
      "x-enum-labels": ["Rill Managed", "Local File"],
      "x-enum-descriptions": [
        "Rill manages your DuckDB infrastructure",
        "Connect to your own DuckDB database file",
      ],
      "x-ui-only": true,
      "x-grouped-fields": {
        "rill-managed": ["managed"],
        "self-hosted": ["path"],
      },
      "x-step": "connector",
    },
    managed: {
      type: "boolean",
      title: "Managed",
      description:
        "This option uses DuckDB as an OLAP engine with Rill-managed infrastructure. No additional configuration is required - Rill will handle the setup and management of your DuckDB instance.",
      default: false,
      "x-informational": true,
      "x-ui-only": true,
      "x-visible-if": {
        connector_type: "rill-managed",
      },
      "x-step": "connector",
    },
    path: {
      type: "string",
      title: "Path",
      description: "Path to external DuckDB database",
      "x-placeholder": "/path/to/main.db",
      "x-visible-if": {
        connector_type: "self-hosted",
      },
      "x-step": "connector",
    },
    sql: {
      type: "string",
      title: "SQL",
      description: "SQL query to run against DuckDB",
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
        required: ["path"],
        properties: {
          managed: { const: false },
        },
      },
    },
  ],
};
