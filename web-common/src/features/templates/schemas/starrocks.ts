import type { MultiStepFormSchema } from "./types";

export const starrocksSchema: MultiStepFormSchema = {
  $schema: "http://json-schema.org/draft-07/schema#",
  type: "object",
  title: "StarRocks",
  "x-category": "olap",
  "x-form-height": "tall",
  properties: {
    connection_mode: {
      type: "string",
      title: "Connection method",
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
          "catalog",
          "database",
          "ssl",
        ],
        dsn: ["dsn"],
      },
    },
    dsn: {
      type: "string",
      title: "Connection string",
      description:
        "MySQL DSN format. If provided, do not set host/port/username/password. Catalog and database should be set separately for external catalogs.",
      "x-placeholder":
        "user:password@tcp(host:9030)/?timeout=30s&readTimeout=300s&parseTime=true",
      "x-secret": true,
    },
    host: {
      type: "string",
      title: "Host",
      description: "Hostname or IP address of the StarRocks FE node",
      "x-placeholder": "localhost",
    },
    port: {
      type: "string",
      title: "Port",
      description: "MySQL protocol port of the StarRocks FE node",
      pattern: "^\\d+$",
      errorMessage: { pattern: "Port must be a number" },
      default: "9030",
      "x-placeholder": "9030",
    },
    username: {
      type: "string",
      title: "Username",
      description: "Username to connect to StarRocks",
      default: "root",
      "x-placeholder": "root",
    },
    password: {
      type: "string",
      title: "Password",
      description: "Password to connect to StarRocks",
      "x-placeholder": "password",
      "x-secret": true,
    },
    catalog: {
      type: "string",
      title: "Catalog",
      description:
        "StarRocks catalog name. Use default_catalog for internal tables, or specify an external catalog (e.g. Iceberg, Hive).",
      default: "default_catalog",
      "x-placeholder": "default_catalog",
    },
    database: {
      type: "string",
      title: "Database",
      description: "Name of the StarRocks database to connect to",
      "x-placeholder": "default",
    },
    ssl: {
      type: "boolean",
      title: "SSL",
      description: "Enable SSL/TLS encryption for the connection",
    },
  },
  required: [],
  allOf: [
    {
      if: {
        properties: { connection_mode: { const: "dsn" } },
      },
      then: { required: ["dsn"] },
      else: { required: ["host"] },
    },
  ],
};
